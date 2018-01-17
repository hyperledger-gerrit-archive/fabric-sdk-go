/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabsdk enables client usage of a Hyperledger Fabric network.
package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicore"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"

	"reflect"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defclient"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
)

// FabricSDK provides access (and context) to clients being managed by the SDK.
type FabricSDK struct {
	opts Options

	configProvider    apiconfig.Config
	stateStore        apifabclient.KeyValueStore
	cryptoSuite       apicryptosuite.CryptoSuite
	discoveryProvider apifabclient.DiscoveryProvider
	selectionProvider apifabclient.SelectionProvider
	signingManager    apifabclient.SigningManager
	fabricProvider    apicore.FabricProvider
}

// Options holds the SDK configuration being manipulated by Option parameters.
type Options struct {
	Core    apisdk.CoreProviderFactory
	Service apisdk.ServiceProviderFactory
	Context apisdk.OrgClientFactory
	Session apisdk.SessionClientFactory
	Logger  apilogging.LoggerProvider
}

// Option configures the SDK.
type Option func(opts *Options) error

// WithCoreProvider ...
func WithCoreProvider(core apisdk.CoreProviderFactory) Option {
	return func(opts *Options) error {
		opts.Core = core
		return nil
	}
}

// WithServiceProvider ...
func WithServiceProvider(service apisdk.ServiceProviderFactory) Option {
	return func(opts *Options) error {
		opts.Service = service
		return nil
	}
}

// WithContextProvider ...
func WithContextProvider(context apisdk.OrgClientFactory) Option {
	return func(opts *Options) error {
		opts.Context = context
		return nil
	}
}

// WithSessionProvider ...
func WithSessionProvider(session apisdk.SessionClientFactory) Option {
	return func(opts *Options) error {
		opts.Session = session
		return nil
	}
}

// WithLoggerProvider ...
func WithLoggerProvider(logger apilogging.LoggerProvider) Option {
	return func(opts *Options) error {
		opts.Logger = logger
		return nil
	}
}

// ChannelClientOpts provides options for creating channel client
type ChannelClientOpts struct {
	OrgName        string
	ConfigProvider apiconfig.Config
}

// ChannelMgmtClientOpts provides options for creating channel management client
type ChannelMgmtClientOpts struct {
	OrgName        string
	ConfigProvider apiconfig.Config
}

// ResourceMgmtClientOpts provides options for creating resource management client
type ResourceMgmtClientOpts struct {
	OrgName        string
	TargetFilter   resmgmt.TargetFilter
	ConfigProvider apiconfig.Config
}

// ProviderInit interface allows for initializing providers
type ProviderInit interface {
	Initialize(sdk *FabricSDK) error
}

var defaultOptions = Options{
	Core:    defcore.NewProviderFactory(),
	Service: defsvc.NewProviderFactory(),
	Context: defclient.NewOrgClientFactory(),
	Session: defclient.NewSessionClientFactory(),
	Logger:  deflogger.LoggerProvider(),
}

// New initializes the SDK based on the set of options provided.
// A package suite containing the SDK implementation must be provided as an option.
// For example using the basic defaults: defpkgsuite.SDKOpt() in def/pkgsuite/defpkgsuite:
// fabsdk.New(defpkgsuite.SDKOpt())
func New(config apiconfig.Config, opts ...Option) (*FabricSDK, error) {

	// A nil struct stored in an interface is no longer nil, requiring reflect to check if it was nil
	if config == nil || reflect.ValueOf(config).IsNil() {
		return nil, errors.New("Missing configuration")
	}

	// Default configuration
	sdk := FabricSDK{
		configProvider: config,
		opts:           defaultOptions,
	}

	for _, option := range opts {
		err := option(&sdk.opts)

		if err != nil {
			return nil, errors.WithMessage(err, "Error in option passed to New")
		}
	}

	logging.InitLogger(sdk.opts.Logger)

	// Initialize crypto provider
	cs, err := sdk.opts.Core.NewCryptoSuiteProvider(sdk.configProvider)

	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize crypto suite")
	}

	sdk.cryptoSuite = cs

	// Setting this cryptosuite as the factory default
	cryptosuite.SetDefault(cs)

	// Initialize state store
	store, err := sdk.opts.Core.NewStateStoreProvider(sdk.configProvider)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize state store")
	}
	sdk.stateStore = store

	// Initialize Signing Manager
	signingMgr, err := sdk.opts.Core.NewSigningManager(sdk.CryptoSuiteProvider(), sdk.configProvider)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize signing manager")
	}
	sdk.signingManager = signingMgr

	// Initialize Fabric Provider
	fabricProvider, err := sdk.opts.Core.NewFabricProvider(sdk.configProvider, sdk.stateStore, sdk.cryptoSuite, sdk.signingManager)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize core fabric provider")
	}
	sdk.fabricProvider = fabricProvider

	// Initialize discovery provider
	discoveryProvider, err := sdk.opts.Service.NewDiscoveryProvider(sdk.configProvider)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize discovery provider")
	}
	if pi, ok := discoveryProvider.(ProviderInit); ok {
		pi.Initialize(&sdk)
	}
	sdk.discoveryProvider = discoveryProvider

	// Initialize selection provider (for selecting endorsing peers)
	selectionProvider, err := sdk.opts.Service.NewSelectionProvider(sdk.configProvider)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize selection provider")
	}
	if pi, ok := selectionProvider.(ProviderInit); ok {
		pi.Initialize(&sdk)
	}
	sdk.selectionProvider = selectionProvider

	return &sdk, nil
}

// ConfigProvider returns the Config provider of sdk.
func (sdk *FabricSDK) ConfigProvider() apiconfig.Config {
	return sdk.configProvider
}

// CryptoSuiteProvider returns the BCCSP provider of sdk.
func (sdk *FabricSDK) CryptoSuiteProvider() apicryptosuite.CryptoSuite {
	return sdk.cryptoSuite
}

// StateStoreProvider returns state store
func (sdk *FabricSDK) StateStoreProvider() apifabclient.KeyValueStore {
	return sdk.stateStore
}

// DiscoveryProvider returns discovery provider
func (sdk *FabricSDK) DiscoveryProvider() apifabclient.DiscoveryProvider {
	return sdk.discoveryProvider
}

// SelectionProvider returns selection provider
func (sdk *FabricSDK) SelectionProvider() apifabclient.SelectionProvider {
	return sdk.selectionProvider
}

// SigningManager returns signing manager
func (sdk *FabricSDK) SigningManager() apifabclient.SigningManager {
	return sdk.signingManager
}

// FabricProvider provides fabric objects such as peer and user
func (sdk *FabricSDK) FabricProvider() apicore.FabricProvider {
	return sdk.fabricProvider
}

// NewContext creates a context from an org
func (sdk *FabricSDK) NewContext(orgID string) (*OrgContext, error) {
	return newOrgContext(sdk.opts.Context, orgID, sdk.configProvider)
}

// NewSession creates a session from a context and a user (TODO)
func (sdk *FabricSDK) NewSession(c apisdk.Org, user apifabclient.User) (*Session, error) {
	return newSession(user, sdk.opts.Session), nil
}

// NewSystemClient returns a new client for the system (operations not on a channel)
// TODO: Reduced immutable interface
// TODO: Parameter for setting up the peers
func (sdk *FabricSDK) NewSystemClient(s apisdk.Session) (apifabclient.FabricClient, error) {
	return sdk.FabricProvider().NewClient(s.Identity())
}

// NewChannelMgmtClient returns a new client for managing channels
func (sdk *FabricSDK) NewChannelMgmtClient(userName string) (chmgmt.ChannelMgmtClient, error) {

	// Read default org name from configuration
	client, err := sdk.configProvider.Client()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to retrieve client from network config")
	}

	if client.Organization == "" {
		return nil, errors.New("must provide default organisation name in configuration")
	}

	opt := &ChannelMgmtClientOpts{OrgName: client.Organization, ConfigProvider: sdk.configProvider}

	return sdk.NewChannelMgmtClientWithOpts(userName, opt)
}

// NewChannelMgmtClientWithOpts returns a new client for managing channels with options
func (sdk *FabricSDK) NewChannelMgmtClientWithOpts(userName string, opt *ChannelMgmtClientOpts) (chmgmt.ChannelMgmtClient, error) {

	if opt == nil || opt.OrgName == "" {
		return nil, errors.New("organization name must be provided")
	}

	session, err := sdk.NewPreEnrolledUserSession(opt.OrgName, userName)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get pre-enrolled user session")
	}

	configProvider := sdk.ConfigProvider()
	if opt.ConfigProvider != nil {
		configProvider = opt.ConfigProvider
	}

	client, err := sdk.opts.Session.NewChannelMgmtClient(sdk, session, configProvider)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create new channel management client")
	}

	return client, nil
}

// NewResourceMgmtClient returns a new client for managing system resources
func (sdk *FabricSDK) NewResourceMgmtClient(userName string) (resmgmt.ResourceMgmtClient, error) {

	// Read default org name from configuration
	client, err := sdk.configProvider.Client()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to retrieve client from network config")
	}

	if client.Organization == "" {
		return nil, errors.New("must provide default organisation name in configuration")
	}

	opt := &ResourceMgmtClientOpts{OrgName: client.Organization, ConfigProvider: sdk.configProvider}

	return sdk.NewResourceMgmtClientWithOpts(userName, opt)
}

// NewResourceMgmtClientWithOpts returns a new resource management client (user has to be pre-enrolled)
func (sdk *FabricSDK) NewResourceMgmtClientWithOpts(userName string, opt *ResourceMgmtClientOpts) (resmgmt.ResourceMgmtClient, error) {

	if opt == nil || opt.OrgName == "" {
		return nil, errors.New("organization name must be provided")
	}

	session, err := sdk.NewPreEnrolledUserSession(opt.OrgName, userName)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get pre-enrolled user session")
	}

	configProvider := sdk.ConfigProvider()
	if opt.ConfigProvider != nil {
		configProvider = opt.ConfigProvider
	}

	client, err := sdk.opts.Session.NewResourceMgmtClient(sdk, session, configProvider, opt.TargetFilter)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to created new resource management client")
	}

	return client, nil
}

// NewChannelClient returns a new client for a channel
func (sdk *FabricSDK) NewChannelClient(channelID string, userName string) (apitxn.ChannelClient, error) {

	// Read default org name from configuration
	client, err := sdk.configProvider.Client()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to retrieve client from network config")
	}

	if client.Organization == "" {
		return nil, errors.New("must provide default organisation name in configuration")
	}

	opt := &ChannelClientOpts{OrgName: client.Organization, ConfigProvider: sdk.configProvider}

	return sdk.NewChannelClientWithOpts(channelID, userName, opt)
}

// NewChannelClientWithOpts returns a new client for a channel (user has to be pre-enrolled)
func (sdk *FabricSDK) NewChannelClientWithOpts(channelID string, userName string, opt *ChannelClientOpts) (apitxn.ChannelClient, error) {

	if opt == nil || opt.OrgName == "" {
		return nil, errors.New("organization name must be provided")
	}

	session, err := sdk.NewPreEnrolledUserSession(opt.OrgName, userName)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get pre-enrolled user session")
	}

	configProvider := sdk.ConfigProvider()
	if opt.ConfigProvider != nil {
		configProvider = opt.ConfigProvider
	}

	client, err := sdk.opts.Session.NewChannelClient(sdk, session, configProvider, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to created new channel client")
	}

	return client, nil
}

// NewPreEnrolledUser returns a new pre-enrolled user
// TODO: Rename this func to NewUser
func (sdk *FabricSDK) NewPreEnrolledUser(orgID string, userName string) (apifabclient.User, error) {

	credentialMgr, err := sdk.opts.Context.NewCredentialManager(orgID, sdk.ConfigProvider(), sdk.CryptoSuiteProvider())
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get credential manager")
	}

	signingIdentity, err := credentialMgr.GetSigningIdentity(userName)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get signing identity")
	}

	user, err := sdk.FabricProvider().NewUser(userName, signingIdentity)
	if err != nil {
		return nil, errors.WithMessage(err, "NewPreEnrolledUser returned error")
	}

	return user, nil
}

// NewPreEnrolledUserSession returns a new pre-enrolled user session
// TODO: Rename this func to NewUserSession
func (sdk *FabricSDK) NewPreEnrolledUserSession(orgID string, userName string) (*Session, error) {

	context, err := sdk.NewContext(orgID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get context for org")
	}

	user, err := sdk.NewPreEnrolledUser(orgID, userName)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get pre-enrolled user")
	}

	session, err := sdk.NewSession(context, user)
	if err != nil {
		return nil, errors.WithMessage(err, "NewSession returned error")
	}

	return session, nil
}
