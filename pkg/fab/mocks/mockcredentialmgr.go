/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
)

// MockIdentityManager is a mock IdentityManager
type MockIdentityManager struct {
}

// NewMockIdentityManager Constructor for a credential manager.
func NewMockIdentityManager(orgName string, config core.Config, cryptoProvider core.CryptoSuite) (fab.IdentityManager, error) {
	mcm := MockIdentityManager{}
	return &mcm, nil
}

// GetSigningIdentity will sign the given object with provided key,
func (mgr *MockIdentityManager) GetSigningIdentity(userName string) (*api.SigningIdentity, error) {

	si := api.SigningIdentity{}
	return &si, nil
}

// CAName ...
func (mgr *MockIdentityManager) CAName() string {
	return ""
}

// Enroll ...
func (mgr *MockIdentityManager) Enroll(enrollmentID string, enrollmentSecret string) (core.Key, []byte, error) {
	return nil, nil, nil
}

// Reenroll ...
func (mgr *MockIdentityManager) Reenroll(user api.User) (core.Key, []byte, error) {
	return nil, nil, nil
}

// Register ...
func (mgr *MockIdentityManager) Register(request *fab.RegistrationRequest) (string, error) {
	return "", nil
}

// Revoke ...
func (mgr *MockIdentityManager) Revoke(request *fab.RevocationRequest) (*fab.RevocationResponse, error) {
	return nil, nil
}
