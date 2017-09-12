/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discoveryprovider

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

/**
 * Discovery Provider is used to discover peers on the network
 */

// DiscoveryProvider implements discovery provider
type DiscoveryProvider struct {
	config apiconfig.Config
}

// DiscoveryService implements discovery service
type DiscoveryService struct {
	config  apiconfig.Config
	channel apifabclient.Channel
}

// NewDiscoveryProvider returns discovery provider
func NewDiscoveryProvider(config apiconfig.Config) (apifabclient.DiscoveryProvider, error) {
	return &DiscoveryProvider{config: config}, nil
}

// NewDiscoveryService return discovery service for specific channel
func (dp *DiscoveryProvider) NewDiscoveryService(channel apifabclient.Channel) (apifabclient.DiscoveryService, error) {
	return &DiscoveryService{channel: channel, config: dp.config}, nil
}

// GetPeers is used to discover eligible peers for chaincode
func (ds *DiscoveryService) GetPeers(chaincodeID string) ([]apifabclient.Peer, error) {

	// TODO: Read config based on channel not hardcoded org ID
	// TODO: Incorporate chaincode policy
	peerConfig, err := ds.config.PeersConfig("peerorg1")
	if err != nil {
		return nil, fmt.Errorf("Error reading peer config: %v", err)
	}

	peers := []apifabclient.Peer{}

	for _, p := range peerConfig {
		peer, err := peer.NewPeerTLSFromCert(fmt.Sprintf("%s:%d", p.Host, p.Port),
			p.TLS.Certificate, p.TLS.ServerHostOverride, ds.config)
		if err != nil {
			return nil, fmt.Errorf("NewPeer return error: %v", err)
		}
		peers = append(peers, peer)
	}

	return peers, nil

}
