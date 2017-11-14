/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defprovider

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	cons "github.com/hyperledger/fabric-sdk-go/api/apitxn/consclient"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	chImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/chclient"
	consImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/consclient"
)

// SessionClientFactory represents the default implementation of a session client.
type SessionClientFactory struct{}

// NewSessionClientFactory creates a new default session client factory.
func NewSessionClientFactory() *SessionClientFactory {
	f := SessionClientFactory{}
	return &f
}

// NewSystemClient returns a new FabricClient.
func (f *SessionClientFactory) NewSystemClient(sdk context.SDK, session context.Session, config apiconfig.Config) (fab.FabricClient, error) {
	client := clientImpl.NewClient(config)

	client.SetCryptoSuite(sdk.CryptoSuiteProvider())
	client.SetStateStore(sdk.StateStoreProvider())
	client.SetUserContext(session.Identity())
	client.SetSigningManager(sdk.SigningManager())

	return client, nil
}

// NewConsortiumClient returns a client that manages consortium (create/join channel)
func (f *SessionClientFactory) NewConsortiumClient(sdk context.SDK, session context.Session, config apiconfig.Config) (cons.ConsortiumClient, error) {
	// For now settings are the same as for system client
	client, err := f.NewSystemClient(sdk, session, config)
	if err != nil {
		return nil, err
	}
	return consImpl.NewConsortiumClient(client, config)
}

// NewChannelClient returns a client that can execute transactions on specified channel
func (f *SessionClientFactory) NewChannelClient(sdk context.SDK, session context.Session, config apiconfig.Config, channelID string) (apitxn.ChannelClient, error) {

	// TODO: Add capablity to override sdk's selection and discovery provider

	client := clientImpl.NewClient(sdk.ConfigProvider())
	client.SetCryptoSuite(sdk.CryptoSuiteProvider())
	client.SetStateStore(sdk.StateStoreProvider())
	client.SetUserContext(session.Identity())
	client.SetSigningManager(sdk.SigningManager())

	channel, err := getChannel(client, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "create channel failed")
	}

	discovery, err := sdk.DiscoveryProvider().NewDiscoveryService(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "create discovery service failed")
	}

	selection, err := sdk.SelectionProvider().NewSelectionService(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "create selection service failed")
	}

	eventHub, err := getEventHub(client, channelID, session)
	if err != nil {
		return nil, errors.WithMessage(err, "getEventHub failed")
	}

	return chImpl.NewChannelClient(client, channel, discovery, selection, eventHub)
}

// getChannel is helper method to initializes and returns a channel based on config
func getChannel(client fab.FabricClient, channelID string) (fab.Channel, error) {

	channel, err := client.NewChannel(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "NewChannel failed")
	}

	chCfg, err := client.Config().ChannelConfig(channel.Name())
	if err != nil || chCfg == nil {
		return nil, errors.Errorf("reading channel config failed: %s", err)
	}

	chOrderers, err := client.Config().ChannelOrderers(channel.Name())
	if err != nil {
		return nil, errors.WithMessage(err, "reading channel orderers failed")
	}

	for _, ordererCfg := range chOrderers {

		orderer, err := orderer.NewOrdererFromConfig(&ordererCfg, client.Config())
		if err != nil {
			return nil, errors.WithMessage(err, "NewOrderer failed")
		}
		err = channel.AddOrderer(orderer)
		if err != nil {
			return nil, errors.WithMessage(err, "adding orderer failed")
		}
	}

	return channel, nil
}

func getEventHub(client fab.FabricClient, channelID string, session context.Session) (*events.EventHub, error) {

	peerConfig, err := client.Config().ChannelPeers(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "read configuration for channel peers failed")
	}

	serverHostOverride := ""
	var eventSource *apiconfig.PeerConfig

	for _, p := range peerConfig {

		if p.EventSource && p.MspID == session.Identity().MspID() {
			serverHostOverride = ""
			if str, ok := p.GRPCOptions["ssl-target-name-override"].(string); ok {
				serverHostOverride = str
			}
			eventSource = &p.PeerConfig
			break
		}
	}

	if eventSource == nil {
		return nil, errors.New("unable to find peer event source for channel")
	}

	// Event source found create event hub
	eventHub, err := events.NewEventHub(client)
	if err != nil {
		return nil, err
	}

	eventHub.SetPeerAddr(eventSource.EventURL, eventSource.TLSCACerts.Path, serverHostOverride)

	return eventHub, nil
}
