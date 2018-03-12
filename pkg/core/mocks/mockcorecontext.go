/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// MockCoreContext is a mock core context
type MockCoreContext struct {
	Xconfig         core.Config
	XcryptoSuite    core.CryptoSuite
	XstateStore     core.KVStore
	XsigningManager core.SigningManager
}

// Config ...
func (m *MockCoreContext) Config() core.Config {
	return m.Xconfig
}

// CryptoSuite ...
func (m *MockCoreContext) CryptoSuite() core.CryptoSuite {
	return m.XcryptoSuite
}

// StateStore ...
func (m *MockCoreContext) StateStore() core.KVStore {
	return m.XstateStore
}

// SigningManager ...
func (m *MockCoreContext) SigningManager() core.SigningManager {
	return m.XsigningManager
}
