/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

const (
	channelConfigFile = "mychannel.tx"
	channelID         = "mychannel"
	orgName           = org1Name
)

func initializeLedgerTests(t *testing.T) (*fabsdk.FabricSDK, []string) {
	// Using shared SDK instance to increase test speed.
	sdk := mainSDK

	//var sdkConfigFile = "../" + integration.ConfigTestFile
	//	sdk, err := fabsdk.New(config.FromFile(sdkConfigFile))
	//	if err != nil {
	//		t.Fatalf("SDK init failed: %s", err)
	//	}
	// Get signing identity that is used to sign create channel request

	orgMspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(orgName))
	if err != nil {
		t.Fatalf("failed to create org2MspClient, err : %s", err)
	}

	adminIdentity, err := orgMspClient.GetSigningIdentity("Admin")
	if err != nil {
		t.Fatalf("failed to load signing identity: %s", err)
	}

	configBackend, err := sdk.Config()
	if err != nil {
		t.Fatalf("failed to get config backend from SDK: %s", err)
	}

	targets, err := integration.OrgTargetPeers([]string{orgName}, configBackend)
	if err != nil {
		t.Fatalf("creating peers failed: %s", err)
	}

	req := resmgmt.SaveChannelRequest{ChannelID: channelID, ChannelConfigPath: path.Join("../../../", metadata.ChannelConfigPath, channelConfigFile), SigningIdentities: []msp.SigningIdentity{adminIdentity}}
	err = integration.InitializeChannel(sdk, orgName, req, targets)
	if err != nil {
		t.Fatalf("failed to ensure channel has been initialized: %s", err)
	}
	return sdk, targets
}

func TestLedgerQueries(t *testing.T) {

	// Setup tests with a random chaincode ID.
	sdk, targets := initializeLedgerTests(t)

	// Using shared SDK instance to increase test speed.
	//defer sdk.Close()

	chaincodeID := integration.GenerateRandomID()
	resp, err := integration.InstallAndInstantiateExampleCC(sdk, fabsdk.WithUser("Admin"), orgName, chaincodeID)
	require.Nil(t, err, "InstallAndInstantiateExampleCC return error")
	require.NotEmpty(t, resp, "instantiate response should be populated")

	//prepare required contexts

	channelClientCtx := sdk.ChannelContext(channelID, fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgName))

	// Get a ledger client.
	ledgerClient, err := ledger.New(channelClientCtx)
	require.Nil(t, err, "ledger new return error")

	// Test Query Info - retrieve values before transaction
	testTargets := targets[0:1]
	bciBeforeTx, err := ledgerClient.QueryInfo(ledger.WithTargetEndpoints(testTargets...))
	if err != nil {
		t.Fatalf("QueryInfo return error: %s", err)
	}

	// Invoke transaction that changes block state
	channelClient, err := channel.New(channelClientCtx)
	if err != nil {
		t.Fatalf("creating channel failed: %s", err)
	}

	txID, expectedQueryValue, err := changeBlockState(t, channelClient, chaincodeID)
	if err != nil {
		t.Fatalf("Failed to change block state (invoke transaction). Return error: %s", err)
	}

	verifyTargetsChangedBlockState(t, channelClient, chaincodeID, targets, expectedQueryValue)

	// Test Query Info - retrieve values after transaction
	bciAfterTx, err := ledgerClient.QueryInfo(ledger.WithTargetEndpoints(testTargets...))
	if err != nil {
		t.Fatalf("QueryInfo return error: %s", err)
	}

	// Test Query Info -- verify block size changed after transaction
	if (bciAfterTx.BCI.Height - bciBeforeTx.BCI.Height) <= 0 {
		t.Fatal("Block size did not increase after transaction")
	}

	testQueryTransaction(t, ledgerClient, txID, targets)

	testQueryBlock(t, ledgerClient, targets)

	testQueryBlockByTxID(t, ledgerClient, txID, targets)

	//prepare context
	clientCtx := sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgName))

	resmgmtClient, err := resmgmt.New(clientCtx)

	require.Nil(t, err, "resmgmt new return error")

	testInstantiatedChaincodes(t, chaincodeID, channelID, resmgmtClient, targets)

	testQueryConfigBlock(t, ledgerClient, targets)
}

func changeBlockState(t *testing.T, client *channel.Client, chaincodeID string) (fab.TransactionID, int, error) {

	req := channel.Request{
		ChaincodeID: chaincodeID,
		Fcn:         "invoke",
		Args:        integration.ExampleCCQueryArgs(),
	}
	resp, err := client.Query(req)
	if err != nil {
		return "", 0, errors.WithMessage(err, "query funds failed")
	}
	value := resp.Payload

	// Start transaction that will change block state
	txID, err := moveFundsAndGetTxID(t, client, chaincodeID)
	if err != nil {
		return "", 0, errors.WithMessage(err, "move funds failed")
	}

	valueInt, _ := strconv.Atoi(string(value))
	valueInt = valueInt + 1

	return txID, valueInt, nil
}

func verifyTargetsChangedBlockState(t *testing.T, client *channel.Client, chaincodeID string, targets []string, expectedValue int) {
	for _, target := range targets {
		verifyTargetChangedBlockState(t, client, chaincodeID, target, expectedValue)
	}
}

func verifyTargetChangedBlockState(t *testing.T, client *channel.Client, chaincodeID string, target string, expectedValue int) {

	const (
		maxRetries = 10
		retrySleep = 500 * time.Millisecond
	)

	for r := 0; r < 10; r++ {
		req := channel.Request{
			ChaincodeID: chaincodeID,
			Fcn:         "invoke",
			Args:        integration.ExampleCCQueryArgs(),
		}

		resp, err := client.Query(req, channel.WithTargetEndpoints(target))
		require.NoError(t, err, "query funds failed")
		valueAfterInvoke := resp.Payload

		// Verify that transaction changed block state
		valueAfterInvokeInt, _ := strconv.Atoi(string(valueAfterInvoke))
		if expectedValue == valueAfterInvokeInt {
			return
		}

		t.Logf("On Attempt [%d / %d]: SendTransaction didn't change the QueryValue %d", r, maxRetries, expectedValue)
		time.Sleep(retrySleep)
	}

	t.Error("Exceeded max retries in verifyPeerChangedBlockState")
}

func testQueryTransaction(t *testing.T, ledgerClient *ledger.Client, txID fab.TransactionID, targets []string) {

	// Test Query Transaction -- verify that valid transaction has been processed
	processedTransaction, err := ledgerClient.QueryTransaction(txID, ledger.WithTargetEndpoints(targets...))
	if err != nil {
		t.Fatalf("QueryTransaction return error: %s", err)
	}

	if processedTransaction.TransactionEnvelope == nil {
		t.Fatal("QueryTransaction failed to return transaction envelope")
	}

	// Test Query Transaction -- Retrieve non existing transaction
	_, err = ledgerClient.QueryTransaction("123ABC", ledger.WithTargetEndpoints(targets...))
	if err == nil {
		t.Fatal("QueryTransaction non-existing didn't return an error")
	}
}

func testQueryBlock(t *testing.T, ledgerClient *ledger.Client, targets []string) {

	// Retrieve current blockchain info
	bci, err := ledgerClient.QueryInfo(ledger.WithTargetEndpoints(targets...))
	if err != nil {
		t.Fatalf("QueryInfo return error: %s", err)
	}

	// Test Query Block by Hash - retrieve current block by hash
	block, err := ledgerClient.QueryBlockByHash(bci.BCI.CurrentBlockHash, ledger.WithTargetEndpoints(targets...))
	if err != nil {
		t.Fatalf("QueryBlockByHash return error: %s", err)
	}

	if block.Data == nil {
		t.Fatal("QueryBlockByHash block data is nil")
	}

	// Test Query Block by Hash - retrieve block by non-existent hash
	_, err = ledgerClient.QueryBlockByHash([]byte("non-existent"), ledger.WithTargetEndpoints(targets...))
	if err == nil {
		t.Fatal("QueryBlockByHash non-existent didn't return an error")
	}

	// Test Query Block - retrieve block by number
	block, err = ledgerClient.QueryBlock(1, ledger.WithTargetEndpoints(targets...))
	if err != nil {
		t.Fatalf("QueryBlock return error: %s", err)
	}
	if block.Data == nil {
		t.Fatal("QueryBlock block data is nil")
	}

	// Test Query Block - retrieve block by non-existent number
	_, err = ledgerClient.QueryBlock(2147483647, ledger.WithTargetEndpoints(targets...))
	if err == nil {
		t.Fatal("QueryBlock non-existent didn't return an error")
	}
}

func testQueryBlockByTxID(t *testing.T, ledgerClient *ledger.Client, txID fab.TransactionID, targets []string) {

	// Test Query Block- retrieve block by non-existent tx ID
	_, err := ledgerClient.QueryBlockByTxID("non-existent", ledger.WithTargetEndpoints(targets...))
	if err == nil {
		t.Fatal("QueryBlockByTxID non-existent didn't return an error")
	}

	// Test Query Block - retrieve block by valid tx ID
	block, err := ledgerClient.QueryBlockByTxID(txID, ledger.WithTargetEndpoints(targets...))
	if err != nil {
		t.Fatalf("QueryBlockByTxID return error: %s", err)
	}
	if block.Data == nil {
		t.Fatal("QueryBlockByTxID block data is nil")
	}

}

func testInstantiatedChaincodes(t *testing.T, ccID string, channelID string, resmgmtClient *resmgmt.Client, targets []string) {

	found := false

	// Test Query Instantiated chaincodes
	chaincodeQueryResponse, err := resmgmtClient.QueryInstantiatedChaincodes(channelID, resmgmt.WithTargetEndpoints(targets...), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatalf("QueryInstantiatedChaincodes return error: %s", err)
	}

	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		t.Logf("**InstantiatedCC: %s", chaincode)
		if chaincode.Name == ccID {
			found = true
		}
	}

	if !found {
		t.Fatalf("QueryInstantiatedChaincodes failed to find instantiated %s chaincode", ccID)
	}
}

// MoveFundsAndGetTxID ...
func moveFundsAndGetTxID(t *testing.T, client *channel.Client, chaincodeID string) (fab.TransactionID, error) {

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	req := channel.Request{
		ChaincodeID:  chaincodeID,
		Fcn:          "invoke",
		Args:         integration.ExampleCCTxArgs(),
		TransientMap: transientDataMap,
	}
	resp, err := client.Execute(req)
	if err != nil {
		return "", errors.WithMessage(err, "execute move funds failed")
	}

	return resp.TransactionID, nil
}

func testQueryConfigBlock(t *testing.T, ledgerClient *ledger.Client, targets []string) {

	// Retrieve current channel configuration
	cfgEnvelope, err := ledgerClient.QueryConfig(ledger.WithTargetEndpoints(targets...))
	if err != nil {
		t.Fatalf("QueryConfig return error: %s", err)
	}

	if cfgEnvelope == nil {
		t.Fatal("QueryConfig config data is nil")
	}

}
