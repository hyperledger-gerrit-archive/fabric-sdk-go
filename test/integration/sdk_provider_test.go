/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"strconv"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context/defprovider"

	ccpolicy "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/ccpolicy"
	selection "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/selection/dynamicselection"
)

func TestDynamicSelection(t *testing.T) {

	testSetup := BaseSetupImpl{
		ConfigFile:      ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(t); err != nil {
		t.Fatalf(err.Error())
	}

	if err := testSetup.InstallAndInstantiateExampleCC(); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	// Dynamic selection requires configuring proper cc policy provider
	// Having dynamic selection without supporting cc policy provider
	// will cause an error when creating channel client
	noCCPolicyProviderFactory := &NoServicePolicyProviderFactory{}

	// Create SDK setup for channel client with dynamic selection and without proper cc policy provider
	sdkOptions := fabapi.Options{
		ConfigFile:      testSetup.ConfigFile,
		ProviderFactory: noCCPolicyProviderFactory,
	}

	sdk, err := fabapi.NewSDK(sdkOptions)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	chClient, err := sdk.NewChannelClient(testSetup.ChannelID, "User1")
	if err == nil {
		t.Fatalf("Should have failed since dynamic selection requires proper cc policy provider")
	}

	// Properly configured factory for using dynamic endorser selection (based on real cc policy data)
	sdkOptions.ProviderFactory = &DynamicSelectionProviderFactory{}
	sdk, err = fabapi.NewSDK(sdkOptions)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	chClient, err = sdk.NewChannelClient(testSetup.ChannelID, "User1")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	value, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: testSetup.ChainCodeID, Fcn: "invoke", Args: queryArgs})
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}

	t.Logf("*** QueryValue before invoke %s", value)

	// Move funds
	_, err = chClient.ExecuteTx(apitxn.ExecuteTxRequest{ChaincodeID: testSetup.ChainCodeID, Fcn: "invoke", Args: txArgs})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	// Verify move funds transaction result
	valueAfterInvoke, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: testSetup.ChainCodeID, Fcn: "invoke", Args: queryArgs})
	if err != nil {
		t.Fatalf("Failed to query funds after transaction: %s", err)
	}

	t.Logf("*** QueryValue after invoke %s", valueAfterInvoke)

	valueInt, _ := strconv.Atoi(string(value))
	valueAfterInvokeInt, _ := strconv.Atoi(string(valueAfterInvoke))
	if valueInt+1 != valueAfterInvokeInt {
		t.Fatalf("ExecuteTx failed. Before: %s, after: %s", value, valueAfterInvoke)
	}

	// Release all channel client resources
	chClient.Close()
}

// DynamicSelectionProviderFactory is configured with dynamic (endorser) selection provider
// and supporting chaincode policy provider with fully imlemented dynamic chaincode policy service
type DynamicSelectionProviderFactory struct {
	defprovider.DefaultProviderFactory
}

// NewSelectionProvider returns a new implementation of dynamic selection provider
func (f *DynamicSelectionProviderFactory) NewSelectionProvider(config apiconfig.Config) (fab.SelectionProvider, error) {
	return selection.NewSelectionProvider(config)
}

// NewCCPolicyProvider returns implementation of chaincode policy service
func (f *DynamicSelectionProviderFactory) NewCCPolicyProvider(config apiconfig.Config) (fab.CCPolicyProvider, error) {
	return ccpolicy.NewCCPolicyProvider(config)
}

// NoServicePolicyProviderFactory represents test provider factory
type NoServicePolicyProviderFactory struct {
	defprovider.DefaultProviderFactory
}

// NewSelectionProvider returns dynamic selection provider
func (f *NoServicePolicyProviderFactory) NewSelectionProvider(config apiconfig.Config) (fab.SelectionProvider, error) {
	return selection.NewSelectionProvider(config)
}
