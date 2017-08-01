/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	fabricTxn "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func TestChainCodeInvoke(t *testing.T) {

	//Testing chaincode invoke for golang user chaincodes
	testChainCodeInvokeByChaincodeType(t, pb.ChaincodeSpec_GOLANG)

	//Testing chaincode invoke for binary user chaincodes
	testChainCodeInvokeByChaincodeType(t, pb.ChaincodeSpec_BINARY)
}

func testChainCodeInvokeByChaincodeType(t *testing.T, ccType pb.ChaincodeSpec_Type) {

	testSetup := BaseSetupImpl{
		ConfigFile:      "../fixtures/config/config_test.yaml",
		ChannelID:       "mychannel",
		OrgID:           "peerorg1",
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	var err error
	if ccType == pb.ChaincodeSpec_BINARY {
		err = testSetup.InstallAndInstantiateBinaryExampleCC()
	} else {
		err = testSetup.InstallAndInstantiateExampleCC()
	}

	if err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC for %v return error: %v", ccType, err)
		return
	}

	// Get Query value before invoke
	value, err := testSetup.QueryAsset()
	if err != nil {
		t.Fatalf("getQueryValue for %v return error: %v", ccType, err)
	}
	fmt.Printf("*** QueryValue before invoke %s\n", value)

	eventID := "test([a-zA-Z]+)"

	// Register callback for chaincode event
	done, rce := fabricTxn.RegisterCCEvent(testSetup.ChainCodeID, eventID, testSetup.EventHub)

	err = moveFunds(&testSetup)
	if err != nil {
		t.Fatalf("Move funds for %v return error: %v", ccType, err)
	}

	select {
	case <-done:
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC for for %v for eventId(%s)\n", ccType, eventID)
	}

	testSetup.EventHub.UnregisterChaincodeEvent(rce)

	valueAfterInvoke, err := testSetup.QueryAsset()
	if err != nil {
		t.Errorf("getQueryValue for %v return error: %v", ccType, err)
		return
	}
	fmt.Printf("*** QueryValue after invoke %s\n", valueAfterInvoke)

	valueInt, _ := strconv.Atoi(value)
	valueInt = valueInt + 1
	valueAfterInvokeInt, _ := strconv.Atoi(valueAfterInvoke)
	if valueInt != valueAfterInvokeInt {
		t.Fatalf("SendTransaction didn't change the QueryValue for %v ", ccType)
	}

}

// moveFunds ...
func moveFunds(setup *BaseSetupImpl) error {
	fcn := "invoke"

	var args []string
	args = append(args, "move")
	args = append(args, "a")
	args = append(args, "b")
	args = append(args, "1")

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	_, err := fabricTxn.InvokeChaincode(setup.Client, setup.Channel, []apitxn.ProposalProcessor{setup.Channel.PrimaryPeer()}, setup.EventHub, setup.ChainCodeID, fcn, args, transientDataMap)
	return err
}
