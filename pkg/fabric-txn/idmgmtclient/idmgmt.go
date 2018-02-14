/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package idmgmtclient enables identity management client
package idmgmtclient

import (
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/core/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// IdentityManager enables managing organization idenitites in Fabric network.
type IdentityManager struct {
	provider fab.ProviderContext
	MspID    string
}

// Context holds the providers and services needed to create an IdentityManager.
type Context struct {
	fab.ProviderContext
	MspID string
}

// New returns an identity management client instance
func New(c Context) (*IdentityManager, error) {
	cc := &IdentityManager{
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
func (cc *IdentityManager) Enroll(req identity.EnrollmentRequest) (*identity.EnrollmentResponse, error) {

	if req.Name == "" || req.Secret == "" {
		return nil, errors.New("must provide user name and enrollment secret")
	}

	logger.Debugf("***** Enrolling user: %s *****\n", req.Name)

	// TODO

	return nil, nil
}
