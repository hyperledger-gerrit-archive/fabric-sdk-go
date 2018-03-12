/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defmsp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/msppvdr"
)

// ProviderFactory represents the default MSP provider factory.
type ProviderFactory struct {
}

// NewProviderFactory returns the default MSP provider factory.
func NewProviderFactory() *ProviderFactory {
	f := ProviderFactory{}
	return &f
}

// CreateIdentityManagerProvider returns a new default implementation of MSP provider
func (f *ProviderFactory) CreateIdentityManagerProvider(config core.Config, cryptoProvider core.CryptoSuite, stateStore core.KVStore) (msp.IdentityManagerProvider, error) {
	return msppvdr.New(config, cryptoProvider, stateStore)
}
