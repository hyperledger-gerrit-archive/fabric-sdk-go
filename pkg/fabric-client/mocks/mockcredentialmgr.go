/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	apiconfig "github.com/hyperledger/fabric-sdk-go/api/config"
)

// MockCredentialManager is a mock CredentialManager
type MockCredentialManager struct {
}

// NewMockCredentialManager Constructor for a credential manager.
func NewMockCredentialManager(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (apifabclient.CredentialManager, error) {
	mcm := MockCredentialManager{}
	return &mcm, nil
}

// GetSigningIdentity will sign the given object with provided key,
func (mgr *MockCredentialManager) GetSigningIdentity(userName string) (*apifabclient.SigningIdentity, error) {

	si := apifabclient.SigningIdentity{}
	return &si, nil
}
