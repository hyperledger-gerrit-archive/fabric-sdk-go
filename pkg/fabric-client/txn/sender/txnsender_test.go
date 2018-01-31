/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sender

import (
	"crypto/rand"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/mocks"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn/env"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
)

func TestCreateTransaction(t *testing.T) {

	//Test Empty proposal response scenario
	_, err := CreateTransaction([]*apitxn.TransactionProposalResponse{})

	if err == nil || err.Error() != "at least one proposal response is necessary" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, for empty proposal response scenario")
	}

	//Test invalid proposal header scenario

	txid := apitxn.TransactionID{
		ID: "1234",
	}

	test := &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				TxnID:          txid,
				Proposal:       &pb.Proposal{Header: []byte("TEST"), Extension: []byte(""), Payload: []byte("")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")},
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
		},
	}

	input := []*apitxn.TransactionProposalResponse{test}

	_, err = CreateTransaction(input)

	if err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, invalid proposal header scenario")
	}

	//Test invalid proposal payload scenario
	test = &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				TxnID:          txid,
				Proposal:       &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("TEST")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")},
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
		},
	}

	input = []*apitxn.TransactionProposalResponse{test}

	_, err = CreateTransaction(input)
	if err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, invalid proposal payload scenario")
	}

	//Test proposal response
	test = &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				Proposal:       &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")}, TxnID: txid,
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
		},
	}

	input = []*apitxn.TransactionProposalResponse{test}
	_, err = CreateTransaction(input)

	if err == nil || err.Error() != "proposal response was not successful, error code 99, msg success" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction")
	}

	//Test repeated field header nil scenario

	test = &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				Proposal:       &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")}, TxnID: txid,
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 200, Payload: []byte("")}},
		},
	}

	_, err = CreateTransaction([]*apitxn.TransactionProposalResponse{test})

	if err == nil || err.Error() != "repeated field endorsements has nil element" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction")
	}

	//TODO: Need actual sample payload for success case

}

func TestSendInstantiateProposal(t *testing.T) {
	//Setup channel
	user := mocks.NewMockUserWithMSPID("test", "1234")
	ctx := mocks.NewMockContext(user)
	channelID := "testChannel"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_apitxn.NewMockProposalProcessor(mockCtrl)

	tp := apitxn.TransactionProposal{SignedProposal: &pb.SignedProposal{}}
	tpr := apitxn.TransactionProposalResult{Endorser: "example.com", Status: 99, Proposal: tp, ProposalResponse: nil}

	proc.EXPECT().ProcessTransactionProposal(gomock.Any()).Return(tpr, nil)
	proc.EXPECT().ProcessTransactionProposal(gomock.Any()).Return(tpr, nil)
	targets := []apitxn.ProposalProcessor{proc}

	tresponse, txnid, err := SendInstantiateProposal(ctx, channelID, "", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	if err == nil || err.Error() != "chaincodeName is required" {
		t.Fatal("Validation for chain code name parameter for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = SendInstantiateProposal(ctx, channelID, "qscc", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	tresponse, txnid, err = SendInstantiateProposal(ctx, channelID, "qscc", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	if err == nil || err.Error() != "chaincodePath is required" {
		t.Fatal("Validation for chain code path for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = SendInstantiateProposal(ctx, channelID, "qscc", nil, "test",
		"", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	if err == nil || err.Error() != "chaincodeVersion is required" {
		t.Fatal("Validation for chain code version for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = SendInstantiateProposal(ctx, channelID, "qscc", nil, "test",
		"1", nil, nil, nil)
	if err == nil || err.Error() != "chaincodePolicy is required" {
		t.Fatal("Validation for chain code policy for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = SendInstantiateProposal(ctx, channelID, "qscc", nil, "test",
		"1", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	if err != nil || len(tresponse) == 0 || txnid.ID == "" {
		t.Fatal("Send Instantiate Proposal Test failed")
	}

	tresponse, txnid, err = SendInstantiateProposal(ctx, channelID, "qscc", nil, "test",
		"1", cauthdsl.SignedByMspMember("Org1MSP"), nil, nil)

	if err == nil || err.Error() != "missing peer objects for chaincode proposal" {
		t.Fatal("Missing peer objects validation is not working as expected")
	}

	// Define the private data collection policy config
	collConfig := []*common.CollectionConfig{
		newCollectionConfig("somecollection", 1, 3, cauthdsl.SignedByAnyMember([]string{"Org1MSP", "Org2MSP"})),
	}
	tresponse, txnid, err = SendInstantiateProposal(ctx, channelID, "qscc", nil, "test",
		"1", cauthdsl.SignedByMspMember("Org1MSP"), collConfig, targets)
	if err != nil || len(tresponse) == 0 || txnid.ID == "" {
		t.Fatal("Send Instantiate Proposal Test failed")
	}
}

func TestSendUpgradeProposal(t *testing.T) {
	//Setup channel
	user := mocks.NewMockUserWithMSPID("test", "1234")
	ctx := mocks.NewMockContext(user)
	channelID := "testChannel"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_apitxn.NewMockProposalProcessor(mockCtrl)

	tp := apitxn.TransactionProposal{SignedProposal: &pb.SignedProposal{}}
	tpr := apitxn.TransactionProposalResult{Endorser: "example.com", Status: 99, Proposal: tp, ProposalResponse: nil}

	proc.EXPECT().ProcessTransactionProposal(gomock.Any()).Return(tpr, nil)
	targets := []apitxn.ProposalProcessor{proc}

	//Add a Peer
	tresponse, txnid, err := SendUpgradeProposal(ctx, channelID, "", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err == nil || err.Error() != "chaincodeName is required" {
		t.Fatal("Validation for chain code name parameter for send Upgrade Proposal failed")
	}

	tresponse, txnid, err = SendUpgradeProposal(ctx, channelID, "qscc", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	tresponse, txnid, err = SendUpgradeProposal(ctx, channelID, "qscc", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err == nil || err.Error() != "chaincodePath is required" {
		t.Fatal("Validation for chain code path for send Upgrade Proposal failed")
	}

	tresponse, txnid, err = SendUpgradeProposal(ctx, channelID, "qscc", nil, "test",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err == nil || err.Error() != "chaincodeVersion is required" {
		t.Fatal("Validation for chain code version for send Upgrade Proposal failed")
	}

	tresponse, txnid, err = SendUpgradeProposal(ctx, channelID, "qscc", nil, "test",
		"2", nil, nil)
	if err == nil || err.Error() != "chaincodePolicy is required" {
		t.Fatal("Validation for chain code policy for send Upgrade Proposal failed")
	}

	tresponse, txnid, err = SendUpgradeProposal(ctx, channelID, "qscc", nil, "test",
		"2", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err != nil || len(tresponse) == 0 || txnid.ID == "" {
		t.Fatal("Send Upgrade Proposal Test failed")
	}

	tresponse, txnid, err = SendUpgradeProposal(ctx, channelID, "qscc", nil, "test",
		"2", cauthdsl.SignedByMspMember("Org1MSP"), nil)
	if err == nil || err.Error() != "missing peer objects for chaincode proposal" {
		t.Fatal("Missing peer objects validation is not working as expected")
	}
}

type mockReader struct {
	err error
}

func (r *mockReader) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	n, _ := rand.Read(p)
	return n, nil
}

func TestBroadcastEnvelope(t *testing.T) {
	lsnr1 := make(chan *fab.SignedEnvelope)
	lsnr2 := make(chan *fab.SignedEnvelope)
	//Create mock orderers
	orderer1 := mocks.NewMockOrderer("1", lsnr1)
	orderer2 := mocks.NewMockOrderer("2", lsnr2)

	orderers := []fab.Orderer{orderer1, orderer2}

	sigEnvelope := &fab.SignedEnvelope{
		Signature: []byte(""),
		Payload:   []byte(""),
	}
	res, err := BroadcastEnvelope(sigEnvelope, orderers)

	if err != nil || res.Err != nil {
		t.Fatalf("Test Broadcast Envelope Failed, cause %v %v", err, res)
	}

	// Ensure only 1 orderer was selected for broadcast
	firstSelected := 0
	secondSelected := 0
	for i := 0; i < 2; i++ {
		select {
		case <-lsnr1:
			firstSelected = 1
		case <-lsnr2:
			secondSelected = 1
		case <-time.After(time.Second):
		}
	}

	if firstSelected+secondSelected != 1 {
		t.Fatal("Both or none orderers were selected for broadcast:", firstSelected+secondSelected)
	}

	// Now make 1 of them fail and repeatedly broadcast
	broadcastCount := 50
	for i := 0; i < broadcastCount; i++ {
		orderer1.(mocks.MockOrderer).EnqueueSendBroadcastError(errors.New("Service Unavailable"))
	}
	// It should always succeed even though one of them has failed
	for i := 0; i < broadcastCount; i++ {
		if res, err := BroadcastEnvelope(sigEnvelope, orderers); err != nil || res.Err != nil {
			t.Fatalf("Test Broadcast Envelope Failed, cause %v %v", err, res)
		}
	}

	// Now, fail both and ensure any attempt fails
	for i := 0; i < broadcastCount; i++ {
		orderer1.(mocks.MockOrderer).EnqueueSendBroadcastError(errors.New("Service Unavailable"))
		orderer2.(mocks.MockOrderer).EnqueueSendBroadcastError(errors.New("Service Unavailable"))
	}

	for i := 0; i < broadcastCount; i++ {
		res, err := BroadcastEnvelope(sigEnvelope, orderers)
		if err != nil {
			t.Fatalf("Test Broadcast sending failed, cause %v", err)
		}
		if res.Err == nil {
			t.Fatal("Test Broadcast succeeded, but it should have failed")
		}
		if !strings.Contains(res.Err.Error(), "Service Unavailable") {
			t.Fatal("Test Broadcast failed but didn't return the correct reason(should contain 'Service Unavailable')")
		}
	}

	emptyOrderers := []fab.Orderer{}
	_, err = BroadcastEnvelope(sigEnvelope, emptyOrderers)

	if err == nil || err.Error() != "orderers not set" {
		t.Fatal("orderers not set validation on broadcast envelope is not working as expected")
	}
}

func TestSendTransaction(t *testing.T) {
	//Setup channel
	user := mocks.NewMockUserWithMSPID("test", "1234")
	ctx := mocks.NewMockContext(user)

	response, err := SendTransaction(ctx, nil, nil)

	//Expect orderer is nil error
	if response != nil || err == nil || err.Error() != "orderers is nil" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'orderers is nil' error")
	}

	//Create mock orderer
	orderer := mocks.NewMockOrderer("", nil)
	orderers := []fab.Orderer{orderer}

	//Call Send Transaction with nil tx
	response, err = SendTransaction(ctx, nil, orderers)

	//Expect tx is nil error
	if response != nil || err == nil || err.Error() != "transaction is nil" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'transaction is nil' error")
	}

	//Create tx with nil proposal
	txn := apitxn.Transaction{
		Proposal: &apitxn.TransactionProposal{
			Proposal: nil,
		},
		Transaction: &pb.Transaction{},
	}

	//Call Send Transaction with nil proposal
	response, err = SendTransaction(ctx, &txn, orderers)

	//Expect proposal is nil error
	if response != nil || err == nil || err.Error() != "proposal is nil" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'proposal is nil' error")
	}

	//Create tx with improper proposal header
	txn = apitxn.Transaction{
		Proposal: &apitxn.TransactionProposal{
			Proposal: &pb.Proposal{Header: []byte("TEST")},
		},
		Transaction: &pb.Transaction{},
	}
	//Call Send Transaction
	response, err = SendTransaction(ctx, &txn, orderers)

	//Expect header unmarshal error
	if response != nil || err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with '...unmarshal...' error")
	}

	//Create tx with proper proposal header
	txn = apitxn.Transaction{
		Proposal: &apitxn.TransactionProposal{
			Proposal: &pb.Proposal{Header: []byte(""), Payload: []byte(""), Extension: []byte("")},
		},
		Transaction: &pb.Transaction{},
	}

	//Call Send Transaction
	response, err = SendTransaction(ctx, &txn, orderers)

	if response == nil || err != nil {
		t.Fatalf("Test SendTransaction failed, reason : '%s'", err.Error())
	}
}

func TestBuildChannelHeader(t *testing.T) {

	header, err := env.BuildChannelHeader(common.HeaderType_CHAINCODE_PACKAGE, "test", "", 1, "1234", time.Time{}, []byte{})

	if err != nil || header == nil {
		t.Fatalf("Test Build Channel Header failed, cause : '%s'", err.Error())
	}

}

func TestSignPayload(t *testing.T) {
	//Setup channel
	user := mocks.NewMockUserWithMSPID("test", "1234")
	ctx := mocks.NewMockContext(user)

	signedEnv, err := env.SignPayload(ctx, []byte(""))

	if err != nil || signedEnv == nil {
		t.Fatal("Test Sign Payload Failed")
	}
}

func TestConcurrentOrderers(t *testing.T) {
	//Setup channel
	user := mocks.NewMockUserWithMSPID("test", "1234")
	ctx := mocks.NewMockContext(user)

	// Determine number of orderers to use - environment can override
	const numOrderersDefault = 2000
	numOrderersEnv := os.Getenv("TEST_MASSIVE_ORDERER_COUNT")
	numOrderers, err := strconv.Atoi(numOrderersEnv)
	if err != nil {
		numOrderers = numOrderersDefault
	}

	orderers := setupMassiveTestOrderer(numOrderers)

	txn := apitxn.Transaction{
		Proposal: &apitxn.TransactionProposal{
			Proposal: &pb.Proposal{},
		},
		Transaction: &pb.Transaction{},
	}
	_, err = SendTransaction(ctx, &txn, orderers)
	if err != nil {
		t.Fatalf("SendTransaction returned error: %s", err)
	}
}

func newCollectionConfig(collName string, requiredPeerCount, maxPeerCount int32, policy *common.SignaturePolicyEnvelope) *common.CollectionConfig {
	return &common.CollectionConfig{
		Payload: &common.CollectionConfig_StaticCollectionConfig{
			StaticCollectionConfig: &common.StaticCollectionConfig{
				Name:              collName,
				RequiredPeerCount: requiredPeerCount,
				MaximumPeerCount:  maxPeerCount,
				MemberOrgsPolicy: &common.CollectionPolicyConfig{
					Payload: &common.CollectionPolicyConfig_SignaturePolicy{
						SignaturePolicy: policy,
					},
				},
			},
		},
	}
}

func setupMassiveTestOrderer(numberOfOrderers int) []fab.Orderer {
	orderers := []fab.Orderer{}

	for i := 0; i < numberOfOrderers; i++ {
		orderer := mocks.NewMockOrderer(fmt.Sprintf("http://mock%d.orderers.r.us", i), nil)
		orderers = append(orderers, orderer)
	}

	return orderers
}
