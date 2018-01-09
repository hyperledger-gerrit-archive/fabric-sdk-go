/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package secondpeer

import (
	"strconv"
	"testing"

	"path"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

const (
	channelID = "mychannel"
	orgName   = "Org1"
	orgAdmin  = "Admin"
	ccID      = "e2eExampleCC"
)

// Peers
var org1Peer0 fab.Peer
var org1Peer1 fab.Peer

func TestSecondPeerEndToEnd(t *testing.T) {
	// Create SDK setup for the integration tests
	sdkOptions := fabapi.Options{
		ConfigFile: "../" + "../fixtures/config/config_secondpeer_test.yaml",
	}

	sdk, err := fabapi.NewSDK(sdkOptions)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	// Channel management client is responsible for managing channels (create/update channel)
	// Supply user that has privileges to create channel (in this case orderer admin)
	chMgmtClient, err := sdk.NewChannelMgmtClientWithOpts("Admin", &fabapi.ChannelMgmtClientOpts{OrgName: "ordererorg"})
	if err != nil {
		t.Fatalf("Failed to create channel management client: %s", err)
	}

	// Org admin user is signing user for creating channel
	orgAdminUser, err := sdk.NewPreEnrolledUser(orgName, orgAdmin)
	if err != nil {
		t.Fatalf("NewPreEnrolledUser failed for %s, %s: %s", orgName, orgAdmin, err)
	}

	// Create channel
	req := chmgmt.SaveChannelRequest{ChannelID: channelID, ChannelConfig: path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"), SigningUser: orgAdminUser}
	if err = chMgmtClient.SaveChannel(req); err != nil {
		t.Fatal(err)
	}

	// Allow orderer to process channel creation
	time.Sleep(time.Second * 3)

	// Org resource management client (Org1 is default org)
	orgResMgmt, err := sdk.NewResourceMgmtClient(orgAdmin)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Org peers join channel
	if err = orgResMgmt.JoinChannel(channelID); err != nil {
		t.Fatalf("Org peers failed to JoinChannel: %s", err)
	}

	// Create chaincode package for example cc
	ccPkg, err := packager.NewCCPackage("github.com/example_cc", "../../fixtures/testdata")
	if err != nil {
		t.Fatal(err)
	}

	// Install example cc to org peers
	installCCReq := resmgmt.InstallCCRequest{Name: ccID, Path: "github.com/example_cc", Version: "0", Package: ccPkg}
	_, err = orgResMgmt.InstallCC(installCCReq)
	if err != nil {
		t.Fatal(err)
	}

	// Set up chaincode policy
	ccPolicy := cauthdsl.SignedByAnyMember([]string{"Org1MSP", "Org2MSP"})

	// Org resource manager will instantiate 'example_cc' on channel
	err = orgResMgmt.InstantiateCC(channelID, resmgmt.InstantiateCCRequest{Name: ccID, Path: "github.com/example_cc", Version: "0", Args: integration.ExampleCCInitArgs(), Policy: ccPolicy})
	if err != nil {
		t.Fatal(err)
	}

	// Channel client is used to query and execute transactions
	chClient, err := sdk.NewChannelClient(channelID, "User1")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Load org peers so we can query individually
	loadOrgPeers(t, sdk)

	peer1QueryOpts := apitxn.QueryOpts{ProposalProcessors: []apitxn.ProposalProcessor{org1Peer0}}
	value, err := chClient.QueryWithOpts(apitxn.QueryRequest{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()}, peer1QueryOpts)
	if err != nil {
		t.Fatalf("Failed to query peer 1 before transaction: %s", err)
	}

	eventID := "test([a-zA-Z]+)"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	notifier := make(chan *apitxn.CCEvent)
	rce := chClient.RegisterChaincodeEvent(notifier, ccID, eventID)

	// Move funds
	peer1ExecuteOpts := apitxn.ExecuteTxOpts{ProposalProcessors: []apitxn.ProposalProcessor{org1Peer0}}
	_, err = chClient.ExecuteTxWithOpts(apitxn.ExecuteTxRequest{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()}, peer1ExecuteOpts)
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case ccEvent := <-notifier:
		t.Logf("Received CC event: %s\n", ccEvent)
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC event for eventId(%s)\n", eventID)
	}

	// Unregister chain code event using registration handle
	err = chClient.UnregisterChaincodeEvent(rce)
	if err != nil {
		t.Fatalf("Unregister cc event failed: %s", err)
	}

	// Query on first peer
	//peer1QueryOpts := apitxn.QueryOpts{ProposalProcessors: []apitxn.ProposalProcessor{org1Peer0}}
	peer1Value, err := chClient.QueryWithOpts(apitxn.QueryRequest{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()}, peer1QueryOpts)
	if err != nil {
		t.Fatalf("Failed to query peer 1 after transaction: %s", err)
	}

	// Query on second peer
	peer2QueryOpts := apitxn.QueryOpts{ProposalProcessors: []apitxn.ProposalProcessor{org1Peer1}}
	peer2Value, err := chClient.QueryWithOpts(apitxn.QueryRequest{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()}, peer2QueryOpts)
	if err != nil {
		t.Fatalf("Failed to query peer 2 after transaction: %s", err)
	}

	peer1ValueInt, _ := strconv.Atoi(string(peer1Value))
	peer2ValueInt, _ := strconv.Atoi(string(peer2Value))
	valueInt, _ := strconv.Atoi(string(value))

	if valueInt+1 != peer1ValueInt {
		t.Fatalf("ExecuteTx failed. Before: %s, after on peer 1: %s", value, peer1Value)
	}

	if peer1ValueInt != peer2ValueInt {
		t.Fatalf("ExecuteTx failed. Before: %s, after on peer 2: %s", value, peer2Value)
	}

	// Release all channel client resources
	chClient.Close()
}

func loadOrgPeers(t *testing.T, sdk *fabapi.FabricSDK) {

	org1Peers, err := sdk.ConfigProvider().PeersConfig(orgName)
	if err != nil {
		t.Fatal(err)
	}

	if len(org1Peers) != 2 {
		t.Fatalf("Incorrect number of org peers: %v", len(org1Peers))
	}

	org1Peer0, err = peer.NewPeerFromConfig(&apiconfig.NetworkPeer{PeerConfig: org1Peers[0]}, sdk.ConfigProvider())
	if err != nil {
		t.Fatal(err)
	}

	org1Peer1, err = peer.NewPeerFromConfig(&apiconfig.NetworkPeer{PeerConfig: org1Peers[1]}, sdk.ConfigProvider())
	if err != nil {
		t.Fatal(err)
	}
}
