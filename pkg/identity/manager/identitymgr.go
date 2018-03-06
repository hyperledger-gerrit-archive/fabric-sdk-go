/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package manager

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	idapi "github.com/hyperledger/fabric-sdk-go/pkg/context/api/identity"
	ca "github.com/hyperledger/fabric-sdk-go/pkg/identity/caclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"

	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

var logger = logging.NewLogger("fabsdk/core")

// IdentityManager implements fab/IdentityManager
type IdentityManager struct {
	orgName         string
	orgMspID        string
	caName          string
	config          core.Config
	cryptoSuite     core.CryptoSuite
	embeddedUsers   map[string]core.TLSKeyPair
	mspPrivKeyStore core.KVStore
	mspCertStore    core.KVStore
	userStore       UserStore
	// CA Client state
	caClient  *ca.Client
	registrar core.EnrollCredentials
}

// New creates a new instance of IdentityManager
// @param {string} organization for this CA
// @param {Config} client config for fabric-ca services
// @returns {IdentityManager} IdentityManager instance
// @returns {error} error, if any
func New(orgName string, stateStore core.KVStore, cryptoSuite core.CryptoSuite, config core.Config) (*IdentityManager, error) {

	netConfig, err := config.NetworkConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "network config retrieval failed")
	}

	// viper keys are case insensitive
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		return nil, errors.New("org config retrieval failed")
	}

	if orgConfig.CryptoPath == "" && len(orgConfig.Users) == 0 {
		return nil, errors.New("Either a cryptopath or an embedded list of users is required")
	}

	var mspPrivKeyStore core.KVStore
	var mspCertStore core.KVStore

	orgCryptoPathTemplate := orgConfig.CryptoPath
	if orgCryptoPathTemplate != "" {
		if !filepath.IsAbs(orgCryptoPathTemplate) {
			orgCryptoPathTemplate = filepath.Join(config.CryptoConfigPath(), orgCryptoPathTemplate)
		}
		mspPrivKeyStore, err = NewFileKeyStore(orgCryptoPathTemplate)
		if err != nil {
			return nil, errors.Wrapf(err, "creating a private key store failed")
		}
		mspCertStore, err = NewFileCertStore(orgCryptoPathTemplate)
		if err != nil {
			return nil, errors.Wrapf(err, "creating a cert store failed")
		}
	} else {
		logger.Warnf("Cryptopath not provided for organization [%s], MSP stores not created", orgName)
	}

	userStore, err := NewCertFileUserStore1(stateStore)
	if err != nil {
		return nil, errors.Wrapf(err, "creating a user store failed")
	}

	var caName string
	var caConfig *core.CAConfig
	var caClient *ca.Client
	var registrar core.EnrollCredentials
	if len(orgConfig.CertificateAuthorities) == 0 {
		logger.Warnln("no CAs configured")
	} else {
		// Currently, an organization can be associated with only one CA
		caName = orgConfig.CertificateAuthorities[0]
		caConfig, err = config.CAConfig(orgName)
		if err == nil {
			caClient, err = ca.New(orgName, caName, cryptoSuite, config)
			if err == nil {
				registrar = caConfig.Registrar
			} else {
				return nil, errors.Wrapf(err, "error initializing CA [%s]", caName)
			}
		} else {
			return nil, errors.Wrapf(err, "error initializing CA [%s]", caName)
		}
	}

	mgr := &IdentityManager{
		orgName:         orgName,
		orgMspID:        orgConfig.MspID,
		caName:          caName,
		config:          config,
		cryptoSuite:     cryptoSuite,
		mspPrivKeyStore: mspPrivKeyStore,
		mspCertStore:    mspCertStore,
		embeddedUsers:   orgConfig.Users,
		userStore:       userStore,
		caClient:        caClient,
		registrar:       registrar,
	}
	return mgr, nil
}

// CAName returns the CA name.
func (im *IdentityManager) CAName() string {
	return im.caName
}

// Enroll a registered user in order to receive a signed X509 certificate.
// A new key pair is generated for the user. The private key and the
// enrollment certificate issued by the CA are stored in SDK stores.
// They can be retrieved by calling IdentityManager.GetSigningIdentity().
//
// enrollmentID The registered ID to use for enrollment
// enrollmentSecret The secret associated with the enrollment ID
func (im *IdentityManager) Enroll(enrollmentID string, enrollmentSecret string) error {

	if im.caClient == nil {
		return fmt.Errorf("no CAs configured for organization: %s", im.orgName)
	}
	if enrollmentID == "" {
		return errors.New("enrollmentID is required")
	}
	if enrollmentSecret == "" {
		return errors.New("enrollmentSecret is required")
	}
	// TODO add attributes
	careq := &idapi.EnrollmentRequest{
		CAName: im.caClient.CAName(),
		Name:   enrollmentID,
		Secret: enrollmentSecret,
	}
	cert, err := im.caClient.Enroll(careq)
	if err != nil {
		return errors.Wrap(err, "enroll failed")
	}
	userData := UserData{
		MspID: im.orgMspID,
		Name:  enrollmentID,
		EnrollmentCertificate: cert,
	}
	err = im.userStore.Store(userData)
	if err != nil {
		return errors.Wrap(err, "enroll failed")
	}
	return nil
}

// Reenroll an enrolled user in order to obtain a new signed X509 certificate
func (im *IdentityManager) Reenroll(enrollmentID string) error {

	if im.caClient == nil {
		return fmt.Errorf("no CAs configured for organization: %s", im.orgName)
	}
	if enrollmentID == "" {
		logger.Infof("invalid re-enroll request, missing enrollmentID")
		return errors.New("user name missing")
	}
	req := &idapi.ReenrollmentRequest{
		CAName: im.caClient.CAName(),
	}

	user, err := im.GetUser(enrollmentID)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve user: %s", enrollmentID)
	}

	cert, err := im.caClient.Reenroll(user.PrivateKey(), user.EnrollmentCertificate(), req)
	if err != nil {
		return errors.Wrap(err, "reenroll failed")
	}
	userData := UserData{
		MspID: im.orgMspID,
		Name:  user.Name(),
		EnrollmentCertificate: cert,
	}
	err = im.userStore.Store(userData)
	if err != nil {
		return errors.Wrap(err, "reenroll failed")
	}

	return nil
}

// Register a User with the Fabric CA
// request: Registration Request
// Returns Enrolment Secret
func (im *IdentityManager) Register(request *idapi.RegistrationRequest) (string, error) {
	if im.caClient == nil {
		return "", fmt.Errorf("no CAs configured for organization: %s", im.orgName)
	}
	if im.registrar.EnrollID == "" {
		return "", idapi.ErrCARegistrarNotFound
	}
	// Validate registration request
	if request == nil {
		return "", errors.New("registration request is required")
	}
	if request.Name == "" {
		return "", errors.New("request.Name is required")
	}
	// Contruct request for Fabric CA client
	var attributes []idapi.Attribute
	for i := range request.Attributes {
		attributes = append(attributes, idapi.Attribute{Name: request.
			Attributes[i].Key, Value: request.Attributes[i].Value})
	}
	var req = idapi.RegistrationRequest{
		CAName:         request.CAName,
		Name:           request.Name,
		Type:           request.Type,
		MaxEnrollments: request.MaxEnrollments,
		Affiliation:    request.Affiliation,
		Secret:         request.Secret,
		Attributes:     attributes,
	}

	registrar, err := im.getRegistrar(im.registrar.EnrollID, im.registrar.EnrollSecret)
	if err != nil {
		return "", err
	}

	secret, err := im.caClient.Register(registrar.PrivateKey, registrar.EnrollmentCert, &req)
	if err != nil {
		return "", errors.Wrap(err, "failed to register user")
	}

	return secret, nil
}

// Revoke a User with the Fabric CA
// registrar: The User that is initiating the revocation
// request: Revocation Request
func (im *IdentityManager) Revoke(request *idapi.RevocationRequest) (*idapi.RevocationResponse, error) {
	if im.caClient == nil {
		return nil, fmt.Errorf("no CAs configured for organization: %s", im.orgName)
	}
	if im.registrar.EnrollID == "" {
		return nil, idapi.ErrCARegistrarNotFound
	}
	// Validate revocation request
	if request == nil {
		return nil, errors.New("revocation request is required")
	}
	// Create revocation request
	var req = idapi.RevocationRequest{
		CAName: request.CAName,
		Name:   request.Name,
		Serial: request.Serial,
		AKI:    request.AKI,
		Reason: request.Reason,
	}

	registrar, err := im.getRegistrar(im.registrar.EnrollID, im.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	resp, err := im.caClient.Revoke(registrar.PrivateKey, registrar.EnrollmentCert, &req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to revoke")
	}
	var revokedCerts []idapi.RevokedCert
	for i := range resp.RevokedCerts {
		revokedCerts = append(
			revokedCerts,
			idapi.RevokedCert{
				Serial: resp.RevokedCerts[i].Serial,
				AKI:    resp.RevokedCerts[i].AKI,
			})
	}

	// TODO complete the response mapping
	return &idapi.RevocationResponse{
		RevokedCerts: revokedCerts,
		CRL:          resp.CRL,
	}, nil
}

func (im *IdentityManager) getRegistrar(enrollID string, enrollSecret string) (*idapi.SigningIdentity, error) {

	if enrollID == "" {
		return nil, idapi.ErrCARegistrarNotFound
	}

	registrar, err := im.GetSigningIdentity(enrollID)
	if err != nil {
		if err != core.ErrUserNotFound {
			return nil, err
		}
		if enrollSecret == "" {
			return nil, idapi.ErrCARegistrarNotFound
		}

		// Attempt to enroll the registrar
		err = im.Enroll(enrollID, enrollSecret)
		if err != nil {
			return nil, err
		}
		registrar, err = im.GetSigningIdentity(enrollID)
		if err != nil {
			return nil, err
		}
	}
	return registrar, nil
}
