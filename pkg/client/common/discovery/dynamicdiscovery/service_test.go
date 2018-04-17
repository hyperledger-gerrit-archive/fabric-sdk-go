/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"testing"
	"time"

	dyndiscmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/dynamicdiscovery/mocks"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	pfab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
)

func TestDiscoveryService(t *testing.T) {
	ctx := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", mspID))
	config := &config{
		EndpointConfig: mocks.NewMockEndpointConfig(),
		peers: []pfab.ChannelPeer{
			{
				NetworkPeer: pfab.NetworkPeer{
					PeerConfig: pfab.PeerConfig{
						URL: "localhost:9999",
					},
				},
			},
		},
	}
	ctx.SetEndpointConfig(config)

	discClient := dyndiscmocks.NewMockDiscoveryClient()
	clientProvider = func(ctx contextAPI.Client) (discoverClient, error) {
		return discClient, nil
	}

	membershipService := newService(
		options{
			refreshInterval: 1 * time.Second,
			responseTimeout: 2 * time.Second,
		},
	)
	defer membershipService.Close()

	chCtx := mocks.NewMockChannelContext(ctx, ch)
	err := membershipService.Initialize(chCtx)
	assert.NoError(t, err)
	// Initialize again should produce no error
	err = membershipService.Initialize(chCtx)
	assert.NoError(t, err)

	peers, err := membershipService.GetPeers()

	// Should have error due to no peers
	assert.Error(t, err)
	assert.Equal(t, 0, len(peers))

	discClient.SetResponses(
		&dyndiscmocks.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
				{
					MSPID:        mspID,
					Endpoint:     "localhost:7051",
					LedgerHeight: 5,
				},
			},
		},
	)

	peers, err = membershipService.GetPeers()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(peers))

	discClient.SetResponses(
		&dyndiscmocks.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
				{
					MSPID:        mspID,
					Endpoint:     "localhost:7051",
					LedgerHeight: 5,
				},
				{
					MSPID:        mspID,
					Endpoint:     "localhost:8051",
					LedgerHeight: 5,
				},
			},
		},
	)

	time.Sleep(3 * time.Second)

	peers, err = membershipService.GetPeers()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(peers))
}
