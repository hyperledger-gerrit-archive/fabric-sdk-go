/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticdiscovery

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

/**
 * Discovery Provider is used to discover peers on the network
 */

// DiscoveryProvider implements discovery provider
type DiscoveryProvider struct {
	config     apiconfig.Config
	peerFilter apifabclient.TargetFilter
}

// discoveryService implements discovery service
type discoveryService struct {
	config apiconfig.Config
	peers  []apifabclient.Peer
}

// MSPFilter is default filter
type MSPFilter struct {
	mspIDs []string
}

// Accept returns true if this peer is to be included in the target list
func (mf *MSPFilter) Accept(peer apifabclient.Peer) bool {
	if len(mf.mspIDs) == 0 {
		return true
	}
	for _, mspID := range mf.mspIDs {
		if mspID == peer.MSPID() {
			return true
		}
	}
	return false
}

// NewDiscoveryProvider returns discovery provider
func NewDiscoveryProvider(config apiconfig.Config, peerFilter apifabclient.TargetFilter) (*DiscoveryProvider, error) {
	return &DiscoveryProvider{config: config, peerFilter: peerFilter}, nil
}

// NewDiscoveryService return discovery service for specific channel
func (dp *DiscoveryProvider) NewDiscoveryService(channelID string) (apifabclient.DiscoveryService, error) {

	peers := []apifabclient.Peer{}
	filter := dp.peerFilter

	if channelID != "" {

		// Use configured channel peers
		chPeers, err := dp.config.ChannelPeers(channelID)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to read configuration for channel peers")
		}

		for _, p := range chPeers {

			newPeer, err := peer.New(dp.config, peer.FromPeerConfig(&p.NetworkPeer))
			if err != nil || newPeer == nil {
				return nil, errors.WithMessage(err, "NewPeer failed")
			}

			peers = append(peers, newPeer)
		}

		if filter == nil {
			// Get channel organizations to connect to their peers
			channelOrgs, err := dp.config.ChannelOrganizations(channelID)
			if err != nil {
				return nil, errors.WithMessage(err, "unable to read configuration for channel organizations")
			}
			mspIDs := make([]string, 0)
			// Get msp id for each organization
			for _, orgID := range channelOrgs {
				mspID, err := dp.config.MspID(orgID)
				if err != nil {
					return nil, errors.WithMessage(err, "unable to read configuration for msp ID")
				}
				mspIDs = append(mspIDs, mspID)
			}
			filter = &MSPFilter{mspIDs: mspIDs}
		}

	} else { // channel id is empty, return all configured peers

		netPeers, err := dp.config.NetworkPeers()
		if err != nil {
			return nil, errors.WithMessage(err, "unable to read configuration for network peers")
		}

		for _, p := range netPeers {
			newPeer, err := peer.New(dp.config, peer.FromPeerConfig(&p))
			if err != nil {
				return nil, errors.WithMessage(err, "NewPeerFromConfig failed")
			}

			peers = append(peers, newPeer)
		}
	}

	peers = filterTargets(peers, filter)

	return &discoveryService{config: dp.config, peers: peers}, nil
}

// GetPeers is used to get peers
func (ds *discoveryService) GetPeers() ([]apifabclient.Peer, error) {

	return ds.peers, nil
}

// filterTargets is helper method to filter peers
func filterTargets(peers []apifabclient.Peer, filter apifabclient.TargetFilter) []apifabclient.Peer {

	if filter == nil {
		return peers
	}

	filteredPeers := []apifabclient.Peer{}
	for _, peer := range peers {
		if filter.Accept(peer) {
			filteredPeers = append(filteredPeers, peer)
		}
	}

	return filteredPeers
}
