/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defidentity

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"

	idapi "github.com/hyperledger/fabric-sdk-go/pkg/context/api/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/idpvdr"
)

// ProviderFactory represents the default identity provider factory.
type ProviderFactory struct {
}

// NewProviderFactory returns the default identity provider factory.
func NewProviderFactory() *ProviderFactory {
	f := ProviderFactory{}
	return &f
}

// CreateIdentityProvider creates the identity provider
func (f *ProviderFactory) CreateIdentityProvider(ctx core.Providers) (idapi.IdentityProvider, error) {
	return idpvdr.New(ctx), nil
}
