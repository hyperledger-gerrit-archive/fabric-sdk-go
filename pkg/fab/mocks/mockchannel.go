/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

// Client supplies the configuration and signing identity to client objects.
type Client struct {
	context.Providers
	msp.SigningIdentity
}

//Channel supplies the configuration for channel context client
type Channel struct {
	context.Client
	discovery      fab.DiscoveryService
	selection      fab.SelectionService
	channelService fab.ChannelService
	channelID      string
}

//Providers returns core providers
func (c *Channel) Providers() context.Client {
	return c
}

//DiscoveryService returns core discovery service
func (c *Channel) DiscoveryService() fab.DiscoveryService {
	return c.discovery
}

//SelectionService returns selection service
func (c *Channel) SelectionService() fab.SelectionService {
	return c.selection
}

//ChannelService returns channel service
func (c *Channel) ChannelService() fab.ChannelService {
	return c.channelService
}

//ChannelID returns channel id
func (c *Channel) ChannelID() string {
	return c.channelID
}

type mockClientContext struct {
	context.Providers
	msp.SigningIdentity
}

//NewMockChannel creates new mock channel
func NewMockChannel(channelID string) (*Channel, error) {

	providerCtx := NewMockProviderContext()

	user := mspmocks.NewMockSigningIdentity("user", "user")

	ctx := &mockClientContext{
		Providers:       providerCtx,
		SigningIdentity: user,
	}

	// Set up mock channel service
	chProvider, err := NewMockChannelProvider(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "mock channel provider creation failed")
	}
	channelService, err := chProvider.ChannelService(ctx, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "mock channel service creation failed")
	}

	peers := []fab.Peer{NewMockPeer("Peer1", "http://peer1.com")}

	// Set up mock discovery service
	mockDiscovery, err := NewMockDiscoveryProvider(nil, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockDiscoveryProvider failed")
	}
	discoveryService, err := mockDiscovery.CreateDiscoveryService(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get discovery service to create channel client")
	}

	// Set up mock selection service
	mockSelection, err := NewMockSelectionProvider(nil, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockSelectinProvider failed")
	}
	selectionService, err := mockSelection.CreateSelectionService("mychannel")
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get discovery service to create channel client")
	}

	channel := &Channel{
		Client:         ctx,
		selection:      selectionService,
		discovery:      discoveryService,
		channelService: channelService,
		channelID:      channelID,
	}

	return channel, nil
}
