/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package discovery

import (
	"fmt"
	"os"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

const (
	org1Name     = "Org1"
	org2Name     = "Org2"
	adminUser    = "Admin"
	org1User     = "User1"
	orgChannelID = "orgchannel"
	ccPath       = "github.com/example_cc"
)

var mainSDK *fabsdk.FabricSDK
var mainTestSetup *integration.BaseSetupImpl

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

	mainSDK = sdk
	mainTestSetup = &testSetup
}

func teardown() {
	integration.CleanupUserData(nil, mainSDK)
	mainSDK.Close()
}
