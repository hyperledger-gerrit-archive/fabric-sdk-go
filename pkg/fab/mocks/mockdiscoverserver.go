/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/gossip"
	"github.com/pkg/errors"
)

// MockDiscoverServer is a mock Discover server
type MockDiscoverServer struct {
	peersByOrg map[string]*discovery.Peers
}

// MockDiscoverServerOpt is an option for the MockDiscoverServer
type MockDiscoverServerOpt func(s *MockDiscoverServer)

// WithDiscoverServerPeers adds a set of mock peers to the MockDiscoverServer
func WithDiscoverServerPeers(peers ...*MockDiscoverPeerEndpoint) MockDiscoverServerOpt {
	return func(s *MockDiscoverServer) {
		peersByOrg := make(map[string]*discovery.Peers)
		for _, p := range peers {
			peers, ok := peersByOrg[p.MSPID]
			if !ok {
				peers = &discovery.Peers{}
				peersByOrg[p.MSPID] = peers
			}

			peers.Peers = append(peers.Peers, asDiscoveryPeer(p))
		}
		s.peersByOrg = peersByOrg
	}
}

// NewMockDiscoverServer returns a new MockDiscoverServer
func NewMockDiscoverServer(opts ...MockDiscoverServerOpt) *MockDiscoverServer {
	s := &MockDiscoverServer{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Discover Discovers a stream of blocks
func (s *MockDiscoverServer) Discover(ctx context.Context, request *discovery.SignedRequest) (*discovery.Response, error) {
	if request == nil {
		return nil, errors.New("nil request")
	}

	req := &discovery.Request{}
	err := proto.Unmarshal(request.Payload, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed parsing request")
	}
	if req.Authentication == nil {
		return nil, errors.New("access denied, no authentication info in request")
	}
	if len(req.Authentication.ClientIdentity) == 0 {
		return nil, errors.New("access denied, client identity wasn't supplied")
	}

	var res []*discovery.QueryResult
	for _, q := range req.Queries {
		if query := q.GetPeerQuery(); query != nil {
			res = append(res, s.getPeerQueryResult(query))
		}
		if query := q.GetConfigQuery(); query != nil {
			res = append(res, s.getConfigQueryResult(query))
		}
		if query := q.GetCcQuery(); query != nil {
			res = append(res, s.getCCQueryResult(query))
		}
	}
	return &discovery.Response{
		Results: res,
	}, nil
}

func (s *MockDiscoverServer) getPeerQueryResult(q *discovery.PeerMembershipQuery) *discovery.QueryResult {
	if s.peersByOrg != nil {
		return &discovery.QueryResult{
			Result: &discovery.QueryResult_Members{
				Members: &discovery.PeerMembershipResult{
					PeersByOrg: s.peersByOrg,
				},
			},
		}
	}
	return &discovery.QueryResult{
		Result: &discovery.QueryResult_Error{
			Error: &discovery.Error{
				Content: "no peers",
			},
		},
	}
}

func (s *MockDiscoverServer) getConfigQueryResult(q *discovery.ConfigQuery) *discovery.QueryResult {
	return &discovery.QueryResult{
		Result: &discovery.QueryResult_Error{
			Error: &discovery.Error{
				Content: "not implemented",
			},
		},
	}
}

func (s *MockDiscoverServer) getCCQueryResult(q *discovery.ChaincodeQuery) *discovery.QueryResult {
	return &discovery.QueryResult{
		Result: &discovery.QueryResult_Error{
			Error: &discovery.Error{
				Content: "not implemented",
			},
		},
	}
}

func asDiscoveryPeer(p *MockDiscoverPeerEndpoint) *discovery.Peer {
	memInfoMsg := &gossip.GossipMessage{
		Content: &gossip.GossipMessage_AliveMsg{
			AliveMsg: &gossip.AliveMessage{
				Membership: &gossip.Member{
					Endpoint: p.Endpoint,
				},
			},
		},
	}
	memInfoPayload, _ := proto.Marshal(memInfoMsg)

	stateInfoMsg := &gossip.GossipMessage{
		Content: &gossip.GossipMessage_StateInfo{
			StateInfo: &gossip.StateInfo{
				Properties: &gossip.Properties{
					Chaincodes:   nil,
					LedgerHeight: p.LedgerHeight,
				},
			},
		},
	}
	stateInfoPayload, _ := proto.Marshal(stateInfoMsg)

	return &discovery.Peer{
		MembershipInfo: &gossip.Envelope{
			Payload: memInfoPayload,
		},
		StateInfo: &gossip.Envelope{
			Payload: stateInfoPayload,
		},
	}
}
