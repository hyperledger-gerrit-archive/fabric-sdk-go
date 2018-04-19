// +build devstable_

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const (
	peer1URL = "peer0.org1.example.com"
	peer2URL = "peer0.org2.example.com"
)

func TestDiscoveryClientPeers(t *testing.T) {
	sdk := mainSDK
	testSetup := mainTestSetup

	ctxProvider := sdk.Context(fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))
	ctx, err := ctxProvider()
	require.NoError(t, err, "error getting channel context")

	var client *discovery.Client
	client, err = discovery.New(ctx)
	require.NoError(t, err, "error creating discovery client")

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	req := discclient.NewRequest().OfChannel(testSetup.ChannelID).AddPeersQuery()

	peerCfg1, err := comm.NetworkPeerConfigFromURL(ctx.EndpointConfig(), peer1URL)
	require.NoErrorf(t, err, "error getting peer config for [%s]", peer1URL)

	responses, err := client.Send(reqCtx, req, peerCfg1.PeerConfig)
	require.NoError(t, err, "error calling discover service send")
	require.NotEmpty(t, responses, "expecting one response but got none")

	resp := responses[0]
	chanResp := resp.ForChannel(testSetup.ChannelID)

	peers, err := chanResp.Peers()
	require.NoError(t, err, "error getting config")
	require.NotEmpty(t, peers, "expecting at least one peer but got none")

	t.Logf("*** Peers for channel %s:\n", testSetup.ChannelID)
	for _, peer := range peers {
		aliveMsg := peer.AliveMessage.GetAliveMsg()
		if !assert.NotNil(t, aliveMsg, "got nil AliveMessage") {
			continue
		}
		if !assert.NotNil(t, aliveMsg.Membership, "got nil Membership") {
			continue
		}

		t.Logf("--- Endpoint: %s\n", aliveMsg.Membership.Endpoint)

		if !assert.NotNil(t, peer.StateInfoMessage, "got nil StateInfoMessage") {
			continue
		}

		stateInfo := peer.StateInfoMessage.GetStateInfo()
		if !assert.NotNil(t, stateInfo, "got nil stateInfo") {
			continue
		}

		if !assert.NotNil(t, stateInfo.Properties, "got nil stateInfo.Properties") {
			continue
		}

		t.Logf("--- Ledger Height: %d\n", stateInfo.Properties.LedgerHeight)
		t.Logf("--- Chaincodes:\n")
		for _, cc := range stateInfo.Properties.Chaincodes {
			t.Logf("------ %s:%s\n", cc.Name, cc.Version)
		}
	}
}

func TestDiscoveryClientLocalPeers(t *testing.T) {
	sdk := mainSDK

	ctxProvider := sdk.Context(fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))
	ctx, err := ctxProvider()
	require.NoError(t, err, "error getting channel context")

	var client *discovery.Client
	client, err = discovery.New(ctx)
	require.NoError(t, err, "error creating discovery client")

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	req := discclient.NewRequest().AddLocalPeersQuery()

	peerCfg1, err := comm.NetworkPeerConfigFromURL(ctx.EndpointConfig(), peer1URL)
	require.NoErrorf(t, err, "error getting peer config for [%s]", peer1URL)

	responses, err := client.Send(reqCtx, req, peerCfg1.PeerConfig)
	require.NoError(t, err, "error calling discover service send")
	require.NotEmpty(t, responses, "No responses")

	resp := responses[0]

	locResp := resp.ForLocal()

	peers, err := locResp.Peers()
	require.NoError(t, err, "error getting local peers")

	t.Logf("*** Local Peers:\n")
	for _, peer := range peers {
		aliveMsg := peer.AliveMessage.GetAliveMsg()
		if !assert.NotNil(t, aliveMsg, "got nil AliveMessage") {
			continue
		}
		if !assert.NotNil(t, aliveMsg.Membership, "got nil Membership") {
			continue
		}

		t.Logf("--- Endpoint: %s\n", aliveMsg.Membership.Endpoint)

		assert.Nil(t, peer.StateInfoMessage, "expected nil StateInfoMessage for local peer")
	}
}
