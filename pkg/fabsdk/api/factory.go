/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
)

// CoreProviderFactory allows overriding of primitives and the fabric core object provider
type CoreProviderFactory interface {
	CreateStateStoreProvider(config core.Config) (core.KVStore, error)
	CreateCryptoSuiteProvider(config core.Config) (core.CryptoSuite, error)
	CreateSigningManager(cryptoProvider core.CryptoSuite, config core.Config) (core.SigningManager, error)
	CreateIdentityManager(orgName string, cryptoProvider core.CryptoSuite, config core.Config) (api.IdentityManager, error)
	CreateFabricProvider(context context.Provider) (fab.FabricProvider, error)
}

// ServiceProviderFactory allows overriding default service providers (such as peer discovery)
type ServiceProviderFactory interface {
	CreateDiscoveryProvider(config core.Config) (fab.DiscoveryProvider, error)
	CreateSelectionProvider(config core.Config) (fab.SelectionProvider, error)
	//CreateChannelProvider(ctx Context, channelID string) (ChannelProvider, error)
}

// SessionClientFactory allows overriding default clients and providers of a session
type SessionClientFactory interface {
	CreateChannelClient(sdk context.BaseProviders, session context.Session, channelID string, targetFilter fab.TargetFilter) (*channel.Client, error)
}
