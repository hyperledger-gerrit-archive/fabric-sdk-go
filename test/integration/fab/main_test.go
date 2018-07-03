/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package fab

import (
	"fmt"
	"os"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

var mainSDK *fabsdk.FabricSDK
var mainTestSetup *integration.BaseSetupImpl
var mainChaincodeID string

const (
	org1Name = "Org1"
)

func TestMain(m *testing.M) {
	setup()
	r := m.Run()
	teardown()
	os.Exit(r)
}

func setup() {
	testSetup := integration.BaseSetupImpl{
		ChannelID:         "mychannel",
		OrgID:             org1Name,
		ChannelConfigFile: integration.GetChannelConfigPath("mychannel.tx"),
	}

	sdk, err := fabsdk.New(integration.ConfigBackend)
	if err != nil {
		panic(fmt.Sprintf("Failed to create new SDK: %s", err))
	}

	if err := testSetup.Initialize(sdk); err != nil {
		panic(err.Error())
	}

	chaincodeID := integration.GenerateExampleID(false)
	if err := integration.PrepareExampleCC(sdk, fabsdk.WithUser("Admin"), testSetup.OrgID, chaincodeID); err != nil {
		panic(fmt.Sprintf("PrepareExampleCC return error: %s", err))
	}

	mainSDK = sdk
	mainTestSetup = &testSetup
	mainChaincodeID = chaincodeID
}

func teardown() {
	mainSDK.Close()
}
