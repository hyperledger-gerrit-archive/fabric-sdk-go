/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defclient

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/chclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/pkg/errors"
)

// SessionClientFactory represents the default implementation of a session client.
type SessionClientFactory struct{}

// NewSessionClientFactory creates a new default session client factory.
func NewSessionClientFactory() *SessionClientFactory {
	f := SessionClientFactory{}
	return &f
}

// NewChannelClient returns a client that can execute transactions on specified channel
func (f *SessionClientFactory) NewChannelClient(providers context.Providers, session context.SessionContext, channelID string, targetFilter context.TargetFilter) (*chclient.ChannelClient, error) {

	chProvider := providers.ChannelProvider()
	chService, err := chProvider.NewChannelService(session, channelID)
	if err != nil {
		return &chclient.ChannelClient{}, errors.WithMessage(err, "create channel service failed")
	}

	discoveryService, err := providers.DiscoveryProvider().NewDiscoveryService(channelID)
	if err != nil {
		return &chclient.ChannelClient{}, errors.WithMessage(err, "create discovery service failed")
	}

	discoveryService = discovery.NewDiscoveryFilterService(discoveryService, targetFilter)

	selection, err := providers.SelectionProvider().NewSelectionService(channelID)
	if err != nil {
		return &chclient.ChannelClient{}, errors.WithMessage(err, "create selection service failed")
	}

	ctx := chclient.Context{
		ProviderContext:  providers,
		DiscoveryService: discoveryService,
		SelectionService: selection,
		ChannelService:   chService,
	}
	return chclient.New(ctx)
}
