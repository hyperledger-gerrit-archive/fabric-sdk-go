// +build channelevents

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	fab "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

type Outcome string
type State string
type NumBlockEvents uint
type NumCCEvents uint

type EventsReceived struct {
	BlockEvents NumBlockEvents
	CCEvents    NumCCEvents
}

const (
	reconnectedOutcome Outcome = "reconnected"
	terminatedOutcome  Outcome = "terminated"
	timedOutOutcome    Outcome = "timeout"
	connectedOutcome   Outcome = "connected"
	errorOutcome       Outcome = "error"

	connected    State = "connected"
	disconnected State = "disconnected"

	firstAttempt  ConnectionAttempt = 1
	secondAttempt ConnectionAttempt = 2
	thirdAttempt  ConnectionAttempt = 3
	fourthAttempt ConnectionAttempt = 4

	expectOneBlockEvent    NumBlockEvents = 1
	expectTwoBlockEvents   NumBlockEvents = 2
	expectThreeBlockEvents NumBlockEvents = 3
	expectFourBlockEvents  NumBlockEvents = 4

	expectOneCCEvent    NumCCEvents = 1
	expectTwoCCEvents   NumCCEvents = 2
	expectThreeCCEvents NumCCEvents = 3
	expectFourCCEvents  NumCCEvents = 4
)

func TestInvalidOptions(t *testing.T) {
	if _, _, err := newClient("", "grpc://localhost:7051", "admin"); err == nil {
		t.Fatalf("expecting error with no channel ID but got none")
	}
	if _, _, err := newClient("mychannel", "", "admin"); err == nil {
		t.Fatalf("expecting error with no peer URL but got none")
	}
}

func TestFailedChannelRegistration(t *testing.T) {
	channelID := "mychannel"
	errMsg := "mock failed channel registration"
	opts := NewClientOpts()
	conn := NewMockConnection(NewResultsOpt(NewResult(RegisterChannel, FailResult, errMsg)))
	opts.connectionProvider = NewMockConnectionProviderFactory().Provider(conn)

	eventClient, err := newClientWithOpts(channelID, "grpc://localhost:7051", "admin", opts)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err == nil {
		t.Fatalf("expecting error connecting channel event client but got none")
	} else if !strings.Contains(err.Error(), errMsg) {
		t.Fatalf("expecting error message to contain [%s] when connecting channel event client but got [%s]", errMsg, err)
	}
}

func TestInvalidChannelRegistrationResponse(t *testing.T) {
	channelID := "mychannel"
	opts := NewClientOpts()
	conn := NewMockConnection(NewResultsOpt(NewResult(RegisterChannel, InvalidChannelResult)))
	opts.connectionProvider = NewMockConnectionProviderFactory().Provider(conn)

	eventClient, err := newClientWithOpts(channelID, "grpc://localhost:7051", "admin", opts)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err == nil {
		t.Fatalf("expecting error connecting channel event client but got none")
	}
}

func TestTimedOutChannelRegistration(t *testing.T) {
	channelID := "mychannel"
	opts := NewClientOpts()
	conn := NewMockConnection(NewResultsOpt(NewResult(RegisterChannel, NoOpResult)))
	opts.connectionProvider = NewMockConnectionProviderFactory().Provider(conn)

	eventClient, err := newClientWithOpts(channelID, "grpc://localhost:7051", "admin", opts)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err == nil {
		t.Fatalf("expecting error connecting channel event client due to no response from channel registration but got none")
	}
}

func TestTimedOutChannelUnregistration(t *testing.T) {
	channelID := "mychannel"
	opts := NewClientOpts()
	conn := NewMockConnection(NewResultsOpt(NewResult(UnregisterChannel, NoOpResult)))
	opts.connectionProvider = NewMockConnectionProviderFactory().Provider(conn)

	eventClient, err := newClientWithOpts(channelID, "grpc://localhost:7051", "admin", opts)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	// Should just disconnect gracefully
	eventClient.Disconnect()
}

func TestCallsOnClosedClient(t *testing.T) {
	eventClient, _, err := newClient("mychannel", "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt(FILTEREDBLOCKEVENT))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	eventClient.Disconnect()

	// Make sure you can call Disconnect again with no issues
	eventClient.Disconnect()

	if err := eventClient.Connect(); err == nil {
		t.Fatalf("expecting error connecting to closed channel event client but got none")
	}

	if _, err := eventClient.registerConnectionEvent(make(chan *apifabclient.ConnectionEvent)); err == nil {
		t.Fatalf("expecting error registering for connection events on closed channel event client but got none")
	}

	if _, err := eventClient.RegisterBlockEvent(make(chan *apifabclient.BlockEvent)); err == nil {
		t.Fatalf("expecting error registering for block events on closed channel event client but got none")
	}

	if _, err := eventClient.RegisterChaincodeEvent("ccid", "event", make(chan *apifabclient.CCEvent)); err == nil {
		t.Fatalf("expecting error registering for chaincode events on closed channel event client but got none")
	}

	if _, err := eventClient.RegisterTxStatusEvent("txid", make(chan *apifabclient.TxStatusEvent)); err == nil {
		t.Fatalf("expecting error registering for TX events on closed channel event client but got none")
	}

	if err := eventClient.Unregister(nil); err == nil {
		t.Fatalf("expecting error unregistering on closed channel event client but got none")
	}
}

func TestInvalidUnregister(t *testing.T) {
	channelID := "mychannel"
	eventClient, _, err := newClient(channelID, "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt(BLOCKEVENT))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	reg := "invalid registration"
	if err := eventClient.Unregister(reg); err == nil {
		t.Fatalf("expecting error unregistering with invalid registration but got none")
	}
}

func TestBlockEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newClient(channelID, "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt(BLOCKEVENT))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	eventch := make(chan *apifabclient.BlockEvent)
	registration, err := eventClient.RegisterBlockEvent(eventch)
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	_, err = eventClient.RegisterBlockEvent(eventch)
	if err == nil {
		t.Fatalf("expecting error registering multiple times for block events: %s", err)
	}
	if err := eventClient.Unregister(registration); err != nil {
		t.Fatalf("error unregistering: %s", err)
	}
	registration, err = eventClient.RegisterBlockEvent(eventch)
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}

	conn.ProduceEvent(newMockBlock(channelID))

	select {
	case _, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed channel")
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for block event")
	}

	if err := eventClient.Unregister(registration); err != nil {
		t.Fatalf("error unregistering: %s", err)
	}
}

func TestBlockEventsUnauthorized(t *testing.T) {
	eventClient, _, err := newClient("mychannel", "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt(FILTEREDBLOCKEVENT))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	eventch := make(chan *apifabclient.BlockEvent)
	if _, err := eventClient.RegisterBlockEvent(eventch); err == nil {
		t.Fatalf("expecting authorization error since client is not authorized to receive block events")
	}
}

func TestTxStatusEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newClient(channelID, "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt(FILTEREDBLOCKEVENT))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	eventch1 := make(chan *apifabclient.TxStatusEvent)
	if _, err := eventClient.RegisterTxStatusEvent("", eventch1); err == nil {
		t.Fatalf("expecting error registering for TxStatus event without a TX ID but got none")
	}
	reg1, err := eventClient.RegisterTxStatusEvent(txID1, eventch1)
	if err != nil {
		t.Fatalf("error registering for TxStatus events: %s", err)
	}
	_, err = eventClient.RegisterTxStatusEvent(txID1, eventch1)
	if err == nil {
		t.Fatalf("expecting error registering multiple times for TxStatus events: %s", err)
	}
	if err := eventClient.Unregister(reg1); err != nil {
		t.Fatalf("error unregistering: %s", err)
	}
	reg1, err = eventClient.RegisterTxStatusEvent(txID1, eventch1)
	if err != nil {
		t.Fatalf("error registering for TxStatus events: %s", err)
	}

	eventch2 := make(chan *apifabclient.TxStatusEvent)
	reg2, err := eventClient.RegisterTxStatusEvent(txID2, eventch2)
	if err != nil {
		t.Fatalf("error registering for TxStatus events: %s", err)
	}

	conn.ProduceEvent(
		newMockFilteredBlock(
			channelID,
			newMockTxEvent(txID1, txCode1),
			newMockTxEvent(txID2, txCode2),
		),
	)

	numReceived := 0
	done := false
	for !done {
		select {
		case event, ok := <-eventch1:
			if !ok {
				t.Fatalf("unexpected closed channel")
			} else {
				checkTxStatusEvent(t, event, txID1, txCode1)
				numReceived++
			}
		case event, ok := <-eventch2:
			if !ok {
				t.Fatalf("unexpected closed channel")
			} else {
				checkTxStatusEvent(t, event, txID2, txCode2)
				numReceived++
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for TxStatus event")
		}

		if numReceived == 2 {
			break
		}
	}

	if err := eventClient.Unregister(reg1); err != nil {
		t.Fatalf("error unregistering: %s", err)
	}
	if err := eventClient.Unregister(reg2); err != nil {
		t.Fatalf("error unregistering: %s", err)
	}

	eventClient.Disconnect()
}

func TestTxStatusEventsUnauthorized(t *testing.T) {
	eventClient, _, err := newClient("mychannel", "grpc://localhost:7051", "admin")
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	eventch := make(chan *apifabclient.TxStatusEvent)
	if _, err := eventClient.RegisterTxStatusEvent("txid", eventch); err == nil {
		t.Fatalf("expecting authorization error since client is not authorized to receive filtered events")
	}
}

func TestCCEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newClient(channelID, "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt(FILTEREDBLOCKEVENT))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}

	ccID1 := "mycc1"
	ccFilter1 := "event1"

	ccID2 := "mycc2"
	ccFilter2 := "event.*"

	txID1 := "1234"
	event1 := "event1"

	txID2 := "5678"
	event2 := "event2"

	txID3 := "9012"
	event3 := "event3"

	eventch1 := make(chan *apifabclient.CCEvent)
	if _, err := eventClient.RegisterChaincodeEvent("", ccFilter1, eventch1); err == nil {
		t.Fatalf("expecting error registering for chaincode events without CC ID but got none")
	}
	if _, err := eventClient.RegisterChaincodeEvent(ccID1, "", eventch1); err == nil {
		t.Fatalf("expecting error registering for chaincode events without event filter but got none")
	}
	if _, err := eventClient.RegisterChaincodeEvent(ccID1, ".(xxx", eventch1); err == nil {
		t.Fatalf("expecting error registering for chaincode events with invalid (regular expression) event filter but got none")
	}
	reg1, err := eventClient.RegisterChaincodeEvent(ccID1, ccFilter1, eventch1)
	if err != nil {
		t.Fatalf("error registering for chaincode events: %s", err)
	}
	_, err = eventClient.RegisterChaincodeEvent(ccID1, ccFilter1, eventch1)
	if err == nil {
		t.Fatalf("expecting error registering multiple times for chaincode events: %s", err)
	}
	if err := eventClient.Unregister(reg1); err != nil {
		t.Fatalf("error unregistering: %s", err)
	}
	reg1, err = eventClient.RegisterChaincodeEvent(ccID1, ccFilter1, eventch1)
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}

	eventch2 := make(chan *apifabclient.CCEvent)
	reg2, err := eventClient.RegisterChaincodeEvent(ccID2, ccFilter2, eventch2)
	if err != nil {
		t.Fatalf("error registering for chaincode events: %s", err)
	}

	conn.ProduceEvent(
		newMockFilteredBlock(
			channelID,
			newMockCCEvent(txID1, ccID1, event1),
			newMockCCEvent(txID2, ccID2, event2),
			newMockCCEvent(txID3, ccID2, event3),
		),
	)

	numReceived := 0
	done := false
	for !done {
		select {
		case event, ok := <-eventch1:
			if !ok {
				t.Fatalf("unexpected closed channel")
			} else {
				checkCCEvent(t, event, ccID1, event1)
				numReceived++
			}
		case event, ok := <-eventch2:
			if !ok {
				t.Fatalf("unexpected closed channel")
			} else {
				checkCCEvent(t, event, ccID2, event2, event3)
				numReceived++
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for TxStatus event")
		}

		if numReceived == 3 {
			break
		}
	}

	if err := eventClient.Unregister(reg1); err != nil {
		t.Fatalf("error unregistering: %s", err)
	}
	if err := eventClient.Unregister(reg2); err != nil {
		t.Fatalf("error unregistering: %s", err)
	}

	eventClient.Disconnect()
}

func TestCCEventsUnauthorized(t *testing.T) {
	eventClient, _, err := newClient("mychannel", "grpc://localhost:7051", "admin")
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	eventch := make(chan *apifabclient.CCEvent)
	if _, err := eventClient.RegisterChaincodeEvent("ccid", ".*", eventch); err == nil {
		t.Fatalf("expecting authorization error since client is not authorized to receive filtered events")
	}
}

// TestConnection tests the ability of the Channel Event Client to retry multiple
// times to connect and reconnect after it has disconnected.
func TestConnection(t *testing.T) {
	eventClient, _, err := newClient("mychannel", "grpc://localhost:7051", "")
	if err != nil {
		t.Fatalf("error creating new channel event client: %s", err)
	}
	if err := eventClient.Connect(); err == nil {
		t.Fatalf("expecting error with no user but got none")
	}

	// (1) Connect
	//     -> should fail to connect on the first and second attempt but succeed on the third attempt
	testConnect(t, 3, connectedOutcome,
		NewConnectResults(
			NewConnectResult(thirdAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
		),
	)

	// (1) Connect
	//     -> should fail to connect on the first attempt and no further attempts are to be made
	testConnect(t, 1, errorOutcome,
		NewConnectResults(),
	)

	// (1) Connect
	//     -> should succeed to connect on the first
	// (2) Disconnect
	//     -> should fail to reconnect on the first and second attempt but succeed on the third attempt
	testReconnect(t, true, 3, reconnectedOutcome,
		NewConnectResults(
			NewConnectResult(firstAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
			NewConnectResult(fourthAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
		),
	)

	// (1) Connect
	//     -> should succeed to connect on the first
	// (2) Disconnect
	//     -> should fail to reconnect after two attempts and then cleanly disconnect
	testReconnect(t, true, 2, terminatedOutcome,
		NewConnectResults(
			NewConnectResult(firstAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
		),
	)

	// (1) Connect
	//     -> should succeed to connect on the first
	// (2) Disconnect
	//     -> should should and not attempt to reconnect and cleanly disconnect
	testReconnect(t, false, 0, terminatedOutcome,
		NewConnectResults(
			NewConnectResult(firstAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
		),
	)
}

// TestReconnectRegistration tests the ability of the Channel Event Client to
// re-establish the existing registrations after a reconnecting.
func TestReconnectRegistration(t *testing.T) {
	// (1) Connect
	//     -> should connect with BLOCKEVENT and FILTEREDBLOCKEVENT permission
	// (2) Register for block events
	// (3) Register for CC events
	// (4) Send one block event
	//     -> should receive one block event
	// (5) Send one CC event
	//     -> should receive one CC event
	// (6) Disconnect
	//     -> should reconnect with BLOCKEVENT and FILTEREDBLOCKEVENT permission
	// (7) Send one block event
	//     -> should receive one block event
	// (8) Send one CC event
	//     -> should receive one CC event
	testReconnectRegistration(
		t, expectTwoBlockEvents, expectTwoCCEvents,
		NewConnectResults(
			NewConnectResult(firstAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
			NewConnectResult(secondAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT)),
	)

	// (1) Connect
	//     -> should connect with BLOCKEVENT and FILTEREDBLOCKEVENT permission
	// (2) Register for block events
	// (3) Register for CC events
	// (4) Send one block event
	//     -> should receive one block event
	// (5) Send one CC event
	//     -> should receive one CC event
	// (6) Disconnect
	//     -> should reconnect with only FILTEREDBLOCKEVENT permission
	// (7) Send one block event
	//     -> should NOT receive the block event
	// (8) Send one CC event
	//     -> should receive one CC event
	testReconnectRegistration(
		t, expectOneBlockEvent, expectTwoCCEvents,
		NewConnectResults(
			NewConnectResult(firstAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
			NewConnectResult(secondAttempt, FILTEREDBLOCKEVENT)),
	)

	// (1) Connect
	//     -> should connect with BLOCKEVENT and FILTEREDBLOCKEVENT permission
	// (2) Register for block events
	// (3) Register for CC events
	// (4) Send one block event
	//     -> should receive one block event
	// (5) Send one CC event
	//     -> should receive one CC event
	// (6) Disconnect
	//     -> should reconnect with only BLOCKEVENT permission
	// (7) Send one block event
	//     -> should receive one block event
	// (8) Send one CC event
	//     -> should NOT receive any CC event
	testReconnectRegistration(
		t, expectTwoBlockEvents, expectOneCCEvent,
		NewConnectResults(
			NewConnectResult(firstAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
			NewConnectResult(secondAttempt, BLOCKEVENT)),
	)
}

func testConnect(t *testing.T, maxConnectAttempts uint, expectedOutcome Outcome, connAttemptResult ConnectAttemptResults) {
	cp := NewMockConnectionProviderFactory()

	opts := &ClientOpts{
		ResponseTimeout:    1 * time.Second,
		MaxConnectAttempts: maxConnectAttempts,
		connectionProvider: cp.FlakeyProvider(connAttemptResult),
	}
	eventClient, err := newClientWithOpts("mychannel", "grpc://localhost:7051", "admin", opts)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}

	var outcome Outcome
	if err := eventClient.Connect(); err != nil {
		outcome = errorOutcome
	} else {
		outcome = connectedOutcome
		defer eventClient.Disconnect()
	}

	if outcome != expectedOutcome {
		t.Fatalf("Expecting that the reconnection attempt would result in [%s] but got [%s]", expectedOutcome, outcome)
	}

}

func testReconnect(t *testing.T, reconnect bool, maxReconnectAttempts uint, expectedOutcome Outcome, connAttemptResult ConnectAttemptResults) {
	cp := NewMockConnectionProviderFactory()

	connectch := make(chan *apifabclient.ConnectionEvent)

	opts := &ClientOpts{
		ResponseTimeout:            3 * time.Second,
		Reconnect:                  reconnect,
		MaxConnectAttempts:         1,
		MaxReconnectAttempts:       maxReconnectAttempts,
		TimeBetweenConnectAttempts: time.Millisecond,
		ConnectEvents:              connectch,
		connectionProvider:         cp.FlakeyProvider(connAttemptResult),
	}
	eventClient, err := newClientWithOpts("mychannel", "grpc://localhost:7051", "admin", opts)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}

	outcomech := make(chan Outcome)
	go listenConnection(connectch, outcomech)

	// Test automatic reconnect handling
	cp.Connection().Disconnect(errors.New("testing reconnect handling"))

	var outcome Outcome

	select {
	case outcome = <-outcomech:
	case <-time.After(5 * time.Second):
		outcome = timedOutOutcome
	}

	if outcome != expectedOutcome {
		t.Fatalf("Expecting that the reconnection attempt would result in [%s] but got [%s]", expectedOutcome, outcome)
	}

	eventClient.Disconnect()
}

// testReconnectRegistration tests the scenario when an events client is registered to receive events and the connection to the
// event service is lost. After the connection is re-established, events should once again be received without the caller having to
// re-register for those events.
func testReconnectRegistration(t *testing.T, expectedBlockEvents NumBlockEvents, expectedCCEvents NumCCEvents, connectResults ConnectAttemptResults) {
	channelID := "mychannel"
	ccID := "mycc"

	cp := NewMockConnectionProviderFactory()

	opts := &ClientOpts{
		ResponseTimeout:            3 * time.Second,
		Reconnect:                  true,
		MaxConnectAttempts:         1,
		MaxReconnectAttempts:       1,
		TimeBetweenConnectAttempts: time.Millisecond,
		connectionProvider:         cp.FlakeyProvider(connectResults),
	}
	eventClient, err := newClientWithOpts(channelID, "grpc://localhost:7051", "admin", opts)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}

	blockch := make(chan *apifabclient.BlockEvent)
	if _, err = eventClient.RegisterBlockEvent(blockch); err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}

	ccch := make(chan *apifabclient.CCEvent)
	if _, err = eventClient.RegisterChaincodeEvent(ccID, ".*", ccch); err != nil {
		t.Fatalf("error registering for chaincode events: %s", err)
	}

	numCh := make(chan EventsReceived)
	go listenEvents(blockch, ccch, 3*time.Second, numCh, expectedBlockEvents, expectedCCEvents)

	time.Sleep(250 * time.Millisecond)

	// Produce a block event
	cp.Connection().ProduceEvent(newMockBlock(channelID))

	// Produce a chaincode event
	cp.Connection().ProduceEvent(
		newMockFilteredBlock(
			channelID,
			newMockCCEvent("tx1", ccID, "event1"),
		),
	)

	time.Sleep(250 * time.Millisecond)

	// Test automatic reconnect handling
	cp.Connection().Disconnect(errors.New("testing reconnect handling"))

	var eventsReceived EventsReceived

	time.Sleep(time.Second)

	// Produce one more block and CC event
	cp.Connection().ProduceEvent(newMockBlock(channelID))
	cp.Connection().ProduceEvent(
		newMockFilteredBlock(
			channelID,
			newMockCCEvent("tx2", ccID, "event2"),
		),
	)

	select {
	case received, ok := <-numCh:
		if !ok {
			t.Fatalf("connection closed prematurely")
		} else {
			eventsReceived = received
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("timed out waiting for events")
	}

	if eventsReceived.BlockEvents != expectedBlockEvents {
		t.Fatalf("Expecting to receive [%d] block events but received [%d]", expectedBlockEvents, eventsReceived.BlockEvents)
	}

	if eventsReceived.CCEvents != expectedCCEvents {
		t.Fatalf("Expecting to receive [%d] CC events but received [%d]", expectedCCEvents, eventsReceived.CCEvents)
	}

	eventClient.Disconnect()
}

func listenConnection(eventch chan *apifabclient.ConnectionEvent, outcome chan Outcome) {
	var state State

	for {
		e, ok := <-eventch
		if !ok {
			outcome <- terminatedOutcome
			break
		}
		if e.Connected {
			if state == disconnected {
				outcome <- reconnectedOutcome
			} else {
			}
			state = connected
		} else {
			state = disconnected
		}
	}
}

func listenEvents(blockch chan *apifabclient.BlockEvent, ccch chan *apifabclient.CCEvent, waitDuration time.Duration, numEventsCh chan EventsReceived, expectedBlockEvents NumBlockEvents, expectedCCEvents NumCCEvents) {
	var numBlockEventsReceived NumBlockEvents = 0
	var numCCEventsReceived NumCCEvents = 0

	for {
		select {
		case _, ok := <-blockch:
			if ok {
				numBlockEventsReceived++
			} else {
				// The channel was closed by the event client. Make a new channel so
				// that we don't get into a tight loop
				blockch = make(chan *apifabclient.BlockEvent)
			}
		case _, ok := <-ccch:
			if ok {
				numCCEventsReceived++
			} else {
				// The channel was closed by the event client. Make a new channel so
				// that we don't get into a tight loop
				ccch = make(chan *apifabclient.CCEvent)
			}
		case <-time.After(waitDuration):
			numEventsCh <- EventsReceived{BlockEvents: numBlockEventsReceived, CCEvents: numCCEventsReceived}
			return
		}
		if numBlockEventsReceived >= expectedBlockEvents && numCCEventsReceived >= expectedCCEvents {
			numEventsCh <- EventsReceived{BlockEvents: numBlockEventsReceived, CCEvents: numCCEventsReceived}
			return
		}
	}
}

func newClient(channelID string, peerURL string, userName string, connOpts ...MockConnOpt) (*Client, *MockConnection, error) {
	conn := NewMockConnection(connOpts...)

	opts := NewClientOpts()
	opts.connectionProvider = NewMockConnectionProviderFactory().Provider(conn)

	client, err := newClientWithOpts(channelID, peerURL, userName, opts)
	return client, conn, err
}

func newClientWithOpts(channelID string, peerURL string, userName string, opts *ClientOpts) (*Client, error) {
	config := mocks.NewMockConfig()

	fabClient := fab.NewClient(config)

	logging.SetLevel(logging.DEBUG, "fabric_sdk_go")

	if userName != "" {
		user := mocks.NewMockUser(userName)
		fabClient.SetUserContext(user)
	}

	fabClient.SetSigningManager(mocks.NewMockSigningManager())

	peerConfig := &apiconfig.PeerConfig{
		URL:         peerURL,
		GRPCOptions: make(map[string]interface{}),
	}

	eventClient, err := NewClientWithOpts(
		fabClient, peerConfig, channelID,
		opts)
	return eventClient, err
}

func checkTxStatusEvent(t *testing.T, event *apifabclient.TxStatusEvent, expectedTxID string, expectedCode pb.TxValidationCode) {
	if event.TxID != expectedTxID {
		t.Fatalf("expecting event for TxID [%s] but received event for TxID [%s]", expectedTxID, event.TxID)
	}
	if event.TxValidationCode != expectedCode {
		t.Fatalf("expecting TxValidationCode [%s] but received [%s]", expectedCode, event.TxValidationCode)
	}
}

func checkCCEvent(t *testing.T, event *apifabclient.CCEvent, expectedCCID string, expectedEventNames ...string) {
	if event.ChaincodeID != expectedCCID {
		t.Fatalf("expecting event for CC [%s] but received event for CC [%s]", expectedCCID, event.ChaincodeID)
	}
	found := false
	for _, eventName := range expectedEventNames {
		if event.EventName == eventName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expecting one of [%v] but received [%s]", expectedEventNames, event.EventName)
	}
}

func newMockBlock(channelID string) *pb.Event_Block {
	return &pb.Event_Block{}
}

func newMockFilteredBlock(channelID string, filteredTx ...*pb.FilteredTransaction) *pb.Event_FilteredBlock {
	return &pb.Event_FilteredBlock{
		FilteredBlock: &pb.FilteredBlock{
			ChannelId:  channelID,
			FilteredTx: filteredTx,
		},
	}
}

func newMockTxEvent(txID string, txValidationCode pb.TxValidationCode) *pb.FilteredTransaction {
	return &pb.FilteredTransaction{
		Txid:             txID,
		TxValidationCode: txValidationCode,
	}
}

func newMockCCEvent(txID, ccID, event string) *pb.FilteredTransaction {
	return &pb.FilteredTransaction{
		Txid: txID,
		CcEvent: &pb.ChaincodeEvent{
			ChaincodeId: ccID,
			EventName:   event,
			TxId:        txID,
		},
	}
}

// func ExampleClient_RegisterBlockEvent() {
// 	var eventClient apifabclient.ChannelEventClient
// 	// eventClient := NeweventClient(...)

// 	eventch := make(chan *apifabclient.BlockEvent)
// 	_, err := eventClient.RegisterBlockEvent(eventch)
// 	if err != nil {
// 		fmt.Printf("error registering for block events: %s\n", err)
// 		return
// 	}

// 	for {
// 		event, ok := <-eventch
// 		if !ok {
// 			// The client has disconnect
// 			return
// 		}
// 		// Handle the event in a separate Go routine so that
// 		// the dispatcher is not blocked from receiving other events.
// 		go func() {
// 			fmt.Printf("Block: %s\n", event.Block)
// 		}()
// 	}
// }

// func ExampleClient_RegisterTxStatusEvent() {
// 	var eventClient apifabclient.ChannelEventClient
// 	// eventClient := NeweventClient(...)

// 	txID := "someTxID"
// 	eventch := make(chan *apifabclient.TxStatusEvent)
// 	registration, err := eventClient.RegisterTxStatusEvent(txID, eventch)
// 	if err != nil {
// 		fmt.Printf("Error registering for TxSTatus event: %s\n", err)
// 		return
// 	}
// 	defer eventClient.Unregister(registration)

// 	event, ok := <-eventch
// 	if !ok {
// 		fmt.Println("The client has disconnected unexpectedly")
// 	} else {
// 		fmt.Printf("Transaction [%s] completed with code [%s]\n", event.TxID, event.TxValidationCode)
// 	}
// }
