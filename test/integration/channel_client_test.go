/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
)

func TestChannelClient(t *testing.T) {

	testSetup := BaseSetupImpl{
		ConfigFile:      "../fixtures/config/config_test.yaml",
		ChannelID:       "mychannel",
		OrgID:           "Org1",
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	if err := testSetup.InstallAndInstantiateExampleCC(); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	// Create SDK setup for the integration tests
	sdkOptions := fabapi.Options{
		ConfigFile: testSetup.ConfigFile,
		StateStoreOpts: opt.StateStoreOpts{
			Path: "/tmp/enroll_user",
		},
	}

	sdk, err := fabapi.NewSDK(sdkOptions)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	opt := &fabapi.ChannelClientOpts{OrgName: testSetup.OrgID}
	chClient, err := sdk.NewChannelClientWithOpts(testSetup.ChannelID, "User1", opt)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Test synchronous query
	result, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: testSetup.ChainCodeID, Fcn: "invoke", Args: []string{"query", "b"}})
	if err != nil {
		t.Fatalf("Failed to invoke example cc: %s", err)
	}

	if result != "200" {
		t.Fatalf("Expecting 200, got %s", result)
	}
	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	// Move funds
	result, err = chClient.ExecuteTx(apitxn.ExecuteTxRequest{ChaincodeID: testSetup.ChainCodeID, Fcn: "invoke", Args: []string{"move", "a", "b", "1"}, TransientMap: transientDataMap})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	// Test asynchronous query
	notifier := make(chan apitxn.QueryResponse)
	result, err = chClient.QueryWithOpts(apitxn.QueryRequest{ChaincodeID: testSetup.ChainCodeID, Fcn: "invoke", Args: []string{"query", "b"}}, apitxn.QueryOpts{Notifier: notifier})
	if err != nil {
		t.Fatalf("Failed to invoke example cc asynchronously: %s", err)
	}
	if result != "" {
		t.Fatalf("Expecting empty, got %s", result)
	}

	select {
	case response := <-notifier:
		if response.Error != nil {
			t.Fatalf("Query returned error: %s", response.Error)
		}
		if response.Response != "201" {
			t.Fatalf("Expecting 201, got %s", response.Response)
		}
	case <-time.After(time.Second * 20):
		t.Fatalf("Query Request timed out")
	}
}
