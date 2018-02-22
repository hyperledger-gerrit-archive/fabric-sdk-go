/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// MockIdentityManager is a mock IdentityManager
type MockIdentityManager struct {
}

// NewMockIdentityManager Constructor for a credential manager.
func NewMockIdentityManager(orgName string, config core.Config, cryptoProvider core.CryptoSuite) (api.IdentityManager, error) {
	mcm := MockIdentityManager{}
	return &mcm, nil
}

// GetSigningIdentity will sign the given object with provided key,
func (mgr *MockIdentityManager) GetSigningIdentity(userName string) (*api.SigningIdentity, error) {

	si := api.SigningIdentity{}
	return &si, nil
}
