/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
)

// filterService implements discovery service
type filterService struct {
	discoveryService context.DiscoveryService
	targetFilter     context.TargetFilter
}

// NewDiscoveryFilterService return discovery service with filter
func NewDiscoveryFilterService(discoveryService context.DiscoveryService, targetFilter context.TargetFilter) context.DiscoveryService {
	return &filterService{discoveryService: discoveryService, targetFilter: targetFilter}
}

// GetPeers is used to get peers
func (fs *filterService) GetPeers() ([]context.Peer, error) {
	peers, err := fs.discoveryService.GetPeers()
	if err != nil {
		return nil, err
	}
	targets := filterTargets(peers, fs.targetFilter)
	return targets, nil
}

// filterTargets is helper method to filter peers
func filterTargets(peers []context.Peer, filter context.TargetFilter) []context.Peer {

	if filter == nil {
		return peers
	}

	filteredPeers := []context.Peer{}
	for _, peer := range peers {
		if filter.Accept(peer) {
			filteredPeers = append(filteredPeers, peer)
		}
	}

	return filteredPeers
}
