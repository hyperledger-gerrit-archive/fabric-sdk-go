// +build experimental

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channelevents"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

func TestChannelEvents(t *testing.T) {
	setup := initializeChannelEventTests(t)

	eventClient, err := newChannelEventClient(&setup)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	smokeTestChannelEvents(t, setup, eventClient)
}

func initializeChannelEventTests(t *testing.T) BaseSetupImpl {
	testSetup := BaseSetupImpl{
		ConfigFile:      "../fixtures/config/config_test.yaml",
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	testSetup.ChainCodeID = GenerateRandomID()

	// Install and Instantiate Events CC
	if err := testSetup.InstallCC(testSetup.ChainCodeID, "github.com/events_cc", "v0", nil); err != nil {
		t.Fatalf("installCC return error: %v", err)
	}

	if err := testSetup.InstantiateCC(testSetup.ChainCodeID, "github.com/events_cc", "v0", nil); err != nil {
		t.Fatalf("instantiateCC return error: %v", err)
	}

	return testSetup
}

func smokeTestChannelEvents(t *testing.T, testSetup BaseSetupImpl, eventClient fab.ChannelEventClient) {
	fcn := "invoke"

	// Arguments for events CC
	var args []string
	args = append(args, "invoke")
	args = append(args, "SEVERE")

	tpResponses, tx, err := testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, fcn, args, []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, nil)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v", err)
	}

	// Register transaction for TX Status event
	txEventChan := make(chan *fab.TxStatusEvent)
	if reg, err := eventClient.RegisterTxStatusEvent(tx.ID, txEventChan); err != nil {
		t.Fatalf("RegisterTxStatusEvent returned error: %v", err)
	} else {
		defer eventClient.Unregister(reg)
	}

	_, err = testSetup.CreateAndSendTransaction(testSetup.Channel, tpResponses)
	if err != nil {
		t.Fatalf("CreateAndSendTransaction failed err: %v", err)
	}

	select {
	case status, ok := <-txEventChan:
		if !ok {
			t.Fatalf("connection with channel event client was unexpectly terminated")
		}
		if status.TxID != tx.ID {
			t.Fatalf("expecting event for TxID [%s] but received event for [%s]", tx.ID, status.TxID)
		}
		if status.TxValidationCode != pb.TxValidationCode_VALID {
			t.Fatalf("expecting TxValidationCode [%s] but received [%s]", pb.TxValidationCode_VALID, status.TxValidationCode)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for TX Status event")
	}
}

func newChannelEventClient(setup *BaseSetupImpl) (fab.ChannelEventClient, error) {
	peersConfig, err := setup.Client.Config().PeersConfig(setup.OrgID)
	if err != nil {
		return nil, fmt.Errorf("error getting peers in org %s: %s", setup.OrgID, err)
	}

	if len(peersConfig) == 0 {
		return nil, fmt.Errorf("no peers in org %s", setup.OrgID)
	}

	eventClient, err := channelevents.NewClient(setup.Client, peersConfig[0], setup.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("could not start channel event client: %s", err)
	}

	if err := eventClient.Connect(); err != nil {
		return nil, fmt.Errorf("could not connect channel event client: %s", err)
	}

	return eventClient, nil
}
