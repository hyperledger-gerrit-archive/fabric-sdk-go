/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
)

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
