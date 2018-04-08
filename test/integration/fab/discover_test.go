// +build devstable

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"fmt"
	"testing"
	"time"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func TestDiscoverService(t *testing.T) {
	sdk := mainSDK
	testSetup := mainTestSetup

	chContextProvider := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))
	chContext, err := chContextProvider()
	if err != nil {
		t.Fatalf("error getting channel context: %s", err)
	}
	service, err := chContext.InfraProvider().CreateDiscoverService(chContext, testSetup.ChannelID)
	if err != nil {
		t.Fatalf("error creating discover service: %s", err)
	}

	reqCtx, cancel := context.NewRequest(chContext, context.WithTimeout(10*time.Second))
	defer cancel()

	req := discclient.NewRequest().OfChannel(testSetup.ChannelID).AddPeersQuery().AddConfigQuery().AddEndorsersQuery(mainChaincodeID)

	resp, err := service.Send(reqCtx, req)
	if err != nil {
		t.Fatalf("error calling discover service send: %s", err)
	}

	fmt.Printf("Response: %#v\n", resp)

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
