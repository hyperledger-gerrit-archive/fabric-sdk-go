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
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/gossip"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	msp "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
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
	config, err := chanResp.Config()
	if err != nil {
		t.Fatalf("error getting config: %s", err)
	}
	PrintConfig(config)

	peers, err := chanResp.Peers()
	if err != nil {
		t.Fatalf("error getting config: %s", err)
	}

	fmt.Printf("***** Peers for channel [%s]:\n", testSetup.ChannelID)
	PrintPeers(peers)

	fmt.Printf("***** Endorsers for chaincode [%s]:\n", mainChaincodeID)
	endorsers, err := chanResp.Endorsers(mainChaincodeID)
	if err != nil {
		t.Fatalf("error getting endorsers: %s", err)
	}
	PrintPeers(endorsers)
}

func PrintPeers(peers []*discclient.Peer) {
	for _, peer := range peers {
		PrintPeer(peer)
	}
}

func PrintPeer(peer *discclient.Peer) {
	PrintAliveMsg(peer.AliveMessage.GetAliveMsg())
	if peer.StateInfoMessage != nil {
		PrintStateInfo(peer.StateInfoMessage.GetStateInfo())
	}
}

func PrintAliveMsg(aliveMsg *gossip.AliveMessage) {
	fmt.Printf("--- Endpoint: %s\n", aliveMsg.Membership.Endpoint)
}

func PrintStateInfo(stateInfo *gossip.StateInfo) {
	fmt.Printf("--- Ledger Height: %d\n", stateInfo.Properties.LedgerHeight)
	fmt.Printf("--- LeftChannel: %t\n", stateInfo.Properties.LeftChannel)
	fmt.Printf("--- Chaincodes:\n")
	for _, cc := range stateInfo.Properties.Chaincodes {
		fmt.Printf("------ %s:%s\n", cc.Name, cc.Version)
	}
}

func PrintConfig(config *discovery.ConfigResult) {
	fmt.Printf("***** Config:\n")
	fmt.Printf("--- MSPs:\n")
	for _, mspConfig := range config.GetMsps() {
		PrintMSPConfig(mspConfig)
	}
	fmt.Printf("--- Orderers:\n")
	for _, endpoints := range config.GetOrderers() {
		PrintEndpoints(endpoints)
	}
}

func PrintMSPConfig(config *msp.FabricMSPConfig) {
	fmt.Printf("------ MSP ID: %s\n", config.GetName())
	fmt.Printf("--------- Config: %#v\n", config)
}

func PrintEndpoints(endpoints *discovery.Endpoints) {
	for _, endpoint := range endpoints.GetEndpoint() {
		PrintEndpoint(endpoint)
	}
}

func PrintEndpoint(endpoint *discovery.Endpoint) {
	fmt.Printf("------ Endpoint: %s:%d\n", endpoint.GetHost(), endpoint.GetPort())
}
