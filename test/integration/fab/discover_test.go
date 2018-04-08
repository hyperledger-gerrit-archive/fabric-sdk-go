// +build devlatest

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const (
	peer1URL = "peer0.org1.example.com"
	peer2URL = "peer0.org2.example.com"
)

func TestDiscoverClient(t *testing.T) {
	sdk := mainSDK
	testSetup := mainTestSetup

	ctxProvider := sdk.Context(fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))
	ctx, err := ctxProvider()
	if err != nil {
		t.Fatalf("error getting channel context: %s", err)
	}

	var client *discovery.Client
	client, err = discovery.New(ctx)
	if err != nil {
		t.Fatalf("error creating discover service: %s", err)
	}

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	req := discclient.NewRequest().OfChannel(testSetup.ChannelID).AddPeersQuery().AddConfigQuery().AddEndorsersQuery(mainChaincodeID)

	peerCfg1, err := comm.NetworkPeerConfigFromURL(ctx.EndpointConfig(), peer1URL)
	if err != nil {
		t.Fatalf("error getting peer config for [%s]", peer1URL, err)
	}

	responses, err := client.Send(reqCtx, req, peerCfg1.PeerConfig)
	if err != nil {
		t.Fatalf("error calling discover service send: %s", err)
	}

	if len(responses) == 0 {
		t.Fatalf("No responses")
	}

	resp := responses[0]
	chanResp := resp.ForChannel(testSetup.ChannelID)

	peers, err := chanResp.Peers()
	if err != nil {
		t.Fatalf("error getting config: %s", err)
	}

	for _, peer := range peers {
		aliveMsg := peer.AliveMessage.GetAliveMsg()
		fmt.Printf("--- Endpoint: %s\n", aliveMsg.Membership.Endpoint)

		if peer.StateInfoMessage != nil {
			stateInfo := peer.StateInfoMessage.GetStateInfo()
			fmt.Printf("--- Ledger Height: %d\n", stateInfo.Properties.LedgerHeight)
			fmt.Printf("--- LeftChannel: %t\n", stateInfo.Properties.LeftChannel)
			fmt.Printf("--- Chaincodes:\n")
			for _, cc := range stateInfo.Properties.Chaincodes {
				fmt.Printf("------ %s:%s\n", cc.Name, cc.Version)
			}
		}
	}
}
