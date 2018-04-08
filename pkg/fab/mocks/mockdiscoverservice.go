/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	reqcontext "context"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/gossip"
)

// MockDiscoverService implements a mock Discover service
type MockDiscoverService struct {
	peers []*discclient.Peer
}

// MockDiscoverPeerEndpoint contains information about a Discover peer endpoint
type MockDiscoverPeerEndpoint struct {
	MSPID        string
	Endpoint     string
	LedgerHeight uint64
}

// MockDiscoverOpt is an option for the MockDiscoverService
type MockDiscoverOpt func(*MockDiscoverService)

// WithDiscoverPeers sets the Discover peers
func WithDiscoverPeers(peerEndpoints ...*MockDiscoverPeerEndpoint) MockDiscoverOpt {
	return func(s *MockDiscoverService) {
		var peers []*discclient.Peer
		for _, endpoint := range peerEndpoints {
			peer := &discclient.Peer{
				MSPID:            endpoint.MSPID,
				AliveMessage:     newAliveMessage(endpoint),
				StateInfoMessage: newStateInfoMessage(endpoint),
			}
			peers = append(peers, peer)
		}
		s.peers = peers
	}
}

// NewMockDiscoverService returns a new mock Discover service
func NewMockDiscoverService(opts ...MockDiscoverOpt) *MockDiscoverService {
	s := &MockDiscoverService{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Send sends a Discovery request
func (m *MockDiscoverService) Send(ctx reqcontext.Context, req *discclient.Request) (discclient.Response, error) {
	return &response{
		peers: m.peers,
	}, nil
}

type response struct {
	peers []*discclient.Peer
}

func (r *response) ForChannel(string) discclient.ChannelResponse {
	return &channelResponse{
		peers: r.peers,
	}
}

type channelResponse struct {
	peers []*discclient.Peer
}

// Config returns a response for a config query, or error if something went wrong
func (cr *channelResponse) Config() (*discovery.ConfigResult, error) {
	panic("not implemented")
}

// Peers returns a response for a peer membership query, or error if something went wrong
func (cr *channelResponse) Peers() ([]*discclient.Peer, error) {
	return cr.peers, nil
}

// Endorsers returns the response for an endorser query for a given
// chaincode in a given channel context, or error if something went wrong.
// The method returns a random set of endorsers, such that signatures from all of them
// combined, satisfy the endorsement policy.
func (cr *channelResponse) Endorsers(string) (discclient.Endorsers, error) {
	panic("not implemented")
}

func newAliveMessage(endpoint *MockDiscoverPeerEndpoint) *gossip.SignedGossipMessage {
	return &gossip.SignedGossipMessage{
		GossipMessage: &gossip.GossipMessage{
			Content: &gossip.GossipMessage_AliveMsg{
				AliveMsg: &gossip.AliveMessage{
					Membership: &gossip.Member{
						Endpoint: endpoint.Endpoint,
					},
				},
			},
		},
	}
}

func newStateInfoMessage(endpoint *MockDiscoverPeerEndpoint) *gossip.SignedGossipMessage {
	return &gossip.SignedGossipMessage{
		GossipMessage: &gossip.GossipMessage{
			Content: &gossip.GossipMessage_StateInfo{
				StateInfo: &gossip.StateInfo{
					Properties: &gossip.Properties{
						LedgerHeight: endpoint.LedgerHeight,
					},
				},
			},
		},
	}
}
