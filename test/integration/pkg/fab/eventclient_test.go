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
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

const eventTimeWindow = 60 * time.Second // the maximum amount of time to watch for events.

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

	t.Log("Testing Deliver events")
	t.Run("Deliver Filtered Block Events", func(t *testing.T) {
		// Filtered block events are the default for the deliver event client
		testEventService(t, testSetup, sdk, chainCodeID, false, eventService)
	})
	t.Run("Deliver Block Events", func(t *testing.T) {
		eventServ, err := chContext.ChannelService().EventService(client.WithBlockEvents())
		if err != nil {
			t.Fatalf("error getting event service: %s", err)
		}
		testEventService(t, testSetup, sdk, chainCodeID, true, eventServ)
	})
}

func testEventService(t *testing.T, testSetup *integration.BaseSetupImpl, sdk *fabsdk.FabricSDK, chainCodeID string, blockEvents bool, eventService fab.EventService) {
	_, cancel, transactor, err := getTransactor(sdk, testSetup.ChannelID, "Admin", testSetup.OrgID)
	if err != nil {
		t.Fatalf("Failed to get channel transactor: %s", err)
	}
	defer cancel()

	tpResponses, prop, txID := sendTxProposal(sdk, testSetup, t, transactor, chainCodeID)

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
	// Commit the transaction to generate events
	_, err = createAndSendTransaction(transactor, prop, tpResponses)
	if err != nil {
		t.Fatalf("First invoke failed err: %s", err)
	}

	if blockEvents {
		go checkBlockEvent(&wg, beventch, t, &numReceived)
	}

	go checkFilteredBlockEvent(&wg, fbeventch, t, &numReceived, txID)
	go checkCCEvent(&wg, cceventch, t, &numReceived, chainCodeID, blockEvents)
	go checkTxStatusEvent(&wg, txstatusch, t, &numReceived, txID)

	wg.Wait()

	if numReceived != numExpected {
		t.Fatalf("expecting %d events but received %d", numExpected, numReceived)
	}
}

func sendTxProposal(sdk *fabsdk.FabricSDK, testSetup *integration.BaseSetupImpl, t *testing.T, transactor fab.Transactor, chainCodeID string) ([]*fab.TransactionProposalResponse, *fab.TransactionProposal, string) {
	peers, err := getProposalProcessors(sdk, "Admin", testSetup.OrgID, testSetup.Targets)
	require.Nil(t, err, "creating peers failed")
	tpResponses, prop, err := createAndSendTransactionProposal(
		transactor,
		chainCodeID,
		"invoke",
		[][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("10")},
		peers,
		nil,
	)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %s", err)
	}
	txID := string(prop.TxnID)
	return tpResponses, prop, txID
}

func checkTxStatusEvent(wg *sync.WaitGroup, txstatusch <-chan *fab.TxStatusEvent, t *testing.T, numReceived *uint32, txID string) {
	defer wg.Done()
	select {
	case txStatus, ok := <-txstatusch:
		if !ok {
			test.Failf(t, "unexpected closed channel while waiting for Tx Status event")
		}
		t.Logf("Received Tx Status event: %#v", txStatus)
		if txStatus.TxID != string(txID) {
			test.Failf(t, "Expecting event for TxID [%s] but got event for TxID [%s]", txID, txStatus.TxID)
		}
		if txStatus.SourceURL == "" {
			test.Failf(t, "Expecting event source URL but got none")
		}
		if txStatus.BlockNumber == 0 {
			test.Failf(t, "Expecting non-zero block number")
		}
		atomic.AddUint32(numReceived, 1)
	case <-time.After(eventTimeWindow):
		return
	}
}

func checkCCEvent(wg *sync.WaitGroup, cceventch <-chan *fab.CCEvent, t *testing.T, numReceived *uint32, chainCodeID string, blockEvents bool) {
	defer wg.Done()
	select {
	case event, ok := <-cceventch:
		if !ok {
			test.Failf(t, "unexpected closed channel while waiting for Tx Status event")
		}
		t.Logf("Received chaincode event: %#v", event)
		if event.ChaincodeID != chainCodeID {
			test.Failf(t, "Expecting event for CC ID [%s] but got event for CC ID [%s]", chainCodeID, event.ChaincodeID)
		}
		if blockEvents {
			expectedPayload := []byte("Test Payload")
			if !bytes.Equal(event.Payload, expectedPayload) {
				test.Failf(t, "Expecting payload [%s] but got [%s]", []byte("Test Payload"), event.Payload)
			}
		} else if event.Payload != nil {
			test.Failf(t, "Expecting nil payload for filtered events but got [%s]", event.Payload)
		}
		if event.SourceURL == "" {
			test.Failf(t, "Expecting event source URL but got none")
		}
		if event.BlockNumber == 0 {
			test.Failf(t, "Expecting non-zero block number")
		}
		atomic.AddUint32(numReceived, 1)
	case <-time.After(eventTimeWindow):
		return
	}
}

func checkFilteredBlockEvent(wg *sync.WaitGroup, fbeventch <-chan *fab.FilteredBlockEvent, t *testing.T, numReceived *uint32, txID string) {
	defer wg.Done()
	for {
		select {
		case event, ok := <-fbeventch:
			if !ok {
				test.Failf(t, "unexpected closed channel while waiting for Tx Status event")
			}
			t.Logf("Received filtered block event: %#v", event)
			if event.FilteredBlock == nil || len(event.FilteredBlock.FilteredTransactions) == 0 {
				test.Failf(t, "Expecting one transaction in filtered block but got none")
			}
			filteredTx := event.FilteredBlock.FilteredTransactions[0]
			if filteredTx.Txid != string(txID) {
				// Not our event
				continue
			}
			t.Logf("Received correct filtered block event: %#v", event)
			atomic.AddUint32(numReceived, 1)
			return
		case <-time.After(eventTimeWindow):
			return
		}
	}
}

func checkBlockEvent(wg *sync.WaitGroup, beventch <-chan *fab.BlockEvent, t *testing.T, numReceived *uint32) {
	defer wg.Done()
	select {
	case event, ok := <-beventch:
		if !ok {
			test.Failf(t, "unexpected closed channel while waiting for Tx Status event")
		}
		t.Logf("Received block event: %#v", event)
		if event.Block == nil {
			test.Failf(t, "Expecting block in block event but got nil")
		}
		atomic.AddUint32(numReceived, 1)
	case <-time.After(eventTimeWindow):
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

func TestMultipleEventsBySeekTypes(t *testing.T) {

	chainCodeID := mainChaincodeID
	testSetup := mainTestSetup

	//Run with seek type default and test behaviour

	for i := 0; i < 4; i++ {
		//create new sdk
		sdk, err := fabsdk.New(integration.ConfigBackend)
		if err != nil {
			t.Fatal("failed to get new sdk instance:", err)
		}

		//create new channel context
		chContextProvider := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))
		chContext, err := chContextProvider()
		if err != nil {
			t.Fatalf("error getting channel context: %s", err)
		}

		//create new event service with default opts
		eventService, err := chContext.ChannelService().EventService()
		if err != nil {
			t.Fatalf("error getting event service: %s", err)
		}

		testChannelEventsSeekOptions(t, testSetup, sdk, chContext, chainCodeID, false, eventService, "")
	}

	//Run with seek type newest and test behaviour
	for i := 0; i < 4; i++ {
		//create new sdk
		sdk, err := fabsdk.New(integration.ConfigBackend)
		if err != nil {
			t.Fatal("failed to get new sdk instance:", err)
		}

		//create new channel context
		chContextProvider := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))
		chContext, err := chContextProvider()
		if err != nil {
			t.Fatalf("error getting channel context: %s", err)
		}

		//create new event service with deliver client opts
		eventService, err := chContext.ChannelService().EventService(deliverclient.WithSeekType(seek.Newest))
		if err != nil {
			t.Fatalf("error getting event service: %s", err)
		}

		testChannelEventsSeekOptions(t, testSetup, sdk, chContext, chainCodeID, false, eventService, seek.Newest)
	}

}

func testChannelEventsSeekOptions(t *testing.T, testSetup *integration.BaseSetupImpl, sdk *fabsdk.FabricSDK, chContext context.Channel, chainCodeID string, blockEvents bool, eventService fab.EventService, seekType seek.Type) {

	defer sdk.Close()

	//get transactor
	_, cancel, transactor, err := getTransactor(sdk, testSetup.ChannelID, "Admin", testSetup.OrgID)
	require.Nil(t, err, "Failed to get channel transactor: %s", err)
	defer cancel()

	//register chanicode event
	ccreg, cceventch, err := eventService.RegisterChaincodeEvent(chainCodeID, ".*")
	require.Nil(t, err, "Error registering for filtered block events: %s", err)
	defer eventService.Unregister(ccreg)

	// prepare and commit the transaction to generate events
	tpResponses, prop, txID := sendTxProposal(sdk, testSetup, t, transactor, chainCodeID)
	_, err = createAndSendTransaction(transactor, prop, tpResponses)
	require.Nil(t, err, "First invoke failed err: %s", err)

	var event *fab.CCEvent
	var ok bool
	select {
	case event, ok = <-cceventch:
		if !ok {
			test.Failf(t, "unexpected closed channel while waiting for Tx Status event")
		}
		if event.ChaincodeID != chainCodeID {
			test.Failf(t, "Expecting event for CC ID [%s] but got event for CC ID [%s]", chainCodeID, event.ChaincodeID)
		}
		if blockEvents {
			expectedPayload := []byte("Test Payload")
			if !bytes.Equal(event.Payload, expectedPayload) {
				test.Failf(t, "Expecting payload [%s] but got [%s]", []byte("Test Payload"), event.Payload)
			}
		} else if event.Payload != nil {
			test.Failf(t, "Expecting nil payload for filtered events but got [%s]", event.Payload)
		}
		if event.SourceURL == "" {
			test.Failf(t, "Expecting event source URL but got none")
		}
		if event.BlockNumber == 0 {
			test.Failf(t, "Expecting non-zero block number")
		}
	case <-time.After(eventTimeWindow):
		return
	}

	//If seek type is newest then the first event we get from event channel is not related to the first transaction happened after registration, it is
	//actually latest block from the chain
	require.Equal(t, seekType == seek.Newest, txID != event.TxID, "for seek type[%s], txID [%s], event.txID[%s] ,condition didn't match", seekType, txID, event.TxID)

	//If seek type is default, then event dispatcher uses first block only for block height calculations, it doesn't publish anything
	//to event channel, and first event we get from event channel actually belongs to first transaction after registration.
	require.Equal(t, seekType == "", txID == event.TxID, "for seek type[%s], txID [%s], event.txID[%s] ,condition didn't match", seekType, txID, event.TxID)
}
