/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"bytes"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

const (
	org1User      = "User1"
	org1AdminUser = "Admin"
)

func TestEventClient(t *testing.T) {
	chainCodeID := mainChaincodeID
	sdk := mainSDK
	testSetup := mainTestSetup

	chContextProvider := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))
	chContext, err := chContextProvider()
	if err != nil {
		t.Fatalf("error getting channel context: %s", err)
	}
	eventService, err := chContext.ChannelService().EventService()
	if err != nil {
		t.Fatalf("error getting event service: %s", err)
	}

	if chContext.EndpointConfig().EventServiceType() == fab.DeliverEventServiceType {
		t.Run("Deliver Filtered Block Events", func(t *testing.T) {
			// Filtered block events are the default for the deliver event client
			testEventService(t, testSetup, sdk, chainCodeID, false, eventService)
		})
		t.Run("Deliver Block Events", func(t *testing.T) {
			eventService, err := chContext.ChannelService().EventService(client.WithBlockEvents())
			if err != nil {
				t.Fatalf("error getting event service: %s", err)
			}
			testEventService(t, testSetup, sdk, chainCodeID, true, eventService)
		})
	} else {
		// Block events are the default for the event hub client
		t.Run("Event Hub Block Events", func(t *testing.T) {
			testEventService(t, testSetup, sdk, chainCodeID, true, eventService)
		})
	}
}

func testEventService(t *testing.T, testSetup *integration.BaseSetupImpl, sdk *fabsdk.FabricSDK, chainCodeID string, blockEvents bool, eventService fab.EventService) {
	_, cancel, transactor, err := getTransactor(sdk, testSetup.ChannelID, "Admin", testSetup.OrgID)
	if err != nil {
		t.Fatalf("Failed to get channel transactor: %s", err)
	}
	defer cancel()

	peers, err := getProposalProcessors(sdk, "Admin", testSetup.OrgID, testSetup.Targets)
	assert.Nil(t, err, "creating peers failed")

	tpResponses, prop, err := createAndSendTransactionProposal(
		transactor,
		chainCodeID,
		"invoke",
		[][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("10")},
		peers,
		nil,
	)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v", err)
	}

	txID := string(prop.TxnID)

	var wg sync.WaitGroup
	var numExpected uint32

	var breg fab.Registration
	var beventch <-chan *fab.BlockEvent
	if blockEvents {
		breg, beventch, err = eventService.RegisterBlockEvent()
		if err != nil {
			t.Fatalf("Error registering for block events: %s", err)
		}
		defer eventService.Unregister(breg)
		numExpected++
		wg.Add(1)
	}

	fbreg, fbeventch, err := eventService.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("Error registering for filtered block events: %s", err)
	}
	defer eventService.Unregister(fbreg)
	numExpected++
	wg.Add(1)

	ccreg, cceventch, err := eventService.RegisterChaincodeEvent(chainCodeID, ".*")
	if err != nil {
		t.Fatalf("Error registering for filtered block events: %s", err)
	}
	defer eventService.Unregister(ccreg)
	numExpected++
	wg.Add(1)

	txReg, txstatusch, err := eventService.RegisterTxStatusEvent(txID)
	if err != nil {
		t.Fatalf("Error registering for Tx Status event: %s", err)
	}
	defer eventService.Unregister(txReg)
	numExpected++
	wg.Add(1)

	var numReceived uint32

	if beventch != nil {
		go func() {
			defer wg.Done()
			select {
			case event, ok := <-beventch:
				if !ok {
					t.Fatalf("unexpected closed channel while waiting for Tx Status event")
				}
				t.Logf("Received block event: %#v", event)
				if event.Block == nil {
					t.Fatalf("Expecting block in block event but got nil")
				}
				atomic.AddUint32(&numReceived, 1)
			case <-time.After(5 * time.Second):
			}
		}()
	}

	go func() {
		defer wg.Done()
		for {
			select {
			case event, ok := <-fbeventch:
				if !ok {
					t.Fatalf("unexpected closed channel while waiting for Tx Status event")
				}
				t.Logf("Received filtered block event: %#v", event)
				if event.FilteredBlock == nil || len(event.FilteredBlock.FilteredTransactions) == 0 {
					t.Fatalf("Expecting one transaction in filtered block but got none")
				}
				filteredTx := event.FilteredBlock.FilteredTransactions[0]
				if filteredTx.Txid != string(txID) {
					// Not our event
					continue
				}
				atomic.AddUint32(&numReceived, 1)
			case <-time.After(5 * time.Second):
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case event, ok := <-cceventch:
			if !ok {
				t.Fatalf("unexpected closed channel while waiting for Tx Status event")
			}
			t.Logf("Received chaincode event: %#v", event)
			if event.ChaincodeID != chainCodeID {
				t.Fatalf("Expecting event for CC ID [%s] but got event for CC ID [%s]", chainCodeID, event.ChaincodeID)
			}
			if blockEvents {
				expectedPayload := []byte("Test Payload")
				if bytes.Compare(event.Payload, expectedPayload) != 0 {
					t.Fatalf("Expecting payload [%s] but got [%s]", []byte("Test Payload"), event.Payload)
				}
			} else if event.Payload != nil {
				t.Fatalf("Expecting nil payload for filtered events but got [%s]", event.Payload)
			}
			if event.SourceURL == "" {
				t.Fatalf("Expecting event source URL but got none")
			}
			if event.BlockNumber == 0 {
				t.Fatalf("Expecting non-zero block number")
			}
			atomic.AddUint32(&numReceived, 1)
		case <-time.After(5 * time.Second):
			return
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case txStatus, ok := <-txstatusch:
			if !ok {
				t.Fatalf("unexpected closed channel while waiting for Tx Status event")
			}
			t.Logf("Received Tx Status event: %#v", txStatus)
			if txStatus.TxID != string(txID) {
				t.Fatalf("Expecting event for TxID [%s] but got event for TxID [%s]", txID, txStatus.TxID)
			}
			if txStatus.SourceURL == "" {
				t.Fatalf("Expecting event source URL but got none")
			}
			if txStatus.BlockNumber == 0 {
				t.Fatalf("Expecting non-zero block number")
			}
			atomic.AddUint32(&numReceived, 1)
		case <-time.After(5 * time.Second):
			return
		}
	}()

	// Commit the transaction to generate events
	_, err = createAndSendTransaction(transactor, prop, tpResponses)
	if err != nil {
		t.Fatalf("First invoke failed err: %v", err)
	}

	wg.Wait()

	if numReceived != numExpected {
		t.Fatalf("expecting %d events but received %d", numExpected, numReceived)
	}
}

// createAndSendTransaction uses transactor to create and send transaction
func createAndSendTransaction(transactor fab.Sender, proposal *fab.TransactionProposal, resps []*fab.TransactionProposalResponse) (*fab.TransactionResponse, error) {

	txRequest := fab.TransactionRequest{
		Proposal:          proposal,
		ProposalResponses: resps,
	}
	tx, err := transactor.CreateTransaction(txRequest)
	if err != nil {
		return nil, errors.WithMessage(err, "CreateTransaction failed")
	}

	transactionResponse, err := transactor.SendTransaction(tx)
	if err != nil {
		return nil, errors.WithMessage(err, "SendTransaction failed")

	}

	return transactionResponse, nil
}
