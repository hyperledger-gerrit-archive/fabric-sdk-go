/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

/**
 * Mock Discovery Provider is used to mock peers on the network
 */

// MockStaticDiscoveryProvider implements mock discovery provider
type MockStaticDiscoveryProvider struct {
	Error error
	Peers []context.Peer
}

// MockStaticDiscoveryService implements mock discovery service
type MockStaticDiscoveryService struct {
	Error error
	Peers []context.Peer
}

// NewMockDiscoveryProvider returns mock discovery provider
func NewMockDiscoveryProvider(err error, peers []context.Peer) (*MockStaticDiscoveryProvider, error) {
	return &MockStaticDiscoveryProvider{Error: err, Peers: peers}, nil
}

// NewDiscoveryService return discovery service for specific channel
func (dp *MockStaticDiscoveryProvider) NewDiscoveryService(channelID string) (context.DiscoveryService, error) {
	if channelID == "error" {
		return nil, errors.New("Generate error when creating new discovery service")
	}
	return &MockStaticDiscoveryService{Error: dp.Error, Peers: dp.Peers}, nil
}

// GetPeers is used to discover eligible peers for chaincode
func (ds *MockStaticDiscoveryService) GetPeers() ([]context.Peer, error) {

	if ds.Error != nil {
		return nil, ds.Error
	}

	if ds.Peers == nil {
		mockPeer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}
		peers := make([]context.Peer, 0)
		peers = append(peers, &mockPeer)
		ds.Peers = peers
	}

	return ds.Peers, nil

}
