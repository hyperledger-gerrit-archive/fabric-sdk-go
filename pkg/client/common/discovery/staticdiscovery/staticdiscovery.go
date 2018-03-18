/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticdiscovery

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	"github.com/pkg/errors"
)

type peerCreator interface {
	CreatePeerFromConfig(peerCfg *core.NetworkPeer) (fab.Peer, error)
}

/**
 * Discovery Provider is used to discover peers on the network
 */

// DiscoveryProvider implements discovery provider
type DiscoveryProvider struct {
	config  core.Config
	fabPvdr peerCreator
	refs    []fab.DiscoveryService
	refLock sync.RWMutex
}

// discoveryService implements discovery service
type discoveryService struct {
	config core.Config
	peers  []fab.Peer
}

// New returns discovery provider
func New(config core.Config, fabPvdr peerCreator) (*DiscoveryProvider, error) {
	return &DiscoveryProvider{config: config, fabPvdr: fabPvdr}, nil
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

			newPeer, err := dp.fabPvdr.CreatePeerFromConfig(&p.NetworkPeer)
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
			newPeer, err := dp.fabPvdr.CreatePeerFromConfig(&p)
			if err != nil {
				return nil, errors.WithMessage(err, "NewPeerFromConfig failed")
			}

			peers = append(peers, newPeer)
		}
	}
	ds := &discoveryService{config: dp.config, peers: peers}

	dp.refLock.Lock()
	dp.refs = append(dp.refs, ds)
	dp.refLock.Unlock()

	return ds, nil
}

// Close the discovery services created by this provider
func (dp *DiscoveryProvider) Close() {
	dp.refLock.Lock()
	defer dp.refLock.Unlock()

	for _, svc := range dp.refs {
		svc.Close()
	}
}

// GetPeers is used to get peers
func (ds *discoveryService) GetPeers() ([]fab.Peer, error) {

	return ds.peers, nil
}

// Close frees up resources held by this service
func (ds *discoveryService) Close() {
}
