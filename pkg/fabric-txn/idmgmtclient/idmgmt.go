/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package idmgmtclient enables identity management client
package idmgmtclient

import (
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	idmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/idmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/api/core/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// IdentityMgmtClient enables managing organization idenitites in Fabric network.
type IdentityMgmtClient struct {
	provider fab.ProviderContext
	MspID    string
	identity identity.Context // Registrar
}

// Context holds the providers and services needed to create an IdentityMgmtClient.
type Context struct {
	fab.ProviderContext
	MspID string
}

// New returns an identity management client instance
func New(c Context) (*IdentityMgmtClient, error) {
	cc := &IdentityMgmtClient{
		provider: c.ProviderContext,
		MspID:    c.MspID,
	}
	if cc.provider == nil || cc.MspID == "" {
		return nil, errors.New("must provide provider and org")
	}

	cc.provider.Config()
	cc.provider.CryptoSuite()

	return cc, nil
}

// Enroll enrolls a registered user with the org's Fabric CA
func (cc *IdentityMgmtClient) Enroll(req idmgmt.EnrollmentRequest) (*idmgmt.EnrollmentResponse, error) {

	if req.Name == "" || req.Secret == "" {
		return nil, errors.New("must provide user name and enrollment secret")
	}

	logger.Debugf("***** Enrolling user: %s *****\n", req.Name)

	// TODO

	return nil, nil
}
