/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/pkg/errors"
)

// Identity supplies the serialized identity and key reference.
type Identity interface {
	MspID() string
	SerializedIdentity() ([]byte, error)
	PrivateKey() core.Key
}

// Session primarily represents the session and identity context
type Session interface {
	Identity
}

// Client supplies the configuration and signing identity to client objects.
type Client interface {
	Providers
	Identity
}

// Providers represents the SDK configured providers context.
type Providers interface {
	core.Providers
	fab.Providers
}

//sdkContext implementation for Providers interface
type sdkContext struct {
	config            core.Config
	stateStore        core.KVStore
	cryptoSuite       core.CryptoSuite
	discoveryProvider fab.DiscoveryProvider
	selectionProvider fab.SelectionProvider
	signingManager    core.SigningManager
	identityManager   map[string]core.IdentityManager
	fabricProvider    fab.InfraProvider
	channelProvider   fab.ChannelProvider
	identity          Identity
}

// Config returns the Config provider of sdk.
func (c *sdkContext) Config() core.Config {
	return c.config
}

// CryptoSuite returns the BCCSP provider of sdk.
func (c *sdkContext) CryptoSuite() core.CryptoSuite {
	return c.cryptoSuite
}

// IdentityManager returns identity manager for organization
func (c *sdkContext) IdentityManager(orgName string) (core.IdentityManager, bool) {
	mgr, ok := c.identityManager[strings.ToLower(orgName)]
	return mgr, ok
}

// SigningManager returns signing manager
func (c *sdkContext) SigningManager() core.SigningManager {
	return c.signingManager
}

// StateStore returns state store
func (c *sdkContext) StateStore() core.KVStore {
	return c.stateStore
}

// DiscoveryProvider returns discovery provider
func (c *sdkContext) DiscoveryProvider() fab.DiscoveryProvider {
	return c.discoveryProvider
}

// SelectionProvider returns selection provider
func (c *sdkContext) SelectionProvider() fab.SelectionProvider {
	return c.selectionProvider
}

// ChannelProvider provides channel services.
func (c *sdkContext) ChannelProvider() fab.ChannelProvider {
	return c.channelProvider
}

// FabricProvider provides fabric objects such as peer and user
func (c *sdkContext) FabricProvider() fab.InfraProvider {
	return c.fabricProvider
}

//MspID returns MSPID
func (c *sdkContext) MspID() string {
	if c.identity == nil {
		return ""
	}
	return c.identity.MspID()
}

//SerializedIdentity returns serialized identity
func (c *sdkContext) SerializedIdentity() ([]byte, error) {
	if c.identity == nil {
		return nil, errors.New("identity not yet initialized")
	}
	return c.identity.SerializedIdentity()
}

//PrivateKey returns private key
func (c *sdkContext) PrivateKey() core.Key {
	if c.identity == nil {
		return nil
	}
	return c.identity.PrivateKey()
}

//SDKContextParams parameter for creating FabContext
type SDKContextParams func(opts *sdkContext)

//WithConfig sets config to FabContext
func WithConfig(config core.Config) SDKContextParams {
	return func(ctx *sdkContext) {
		ctx.config = config
	}
}

//WithStateStore sets state store to FabContext
func WithStateStore(stateStore core.KVStore) SDKContextParams {
	return func(ctx *sdkContext) {
		ctx.stateStore = stateStore
	}
}

//WithCryptoSuite sets cryptosuite parameter to FabContext
func WithCryptoSuite(cryptoSuite core.CryptoSuite) SDKContextParams {
	return func(ctx *sdkContext) {
		ctx.cryptoSuite = cryptoSuite
	}
}

//WithDiscoveryProvider sets discoveryProvider to FabContext
func WithDiscoveryProvider(discoveryProvider fab.DiscoveryProvider) SDKContextParams {
	return func(ctx *sdkContext) {
		ctx.discoveryProvider = discoveryProvider
	}
}

//WithSelectionProvider sets selectionProvider to FabContext
func WithSelectionProvider(selectionProvider fab.SelectionProvider) SDKContextParams {
	return func(ctx *sdkContext) {
		ctx.selectionProvider = selectionProvider
	}
}

//WithSigningManager sets signingManager to FabContext
func WithSigningManager(signingManager core.SigningManager) SDKContextParams {
	return func(ctx *sdkContext) {
		ctx.signingManager = signingManager
	}
}

//WithIdentityManager sets identityManagers maps to FabContext
func WithIdentityManager(identityManagers map[string]core.IdentityManager) SDKContextParams {
	return func(ctx *sdkContext) {
		ctx.identityManager = identityManagers
	}
}

//WithFabricProvider sets fabricProvider maps to FabContext
func WithFabricProvider(fabricProvider fab.InfraProvider) SDKContextParams {
	return func(ctx *sdkContext) {
		ctx.fabricProvider = fabricProvider
	}
}

//WithChannelProvider sets channelProvider to FabContext
func WithChannelProvider(channelProvider fab.ChannelProvider) SDKContextParams {
	return func(ctx *sdkContext) {
		ctx.channelProvider = channelProvider
	}
}

//Create creates context client
func Create(ic Identity, params ...SDKContextParams) Client {
	fabCtx := sdkContext{identity: ic}
	for _, param := range params {
		param(&fabCtx)
	}
	return &fabCtx
}
