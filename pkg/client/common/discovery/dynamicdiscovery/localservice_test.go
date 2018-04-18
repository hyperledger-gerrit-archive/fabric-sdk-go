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

func TestLocalDiscoveryService(t *testing.T) {
	ctx := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", mspID1))
	config := &mocks.MockConfig{}
	ctx.SetEndpointConfig(config)

	localCtx := mocks.NewMockLocalContext(ctx, nil)

	peer1 := pfab.NetworkPeer{
		PeerConfig: pfab.PeerConfig{
			URL: "localhost:9999",
		},
		MSPID: mspID1,
	}
	config.SetCustomNetworkPeerCfg([]pfab.NetworkPeer{peer1})

	discClient := dyndiscmocks.NewMockDiscoveryClient()
	discClient.SetResponses(
		&dyndiscmocks.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{},
		},
	)

	clientProvider = func(ctx contextAPI.Client) (discoverClient, error) {
		return discClient, nil
	}

	service := newLocalService(
		options{
			refreshInterval: 500 * time.Millisecond,
			responseTimeout: 2 * time.Second,
		},
	)
	defer service.Close()

	err := service.Initialize(localCtx)
	assert.NoError(t, err)
	// Initialize again should produce no error
	err = service.Initialize(localCtx)
	assert.NoError(t, err)

	peers, err := service.GetPeers()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(peers))

	discClient.SetResponses(
		&dyndiscmocks.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
				{
					MSPID:        mspID1,
					Endpoint:     "localhost:7051",
					LedgerHeight: 5,
				},
			},
		},
	)

	time.Sleep(time.Second)

	peers, err = service.GetPeers()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(peers))

	discClient.SetResponses(
		&dyndiscmocks.MockDiscoverEndpointResponse{
			PeerEndpoints: []*discmocks.MockDiscoveryPeerEndpoint{
				{
					MSPID:        mspID1,
					Endpoint:     "localhost:7051",
					LedgerHeight: 5,
				},
				{
					MSPID:        mspID1,
					Endpoint:     "localhost:8051",
					LedgerHeight: 5,
				},
			},
		},
	)

	time.Sleep(time.Second)

	peers, err = service.GetPeers()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(peers))
}
