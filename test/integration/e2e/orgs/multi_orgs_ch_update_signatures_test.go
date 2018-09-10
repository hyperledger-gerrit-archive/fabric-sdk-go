/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package orgs

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dsChannel              = "dschannel"
	adminUser              = "Admin"
	user1                  = "User1"
	dsChannelOrg1SDKConfig = "overrides/org1_dschannel.yaml"
	dsChannelOrg2SDKConfig = "overrides/org2_dschannel.yaml"
	mainConfigFilename     = "config_e2e_multiorg_bootstrap.yaml"
	isSDKSigning           = true
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

	sigDir, err := ioutil.TempDir(".", "externalsignatures")
	require.NoError(t, err, "Failed to create temporary directory")
	defer func() {
		err = os.RemoveAll(sigDir)
		require.NoError(t, err, "Failed to remove temporary directory")
	}()
	t.Logf("created tempDir: %s", sigDir)

	// create org1 ConfigSignature
	// for now, use SDK signing
	// TODO fix externalSigning issue
	org1DsChannelCfgSig := executeSigning(t, org1ClCtx, chConfigPath, adminUser, sigDir, isSDKSigning)
	t.Logf("org1DsChannelCfgSig:[%+v]", org1DsChannelCfgSig)

	// create org2 ConfigSignature
	org2DsChannelCfgSig := executeSigning(t, org2ClCtx, chConfigPath, adminUser, sigDir, isSDKSigning)
	t.Logf("org2DsChannelCfgSig:[%+v]", org2DsChannelCfgSig)

	// create signature for anchor peer of org1
	chConfigOrg1MSPPath := integration.GetChannelConfigPath(fmt.Sprintf("%s%sMSPanchors.tx", dsChannel, org1))
	org1MSPDsChannelCfgSig := executeSigning(t, org1ClCtx, chConfigOrg1MSPPath, adminUser, sigDir, isSDKSigning)
	t.Logf("org1MSPDsChannelCfgSig:[%+v]", org1MSPDsChannelCfgSig)

	// create signature for anchor peer of org2
	chConfigOrg2MSPPath := integration.GetChannelConfigPath(fmt.Sprintf("%s%sMSPanchors.tx", dsChannel, org2))
	org2MSPDsChannelCfgSig := executeSigning(t, org2ClCtx, chConfigOrg2MSPPath, adminUser, sigDir, isSDKSigning)
	t.Logf("org2MSPDsChannelCfgSig:[%+v]", org2MSPDsChannelCfgSig)

	// create channel on the orderer
	req := resmgmt.SaveChannelRequest{ChannelID: dsChannel, ChannelConfigPath: chConfigPath}
	txID, err := ordererClCtx.rsCl.SaveChannel(req, resmgmt.WithConfigSignatures(org1DsChannelCfgSig, org2DsChannelCfgSig), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "error creating channel %s signatures for %s", dsChannel, ordererOrgName)
	require.NotEmpty(t, txID, "transaction ID should be populated for Channel create for %s", ordererOrgName)

	var lastConfigBlock uint64
	lastConfigBlock = integration.WaitForOrdererConfigUpdate(t, org1ClCtx.rsCl, dsChannel, true, lastConfigBlock)

	// create channel on anchor peer of org1
	req.ChannelConfigPath = chConfigOrg1MSPPath
	txID, err = org1ClCtx.rsCl.SaveChannel(req, resmgmt.WithConfigSignatures(org1MSPDsChannelCfgSig), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "error creating channel %s for anchor peer of %s", dsChannel, org1)
	require.NotEmpty(t, txID, "transaction ID should be populated for Channel create for anchor peer of %s", org1)

	lastConfigBlock = integration.WaitForOrdererConfigUpdate(t, org1ClCtx.rsCl, dsChannel, false, lastConfigBlock)

	// create channel on anchor peer of org2
	req.ChannelConfigPath = chConfigOrg2MSPPath
	txID, err = org2ClCtx.rsCl.SaveChannel(req, resmgmt.WithConfigSignatures(org2MSPDsChannelCfgSig), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.NoError(t, err, "error creating channel %s for anchor peer of %s", dsChannel, org2)
	require.NotEmpty(t, txID, "transaction ID should be populated for Channel create for anchor peer of %s", org2)

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

func executeSigning(t *testing.T, dsCtx *dsClientCtx, chConfigPath, user, sigDir string, isSDK bool) *common.ConfigSignature {
	if isSDK {
		return executeSDKSigning(t, dsCtx, chConfigPath, user, sigDir)
	}
	return executeExternalSigning(t, dsCtx, chConfigPath, user, sigDir)
}

func executeSDKSigning(t *testing.T, dsCtx *dsClientCtx, chConfigPath, user, sigDir string) *common.ConfigSignature {
	chCfgName := filepath.Base(chConfigPath)
	chCfgName = strings.Trim(chCfgName, ".tx")

	channelCfgSigSDK := createSignatureFromSDK(t, dsCtx, chConfigPath, user)
	f, err := os.Create(fmt.Sprintf("%s/%s_%s_%s_sbytes.txt.sha256", sigDir, chCfgName, dsCtx.org, user))
	require.NoError(t, err, "Failed to create temporary file")
	defer func() {
		err = f.Close()
		require.NoError(t, err, "Failed to close signature file")
	}()
	bufferedWriter := bufio.NewWriter(f)
	_, err = bufferedWriter.Write(channelCfgSigSDK.Signature)
	assert.NoError(t, err, "must be able to write signature of [%s-%s] to buffer", dsCtx.org, user)
	bufferedWriter.Flush()
	shf, err := os.Create(fmt.Sprintf("%s/%s_%s_%s_sHeaderbytes.txt", sigDir, chCfgName, dsCtx.org, user))
	require.NoError(t, err, "Failed to create temporary file")
	defer func() {
		err = shf.Close()
		require.NoError(t, err, "Failed to close signature header file")
	}()
	bufferedWriter = bufio.NewWriter(shf)
	_, err = bufferedWriter.Write(channelCfgSigSDK.SignatureHeader)
	assert.NoError(t, err, "must be able to write signature header of [%s-%s] to buffer", dsCtx.org, user)
	bufferedWriter.Flush()
	// now that signature is stored in the filesystem, load it to simulate external signature read
	return loadExternalSignature(t, dsCtx.org, chConfigPath, user, sigDir)
}

func createSignatureFromSDK(t *testing.T, dsCtx *dsClientCtx, chConfigPath string, user string) *common.ConfigSignature {
	mspClient, err := mspclient.New(dsCtx.sdk.Context(), mspclient.WithOrg(dsCtx.org))
	require.NoError(t, err, "error creating a new msp management client for %s", dsCtx.org)
	usr, err := mspClient.GetSigningIdentity(user)
	require.NoError(t, err, "error creating a new SigningIdentity for %s", dsCtx.org)

	signature, err := dsCtx.rsCl.CreateConfigSignature(usr, chConfigPath)
	require.NoError(t, err, "error creating a new ConfigSignature for %s", org1)

	return signature
}

func executeExternalSigning(t *testing.T, clCtx *dsClientCtx, chConfigPath, user string, sigDir string) *common.ConfigSignature {
	// example generating and loading an external signature (not signed by the SDK)
	generateChConfigData(t, clCtx, chConfigPath, user, sigDir)

	// sign signature data with external tool (script running openssl)
	generateExternalChConfigSignature(t, clCtx.org, user, chConfigPath, sigDir)

	cs := loadExternalSignature(t, clCtx.org, chConfigPath, user, sigDir)

	return cs
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

		configBackends, err := config.FromFile(integration.GetConfigPath(mainConfigFilename))()
		require.NoError(t, err, "failed to read config backend from file for org %s", org)

		configBackendsOverrides, err := config.FromFile(integration.GetConfigPath(configFileOverride))()
		require.NoError(t, err, "failed to read config backend from file for org %s", org)

		return append(configBackendsOverrides, configBackends...), nil
	}
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

// generateChConfigData will generate serialized Channel Config Data for external signing
func generateChConfigData(t *testing.T, dsCtx *dsClientCtx, chConfigPath, user, sigDir string) {
	mspClient, err := mspclient.New(dsCtx.sdk.Context(), mspclient.WithOrg(dsCtx.org))
	require.NoError(t, err, "error creating a new msp management client for %s", dsCtx.org)
	u, err := mspClient.GetSigningIdentity(user)
	require.NoError(t, err, "error creating a new SigningIdentity for %s", dsCtx.org)

	d, err := dsCtx.rsCl.CreateConfigSignatureData(u, chConfigPath)
	require.NoError(t, err, "Failed to fetch Channel config data for signing")

	chCfgName := filepath.Base(chConfigPath)
	chCfgName = strings.Trim(chCfgName, ".tx")

	// create a temporary file and save the channel config data in that file
	f, err := os.Create(fmt.Sprintf("%s/%s_%s_%s_sbytes.txt", sigDir, chCfgName, dsCtx.org, user))
	require.NoError(t, err, "Failed to create temporary file")
	defer func() {
		err = f.Sync() // ensure data is flushed into storage
		require.NoError(t, err, "Failed to sync signature file")
		err = f.Close()
		require.NoError(t, err, "Failed to close signature file")
	}()

	bufferedWriter := bufio.NewWriter(f)
	_, err = bufferedWriter.Write(d.SigningBytes)
	assert.NoError(t, err, "must be able to write signature of [%s-%s] to buffer", dsCtx.org, user)

	err = bufferedWriter.Flush()
	assert.NoError(t, err, "must be able to flush buffer for signature of [%s-%s] to buffer", dsCtx.org, user)

	// marshal signatureHeader struct for later use
	shf, err := os.Create(fmt.Sprintf("%s/%s_%s_%s_sHeaderbytes.txt", sigDir, chCfgName, dsCtx.org, user))
	require.NoError(t, err, "Failed to create temporary file")
	defer func() {
		err = shf.Sync() // ensure data is flushed into storage
		require.NoError(t, err, "Failed to close signature header file")
		err = shf.Close()
		require.NoError(t, err, "Failed to close signature header file")
	}()

	bufferedWriter = bufio.NewWriter(shf)
	_, err = bufferedWriter.Write(d.SignatureHeaderBytes)
	assert.NoError(t, err, "must be able to write signature header of [%s-%s] to buffer", dsCtx.org, user)

	err = bufferedWriter.Flush()
	assert.NoError(t, err, "must be able to flush buffer for signature of [%s-%s] to buffer", dsCtx.org, user)
}

func generateExternalChConfigSignature(t *testing.T, org, user, chConfigPath, sigDir string) {
	chCfgName := filepath.Base(chConfigPath)
	chCfgName = strings.Trim(chCfgName, ".tx")

	cmd := exec.Command("scripts/generate_signature.sh", org, user, chCfgName, sigDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	assert.NoError(t, err, "Failed to create external signature for [%s, %s, %s], script error: [%s]", org, user, chCfgName, stderr.String())

	t.Logf("running generate_signature.sh script output: %s", string(b))
}

func loadExternalSignature(t *testing.T, org, chConfigPath, user, sigDir string) *common.ConfigSignature {
	chCfgName := filepath.Base(chConfigPath)
	chCfgName = strings.Trim(chCfgName, ".tx")

	f, err := os.Open(fmt.Sprintf("%s/%s_%s_%s_sbytes.txt.sha256", sigDir, chCfgName, org, user))
	defer func() {
		err = f.Close()
		require.NoError(t, err, "Failed to close signature file")
	}()
	require.NoError(t, err, "Failed to open signature file")

	sig, err := ioutil.ReadAll(f)
	require.NoError(t, err, "Failed to read signature data")

	fh, err := os.Open(fmt.Sprintf("%s/%s_%s_%s_sHeaderbytes.txt", sigDir, chCfgName, org, user))
	defer func() {
		err = fh.Close()
		require.NoError(t, err, "Failed to close signature header file")
	}()
	require.NoError(t, err, "Failed to open signature header file")
	sigHeader, err := ioutil.ReadAll(fh)
	require.NoError(t, err, "Failed to read signature header data")

	cs := &common.ConfigSignature{
		Signature:       sig,
		SignatureHeader: sigHeader,
	}
	return cs
}
