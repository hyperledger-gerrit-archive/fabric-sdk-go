/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/apicryptosuite"
)

// MockCredentialManager is a mock CredentialManager
type MockCredentialManager struct {
}

// NewMockCredentialManager Constructor for a credential manager.
func NewMockCredentialManager(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (context.CredentialManager, error) {
	mcm := MockCredentialManager{}
	return &mcm, nil
}

// GetSigningIdentity will sign the given object with provided key,
func (mgr *MockCredentialManager) GetSigningIdentity(userName string) (*context.SigningIdentity, error) {

	si := context.SigningIdentity{}
	return &si, nil
}
