/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/ca"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
)

// Providers represents the SDK configured providers context.
type Providers interface {
	core.Providers
	ca.Providers
	fab.Providers
}

// CoreProviderFactory allows overriding of primitives and the fabric core object provider
type CoreProviderFactory interface {
	CreateStateStoreProvider(config core.Config) (core.KVStore, error)
	CreateCryptoSuiteProvider(config core.Config) (core.CryptoSuite, error)
	CreateSigningManager(cryptoProvider core.CryptoSuite, config core.Config) (core.SigningManager, error)
	CreateIdentityManager(StateStore core.KVStore, cryptoProvider core.CryptoSuite, config core.Config) (core.IdentityManager, error)
	CreateInfraProvider(context Providers) (fab.InfraProvider, error)
}

// CAProviderFactory allows overriding default identity providers
type CAProviderFactory interface {
	CreateCAProvider(context core.Providers) (ca.Provider, error)
}

// ServiceProviderFactory allows overriding default service providers (such as peer discovery)
type ServiceProviderFactory interface {
	CreateDiscoveryProvider(config core.Config, fabPvdr fab.InfraProvider) (fab.DiscoveryProvider, error)
	CreateSelectionProvider(config core.Config) (fab.SelectionProvider, error)
}
