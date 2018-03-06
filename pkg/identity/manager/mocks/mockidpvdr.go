/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	idapi "github.com/hyperledger/fabric-sdk-go/pkg/context/api/identity"
)

// MockIdentityProvider is a mock IdentityProvider
type MockIdentityProvider struct {
}

// NewMockIdentityProvider Constructor for the identity provider.
func NewMockIdentityProvider(context context.Providers) (idapi.IdentityProvider, error) {
	mcm := MockIdentityProvider{}
	return &mcm, nil
}

// CreateIdentityManager ...
func (mgr *MockIdentityProvider) CreateIdentityManager(orgName string) (idapi.IdentityManager, error) {

	mim, _ := NewMockIdentityManager("", nil, nil)
	return mim, nil
}

// CreateIdentityManager ...
func (mgr *MockIdentityProvider) CreateEnrollmentService(orgName string) (idapi.EnrollmentService, error) {

	mim, _ := NewMockEnrollmentService("", nil, nil)
	return mim, nil
}
