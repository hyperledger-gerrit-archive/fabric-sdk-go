/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

// MockDiscoveryService is a mock discovery service used for event endpoint discovery
type MockDiscoveryService struct {
	peers []apifabclient.Peer
}

// NewDiscoveryService returns a new mock discovery service
func NewDiscoveryService(peers ...apifabclient.Peer) apifabclient.DiscoveryService {
	return &MockDiscoveryService{
		peers: peers,
	}
}

// GetPeers returns a list of discovered peers
func (s *MockDiscoveryService) GetPeers() ([]apifabclient.Peer, error) {
	return s.peers, nil
}
