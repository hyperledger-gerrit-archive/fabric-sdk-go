/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/pkg/errors"
)

//FabricProvider interface
type FabricProvider interface {
	CreateChannelClient(user IdentityContext, cfg fab.ChannelCfg) (fab.Channel, error)
	CreateChannelLedger(ic IdentityContext, name string) (fab.ChannelLedger, error)
	CreateChannelConfig(user IdentityContext, name string) (fab.ChannelConfig, error)
	//CreateResourceClient(user IdentityContext) (Resource, error)
	CreateChannelTransactor(ic IdentityContext, cfg fab.ChannelCfg) (fab.Transactor, error)
	CreateEventHub(ic IdentityContext, name string) (fab.EventHub, error)
	CreatePeerFromConfig(peerCfg *core.NetworkPeer) (fab.Peer, error)
	CreateOrdererFromConfig(cfg *core.OrdererConfig) (fab.Orderer, error)
	CreateUser(name string, signingIdentity *api.SigningIdentity) (api.User, error)
}

// Providers represents the SDK configured providers context.
type Providers interface {
	CoreProviders
	SvcProviders
}

// CoreProviders represents the SDK configured core providers context.
type CoreProviders interface {
	CryptoSuite() core.CryptoSuite
	StateStore() api.KVStore
	Config() core.Config
	SigningManager() api.SigningManager
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
	SigningManager() api.SigningManager
	Config() core.Config
	CryptoSuite() core.CryptoSuite
}

// IdentityContext supplies the serialized identity and key reference.
//
// TODO - refactor SigningIdentity and this interface.
type IdentityContext interface {
	MspID() string
	Identity() ([]byte, error)
	PrivateKey() core.Key
}

//ProviderOption provides parameters for creating core
type ProviderOption func(opts *ProviderOptions) error

//ClientOption provides parameters for creating core
type ClientOption func(opts *ClientOptions) error

//ProviderOptions supplies identity options to caller
type ProviderOptions struct {
	providers Providers
}

//ClientOptions supplies identity options to caller
type ClientOptions struct {
	orgID    string
	identity IdentityContext
	ok       bool
}

//Provider returns core context
type Provider struct {
	Provider Providers
}

//Client exposes providers and identity context
type Client struct {
	Provider
	IdentityContext
}

// NewProvider returns identity context takes providers
func NewProvider(options ...ProviderOption) (*Provider, error) {

	opts := ProviderOptions{} //

	for _, option := range options {
		err := option(&opts)
		if err != nil {
			return nil, errors.WithMessage(err, "Error in option passed to client")
		}
	}
	cc := Provider{}
	cc.Provider = opts.providers
	return &cc, nil
}

//NewClient new client
func NewClient(coreProvider *Provider, options ...ClientOption) (*Client, error) {
	opts := ClientOptions{}

	for _, option := range options {
		err := option(&opts)
		if err != nil {
			return nil, errors.WithMessage(err, "Error in option passed to client")
		}
	}
	if !opts.ok {
		return nil, errors.New("Missing identity")
	}
	cc := Client{}
	cc.Provider = *coreProvider
	cc.IdentityContext = opts.identity

	return &cc, nil
}

// WithIdentity returns option with identity
func WithIdentity(identity IdentityContext) ClientOption {
	return func(o *ClientOptions) error {
		if o.ok {
			return errors.New("Identity already determined")
		}
		o.identity = identity
		o.orgID = identity.MspID()
		o.ok = true
		return nil
	}
}

// WithProvider returns option with provider
func WithProvider(provider Providers) ProviderOption {
	return func(o *ProviderOptions) error {
		o.providers = provider
		return nil
	}
}
