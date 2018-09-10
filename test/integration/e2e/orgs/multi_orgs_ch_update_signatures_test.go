/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package orgs

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/stretchr/testify/require"
)

const (
	dsChannel              = "dschannel"
	adminUser              = "Admin"
	user1                  = "User1"
	dsChannelOrg1SDKConfig = "config_e2e_org1_dschannel.yaml"
	dsChannelOrg2SDKConfig = "config_e2e_org2_dschannel.yaml"
)

// testDistributedSignatures will create at least 2 clients, each from 2 different orgs and creates a channel where these 2 orgs are members
func testDistributedSignatures(t *testing.T, examplecc string) {
	sdk, err := fabsdk.New(integration.ConfigBackend)
	require.NoError(t, err, "error creating a new sdk client for %s", ordererOrgName)
	defer sdk.Close()
	ordererCtx := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(ordererOrgName))
	// create Channel management client for OrdererOrg
	chMgmtClient, err := resmgmt.New(ordererCtx)
	require.NoError(t, err, "error creating a new resource management client for %s", ordererOrgName)

	//Config containing references to org1 only
	configProvider := config.FromFile(integration.GetConfigPath(dsChannelOrg1SDKConfig))
	//if local test, add entity matchers to override URLs to localhost
	if integration.IsLocal() {
		configProvider = integration.AddLocalEntityMapping(configProvider)
	}

	org1sdk, err := fabsdk.New(configProvider)
	require.NoError(t, err, "error creating a new sdk client for %s", org1)
	defer org1sdk.Close()
	org1Ctx := org1sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(org1))
	org1ChClient, err := resmgmt.New(org1Ctx)
	require.NoError(t, err, "error creating a new resource management client for %s", org1)

	configProvider2 := config.FromFile(integration.GetConfigPath(dsChannelOrg2SDKConfig))
	if integration.IsLocal() {
		configProvider2 = integration.AddLocalEntityMapping(configProvider)
	}

	org2sdk, err := fabsdk.New(configProvider2)
	require.NoError(t, err, "error creating a new sdk client")
	defer org2sdk.Close()
	org2Ctx := org2sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(org2))
	org2ChClient, err := resmgmt.New(org2Ctx)
	require.NoError(t, err, "error creating a new resource management client for %s", org2)

	//// temp code.. remove once signatures are done by an external tool (ie openssl) and replace all signature variables in the code below
	org1MspClient, err := mspclient.New(org1sdk.Context(), mspclient.WithOrg(org1))
	require.NoError(t, err, "error creating a new msp management client for %s", org1)
	org1User, err := org1MspClient.GetSigningIdentity(adminUser)
	require.NoError(t, err, "error creating a new SigningIdentity for %s", org1)
	org2MspClient, err := mspclient.New(org2sdk.Context(), mspclient.WithOrg(org2))
	require.NoError(t, err, "error creating a new msp management client for %s", org2)
	org2User, err := org2MspClient.GetSigningIdentity(adminUser)
	require.NoError(t, err, "error creating a new SigningIdentity for %s", org2)
	// TODO remove temp code

	// retrieve orderererOrg ConfigSignature for Admin
	chConfigPath := integration.GetChannelConfigPath(fmt.Sprintf("%s.tx", dsChannel))

	// retrieve org1 ConfigSignature
	org1DsChannelCfgSig, err := org1ChClient.CreateConfigSignature(org1User, chConfigPath)
	require.NoError(t, err, "error creating a new ConfigSignature for %s", org1)

	// retrieve org2 ConfigSignature
	org2DsChannelCfgSig, err := org2ChClient.CreateConfigSignature(org2User, chConfigPath)
	require.NoError(t, err, "error creating a new ConfigSignature for %s", org2)

	// create channel on the orderer
	req := resmgmt.SaveChannelRequest{ChannelID: dsChannel, ChannelConfigPath: chConfigPath}

	txID, err := chMgmtClient.SaveChannel(req, resmgmt.WithConfigSignatures(org1DsChannelCfgSig, org2DsChannelCfgSig), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "error creating channel %s signatures for %s", dsChannel, ordererOrgName)
	require.NotEmpty(t, txID, "transaction ID should be populated for Channel create for %s", ordererOrgName)

	var lastConfigBlock uint64
	lastConfigBlock = integration.WaitForOrdererConfigUpdate(t, org1ChClient, dsChannel, true, lastConfigBlock)

	// create channel on anchor peer of org1
	chConfigOrg1MSPPath := integration.GetChannelConfigPath(fmt.Sprintf("%s%sMSPanchors.tx", dsChannel, org1))
	org1MSPDsChannelCfgSig, err := org1ChClient.CreateConfigSignature(org1User, chConfigOrg1MSPPath)
	require.NoError(t, err, "error creating a new ConfigSignature for %s", org1)

	req.ChannelConfigPath = chConfigOrg1MSPPath
	txID, err = org1ChClient.SaveChannel(req, resmgmt.WithConfigSignatures(org1MSPDsChannelCfgSig), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "error creating channel %s signatures for %s", dsChannel, org1)
	require.NotEmpty(t, txID, "transaction ID should be populated for Channel create for %s", org1)
	lastConfigBlock = integration.WaitForOrdererConfigUpdate(t, org1ChClient, dsChannel, false, lastConfigBlock)

	// create channel on anchor peer of org2
	chConfigOrg2MSPPath := integration.GetChannelConfigPath(fmt.Sprintf("%s%sMSPanchors.tx", dsChannel, org2))
	org2MSPDsChannelCfgSig, err := org2ChClient.CreateConfigSignature(org2User, chConfigOrg2MSPPath)
	require.NoError(t, err, "error creating a new ConfigSignature for %s", org2)

	req.ChannelConfigPath = chConfigOrg2MSPPath
	txID, err = org2ChClient.SaveChannel(req, resmgmt.WithConfigSignatures(org2MSPDsChannelCfgSig), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "error creating channel %s signatures for %s", dsChannel, org2)
	require.NotEmpty(t, txID, "transaction ID should be populated for Channel create for %s", org2)
	integration.WaitForOrdererConfigUpdate(t, org1ChClient, dsChannel, false, lastConfigBlock)

	// Org1 peers join channel
	err = org1ChClient.JoinChannel(dsChannel, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "Org1 peers failed to JoinChannel")

	// Org2 peers join channel
	err = org2ChClient.JoinChannel(dsChannel, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "Org2 peers failed to JoinChannel")

	// Ensure that Gossip has propagated it's view of local peers before invoking
	// install since some peers may be missed if we call InstallCC too early
	/*	org1Peers, err := integration.DiscoverLocalPeers(org1Ctx, 2)
		require.NoError(t, err)
		org2Peers, err := integration.DiscoverLocalPeers(org2Ctx, 2)
		require.NoError(t, err)
	*/
	//TODO remove below two lines and uncomment above peer discovery code once fixed (currently it always returns 1 peer per org).
	org1Peers := []fab.Peer{orgTestPeer0}
	org2Peers := []fab.Peer{orgTestPeer1}

	ccVersion := "1" // ccVersion= 1 because previous test increase the ccVersion # on the peers.

	// instantiate example_CC on dschannel
	instantiateCC(t, org1ChClient, exampleCC, ccVersion, dsChannel)

	// Ensure the CC is instantiated on all peers in both orgs
	found := queryInstantiatedCC(t, org1, org1ChClient, dsChannel, exampleCC, ccVersion, org1Peers)
	require.True(t, found, "Failed to find instantiated chaincode [%s:%s] in at least one peer in Org1 on channel [%s]", exampleCC, ccVersion, dsChannel)

	found = queryInstantiatedCC(t, org2, org2ChClient, dsChannel, exampleCC, ccVersion, org2Peers)
	require.True(t, found, "Failed to find instantiated chaincode [%s:%s] in at least one peer in Org2 on channel [%s]", exampleCC, ccVersion, dsChannel)

	// test regular querying on dschannel from org1 and org2
	testQueryingOrgs(t, org1sdk, org2sdk, examplecc)
}

func testQueryingOrgs(t *testing.T, org1sdk *fabsdk.FabricSDK, org2sdk *fabsdk.FabricSDK, examplecc string) {
	//prepare context
	org1ChannelClientContext := org1sdk.ChannelContext(dsChannel, fabsdk.WithUser(user1), fabsdk.WithOrg(org1))
	org2ChannelClientContext := org2sdk.ChannelContext(dsChannel, fabsdk.WithUser(user1), fabsdk.WithOrg(org2))

	// Org1 user connects to 'dschannel'
	chClientOrg1User, err := channel.New(org1ChannelClientContext)
	require.NoError(t, err, "Failed to create new channel client for Org1 user: %s", err)

	// Org2 user connects to 'dschannel'
	chClientOrg2User, err := channel.New(org2ChannelClientContext)
	require.NoError(t, err, "Failed to create new channel client for Org1 user: %s", err)

	req := channel.Request{
		ChaincodeID: examplecc,
		Fcn:         "invoke",
		Args:        integration.ExampleCCDefaultQueryArgs(),
	}

	// query org1
	resp, err := chClientOrg1User.Query(req, channel.WithRetry(retry.DefaultChannelOpts))
	require.NoError(t, err, "query funds failed")

	foundOrg2Endorser := false
	for _, v := range resp.Responses {
		//check if response endorser is org2 peer and MSP ID 'Org2MSP' is found
		if strings.Contains(string(v.Endorsement.Endorser), "Org2MSP") {
			foundOrg2Endorser = true
			break
		}
	}

	require.True(t, foundOrg2Endorser, "Org2 MSP ID was not in the endorsement")

	//query org2
	resp, err = chClientOrg2User.Query(req, channel.WithRetry(retry.DefaultChannelOpts))
	require.NoError(t, err, "query funds failed")

	foundOrg1Endorser := false
	for _, v := range resp.Responses {
		//check if response endorser is org1 peer and MSP ID 'Org1MSP' is found
		if strings.Contains(string(v.Endorsement.Endorser), "Org1MSP") {
			foundOrg1Endorser = true
			break
		}
	}

	require.True(t, foundOrg1Endorser, "Org1 MSP ID was not in the endorsement")
}
