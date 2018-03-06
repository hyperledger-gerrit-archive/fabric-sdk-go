/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	idapi "github.com/hyperledger/fabric-sdk-go/pkg/context/api/identity"
)

// MockIdentityManager is a mock Identity Manager
type MockIdentityManager struct {
}

// NewMockIdentityManager Constructor for a identity manager.
func NewMockIdentityManager(orgName string, cryptoProvider core.CryptoSuite, config core.Config) (idapi.Manager, error) {
	mcm := MockIdentityManager{}
	return &mcm, nil
}

// GetSigningIdentity will return an identity that can be used to cryptographically sign an object
func (mgr *MockIdentityManager) GetSigningIdentity(userName string) (*idapi.SigningIdentity, error) {

	si := idapi.SigningIdentity{
		MspID: "Org1MSP",
	}
	return &si, nil
}

// GetUser will return a user for a given user name
func (mgr *MockIdentityManager) GetUser(userName string) (core.User, error) {
	return nil, nil
}

// CAName return the name of a CA associated with this identity manager
func (mgr *MockIdentityManager) CAName() string {
	return ""
}
