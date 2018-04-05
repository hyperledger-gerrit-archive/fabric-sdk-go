/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticdiscovery

import (
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"

	"github.com/pkg/errors"
)

/**
 * Discovery Provider is used to discover peers on the network
 */

// DiscoveryProvider implements discovery provider
type DiscoveryProvider struct {
	config  fab.EndpointConfig
	fabPvdr contextAPI.Providers
}

// discoveryService implements discovery service
type discoveryService struct {
	config fab.EndpointConfig
	peers  []fab.Peer
}

// New returns discovery provider
func New(config fab.EndpointConfig) (*DiscoveryProvider, error) {
	return &DiscoveryProvider{config: config}, nil
}

// Initialize initializes the DiscoveryProvider
func (dp *DiscoveryProvider) Initialize(fabPvdr contextAPI.Providers) error {
	dp.fabPvdr = fabPvdr
	return nil
}

// CreateDiscoveryService return discovery service for specific channel
func (dp *DiscoveryProvider) CreateDiscoveryService(channelID string) (fab.DiscoveryService, error) {

	peers := []fab.Peer{}

	if channelID != "" {

		// Use configured channel peers
		chPeers, err := dp.config.ChannelPeers(channelID)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to read configuration for channel peers")
		}

		for _, p := range chPeers {
			newPeer, err := peerImpl.New(dp.fabPvdr.EndpointConfig(), peerImpl.FromPeerConfig(&p.NetworkPeer))
			if err != nil || newPeer == nil {
				return nil, errors.WithMessage(err, "NewPeer failed")
			}

			peers = append(peers, newPeer)
		}

	} else { // channel id is empty, return all configured peers

		netPeers, err := dp.config.NetworkPeers()
		if err != nil {
			return nil, errors.WithMessage(err, "unable to read configuration for network peers")
		}

		for _, p := range netPeers {
			newPeer, err := peerImpl.New(dp.fabPvdr.EndpointConfig(), peerImpl.FromPeerConfig(&p))
			if err != nil {
				return nil, errors.WithMessage(err, "NewPeerFromConfig failed")
			}

			peers = append(peers, newPeer)
		}
	}

	return &discoveryService{config: dp.config, peers: peers}, nil
}

// GetPeers is used to get peers
func (ds *discoveryService) GetPeers() ([]fab.Peer, error) {

	return ds.peers, nil
}
