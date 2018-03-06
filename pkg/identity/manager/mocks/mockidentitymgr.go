/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// MockIdentityManager is a mock Identity Manager
type MockIdentityManager struct {
}

// NewMockIdentityManager Constructor for a identity manager.
func NewMockIdentityManager() core.IdentityManager {
	mcm := MockIdentityManager{}
	return &mcm
}

// GetSigningIdentity will return an identity that can be used to cryptographically sign an object
func (mgr *MockIdentityManager) GetSigningIdentity(orgName string, iserName string) (*core.SigningIdentity, error) {

	si := core.SigningIdentity{
		MspID: "Org1MSP",
	}
	return &si, nil
}

// GetUser will return a user for a given user name
func (mgr *MockIdentityManager) GetUser(orgName string, iserName string) (core.User, error) {
	return nil, nil
}
