/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package resmgmt

import (
	"fmt"
	"os"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

const (
	org1Name      = "Org1"
	org2Name      = "Org2"
	adminUser     = "Admin"
	org1AdminUser = "Admin"
)

var mainSDK *fabsdk.FabricSDK
var mainTestSetup *integration.BaseSetupImpl
var mainChaincodeID string

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

	// Delete all private keys from the crypto suite store
	// and users from the user store
	integration.CleanupUserData(nil, sdk)

	if err := testSetup.Initialize(sdk); err != nil {
		panic(err.Error())
	}

	chaincodeID := integration.GenerateExampleID()
	if err := integration.PrepareExampleCC(sdk, fabsdk.WithUser("Admin"), testSetup.OrgID, chaincodeID); err != nil {
		panic(fmt.Sprintf("PrepareExampleCC return error: %s", err))
	}

	mainSDK = sdk
	mainTestSetup = &testSetup
	mainChaincodeID = chaincodeID
}

func teardown() {
	integration.CleanupUserData(nil, mainSDK)
	mainSDK.Close()
}

func contains(list []string, value string) bool {
	for _, e := range list {
		if e == value {
			return true
		}
	}
	return false
}
