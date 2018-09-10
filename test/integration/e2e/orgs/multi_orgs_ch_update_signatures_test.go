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
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/stretchr/testify/require"
)

const (
	dsChannel              = "dschannel"
	adminUser              = "Admin"
	user1                  = "User1"
	dsChannelOrg1SDKConfig = "overrides/org1_dschannel.yaml"
	dsChannelOrg2SDKConfig = "overrides/org2_dschannel.yaml"
)

var orgCfg = map[string]string{
	org1: dsChannelOrg1SDKConfig,
	org2: dsChannelOrg2SDKConfig,
}

type dsClientCtx struct {
	org   string
	sdk   *fabsdk.FabricSDK
	clCtx contextApi.ClientProvider
	rsCl  *resmgmt.Client
}

// testDistributedSignatures will create at least 2 clients, each from 2 different orgs and creates a channel where these 2 orgs are members
func testDistributedSignatures(t *testing.T, examplecc string) {
	ordererClCtx := createDSClientCtx(t, ordererOrgName)
	defer ordererClCtx.sdk.Close()

	org1ClCtx := createDSClientCtx(t, org1)
	defer org1ClCtx.sdk.Close()

	org2ClCtx := createDSClientCtx(t, org2)
	defer org2ClCtx.sdk.Close()

	chConfigPath := integration.GetChannelConfigPath(fmt.Sprintf("%s.tx", dsChannel))

	// create org1 ConfigSignature
	org1DsChannelCfgSig := createSignature(t, org1ClCtx, chConfigPath, adminUser)
	// create org2 ConfigSignature
	org2DsChannelCfgSig := createSignature(t, org2ClCtx, chConfigPath, adminUser)

	// create channel on the orderer
	req := resmgmt.SaveChannelRequest{ChannelID: dsChannel, ChannelConfigPath: chConfigPath}

	txID, err := ordererClCtx.rsCl.SaveChannel(req, resmgmt.WithConfigSignatures(org1DsChannelCfgSig, org2DsChannelCfgSig), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "error creating channel %s signatures for %s", dsChannel, ordererOrgName)
	require.NotEmpty(t, txID, "transaction ID should be populated for Channel create for %s", ordererOrgName)

	var lastConfigBlock uint64
	lastConfigBlock = integration.WaitForOrdererConfigUpdate(t, org1ClCtx.rsCl, dsChannel, true, lastConfigBlock)

	// create channel on anchor peer of org1
	chConfigOrg1MSPPath := integration.GetChannelConfigPath(fmt.Sprintf("%s%sMSPanchors.tx", dsChannel, org1))
	org1MSPDsChannelCfgSig := createSignature(t, org1ClCtx, chConfigOrg1MSPPath, user1)

	req.ChannelConfigPath = chConfigOrg1MSPPath

	txID, err = org1ClCtx.rsCl.SaveChannel(req, resmgmt.WithConfigSignatures(org1MSPDsChannelCfgSig), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "error creating channel %s signatures for %s", dsChannel, org1)
	require.NotEmpty(t, txID, "transaction ID should be populated for Channel create for %s", org1)

	lastConfigBlock = integration.WaitForOrdererConfigUpdate(t, org1ClCtx.rsCl, dsChannel, false, lastConfigBlock)

	// create channel on anchor peer of org2
	chConfigOrg2MSPPath := integration.GetChannelConfigPath(fmt.Sprintf("%s%sMSPanchors.tx", dsChannel, org2))
	org2MSPDsChannelCfgSig := createSignature(t, org2ClCtx, chConfigOrg2MSPPath, user1)
	require.NoError(t, err, "error creating a new ConfigSignature for %s", org2)

	req.ChannelConfigPath = chConfigOrg2MSPPath

	txID, err = org2ClCtx.rsCl.SaveChannel(req, resmgmt.WithConfigSignatures(org2MSPDsChannelCfgSig), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "error creating channel %s signatures for %s", dsChannel, org2)
	require.NotEmpty(t, txID, "transaction ID should be populated for Channel create for %s", org2)

	integration.WaitForOrdererConfigUpdate(t, org1ClCtx.rsCl, dsChannel, false, lastConfigBlock)

	// Saving Channel is successful with distributed signatures, now let's join the peers to the channel and run some queries

	// Org1 peers join channel
	err = org1ClCtx.rsCl.JoinChannel(dsChannel, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "Org1 peers failed to JoinChannel")

	// Org2 peers join channel
	err = org2ClCtx.rsCl.JoinChannel(dsChannel, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "Org2 peers failed to JoinChannel")

	// Ensure that Gossip has propagated it's view of local peers before invoking
	// install since some peers may be missed if we call InstallCC too early
	org1Peers, err := integration.DiscoverLocalPeers(org1ClCtx.clCtx, 2)
	require.NoError(t, err)
	org2Peers, err := integration.DiscoverLocalPeers(org2ClCtx.clCtx, 2)
	require.NoError(t, err)

	ccVersion := "1" // ccVersion= 1 because previous test increased the ccVersion # on the peers.

	// instantiate example_CC on dschannel
	instantiateCC(t, org1ClCtx.rsCl, exampleCC, ccVersion, dsChannel)

	// Ensure the CC is instantiated on all peers in both orgs
	found := queryInstantiatedCC(t, org1, org1ClCtx.rsCl, dsChannel, exampleCC, ccVersion, org1Peers)
	require.True(t, found, "Failed to find instantiated chaincode [%s:%s] in at least one peer in Org1 on channel [%s]", exampleCC, ccVersion, dsChannel)

	found = queryInstantiatedCC(t, org2, org2ClCtx.rsCl, dsChannel, exampleCC, ccVersion, org2Peers)
	require.True(t, found, "Failed to find instantiated chaincode [%s:%s] in at least one peer in Org2 on channel [%s]", exampleCC, ccVersion, dsChannel)

	// test regular querying on dschannel from org1 and org2
	testQueryingOrgs(t, org1ClCtx.sdk, org2ClCtx.sdk, examplecc)
}

func createDSClientCtx(t *testing.T, org string) *dsClientCtx {
	if org == ordererOrgName {
		return createOrderDsClientCtx(t)
	}

	d := &dsClientCtx{org: org}

	var err error
	d.sdk, err = fabsdk.New(getCustomConfigBackend(t, org))
	require.NoError(t, err, "error creating a new sdk client for %s", org)

	d.clCtx = d.sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(org))
	d.rsCl, err = resmgmt.New(d.clCtx)
	require.NoError(t, err, "error creating a new resource management client for %s", org)
	return d
}

func createOrderDsClientCtx(t *testing.T) *dsClientCtx {
	sdk, err := fabsdk.New(integration.ConfigBackend)
	require.NoError(t, err, "error creating a new sdk client for %s", ordererOrgName)

	ordererCtx := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(ordererOrgName))

	// create Channel management client for OrdererOrg
	chMgmtClient, err := resmgmt.New(ordererCtx)
	require.NoError(t, err, "error creating a new resource management client for %s", ordererOrgName)

	return &dsClientCtx{
		org:   ordererOrgName,
		sdk:   sdk,
		clCtx: ordererCtx,
		rsCl:  chMgmtClient,
	}
}

func getCustomConfigBackend(t *testing.T, org string) core.ConfigProvider {
	return func() ([]core.ConfigBackend, error) {
		configFileOverride, ok := orgCfg[org]
		require.True(t, ok, "org config mapping should exist for %s", org)

		configBackends, err := config.FromFile(integration.GetConfigPath(configFilename))()
		require.NoError(t, err, "failed to read config backend from file for org %s", org)

		configBackendsOverrides, err := config.FromFile(integration.GetConfigPath(configFileOverride))()
		require.NoError(t, err, "failed to read config backend from file for org %s", org)

		return append(configBackendsOverrides, configBackends...), nil
	}
}

// TODO replace code below once signatures are created by an external tool (ie openssl)
// TODO with the help of resmgmt client's CreateConfigSignatureData() function
// TODO the below code generates signature directly from the SDK
func createSignature(t *testing.T, dsCtx *dsClientCtx, chConfigPath string, user string) *common.ConfigSignature {
	mspClient, err := mspclient.New(dsCtx.sdk.Context(), mspclient.WithOrg(dsCtx.org))
	require.NoError(t, err, "error creating a new msp management client for %s", dsCtx.org)
	usr, err := mspClient.GetSigningIdentity(user)
	require.NoError(t, err, "error creating a new SigningIdentity for %s", dsCtx.org)

	signature, err := dsCtx.rsCl.CreateConfigSignature(usr, chConfigPath)
	require.NoError(t, err, "error creating a new ConfigSignature for %s", org1)

	return signature
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
