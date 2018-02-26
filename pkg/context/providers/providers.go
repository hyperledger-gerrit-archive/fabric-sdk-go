/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package providers

import (
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/api"
	//"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/api"
)

//IdentityContext interface
type IdentityContext interface {
	MspID() string
	Identity() ([]byte, error)
	PrivateKey() core.Key
}

//FabricProvider interface
type FabricProvider interface {
	CreateChannelClient(user IdentityContext, cfg fab.ChannelCfg) (fab.Channel, error)
	CreateChannelLedger(ic IdentityContext, name string) (fab.ChannelLedger, error)
	CreateChannelConfig(user IdentityContext, name string) (fab.ChannelConfig, error)
	CreateResourceClient(user IdentityContext) (api.Resource, error)
	CreateChannelTransactor(ic IdentityContext, cfg fab.ChannelCfg) (fab.Transactor, error)
	CreateEventHub(ic IdentityContext, name string) (fab.EventHub, error)
	CreatePeerFromConfig(peerCfg *core.NetworkPeer) (fab.Peer, error)
	CreateOrdererFromConfig(cfg *core.OrdererConfig) (fab.Orderer, error)
	CreateUser(name string, signingIdentity *contextApi.SigningIdentity) (contextApi.User, error)
}

// Providers represents the SDK configured providers context.
type Providers interface {
	CoreProviders
	SvcProviders
}

// CoreProviders represents the SDK configured core providers context.
type CoreProviders interface {
	CryptoSuite() core.CryptoSuite
	StateStore() contextApi.KVStore
	Config() core.Config
	SigningManager() contextApi.SigningManager
	FabricProvider() FabricProvider
}

// SvcProviders represents the SDK configured service providers context.
type SvcProviders interface {
	DiscoveryProvider() fab.DiscoveryProvider
	SelectionProvider() fab.SelectionProvider
	ChannelProvider() fab.ChannelProvider
}

// SessionContext primarily represents the session and identity context
type SessionContext interface {
	IdentityContext
}

//Context supplies the configuration and signing identity to client objects.
type Context interface {
	ProviderContext
	IdentityContext
}

//ProviderContext supplies the configuration to client objects.
type ProviderContext interface {
	SigningManager() contextApi.SigningManager
	Config() core.Config
	CryptoSuite() core.CryptoSuite
}
