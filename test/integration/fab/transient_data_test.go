/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/pkg/errors"
)

// TestTransient ...
func TestTransient(t *testing.T) {
	// Using shared SDK instance to increase test speed.
	sdk := mainSDK
	testSetup := mainTestSetup
	chaincodeID := mainChaincodeID

	//testSetup := integration.BaseSetupImpl{
	//	ConfigFile:    "../" + integration.ConfigTestFile,
	//	ChannelID:     "mychannel",
	//	OrgID:         org1Name,
	//	ChannelConfig: path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
	//}

	//sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile))
	//if err != nil {
	//	t.Fatalf("Failed to create new SDK: %s", err)
	//}
	//defer sdk.Close()

	//if err := testSetup.Initialize(sdk); err != nil {
	//	t.Fatalf(err.Error())
	//}

	//chaincodeID := integration.GenerateRandomID()
	//if err := integration.InstallAndInstantiateExampleCC(sdk, fabsdk.WithUser("Admin"), testSetup.OrgID, chaincodeID); err != nil {
	//	t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	//}

	fcn := "invoke"
	transientData := "Transient data test..."

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte(transientData)

	transactor, err := getTransactor(sdk, testSetup.ChannelID, "Admin", testSetup.OrgID)
	if err != nil {
		t.Fatalf("Failed to get channel transactor: %s", err)
	}

	transactionProposalResponse, _, err := createAndSendTransactionProposal(transactor, chaincodeID, fcn, integration.ExampleCCTxArgs(), testSetup.Targets[:1], transientDataMap)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v", err)
	}
	strResponse := string(transactionProposalResponse[0].ProposalResponse.GetResponse().Payload)
	//validate transient data exists in proposal
	if len(strResponse) == 0 {
		t.Fatalf("Transient data does not exist: expected %s", transientData)
	}

	//verify transient data content
	if strResponse != transientData {
		t.Fatalf("Expected '%s' in transient data field. Received '%s' ", transientData, strResponse)
	}
	//transient data null
	transientDataMap["result"] = []byte{}
	transactionProposalResponse, _, err = createAndSendTransactionProposal(transactor, chaincodeID, fcn, integration.ExampleCCTxArgs(), testSetup.Targets[:1], transientDataMap)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal with empty transient data return an error: %v", err)
	}
	//validate that transient data does not exist in proposal
	strResponse = string(transactionProposalResponse[0].ProposalResponse.GetResponse().Payload)
	if len(strResponse) != 0 {
		t.Fatalf("Transient data validation has failed. An empty transient data was expected but %s was returned", strResponse)
	}

}

// createAndSendTransactionProposal uses transactor to create and send transaction proposal
func createAndSendTransactionProposal(transactor fab.ProposalSender, chainCodeID string,
	fcn string, args [][]byte, targets []fab.ProposalProcessor, transientData map[string][]byte) ([]*fab.TransactionProposalResponse, *fab.TransactionProposal, error) {

	propReq := fab.ChaincodeInvokeRequest{
		Fcn:          fcn,
		Args:         args,
		TransientMap: transientData,
		ChaincodeID:  chainCodeID,
	}

	txh, err := transactor.CreateTransactionHeader()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "creating transaction header failed")
	}

	tp, err := txn.CreateChaincodeInvokeProposal(txh, propReq)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "creating transaction proposal failed")
	}

	tpr, err := transactor.SendTransactionProposal(tp, targets)
	return tpr, tp, err
}

func getTransactor(sdk *fabsdk.FabricSDK, channelID string, user string, orgName string) (fab.Transactor, error) {

	clientChannelContextProvider := sdk.ChannelContext(channelID, fabsdk.WithUser(user), fabsdk.WithOrg(orgName))

	channelContext, err := clientChannelContextProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "channel service creation failed")
	}
	chService := channelContext.ChannelService()

	return chService.Transactor()
}