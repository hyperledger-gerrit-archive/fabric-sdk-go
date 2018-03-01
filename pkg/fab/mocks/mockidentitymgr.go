/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/pkg/errors"
)

// MockIdentityManager is a mock IdentityManager
type MockIdentityManager struct {
}

// NewMockIdentityManager Constructor for a identity manager.
func NewMockIdentityManager(orgName string, cryptoProvider core.CryptoSuite, config core.Config) (api.IdentityManager, error) {
	mcm := MockIdentityManager{}
	return &mcm, nil
}

// GetUser will return a user for a given user name
func (mgr *MockIdentityManager) GetUser(userName string) (api.User, error) {
	return nil, nil
}

// Enroll enrolls a user with a Fabric network
func (mgr *MockIdentityManager) Enroll(enrollmentID string, enrollmentSecret string) error {
	return errors.New("not implemented")
}

// Reenroll re-enrolls a user
func (mgr *MockIdentityManager) Reenroll(user api.User) error {
	return errors.New("not implemented")
}

// Register registers a user with a Fabric network
func (mgr *MockIdentityManager) Register(request *api.RegistrationRequest) (string, error) {
	return "", errors.New("not implemented")
}

// Revoke revokes a user
func (mgr *MockIdentityManager) Revoke(request *api.RevocationRequest) (*api.RevocationResponse, error) {
	return nil, errors.New("not implemented")
}

// CAName return the name of a CA associated with this identity manager
func (mgr *MockIdentityManager) CAName() string {
	return ""
}
