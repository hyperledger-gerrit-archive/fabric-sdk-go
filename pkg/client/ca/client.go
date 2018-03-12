/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package ca enables access to CA services.
package ca

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
	mspapi "github.com/hyperledger/fabric-sdk-go/pkg/msp/api"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/client")

// Client enables access to CA services client implementation
type Client struct {
	orgName string
	client  mspapi.CAClient
}

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*Client) error

// WithOrg option
func WithOrg(orgName string) ClientOption {
	return func(client *Client) error {
		client.orgName = orgName
		return nil
	}
}

// New returns a Client instance.
func New(clientProvider context.ClientProvider, opts ...ClientOption) (*Client, error) {

	ctx, err := clientProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create CA client")
	}

	client := Client{}

	for _, param := range opts {
		err := param(&client)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create CA client")
		}
	}

	if client.orgName == "" {
		clientConfig, err := ctx.Config().Client()
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create CA Client")
		}
		client.orgName = clientConfig.Organization
	}
	identityManager, ok := ctx.IdentityManager(client.orgName)
	if !ok {
		return nil, fmt.Errorf("identity managet not found for organization '%s", client.orgName)
	}
	caClient, err := msp.NewCAClient(client.orgName, identityManager, ctx.StateStore(), ctx.CryptoSuite(), ctx.Config())
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create CA Client")
	}

	client.client = caClient

	return &client, nil
}

// Enroll enrolls a registered user in order to receive a signed X509 certificate.
// A new key pair is generated for the user. The private key and the
// enrollment certificate issued by the CA are stored in SDK stores.
// They can be retrieved by calling IdentityManager.GetSigningIdentity().
//
// enrollmentID enrollment ID of a registered user
// enrollmentSecret secret associated with the enrollment ID
func (c *Client) Enroll(enrollmentID string, enrollmentSecret string) error {
	return c.client.Enroll(enrollmentID, enrollmentSecret)
}

// Reenroll reenrolls an enrolled user in order to obtain a new signed X509 certificate
func (c *Client) Reenroll(enrollmentID string) error {
	return c.client.Reenroll(enrollmentID)
}

// Register registers a User with the Fabric CA
// request: Registration Request
// Returns Enrolment Secret
func (c *Client) Register(request *mspapi.RegistrationRequest) (string, error) {
	return c.client.Register(request)
}

// Revoke revokes a User with the Fabric CA
// request: Revocation Request
func (c *Client) Revoke(request *mspapi.RevocationRequest) (*mspapi.RevocationResponse, error) {
	return c.client.Revoke(request)
}
