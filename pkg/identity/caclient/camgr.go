/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package caclient

import (
	"fmt"

	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/ca"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/identity/manager"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/ca")

// CAManager implements Identity Manager
type CAManager struct {
	orgName         string
	orgMspID        string
	caName          string
	config          core.Config
	cryptoSuite     core.CryptoSuite
	identityManager core.IdentityManager
	userStore       manager.UserStore
	caClient        *Client
	registrar       core.EnrollCredentials
}

// New creates a new instance of CA Manager
// @param {string} organization for this CA
// @param {Config} client config for fabric-ca services
// @returns {Manager} Identity Manager instance
// @returns {error} error, if any
func New(orgName string, identityManager core.IdentityManager, stateStore core.KVStore, cryptoSuite core.CryptoSuite, config core.Config) (*CAManager, error) {

	userStore, err := manager.NewCertFileUserStore1(stateStore)
	if err != nil {
		return nil, errors.Wrapf(err, "creating a user store failed")
	}

	netConfig, err := config.NetworkConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "network config retrieval failed")
	}

	if orgName == "" {
		clientConfig, err := config.Client()
		if err != nil {
			return nil, errors.Wrapf(err, "client config retrieval failed")
		}
		orgName = clientConfig.Organization
	}

	if orgName == "" {
		return nil, errors.New("organization is missing")
	}

	// viper keys are case insensitive
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		return nil, errors.New("org config retrieval failed")
	}

	var caName string
	var caConfig *core.CAConfig
	var caClient *Client
	var registrar core.EnrollCredentials
	if len(orgConfig.CertificateAuthorities) == 0 {
		logger.Warnln("no CAs configured")
	} else {
		// Currently, an organization can be associated with only one CA
		caName = orgConfig.CertificateAuthorities[0]
		caConfig, err = config.CAConfig(orgName)
		if err == nil {
			caClient, err = newCAlient(orgName, caName, cryptoSuite, config)
			if err == nil {
				registrar = caConfig.Registrar
			} else {
				return nil, errors.Wrapf(err, "error initializing CA [%s]", caName)
			}
		} else {
			return nil, errors.Wrapf(err, "error initializing CA [%s]", caName)
		}
	}

	mgr := &CAManager{
		orgName:         orgName,
		orgMspID:        orgConfig.MspID,
		caName:          caName,
		config:          config,
		cryptoSuite:     cryptoSuite,
		identityManager: identityManager,
		userStore:       userStore,
		caClient:        caClient,
		registrar:       registrar,
	}
	return mgr, nil
}

// CAName returns the CA name.
func (im *CAManager) CAName() string {
	return im.caName
}

// Enroll a registered user in order to receive a signed X509 certificate.
// A new key pair is generated for the user. The private key and the
// enrollment certificate issued by the CA are stored in SDK stores.
// They can be retrieved by calling IdentityManager.GetSigningIdentity().
//
// enrollmentID The registered ID to use for enrollment
// enrollmentSecret The secret associated with the enrollment ID
func (im *CAManager) Enroll(enrollmentID string, enrollmentSecret string) error {

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
	careq := &ca.EnrollmentRequest{
		CAName: im.caClient.CAName(),
		Name:   enrollmentID,
		Secret: enrollmentSecret,
	}
	cert, err := im.caClient.Enroll(careq)
	if err != nil {
		return errors.Wrap(err, "enroll failed")
	}
	userData := manager.UserData{
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
func (im *CAManager) Reenroll(enrollmentID string) error {

	if im.caClient == nil {
		return fmt.Errorf("no CAs configured for organization: %s", im.orgName)
	}
	if enrollmentID == "" {
		logger.Infof("invalid re-enroll request, missing enrollmentID")
		return errors.New("user name missing")
	}
	req := &ca.ReenrollmentRequest{
		CAName: im.caClient.CAName(),
	}

	user, err := im.identityManager.GetUser(im.orgMspID, enrollmentID)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve user: %s", enrollmentID)
	}

	cert, err := im.caClient.Reenroll(user.PrivateKey(), user.EnrollmentCertificate(), req)
	if err != nil {
		return errors.Wrap(err, "reenroll failed")
	}
	userData := manager.UserData{
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
func (im *CAManager) Register(request *ca.RegistrationRequest) (string, error) {
	if im.caClient == nil {
		return "", fmt.Errorf("no CAs configured for organization: %s", im.orgName)
	}
	if im.registrar.EnrollID == "" {
		return "", ca.ErrCARegistrarNotFound
	}
	// Validate registration request
	if request == nil {
		return "", errors.New("registration request is required")
	}
	if request.Name == "" {
		return "", errors.New("request.Name is required")
	}

	registrar, err := im.getRegistrar(im.registrar.EnrollID, im.registrar.EnrollSecret)
	if err != nil {
		return "", err
	}

	secret, err := im.caClient.Register(registrar.PrivateKey, registrar.EnrollmentCert, request)
	if err != nil {
		return "", errors.Wrap(err, "failed to register user")
	}

	return secret, nil
}

// Revoke a User with the Fabric CA
// registrar: The User that is initiating the revocation
// request: Revocation Request
func (im *CAManager) Revoke(request *ca.RevocationRequest) (*ca.RevocationResponse, error) {
	if im.caClient == nil {
		return nil, fmt.Errorf("no CAs configured for organization: %s", im.orgName)
	}
	if im.registrar.EnrollID == "" {
		return nil, ca.ErrCARegistrarNotFound
	}
	// Validate revocation request
	if request == nil {
		return nil, errors.New("revocation request is required")
	}

	registrar, err := im.getRegistrar(im.registrar.EnrollID, im.registrar.EnrollSecret)
	if err != nil {
		return nil, err
	}

	resp, err := im.caClient.Revoke(registrar.PrivateKey, registrar.EnrollmentCert, request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to revoke")
	}
	return resp, nil
}

func (im *CAManager) getRegistrar(enrollID string, enrollSecret string) (*core.SigningIdentity, error) {

	if enrollID == "" {
		return nil, ca.ErrCARegistrarNotFound
	}

	registrar, err := im.identityManager.GetSigningIdentity(im.orgMspID, enrollID)
	if err != nil {
		if err != core.ErrUserNotFound {
			return nil, err
		}
		if enrollSecret == "" {
			return nil, ca.ErrCARegistrarNotFound
		}

		// Attempt to enroll the registrar
		err = im.Enroll(enrollID, enrollSecret)
		if err != nil {
			return nil, err
		}
		registrar, err = im.identityManager.GetSigningIdentity(im.orgMspID, enrollID)
		if err != nil {
			return nil, err
		}
	}
	return registrar, nil
}
