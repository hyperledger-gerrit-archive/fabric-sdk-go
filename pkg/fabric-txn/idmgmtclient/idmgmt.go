/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package idmgmtclient enables identity management client
package idmgmtclient

import (
	"path"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	idapi "github.com/hyperledger/fabric-sdk-go/api/core/identity"
	"github.com/hyperledger/fabric-sdk-go/api/kvstore"
	caapi "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/api"
	fabric_ca "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	fabricclient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// IdentityManager enables managing organization idenitites in Fabric network.
type IdentityManager struct {
	mspID       string
	config      config.Config
	cryptoSuite apicryptosuite.CryptoSuite
	userStore   kvstore.KVStore
	caClient    *fabric_ca.Client
	registrar   struct {
		EnrollID     string
		EnrollSecret string
	}
}

// Context holds the providers and services needed to create an IdentityManager.
type Context struct {
	MspID       string
	Config      config.Config
	CryptoSuite apicryptosuite.CryptoSuite
}

// New returns an identity management client instance
func New(c Context) (*IdentityManager, error) {
	if c.MspID == "" {
		return nil, errors.New("without oganization information")
	}
	if c.Config == nil {
		return nil, errors.New("must provide config")
	}
	if c.CryptoSuite == nil {
		return nil, errors.New("must provide crypto suite")
	}
	userStorePath := c.Config.UserStorePath()
	userStore, err := kvs.NewFileKeyValueStore(&kvs.FileKeyValueStoreOptions{
		Path: userStorePath,
		KeySerializer: func(key interface{}) (string, error) {
			keyString, ok := key.(string)
			if !ok {
				return "", errors.New("converting key to string failed")
			}
			return path.Join(userStorePath, keyString+".json"), nil
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get user store")
	}
	caClient, err := newClient(c.MspID, c.Config, c.CryptoSuite)
	if err != nil {
		return nil, err
	}
	orgConfig, err := c.Config.CAConfig(c.MspID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get CA configurtion for msp: %s", c.MspID)
	}
	registrar := orgConfig.Registrar
	mgr := &IdentityManager{
		mspID:       c.MspID,
		config:      c.Config,
		cryptoSuite: c.CryptoSuite,
		caClient:    caClient,
		userStore:   userStore,
		registrar:   registrar,
	}
	return mgr, nil
}

func newClient(org string, config config.Config, cryptoSuite apicryptosuite.CryptoSuite) (*fabric_ca.Client, error) {
	if org == "" || config == nil || cryptoSuite == nil {
		return nil, errors.New("organization, config and cryptoSuite are required to load CA config")
	}

	// Create new Fabric-ca client without configs
	c := &fabric_ca.Client{
		Config: &fabric_ca.ClientConfig{},
	}

	conf, err := config.CAConfig(org)
	if err != nil {
		return nil, err
	}

	if conf == nil {
		return nil, errors.Errorf("Orgnization %s have no corresponding CA in the configs", org)
	}

	//set server CAName
	c.Config.CAName = conf.CAName
	//set server URL
	c.Config.URL = urlutil.ToAddress(conf.URL)
	//certs file list
	c.Config.TLS.CertFiles, err = config.CAServerCertPaths(org)
	if err != nil {
		return nil, err
	}

	// set key file and cert file
	c.Config.TLS.Client.CertFile, err = config.CAClientCertPath(org)
	if err != nil {
		return nil, err
	}

	c.Config.TLS.Client.KeyFile, err = config.CAClientKeyPath(org)
	if err != nil {
		return nil, err
	}

	// get Client configs
	_, err = config.Client()
	if err != nil {
		return nil, err
	}

	//TLS flag enabled/disabled
	c.Config.TLS.Enabled = urlutil.IsTLSEnabled(conf.URL)
	c.Config.MSPDir = config.CAKeyStorePath()

	//Factory opts
	c.Config.CSP = cryptoSuite

	err = c.Init()
	if err != nil {
		return nil, errors.Wrap(err, "init failed")
	}

	return c, nil
}

// CAName returns the of the associated CA.
func (im *IdentityManager) CAName() string {
	return im.caClient.Config.CAName
}

// Enroll a registered user in order to receive a signed X509 certificate.
// enrollmentID The registered ID to use for enrollment
// enrollmentSecret The secret associated with the enrollment ID
// Returns X509 certificate
func (im *IdentityManager) Enroll(enrollmentID string, enrollmentSecret string, attrs ...idapi.AttributeRequest) (apicryptosuite.Key, []byte, error) {
	if enrollmentID == "" {
		return nil, nil, errors.New("enrollmentID is required")
	}
	if enrollmentSecret == "" {
		return nil, nil, errors.New("enrollmentSecret is required")
	}
	// TODO add attributes
	careq := &caapi.EnrollmentRequest{
		CAName: im.caClient.Config.CAName,
		Name:   enrollmentID,
		Secret: enrollmentSecret,
	}
	caresp, err := im.caClient.Enroll(careq)
	if err != nil {
		return nil, nil, errors.Wrap(err, "enroll failed")
	}
	user := identity.NewUser(enrollmentID, im.mspID)
	user.SetEnrollmentCertificate(caresp.Identity.GetECert().Cert())
	user.SetPrivateKey(caresp.Identity.GetECert().Key())
	err = im.storeUser(user)
	if err != nil {
		return nil, nil, errors.Wrap(err, "enroll failed")
	}
	return caresp.Identity.GetECert().Key(), caresp.Identity.GetECert().Cert(), nil
}

// Reenroll an enrolled user in order to receive a signed X509 certificate
// Returns X509 certificate
func (im *IdentityManager) Reenroll(user idapi.User) (apicryptosuite.Key, []byte, error) {
	if user == nil {
		return nil, nil, errors.New("user required")
	}
	if user.Name() == "" {
		logger.Infof("Invalid re-enroll request, missing argument user")
		return nil, nil, errors.New("user name missing")
	}
	req := &caapi.ReenrollmentRequest{
		CAName: im.caClient.Config.CAName,
	}
	// Create signing identity
	identity, err := im.createSigningIdentity(user)
	if err != nil {
		logger.Debugf("Invalid re-enroll request, %s is not a valid user  %s\n", user.Name(), err)
		return nil, nil, errors.Wrap(err, "createSigningIdentity failed")
	}

	reenrollmentResponse, err := identity.Reenroll(req)
	if err != nil {
		return nil, nil, errors.Wrap(err, "reenroll failed")
	}
	return reenrollmentResponse.Identity.GetECert().Key(), reenrollmentResponse.Identity.GetECert().Cert(), nil
}

// Register a User with the Fabric CA
// registrar: The User that is initiating the registration
// request: Registration Request
// Returns Enrolment Secret
func (im *IdentityManager) Register(request *idapi.RegistrationRequest) (string, error) {
	// Validate registration request
	if request == nil {
		return "", errors.New("registration request is required")
	}
	if request.Name == "" {
		return "", errors.New("request.Name is required")
	}
	registrar, err := im.getRegistrar()
	if err != nil {
		return "", errors.Wrapf(err, "failed to get registrar")
	}
	// Create request signing identity
	identity, err := im.createSigningIdentity(registrar)
	if err != nil {
		return "", errors.Wrap(err, "failed to create request for signing identity")
	}
	// Contruct request for Fabric CA client
	var attributes []caapi.Attribute
	for i := range request.Attributes {
		attributes = append(attributes, caapi.Attribute{Name: request.
			Attributes[i].Key, Value: request.Attributes[i].Value})
	}
	var req = caapi.RegistrationRequest{
		CAName:         request.CAName,
		Name:           request.Name,
		Type:           request.Type,
		MaxEnrollments: request.MaxEnrollments,
		Affiliation:    request.Affiliation,
		Secret:         request.Secret,
		Attributes:     attributes}
	// Make registration request
	response, err := identity.Register(&req)
	if err != nil {
		return "", errors.Wrap(err, "failed to register user")
	}

	return response.Secret, nil
}

func (im *IdentityManager) getRegistrar() (idapi.User, error) {
	r, err := im.loadUser(im.registrar.EnrollID)
	if err != nil {
		if err != idapi.ErrUserNotFound {
			return nil, err
		}
		if im.registrar.EnrollSecret == "" {
			return nil, errors.New("registrar not found and cannot be enrolled because enrollment secret is not present")
		}
		_, _, err := im.Enroll(im.registrar.EnrollID, im.registrar.EnrollSecret)
		if err != nil {
			return nil, err
		}
		return im.loadUser(im.registrar.EnrollID)
	}
	return r, nil
}

func (im *IdentityManager) loadUser(username string) (idapi.User, error) {
	userManager := &fabricclient.Client{}
	userManager.SetStateStore(im.userStore)
	userManager.SetCryptoSuite(im.cryptoSuite)
	user, err := userManager.LoadUserFromStateStore(username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (im *IdentityManager) storeUser(user idapi.User) error {
	userManager := &fabricclient.Client{}
	userManager.SetStateStore(im.userStore)
	userManager.SetCryptoSuite(im.cryptoSuite)
	err := userManager.SaveUserToStateStore(user)
	if err != nil {
		return errors.Wrapf(err, "failed to store user: %s", user.Name())
	}
	return nil
}

// Revoke a User with the Fabric CA
// registrar: The User that is initiating the revocation
// request: Revocation Request
func (im *IdentityManager) Revoke(request *idapi.RevocationRequest) (*idapi.RevocationResponse, error) {
	// Validate revocation request
	if request == nil {
		return nil, errors.New("revocation request is required")
	}
	registrar, err := im.getRegistrar()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to ret registrar")
	}
	// Create request signing identity
	identity, err := im.createSigningIdentity(registrar)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request for signing identity")
	}
	// Create revocation request
	var req = caapi.RevocationRequest{
		CAName: request.CAName,
		Name:   request.Name,
		Serial: request.Serial,
		AKI:    request.AKI,
		Reason: request.Reason,
	}

	_, err = identity.Revoke(&req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to revoke")
	}

	// TODO complete the response mapping
	return &idapi.RevocationResponse{}, nil
}

// createSigningIdentity creates an identity to sign Fabric CA requests with
func (im *IdentityManager) createSigningIdentity(user idapi.
	User) (*fabric_ca.Identity, error) {
	// Validate user
	if user == nil {
		return nil, errors.New("user required")
	}
	// Validate enrolment information
	cert := user.EnrollmentCertificate()
	key := user.PrivateKey()
	if key == nil || cert == nil {
		return nil, errors.New(
			"Unable to read user enrolment information to create signing identity")
	}
	return im.caClient.NewIdentity(key, cert)
}
