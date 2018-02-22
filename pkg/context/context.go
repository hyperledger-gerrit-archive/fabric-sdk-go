/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/apicryptosuite"
)

// Context supplies the configuration and signing identity to client objects.
type Context interface {
	ProviderContext
	IdentityContext
}

// ProviderContext supplies the configuration to client objects.
type ProviderContext interface {
	SigningManager() SigningManager
	Config() apiconfig.Config
	CryptoSuite() apicryptosuite.CryptoSuite
}

// ChannelProvider supplies Channel related-objects for the named channel.
type ChannelProvider interface {
	NewChannelService(ic IdentityContext, channelID string) (ChannelService, error)
}

// ChannelService supplies services related to a channel.
type ChannelService interface {
	Config() (ChannelConfig, error)
	Ledger() (ChannelLedger, error)
	Channel() (Channel, error) // TODO remove
	Transactor() (Transactor, error)
	EventHub() (EventHub, error) // TODO support new event delivery
}

// Transactor supplies methods for sending transaction proposals and transactions.
type Transactor interface {
	Sender
	ProposalSender
}

// Providers represents the SDK configured providers context.
type Providers interface {
	CoreProviders
	SvcProviders
}

// CoreProviders represents the SDK configured core providers context.
type CoreProviders interface {
	CryptoSuite() apicryptosuite.CryptoSuite
	StateStore() KVStore
	Config() apiconfig.Config
	SigningManager() SigningManager
	FabricProvider() FabricProvider
}

// SvcProviders represents the SDK configured service providers context.
type SvcProviders interface {
	DiscoveryProvider() DiscoveryProvider
	SelectionProvider() SelectionProvider
	ChannelProvider() ChannelProvider
}
