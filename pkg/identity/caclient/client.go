/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package caclient

import (
	"github.com/pkg/errors"

	caapi "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/api"
	calib "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	idapi "github.com/hyperledger/fabric-sdk-go/pkg/context/api/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabsdk/core")

// Client handles communication with Fabric CA
type Client struct {
	caName      string
	config      core.Config
	cryptoSuite core.CryptoSuite
	caClient    *calib.Client
}

// New creates a new instance of Client
// @param {Config} client config for fabric-ca services
// @returns {IdentityManager} IdentityManager instance
// @returns {error} error, if any
func New(orgName string, caName string, cryptoSuite core.CryptoSuite, config core.Config) (*Client, error) {

	caClient, err := newCAClient(orgName, cryptoSuite, config)
	if err != nil {
		return nil, err
	}

	c := &Client{
		caName:      caName,
		config:      config,
		cryptoSuite: cryptoSuite,
		caClient:    caClient,
	}
	return c, nil
}

// CAName returns the CA name.
func (c *Client) CAName() string {
	return c.caName
}

// Enroll a registered user in order to receive a signed X509 certificate. //
// enrollmentID The registered ID to use for enrollment
// enrollmentSecret The secret associated with the enrollment ID
func (c *Client) Enroll(req *idapi.EnrollmentRequest) ([]byte, error) {

	logger.Debugf("Enrolling user [%s]", req.Name)

	// TODO add attributes
	careq := &caapi.EnrollmentRequest{
		CAName: c.caClient.Config.CAName,
		Name:   req.Name,
		Secret: req.Secret,
	}
	caresp, err := c.caClient.Enroll(careq)
	if err != nil {
		return nil, errors.WithMessage(err, "enroll failed")
	}
	return caresp.Identity.GetECert().Cert(), nil
}

// Reenroll an enrolled user in order to receive a signed X509 certificate
// key: user private key
// cert: user enrollment certificate
// Returns X509 certificate
func (c *Client) Reenroll(key core.Key, cert []byte, req *idapi.ReenrollmentRequest) ([]byte, error) {

	logger.Debugf("Enrolling user [%s]")

	careq := &caapi.ReenrollmentRequest{
		CAName: c.caClient.Config.CAName,
	}
	caidentity, err := c.caClient.NewIdentity(key, cert)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create CA signing identity")
	}

	caresp, err := caidentity.Reenroll(careq)
	if err != nil {
		return nil, errors.WithMessage(err, "reenroll failed")
	}

	return caresp.Identity.GetECert().Cert(), nil
}

// Register a User with the Fabric CA
// key: registrar private key
// cert: registrar enrollment certificate
// request: Registration Request
// Returns Enrolment Secret
func (c *Client) Register(key core.Key, cert []byte, request *idapi.RegistrationRequest) (string, error) {
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

	registrar, err := c.caClient.NewIdentity(key, cert)
	if err != nil {
		return "", errors.Wrap(err, "failed to create CA signing identity")
	}

	response, err := registrar.Register(&req)
	if err != nil {
		return "", errors.Wrap(err, "failed to register user")
	}

	return response.Secret, nil
}

// Revoke a User with the Fabric CA
// key: registrar private key
// cert: registrar enrollment certificate
// request: Revocation Request
func (c *Client) Revoke(key core.Key, cert []byte, request *idapi.RevocationRequest) (*idapi.RevocationResponse, error) {
	// Create revocation request
	var req = caapi.RevocationRequest{
		CAName: request.CAName,
		Name:   request.Name,
		Serial: request.Serial,
		AKI:    request.AKI,
		Reason: request.Reason,
	}

	registrar, err := c.caClient.NewIdentity(key, cert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CA signing identity")
	}

	resp, err := registrar.Revoke(&req)
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

	return &idapi.RevocationResponse{
		RevokedCerts: revokedCerts,
		CRL:          resp.CRL,
	}, nil
}
