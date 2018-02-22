/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"path"
	"strconv"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/apiconfig"

	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/chclient/api"
	selection "github.com/hyperledger/fabric-sdk-go/pkg/client/selection/dynamicselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func TestDynamicSelection(t *testing.T) {

	testSetup := integration.BaseSetupImpl{
		ConfigFile:      "../" + integration.ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   path.Join("../../", metadata.ChannelConfigPath, "mychannel.tx"),
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	chainCodeID := integration.GenerateRandomID()
	if err := integration.InstallAndInstantiateExampleCC(testSetup.SDK, fabsdk.WithUser("Admin"), testSetup.OrgID, chainCodeID); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	// Specify user that will be used by dynamic selection service (to retrieve chanincode policy information)
	// This user has to have privileges to query lscc for chaincode data
	mychannelUser := selection.ChannelUser{ChannelID: testSetup.ChannelID, UserName: "User1", OrgName: "Org1"}

	// Create SDK setup for channel client with dynamic selection
	sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile),
		fabsdk.WithServicePkg(&DynamicSelectionProviderFactory{ChannelUsers: []selection.ChannelUser{mychannelUser}}))

	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	chClient, err := sdk.NewClient(fabsdk.WithUser("User1")).Channel(testSetup.ChannelID)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Release all channel client resources
	defer chClient.Close()

	response, err := chClient.Query(api.Request{ChaincodeID: chainCodeID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}
	value := response.Payload

	// Move funds
	response, err = chClient.Execute(api.Request{ChaincodeID: chainCodeID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	// Verify move funds transaction result
	response, err = chClient.Query(api.Request{ChaincodeID: chainCodeID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds after transaction: %s", err)
	}

	valueInt, _ := strconv.Atoi(string(value))
	valueAfterInvokeInt, _ := strconv.Atoi(string(response.Payload))
	if valueInt+1 != valueAfterInvokeInt {
		t.Fatalf("Execute failed. Before: %s, after: %s", value, response.Payload)
	}

}

// DynamicSelectionProviderFactory is configured with dynamic (endorser) selection provider
type DynamicSelectionProviderFactory struct {
	defsvc.ProviderFactory
	ChannelUsers []selection.ChannelUser
}

// NewSelectionProvider returns a new implementation of dynamic selection provider
func (f *DynamicSelectionProviderFactory) NewSelectionProvider(config apiconfig.Config) (context.SelectionProvider, error) {
	return selection.NewSelectionProvider(config, f.ChannelUsers, nil)
}
