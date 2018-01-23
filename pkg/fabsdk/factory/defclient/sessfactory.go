/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defclient

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"

	apichmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	apiresmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/chclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/chmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/resmgmtclient"
)

// SessionClientFactory represents the default implementation of a session client.
type SessionClientFactory struct{}

// NewSessionClientFactory creates a new default session client factory.
func NewSessionClientFactory() *SessionClientFactory {
	f := SessionClientFactory{}
	return &f
}

// NewChannelMgmtClient returns a client that manages channels (create/join channel)
func (f *SessionClientFactory) NewChannelMgmtClient(providers apisdk.Providers, session apisdk.SessionSvc) (apichmgmt.ChannelMgmtClient, error) {
	// For now settings are the same as for system client
	resource, err := providers.FabricProvider().NewResourceClient(session.Identity())
	if err != nil {
		return nil, err
	}
	ctx := chmgmtclient.Context{
		ProviderContext: providers,
		IdentityContext: session.Identity(),
		Resource:        resource,
	}
	return chmgmtclient.New(ctx)
}

// NewResourceMgmtClient returns a client that manages resources
func (f *SessionClientFactory) NewResourceMgmtClient(providers apisdk.Providers, session apisdk.SessionSvc, filter apiresmgmt.TargetFilter) (apiresmgmt.ResourceMgmtClient, error) {

	resource, err := providers.FabricProvider().NewResourceClient(session.Identity())
	if err != nil {
		return nil, err
	}

	discovery := providers.DiscoveryProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "create discovery provider failed")
	}

	ctx := resmgmtclient.Context{
		ProviderContext:   providers,
		IdentityContext:   session.Identity(),
		Resource:          resource,
		DiscoveryProvider: discovery,
		ChannelService:    session,
	}
	return resmgmtclient.New(ctx, filter)
}

// NewChannelClient returns a client that can execute transactions on specified channel
// TODO - better refactoring for testing and/or extract getChannelImpl to another package
func (f *SessionClientFactory) NewChannelClient(providers apisdk.Providers, session apisdk.SessionSvc, channelID string) (apitxn.ChannelClient, error) {
	// TODO: Add capablity to override sdk's selection and discovery provider

	channel, err := session.Channel(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "create channel failed")
	}

	discovery, err := providers.DiscoveryProvider().NewDiscoveryService(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "create discovery service failed")
	}

	selection, err := providers.SelectionProvider().NewSelectionService(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "create selection service failed")
	}

	eventHub, err := getEventHub(providers, channelID, session)
	if err != nil {
		return nil, errors.WithMessage(err, "getEventHub failed")
	}

	ctx := chclient.Context{
		ProviderContext:  providers,
		Channel:          channel,
		DiscoveryService: discovery,
		SelectionService: selection,
		EventHub:         eventHub,
	}
	return chclient.New(ctx)
}

func getEventHub(client fab.ProviderContext, channelID string, session apisdk.Session) (*events.EventHub, error) {

	peerConfig, err := client.Config().ChannelPeers(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "read configuration for channel peers failed")
	}

	var eventSource *apiconfig.PeerConfig

	for _, p := range peerConfig {
		if p.EventSource && p.MspID == session.Identity().MspID() {
			eventSource = &p.PeerConfig
			break
		}
	}

	if eventSource == nil {
		return nil, errors.New("unable to find peer event source for channel")
	}

	return events.NewEventHubFromConfig(client, session.Identity(), eventSource)

}
