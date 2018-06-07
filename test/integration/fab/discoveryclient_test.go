// +build devstable

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
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	fabdiscovery "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

const (
	peer1URL      = "peer0.org1.example.com"
	peer2URL      = "peer0.org2.example.com"
	org1AdminUser = "Admin"
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

	peerCfg1, err := comm.NetworkPeerConfig(ctx.EndpointConfig(), peer1URL)
	require.NoErrorf(t, err, "error getting peer config for [%s]", peer1URL)

	responses, err := client.Send(reqCtx, req, peerCfg1.PeerConfig)
	require.NoError(t, err, "error calling discover service send")
	require.NotEmpty(t, responses, "expecting one response but got none")

	resp := responses[0]
	chanResp := resp.ForChannel(testSetup.ChannelID)

	peers, err := chanResp.Peers()
	require.NoError(t, err, "error getting peers")
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
		t.Logf("--- LeftChannel: %t\n", stateInfo.Properties.LeftChannel)
		t.Log("--- Chaincodes:\n")
		for _, cc := range stateInfo.Properties.Chaincodes {
			t.Logf("------ %s:%s\n", cc.Name, cc.Version)
		}
	}
}

func TestDiscoveryClientLocalPeers(t *testing.T) {
	sdk := mainSDK

	// By default, query for local peers (outside of a channel) requires admin privileges.
	// To bypass this restriction, set peer.discovery.orgMembersAllowedAccess=true in core.yaml.
	ctxProvider := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))
	ctx, err := ctxProvider()
	require.NoError(t, err, "error getting channel context")

	var client *discovery.Client
	client, err = discovery.New(ctx)
	require.NoError(t, err, "error creating discovery client")

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	req := discclient.NewRequest().AddLocalPeersQuery()

	peerCfg1, err := comm.NetworkPeerConfig(ctx.EndpointConfig(), peer1URL)
	require.NoErrorf(t, err, "error getting peer config for [%s]", peer1URL)

	responses, err := client.Send(reqCtx, req, peerCfg1.PeerConfig)
	require.NoError(t, err, "error calling discover service send")
	require.NotEmpty(t, responses, "No responses")

	resp := responses[0]

	locResp := resp.ForLocal()

	peers, err := locResp.Peers()
	require.NoError(t, err, "error getting local peers")

	t.Log("*** Local Peers:\n")
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

func TestDiscoveryClientSelectEndorsers(t *testing.T) {
	sdk := mainSDK
	testSetup := mainTestSetup

	ctxProvider := sdk.Context(fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))
	ctx, err := ctxProvider()
	require.NoError(t, err, "error getting channel context")

	chaincodeID := integration.GenerateRandomID()
	initArgs := [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b"), []byte(integration.ExampleCCInitB)}

	instResp, err := integration.InstallAndInstantiateCC(
		sdk, fabsdk.WithUser("Admin"), orgName,
		chaincodeID, "github.com/example_cc", "v0",
		integration.GetDeployPath(), initArgs,
		func() (*cb.SignaturePolicyEnvelope, error) {
			return cauthdsl.FromString("AND ('Org1MSP.member','Org2MSP.member')")
		},
	)

	require.Nil(t, err, "InstallAndInstantiateExampleCC returned error")
	require.NotEmpty(t, instResp, "instantiate response should be populated")

	// Wait for Gossip to propagate the instantiated blocks to all peers
	time.Sleep(5 * time.Second)

	interest := &fabdiscovery.ChaincodeInterest{
		Chaincodes: []*fabdiscovery.ChaincodeCall{
			{
				Name:            chaincodeID,
				CollectionNames: nil,
			},
		},
	}

	var client *discovery.Client
	client, err = discovery.New(ctx)
	require.NoError(t, err, "error creating discovery client")

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(30*time.Second))
	defer cancel()

	req, err := discclient.NewRequest().OfChannel(testSetup.ChannelID).AddEndorsersQuery(interest)
	require.NoError(t, err, "error adding endorsers query")

	peerCfg1, err := comm.NetworkPeerConfig(ctx.EndpointConfig(), peer1URL)
	require.NoErrorf(t, err, "error getting peer config for [%s]", peer1URL)

	responses, err := client.Send(reqCtx, req, peerCfg1.PeerConfig)
	require.NoError(t, err, "error calling discover service send")
	require.NotEmpty(t, responses, "expecting one response but got none")

	resp := responses[0]
	chanResp := resp.ForChannel(testSetup.ChannelID)

	endorsers, err := chanResp.Endorsers(interest.Chaincodes, discclient.NoPriorities, discclient.NoExclusion)
	require.NoError(t, err, "error getting endorsers")
	require.NotEmpty(t, endorsers, "expecting at least one endorser but got none")

	t.Logf("*** Endorsers for chaincode [%s] channel %s:\n", chaincodeID, testSetup.ChannelID)
	for _, endorser := range endorsers {
		aliveMsg := endorser.AliveMessage.GetAliveMsg()
		if !assert.NotNil(t, aliveMsg, "got nil AliveMessage") {
			continue
		}
		if !assert.NotNil(t, aliveMsg.Membership, "got nil Membership") {
			continue
		}
		t.Logf("--- Endpoint: %s\n", aliveMsg.Membership.Endpoint)
	}
}
