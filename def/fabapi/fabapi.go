/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabapi enables client usage of a Hyperledger Fabric network
package fabapi

import (
	"fmt"

	"github.com/hyperledger/fabric/bccsp"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/org"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/provider"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/session"
)

// Options encapsulates configuration for the SDK
type Options struct {
	// Quick access options
	ConfigFile string
	//OrgID      string // TODO: separate into context options

	// Options for default providers
	ConfigOpts     opt.ConfigOpts
	StateStoreOpts opt.StateStoreOpts

	// Factory methods to create clients and providers
	ProviderFactory provider.Factory
	ContextFactory  org.ProviderFactory
	SessionFactory  session.ProviderFactory

	// TODO extract hard-coded logger
}

// FabricSDK provides access (and context) to clients being managed by the SDK
type FabricSDK struct {
	Options

	// Implementations of client functionality (defaults are used if not specified)
	configProvider apiconfig.Config
	stateStore     apifabclient.KeyValueStore
	cryptoSuite    bccsp.BCCSP // TODO - maybe copy this interface into the API package

	// TODO: move context out?
	//Context *Context
}

// NewSDK initializes default clients
func NewSDK(options Options) (*FabricSDK, error) {
	// Construct SDK opts from the quick access options in setup
	sdkOpts := opt.SDKOpts{
		ConfigFile: options.ConfigFile,
	}

	sdk := FabricSDK{
		Options: options,
	}

	// Initialize default factories (if needed)
	if sdk.ProviderFactory == nil {
		sdk.ProviderFactory = provider.NewDefaultProviderFactory()
	}
	if sdk.ContextFactory == nil {
		sdk.ContextFactory = org.NewDefaultContextFactory()
	}
	if sdk.SessionFactory == nil {
		sdk.SessionFactory = session.NewDefaultSessionFactory()
	}

	// Initialize config provider
	config, err := sdk.ProviderFactory.NewConfigProvider(sdk.ConfigOpts, sdkOpts)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize config [%s]", err)
	}
	sdk.configProvider = config

	// Initialize crypto provider
	cryptosuite, err := sdk.ProviderFactory.NewCryptoSuiteProvider(sdk.configProvider.CSPConfig())
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize crypto suite [%s]", err)
	}
	sdk.cryptoSuite = cryptosuite

	// Initialize state store
	store, err := sdk.ProviderFactory.NewStateStoreProvider(sdk.StateStoreOpts, sdk.configProvider)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize state store [%s]", err)
	}
	sdk.stateStore = store

	// TODO: make creation of context explicit and separate from the SDK...?
	//context, err := NewContext(&sdk)
	//if err != nil {
	//	return nil, fmt.Errorf("Failed to initialize context [%s]", err)
	//}
	//sdk.Context = context

	return &sdk, nil
}

// ConfigProvider returns the Config provider of sdk.
func (sdk *FabricSDK) ConfigProvider() apiconfig.Config {
	return sdk.configProvider
}

// CryptoSuiteProvider returns the BCCSP provider of sdk.
func (sdk *FabricSDK) CryptoSuiteProvider() bccsp.BCCSP {
	return sdk.cryptoSuite
}

// StateStoreProvider returns the BCCSP provider of sdk.
func (sdk *FabricSDK) StateStoreProvider() apifabclient.KeyValueStore {
	return sdk.stateStore
}

// NewContext creates a context from an org
func (sdk *FabricSDK) NewContext(orgID string) (*org.Context, error) {
	return org.NewContext(sdk.ContextFactory, orgID, sdk.configProvider)
}

// NewSession creates a session from a context and a user (TODO)
func (sdk *FabricSDK) NewSession(user apifabclient.User) *session.Session {
	return session.NewSession(user, sdk.SessionFactory)
}

// NewSystemClient returns a new client for the system (operations not on a channel)
// TODO: Reduced immutable interface
// TODO: Parameter for setting up the peers
// TODO: session interface paramater
func (sdk *FabricSDK) NewSystemClient(session *session.Session) apifabclient.FabricClient {
	client := NewSystemClient(sdk.configProvider)

	client.SetCryptoSuite(sdk.cryptoSuite)
	client.SetStateStore(sdk.stateStore)

	return client
}

/*
TODO
func (sdk *FabricSDK) NewChannelClient(session Session) {

}*/
