/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defca

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/ca"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/capvdr"
)

// ProviderFactory represents the default identity provider factory.
type ProviderFactory struct {
}

// NewProviderFactory returns the default identity provider factory.
func NewProviderFactory() *ProviderFactory {
	f := ProviderFactory{}
	return &f
}

// CreateCAProvider creates the identity provider
func (f *ProviderFactory) CreateCAProvider(ctx core.Providers) (ca.Provider, error) {
	return capvdr.New(ctx), nil
}
