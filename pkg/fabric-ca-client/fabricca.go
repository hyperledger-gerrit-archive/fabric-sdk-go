/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricca

import (
	"path"

	fabricclient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	"github.com/pkg/errors"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	idapi "github.com/hyperledger/fabric-sdk-go/api/core/identity"
	"github.com/hyperledger/fabric-sdk-go/api/kvstore"
	api "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/api"
	fabric_ca "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
)

var logger = logging.NewLogger("fabric_sdk_go")

// FabricCA represents a client to Fabric CA.
type FabricCA struct {
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

// NewFabricCAClient creates a new fabric-ca client
// @param {string} organization for this CA
// @param {api.Config} client config for fabric-ca services
// @returns {api.FabricCAClient} FabricCAClient implementation
// @returns {error} error, if any
func NewFabricCAClient(org string, config config.Config, cryptoSuite apicryptosuite.CryptoSuite) (*FabricCA, error) {
	if org == "" {
		return nil, errors.New("without oganization information")
	}
	if config == nil {
		return nil, errors.New("must provide config")
	}
	if cryptoSuite == nil {
		return nil, errors.New("must provide crypto suite")
	}
	userStorePath := config.UserStorePath()
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
	caClient, err := newClient(org, config, cryptoSuite)
	if err != nil {
		return nil, err
	}
	orgConfig, err := config.CAConfig(org)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get CA configurtion for msp: %s", org)
	}
	registrar := orgConfig.Registrar
	client := &FabricCA{
		mspID:       org,
		config:      config,
		cryptoSuite: cryptoSuite,
		caClient:    caClient,
		userStore:   userStore,
		registrar:   registrar,
	}
	return client, nil
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

// CAName returns the CA name.
func (im *FabricCA) CAName() string {
	return im.caClient.Config.CAName
}

// Enroll a registered user in order to receive a signed X509 certificate.
// enrollmentID The registered ID to use for enrollment
// enrollmentSecret The secret associated with the enrollment ID
// Returns X509 certificate
func (im *FabricCA) Enroll(enrollmentID string, enrollmentSecret string) (apicryptosuite.Key, []byte, error) {
	if enrollmentID == "" {
		return nil, nil, errors.New("enrollmentID is required")
	}
	if enrollmentSecret == "" {
		return nil, nil, errors.New("enrollmentSecret is required")
	}
	// TODO add attributes
	careq := &api.EnrollmentRequest{
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
func (im *FabricCA) Reenroll(user idapi.User) (apicryptosuite.Key, []byte, error) {
	if user == nil {
		return nil, nil, errors.New("user required")
	}
	if user.Name() == "" {
		logger.Infof("Invalid re-enroll request, missing argument user")
		return nil, nil, errors.New("user name missing")
	}
	req := &api.ReenrollmentRequest{
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
// request: Registration Request
// Returns Enrolment Secret
func (im *FabricCA) Register(request *idapi.RegistrationRequest) (string, error) {
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
	var attributes []api.Attribute
	for i := range request.Attributes {
		attributes = append(attributes, api.Attribute{Name: request.
			Attributes[i].Key, Value: request.Attributes[i].Value})
	}
	var req = api.RegistrationRequest{
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

// Revoke a User with the Fabric CA
// registrar: The User that is initiating the revocation
// request: Revocation Request
func (im *FabricCA) Revoke(request *idapi.RevocationRequest) (*idapi.RevocationResponse, error) {
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
	var req = api.RevocationRequest{
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

func (im *FabricCA) getRegistrar() (idapi.User, error) {
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

func (im *FabricCA) loadUser(username string) (idapi.User, error) {
	userManager := &fabricclient.Client{}
	userManager.SetStateStore(im.userStore)
	userManager.SetCryptoSuite(im.cryptoSuite)
	user, err := userManager.LoadUserFromStateStore(username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (im *FabricCA) storeUser(user idapi.User) error {
	userManager := &fabricclient.Client{}
	userManager.SetStateStore(im.userStore)
	userManager.SetCryptoSuite(im.cryptoSuite)
	err := userManager.SaveUserToStateStore(user)
	if err != nil {
		return errors.Wrapf(err, "failed to store user: %s", user.Name())
	}
	return nil
}

// createSigningIdentity creates an identity to sign Fabric CA requests with
func (im *FabricCA) createSigningIdentity(user idapi.
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
