// +build devstable

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"strconv"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/test/integration"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/dynamicdiscovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/stretchr/testify/assert"
)

func TestDynamicDiscovery(t *testing.T) {

	// Using shared SDK instance to increase test speed.
	sdk := mainSDK
	testSetup := mainTestSetup

	// Create SDK setup for channel client with dynamic selection
	sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile),
		fabsdk.WithServicePkg(&DynamicSelectionProviderFactory{}))

	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	if err := testSetup.Initialize(sdk); err != nil {
		t.Fatalf(err.Error())
	}

	chainCodeID := integration.GenerateRandomID()
	resp, err := integration.InstallAndInstantiateExampleCC(sdk, fabsdk.WithUser("Admin"), testSetup.OrgID, chainCodeID)
	assert.Nil(t, err, "InstallAndInstantiateExampleCC return error")
	assert.NotEmpty(t, resp, "instantiate response should be populated")

	//prepare contexts
	org1ChannelClientContext := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))

	chClient, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	response, err := chClient.Query(channel.Request{ChaincodeID: chainCodeID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()},
		channel.WithRetry(retry.DefaultChClientOpts))
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}
	value := response.Payload

	// Move funds
	response, err = chClient.Execute(channel.Request{ChaincodeID: chainCodeID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()},
		channel.WithRetry(retry.DefaultChClientOpts))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	// Verify move funds transaction result
	response, err = chClient.Query(channel.Request{ChaincodeID: chainCodeID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
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
}

// CreateDiscoveryProvider returns a new dynamic discovery provider
func (f *DynamicSelectionProviderFactory) CreateDiscoveryProvider(config fab.EndpointConfig) (fab.DiscoveryProvider, error) {
	return dynamicdiscovery.New(), nil
}
