/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"github.com/hyperledger/fabric-sdk-go/api/apicore"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// ChannelProvider keeps context across ChannelService instances.
//
// TODO: add cache for dynamic channel configuration. This cache is updated
// by channel services, as only channel service have an identity context.
// TODO: add listener for channel config changes. Upon channel config change,
// underlying channel services need to recreate their channel clients.
type ChannelProvider struct {
	fabricProvider    apicore.FabricProvider
	discoveryProvider apifabclient.DiscoveryProvider
}

// New creates a ChannelProvider based on a context
func New(fabricProvider apicore.FabricProvider, discoveryProvider apifabclient.DiscoveryProvider) (*ChannelProvider, error) {
	cp := ChannelProvider{
		fabricProvider,
		discoveryProvider,
	}
	return &cp, nil
}

// NewChannelService creates a ChannelService for an identity
func (cp *ChannelProvider) NewChannelService(ic apifabclient.IdentityContext, channelID string) (apifabclient.ChannelService, error) {
	ds, err := cp.discoveryProvider.NewDiscoveryService(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create discovery service")
	}

	cs := ChannelService{
		fabricProvider:   cp.fabricProvider,
		discoveryService: ds,
		identityContext:  ic,
		channelID:        channelID,
	}
	return &cs, nil
}

// ChannelService provides Channel clients and maintains contexts for them.
// the identity context is used
//
// TODO: add cache for channel rather than reconstructing each time.
type ChannelService struct {
	fabricProvider   apicore.FabricProvider
	discoveryService apifabclient.DiscoveryService
	identityContext  apifabclient.IdentityContext
	channelID        string
}

// Channel returns the named Channel client.
func (cs *ChannelService) Channel() (apifabclient.Channel, error) {
	channel, err := cs.fabricProvider.NewChannelClient(cs.identityContext, cs.channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create channel client")
	}

	peers, err := cs.discoveryService.GetPeers()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to discover channel peers")
	}

	for _, p := range peers {
		logger.Infof("adding peer: %v", p)
		channel.AddPeer(p)
	}

	// TODO - replace with channel config!
	err = channel.Initialize([]byte{})
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize channel client")
	}

	return channel, nil
}

// EventHub returns the EventHub for the named channel.
func (cs *ChannelService) EventHub() (apifabclient.EventHub, error) {
	return cs.fabricProvider.NewEventHub(cs.identityContext, cs.channelID)
}

// Ledger provides access to ledger blocks.
func (cs *ChannelService) Ledger() (apifabclient.ChannelLedger, error) {
	return cs.fabricProvider.NewChannelClient(cs.identityContext, cs.channelID)
}
