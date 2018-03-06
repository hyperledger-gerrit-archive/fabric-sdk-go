/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/ca"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// MockCAProvider is a mock CAProvider
type MockCAProvider struct {
}

// NewMockCAProvider Constructor for the CA provider.
func NewMockCAProvider(context context.Providers) (ca.Provider, error) {
	mcm := MockCAProvider{}
	return &mcm, nil
}

// CreateIdentityManager ...
func (mgr *MockCAProvider) CreateIdentityManager(orgName string) (core.IdentityManager, error) {

	mim := NewMockIdentityManager()
	return mim, nil
}

// CreateCAService ...
func (mgr *MockCAProvider) CreateCAService(orgName string) (ca.Client, error) {

	mim, _ := NewMockCAService("", nil, nil)
	return mim, nil
}
