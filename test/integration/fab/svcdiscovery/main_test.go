/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package svcdiscovery

import (
	"fmt"
	"os"
	"testing"

	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/stretchr/testify/require"
)

const (
	org1Name     = "Org1"
	org2Name     = "Org2"
	adminUser    = "Admin"
	org1User     = "User1"
	orgChannelID = "orgchannel"
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

func setupMultiOrgContext(t *testing.T, sdk *fabsdk.FabricSDK) []*integration.OrgContext {
	org1AdminContext := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(org1Name))
	org1ResMgmt, err := resmgmt.New(org1AdminContext)
	require.NoError(t, err)

	org1MspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org1Name))
	require.NoError(t, err)
	org1Admin, err := org1MspClient.GetSigningIdentity(adminUser)
	require.NoError(t, err)

	org2AdminContext := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(org2Name))
	org2ResMgmt, err := resmgmt.New(org2AdminContext)
	require.NoError(t, err)

	org2MspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org2Name))
	require.NoError(t, err)
	org2Admin, err := org2MspClient.GetSigningIdentity(adminUser)
	require.NoError(t, err)

	// Ensure that Gossip has propagated its view of local peers before invoking
	// install since some peers may be missed if we call InstallCC too early
	org1Peers, err := integration.DiscoverLocalPeers(org1AdminContext, 2)
	require.NoError(t, err)
	org2Peers, err := integration.DiscoverLocalPeers(org2AdminContext, 1)
	require.NoError(t, err)

	return []*integration.OrgContext{
		{
			OrgID:                org1Name,
			CtxProvider:          org1AdminContext,
			ResMgmt:              org1ResMgmt,
			Peers:                org1Peers,
			SigningIdentity:      org1Admin,
			AnchorPeerConfigFile: "orgchannelOrg1MSPanchors.tx",
		},
		{
			OrgID:                org2Name,
			CtxProvider:          org2AdminContext,
			ResMgmt:              org2ResMgmt,
			Peers:                org2Peers,
			SigningIdentity:      org2Admin,
			AnchorPeerConfigFile: "orgchannelOrg2MSPanchors.tx",
		},
	}
}

type closable interface {
	Close()
}

func contains(list []string, value string) bool {
	for _, e := range list {
		if e == value {
			return true
		}
	}
	return false
}
