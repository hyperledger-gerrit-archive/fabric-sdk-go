// +build channelevents

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	fab "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

type Outcome string
type State int32
type NumBlockEvents uint
type NumCCEvents uint

type EventsReceived struct {
	BlockEvents NumBlockEvents
	CCEvents    NumCCEvents
}

const (
	initialState State = -1

	reconnectedOutcome Outcome = "reconnected"
	terminatedOutcome  Outcome = "terminated"
	timedOutOutcome    Outcome = "timeout"
	connectedOutcome   Outcome = "connected"
	errorOutcome       Outcome = "error"

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

func TestInvalidOptionsInNewClient(t *testing.T) {
	config := mocks.NewMockConfig()

	fabClient := fab.NewClient(config)
	user := mocks.NewMockUser("admin")
	fabClient.SetUserContext(user)
	fabClient.SetSigningManager(mocks.NewMockSigningManager())

	// Client
	if _, err := NewClient(fabClient, newPeerConfig("grpc://localhost:7051"), ""); err == nil {
		t.Fatalf("expecting error with no channel ID but got none")
	}
	if _, err := NewClient(fabClient, newPeerConfig(""), "channelid"); err == nil {
		t.Fatalf("expecting error with no peer URL but got none")
	}

	// Admin client
	if _, err := NewAdminClient(fabClient, newPeerConfig("grpc://localhost:7051"), ""); err == nil {
		t.Fatalf("expecting error with no channel ID but got none")
	}
	if _, err := NewAdminClient(fabClient, newPeerConfig(""), "channelid"); err == nil {
		t.Fatalf("expecting error with no peer URL but got none")
	}
}

func TestClientConnect(t *testing.T) {
	opts := newMockClientOpts()
	opts.connectionProvider = NewMockConnectionProviderFactory().Provider(NewMockConnection())

	eventClient, err := newClientWithMockConnAndOpts("mychannel", "grpc://localhost:7051", "admin", opts)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if eventClient.ConnectionState() != disconnected {
		t.Fatalf("expecting connection state %s but got %s", disconnected, eventClient.ConnectionState())
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting: %s", err)
	}
	time.Sleep(500 * time.Millisecond)
	if eventClient.ConnectionState() != connected {
		t.Fatalf("expecting connection state %s but got %s", connected, eventClient.ConnectionState())
	}
	eventClient.Disconnect()
	if eventClient.ConnectionState() != disconnected {
		t.Fatalf("expecting connection state %s but got %s", disconnected, eventClient.ConnectionState())
	}
}

func TestFailedChannelRegistration(t *testing.T) {
	channelID := "mychannel"
	errMsg := "mock failed channel registration"
	opts := newMockClientOpts()
	opts.connectionProvider = NewMockConnectionProviderFactory().Provider(
		NewMockConnection(
			NewResultsOpt(
				NewResult(RegisterChannel, FailResult, errMsg),
			),
		),
	)

	eventClient, err := newClientWithMockConnAndOpts(channelID, "grpc://localhost:7051", "admin", opts)
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
	opts := newMockClientOpts()
	opts.connectionProvider = NewMockConnectionProviderFactory().Provider(
		NewMockConnection(
			NewResultsOpt(
				NewResult(RegisterChannel, InvalidChannelResult),
			),
		),
	)

	eventClient, err := newClientWithMockConnAndOpts(channelID, "grpc://localhost:7051", "admin", opts)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err == nil {
		t.Fatalf("expecting error connecting channel event client but got none")
	}
}

func TestTimedOutChannelRegistration(t *testing.T) {
	channelID := "mychannel"
	opts := newMockClientOpts()
	opts.connectionProvider = NewMockConnectionProviderFactory().Provider(
		NewMockConnection(
			NewResultsOpt(
				NewResult(RegisterChannel, NoOpResult),
			),
		),
	)
	eventClient, err := newClientWithMockConnAndOpts(channelID, "grpc://localhost:7051", "admin", opts)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err == nil {
		t.Fatalf("expecting error connecting channel event client due to no response from channel registration but got none")
	}
}

func TestTimedOutChannelUnregistration(t *testing.T) {
	channelID := "mychannel"
	opts := newMockClientOpts()
	opts.connectionProvider = NewMockConnectionProviderFactory().Provider(
		NewMockConnection(
			NewResultsOpt(
				NewResult(UnregisterChannel, NoOpResult),
			),
		),
	)
	eventClient, err := newClientWithMockConnAndOpts(channelID, "grpc://localhost:7051", "admin", opts)
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
	eventClient, _, err := newClientWithMockConn("mychannel", "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt(FILTEREDBLOCKEVENT))
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

	if _, _, err := eventClient.registerConnectionEvent(); err == nil {
		t.Fatalf("expecting error registering for connection events on closed channel event client but got none")
	}

	if _, _, err := eventClient.RegisterFilteredBlockEvent(); err == nil {
		t.Fatalf("expecting error registering for block events on closed channel event client but got none")
	}

	if _, _, err := eventClient.RegisterChaincodeEvent("ccid", "event"); err == nil {
		t.Fatalf("expecting error registering for chaincode events on closed channel event client but got none")
	}

	if _, _, err := eventClient.RegisterTxStatusEvent("txid"); err == nil {
		t.Fatalf("expecting error registering for TX events on closed channel event client but got none")
	}

	// Make sure the client doesn't panic when calling unregister on disconnected client
	eventClient.Unregister(nil)
}

func TestInvalidUnregister(t *testing.T) {
	channelID := "mychannel"
	eventClient, _, err := newClientWithMockConn(channelID, "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt(BLOCKEVENT))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	// Make sure the client doesn't panic with invalid registration
	eventClient.Unregister("invalid registration")
}

func TestBlockEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newAdminClient(channelID, "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt(BLOCKEVENT))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	registration, eventch, err := eventClient.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	_, eventch, err = eventClient.RegisterBlockEvent()
	if err == nil {
		t.Fatalf("expecting error registering multiple times for block events: %s", err)
	}
	eventClient.Unregister(registration)

	registration, eventch, err = eventClient.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	defer eventClient.Unregister(registration)

	conn.ProduceEvent(newMockBlock(channelID))

	select {
	case _, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed channel")
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for block event")
	}
}

func TestBlockEventsUnauthorized(t *testing.T) {
	eventClient, _, err := newAdminClient("mychannel", "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt(FILTEREDBLOCKEVENT))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	if _, _, err := eventClient.RegisterBlockEvent(); err == nil {
		t.Fatalf("expecting authorization error since client is not authorized to receive block events")
	}
}

func TestFilteredBlockEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newClientWithMockConn(channelID, "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt(FILTEREDBLOCKEVENT))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	registration, _, err := eventClient.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("error registering for filtered block events: %s", err)
	}
	_, _, err = eventClient.RegisterFilteredBlockEvent()
	if err == nil {
		t.Fatalf("expecting error registering multiple times for filtered block events: %s", err)
	}
	eventClient.Unregister(registration)

	registration, eventch, err := eventClient.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("error registering for filtered block events: %s", err)
	}
	defer eventClient.Unregister(registration)

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	conn.ProduceEvent(newMockFilteredBlock(
		channelID,
		newMockTxEvent(txID1, txCode1),
		newMockTxEvent(txID2, txCode2),
	))

	select {
	case fbevent, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed channel")
		}
		if fbevent.FilteredBlock == nil {
			t.Fatalf("Expecting filtered block but got nil")
		}
		if fbevent.FilteredBlock.ChannelId != channelID {
			t.Fatalf("Expecting channel [%s] but got [%s]", channelID, fbevent.FilteredBlock.ChannelId)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for filtered block event")
	}
}

func TestFilteredBlockEventsUnauthorized(t *testing.T) {
	eventClient, _, err := newClientWithMockConn("mychannel", "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt())
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	if _, _, err := eventClient.RegisterFilteredBlockEvent(); err == nil {
		t.Fatalf("expecting authorization error since client is not authorized to receive filtered block events")
	}
}

func TestBlockAndFilteredBlockEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newAdminClient(channelID, "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt(BLOCKEVENT, FILTEREDBLOCKEVENT))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	// First register for filtered block events
	fbreg, fbeventch, err := eventClient.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("error registering for filtered block events: %s", err)
	}
	defer eventClient.Unregister(fbreg)

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	conn.ProduceEvent(newMockFilteredBlock(
		channelID,
		newMockTxEvent(txID1, txCode1),
		newMockTxEvent(txID2, txCode2),
	))

	select {
	case fbevent, ok := <-fbeventch:
		if !ok {
			t.Fatalf("unexpected closed channel")
		}
		if fbevent.FilteredBlock == nil {
			t.Fatalf("Expecting filtered block but got nil")
		}
		if fbevent.FilteredBlock.ChannelId != channelID {
			t.Fatalf("Expecting channel [%s] but got [%s]", channelID, fbevent.FilteredBlock.ChannelId)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for filtered block event")
	}

	// Now register for block events
	breg, beventch, err := eventClient.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	defer eventClient.Unregister(breg)

	conn.ProduceEvent(newMockBlock(channelID))
	conn.ProduceEvent(newMockFilteredBlock(
		channelID,
		newMockTxEvent(txID1, txCode1),
		newMockTxEvent(txID2, txCode2),
	))

	numEventsReceived := 0
	expectedEvents := 2

	for {
		select {
		case _, ok := <-fbeventch:
			if !ok {
				t.Fatalf("unexpected closed channel")
			}
			numEventsReceived++
		case _, ok := <-beventch:
			if !ok {
				t.Fatalf("unexpected closed channel")
			}
			numEventsReceived++
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for events")
		}
		if numEventsReceived == expectedEvents {
			break
		}
	}
}

func TestTxStatusEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newClientWithMockConn(channelID, "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt(FILTEREDBLOCKEVENT))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	if _, _, err := eventClient.RegisterTxStatusEvent(""); err == nil {
		t.Fatalf("expecting error registering for TxStatus event without a TX ID but got none")
	}
	reg1, _, err := eventClient.RegisterTxStatusEvent(txID1)
	if err != nil {
		t.Fatalf("error registering for TxStatus events: %s", err)
	}
	_, _, err = eventClient.RegisterTxStatusEvent(txID1)
	if err == nil {
		t.Fatalf("expecting error registering multiple times for TxStatus events: %s", err)
	}
	eventClient.Unregister(reg1)

	reg1, eventch1, err := eventClient.RegisterTxStatusEvent(txID1)
	if err != nil {
		t.Fatalf("error registering for TxStatus events: %s", err)
	}
	defer eventClient.Unregister(reg1)

	reg2, eventch2, err := eventClient.RegisterTxStatusEvent(txID2)
	if err != nil {
		t.Fatalf("error registering for TxStatus events: %s", err)
	}
	defer eventClient.Unregister(reg2)

	conn.ProduceEvent(
		newMockFilteredBlock(
			channelID,
			newMockTxEvent(txID1, txCode1),
			newMockTxEvent(txID2, txCode2),
		),
	)

	numExpected := 2
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
			t.Fatalf("timed out waiting for [%d] TxStatus events. Only received [%d]", numExpected, numReceived)
		}

		if numReceived == numExpected {
			break
		}
	}
}

func TestTxStatusEventsUnauthorized(t *testing.T) {
	eventClient, _, err := newClientWithMockConn("mychannel", "grpc://localhost:7051", "admin")
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	if _, _, err := eventClient.RegisterTxStatusEvent("txid"); err == nil {
		t.Fatalf("expecting authorization error since client is not authorized to receive filtered events")
	}
}

func TestCCEvents(t *testing.T) {
	channelID := "mychannel"
	eventClient, conn, err := newClientWithMockConn(channelID, "grpc://localhost:7051", "admin", NewAuthorizedEventsOpt(FILTEREDBLOCKEVENT))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	ccID1 := "mycc1"
	ccID2 := "mycc2"
	ccFilter1 := "event1"
	ccFilter2 := "event.*"
	event1 := "event1"
	event2 := "event2"
	event3 := "event3"

	if _, _, err := eventClient.RegisterChaincodeEvent("", ccFilter1); err == nil {
		t.Fatalf("expecting error registering for chaincode events without CC ID but got none")
	}
	if _, _, err := eventClient.RegisterChaincodeEvent(ccID1, ""); err == nil {
		t.Fatalf("expecting error registering for chaincode events without event filter but got none")
	}
	if _, _, err := eventClient.RegisterChaincodeEvent(ccID1, ".(xxx"); err == nil {
		t.Fatalf("expecting error registering for chaincode events with invalid (regular expression) event filter but got none")
	}
	reg1, _, err := eventClient.RegisterChaincodeEvent(ccID1, ccFilter1)
	if err != nil {
		t.Fatalf("error registering for chaincode events: %s", err)
	}
	_, _, err = eventClient.RegisterChaincodeEvent(ccID1, ccFilter1)
	if err == nil {
		t.Fatalf("expecting error registering multiple times for chaincode events: %s", err)
	}
	eventClient.Unregister(reg1)

	reg1, eventch1, err := eventClient.RegisterChaincodeEvent(ccID1, ccFilter1)
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	defer eventClient.Unregister(reg1)

	reg2, eventch2, err := eventClient.RegisterChaincodeEvent(ccID2, ccFilter2)
	if err != nil {
		t.Fatalf("error registering for chaincode events: %s", err)
	}
	defer eventClient.Unregister(reg2)

	conn.ProduceEvent(
		newMockFilteredBlock(
			channelID,
			newMockCCEvent("txid1", ccID1, event1),
			newMockCCEvent("txid2", ccID2, event2),
			newMockCCEvent("txid3", ccID2, event3),
		),
	)

	numExpected := 3
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
			t.Fatalf("timed out waiting for [%d] CC events. Only received [%d]", numExpected, numReceived)
		}

		if numReceived == numExpected {
			break
		}
	}
}

func TestCCEventsUnauthorized(t *testing.T) {
	eventClient, _, err := newClientWithMockConn("mychannel", "grpc://localhost:7051", "admin")
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	if _, _, err := eventClient.RegisterChaincodeEvent("ccid", ".*"); err == nil {
		t.Fatalf("expecting authorization error since client is not authorized to receive filtered events")
	}
}

// TestReconnect tests the ability of the Channel Event Client to retry multiple
// times to connect, and reconnect after it has disconnected.
func TestReconnect(t *testing.T) {
	// Test Connect with invalid user
	eventClient, _, err := newClientWithMockConn("mychannel", "grpc://localhost:7051", "")
	if err != nil {
		t.Fatalf("error creating new channel event client: %s", err)
	}
	if err := eventClient.Connect(); err == nil {
		t.Fatalf("expecting error with no user but got none")
	}

	// (1) Connect
	//     -> should fail to connect on the first and second attempt but succeed on the third attempt
	t.Run("#1", func(t *testing.T) {
		t.Parallel()
		testConnect(t, 3, connectedOutcome,
			NewConnectResults(
				NewConnectResult(thirdAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
			),
		)
	})

	// (1) Connect
	//     -> should fail to connect on the first attempt and no further attempts are to be made
	t.Run("#2", func(t *testing.T) {
		t.Parallel()
		testConnect(t, 1, errorOutcome,
			NewConnectResults(),
		)
	})

	// (1) Connect
	//     -> should succeed to connect on the first attempt
	// (2) Disconnect
	//     -> should fail to reconnect on the first and second attempt but succeed on the third attempt
	t.Run("#3", func(t *testing.T) {
		t.Parallel()
		testReconnect(t, true, 3, reconnectedOutcome,
			NewConnectResults(
				NewConnectResult(firstAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
				NewConnectResult(fourthAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
			),
		)
	})

	// (1) Connect
	//     -> should succeed to connect on the first attempt
	// (2) Disconnect
	//     -> should fail to reconnect after two attempts and then cleanly disconnect
	t.Run("#4", func(t *testing.T) {
		t.Parallel()
		testReconnect(t, true, 2, terminatedOutcome,
			NewConnectResults(
				NewConnectResult(firstAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
			),
		)
	})

	// (1) Connect
	//     -> should succeed to connect on the first attempt
	// (2) Disconnect
	//     -> should fail and not attempt to reconnect and then cleanly disconnect
	t.Run("#5", func(t *testing.T) {
		t.Parallel()
		testReconnect(t, false, 0, terminatedOutcome,
			NewConnectResults(
				NewConnectResult(firstAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
			),
		)
	})
}

// TestReconnectRegistration tests the ability of the Channel Event Client to
// re-establish the existing registrations after reconnecting.
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
	t.Run("#1", func(t *testing.T) {
		t.Parallel()
		testReconnectRegistration(
			t, expectTwoBlockEvents, expectTwoCCEvents,
			NewConnectResults(
				NewConnectResult(firstAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
				NewConnectResult(secondAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT)),
		)
	})

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
	t.Run("#2", func(t *testing.T) {
		t.Parallel()
		testReconnectRegistration(
			t, expectOneBlockEvent, expectTwoCCEvents,
			NewConnectResults(
				NewConnectResult(firstAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
				NewConnectResult(secondAttempt, FILTEREDBLOCKEVENT)),
		)
	})

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
	t.Run("#3", func(t *testing.T) {
		t.Parallel()
		testReconnectRegistration(
			t, expectTwoBlockEvents, expectOneCCEvent,
			NewConnectResults(
				NewConnectResult(firstAttempt, BLOCKEVENT, FILTEREDBLOCKEVENT),
				NewConnectResult(secondAttempt, BLOCKEVENT)),
		)
	})
}

// TestConcurrentEvents ensures that the channel event client is thread-safe
func TestConcurrentEvents(t *testing.T) {
	numEvents := 1000
	channelID := "mychannel"
	opts := newMockClientOpts()
	opts.EventQueueSize = numEvents * 4
	eventClient, conn, err := newAdminClientWithOpts(channelID, "grpc://localhost:7051", "admin", opts, NewAuthorizedEventsOpt(BLOCKEVENT, FILTEREDBLOCKEVENT))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}

	t.Run("Block Events", func(t *testing.T) {
		t.Parallel()
		if err := testConcurrentBlockEvents(channelID, numEvents, eventClient, conn); err != nil {
			t.Fatalf("error in testConcurrentBlockEvents: %s", err)
		}
	})
	t.Run("Filtered Block Events", func(t *testing.T) {
		t.Parallel()
		if err := testConcurrentFilteredBlockEvents(channelID, numEvents, eventClient, conn); err != nil {
			t.Fatalf("error in testConcurrentBlockEvents: %s", err)
		}
	})
	t.Run("Chaincode Events", func(t *testing.T) {
		t.Parallel()
		if err := testConcurrentCCEvents(channelID, numEvents, eventClient, conn); err != nil {
			t.Fatalf("error in testConcurrentBlockEvents: %s", err)
		}
	})
	t.Run("Tx Status Events", func(t *testing.T) {
		t.Parallel()
		if err := testConcurrentTxStatusEvents(channelID, numEvents, eventClient, conn); err != nil {
			t.Fatalf("error in testConcurrentBlockEvents: %s", err)
		}
	})
}

func testConnect(t *testing.T, maxConnectAttempts uint, expectedOutcome Outcome, connAttemptResult ConnectAttemptResults) {
	cp := NewMockConnectionProviderFactory()

	opts := DefaultClientOpts()
	opts.ResponseTimeout = 1 * time.Second
	opts.MaxConnectAttempts = maxConnectAttempts
	opts.connectionProvider = cp.FlakeyProvider(connAttemptResult)

	eventClient, err := newClientWithMockConnAndOpts("mychannel", "grpc://localhost:7051", "admin", opts)
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

	opts := DefaultClientOpts()
	opts.ResponseTimeout = 3 * time.Second
	opts.Reconnect = reconnect
	opts.MaxConnectAttempts = 1
	opts.MaxReconnectAttempts = maxReconnectAttempts
	opts.TimeBetweenConnectAttempts = time.Millisecond
	opts.ConnectEventCh = connectch
	opts.connectionProvider = cp.FlakeyProvider(connAttemptResult)

	eventClient, err := newClientWithMockConnAndOpts("mychannel", "grpc://localhost:7051", "admin", opts)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

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
}

// testReconnectRegistration tests the scenario when an events client is registered to receive events and the connection to the
// event service is lost. After the connection is re-established, events should once again be received without the caller having to
// re-register for those events.
func testReconnectRegistration(t *testing.T, expectedBlockEvents NumBlockEvents, expectedCCEvents NumCCEvents, connectResults ConnectAttemptResults) {
	channelID := "mychannel"
	ccID := "mycc"

	cp := NewMockConnectionProviderFactory()

	opts := DefaultClientOpts()
	opts.ResponseTimeout = 3 * time.Second
	opts.MaxConnectAttempts = 1
	opts.MaxReconnectAttempts = 1
	opts.TimeBetweenConnectAttempts = time.Millisecond
	opts.connectionProvider = cp.FlakeyProvider(connectResults)

	eventClient, _, err := newAdminClientWithOpts(channelID, "grpc://localhost:7051", "admin", opts)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	if err := eventClient.Connect(); err != nil {
		t.Fatalf("error connecting channel event client: %s", err)
	}
	defer eventClient.Disconnect()

	_, blockch, err := eventClient.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}

	_, ccch, err := eventClient.RegisterChaincodeEvent(ccID, ".*")
	if err != nil {
		t.Fatalf("error registering for chaincode events: %s", err)
	}

	numCh := make(chan EventsReceived)
	go listenEvents(blockch, ccch, 3*time.Second, numCh, expectedBlockEvents, expectedCCEvents)

	time.Sleep(500 * time.Millisecond)

	// Produce a block event
	cp.Connection().ProduceEvent(newMockBlock(channelID))

	// Produce a chaincode event
	cp.Connection().ProduceEvent(
		newMockFilteredBlock(
			channelID,
			newMockCCEvent("tx1", ccID, "event1"),
		),
	)

	// Wait a while for the subscriber to receive the event
	time.Sleep(500 * time.Millisecond)

	// Simulate a connection error
	cp.Connection().Disconnect(errors.New("testing reconnect handling"))

	time.Sleep(time.Second)

	// Produce one more block and CC event
	cp.Connection().ProduceEvent(newMockBlock(channelID))
	cp.Connection().ProduceEvent(
		newMockFilteredBlock(
			channelID,
			newMockCCEvent("tx2", ccID, "event2"),
		),
	)

	var eventsReceived EventsReceived

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
}

func testConcurrentBlockEvents(channelID string, numEvents int, eventClient apifabclient.ChannelEventAdminClient, conn *MockConnection) error {
	registration, eventch, err := eventClient.RegisterBlockEvent()
	if err != nil {
		return errors.Errorf("error registering for block events: %s", err)
	}

	go func() {
		for i := 0; i < numEvents+10; i++ {
			conn.ProduceEvent(newMockBlock(channelID))
		}
	}()

	numReceived := 0
	done := false

	for !done {
		select {
		case _, ok := <-eventch:
			if !ok {
				fmt.Printf("Block events channel was closed \n")
				done = true
			} else {
				numReceived++
				if numReceived == numEvents {
					// Unregister will close the event channel
					// and done will be set to true
					eventClient.Unregister(registration)
				}
			}
		case <-time.After(5 * time.Second):
			if numReceived < numEvents {
				return errors.Errorf("Expected [%d] events but received [%d]", numEvents, numReceived)
			}
		}
	}

	return nil
}

func testConcurrentFilteredBlockEvents(channelID string, numEvents int, eventClient apifabclient.ChannelEventClient, conn *MockConnection) error {
	registration, eventch, err := eventClient.RegisterFilteredBlockEvent()
	if err != nil {
		return errors.Errorf("error registering for filtered block events: %s", err)
	}
	defer eventClient.Unregister(registration)

	for i := 0; i < numEvents; i++ {
		txID := fmt.Sprintf("txid_fb_%d", i)
		conn.ProduceEvent(newMockFilteredBlock(
			channelID,
			newMockTxEvent(txID, pb.TxValidationCode_VALID),
		))
	}

	numReceived := 0
	done := false

	for !done {
		select {
		case fbevent, ok := <-eventch:
			if !ok {
				fmt.Printf("Filtered block events channel was closed \n")
				done = true
			} else {
				if fbevent.FilteredBlock == nil {
					return errors.New("Expecting filtered block but got nil")
				}
				if fbevent.FilteredBlock.ChannelId != channelID {
					return errors.Errorf("Expecting channel [%s] but got [%s]", channelID, fbevent.FilteredBlock.ChannelId)
				}
				numReceived++
				if numReceived == numEvents {
					// Unregister will close the event channel and done will be set to true
					return nil
					// eventClient.Unregister(registration)
				}
			}
		case <-time.After(5 * time.Second):
			if numReceived < numEvents {
				return errors.Errorf("Expected [%d] events but received [%d]", numEvents, numReceived)
			}
		}
	}

	return nil
}

func testConcurrentCCEvents(channelID string, numEvents int, eventClient apifabclient.ChannelEventClient, conn *MockConnection) error {
	ccID := "mycc1"
	ccFilter := "event.*"
	event1 := "event1"

	reg, eventch, err := eventClient.RegisterChaincodeEvent(ccID, ccFilter)
	if err != nil {
		return errors.New("error registering for chaincode events")
	}

	for i := 0; i < numEvents+10; i++ {
		txID := fmt.Sprintf("txid_cc_%d", i)
		conn.ProduceEvent(
			newMockFilteredBlock(
				channelID,
				newMockCCEvent(txID, ccID, event1),
			),
		)
	}

	numReceived := 0
	done := false
	for !done {
		select {
		case _, ok := <-eventch:
			if !ok {
				fmt.Printf("CC events channel was closed \n")
				done = true
			} else {
				numReceived++
			}
		case <-time.After(5 * time.Second):
			if numReceived < numEvents {
				return errors.Errorf("timed out waiting for [%d] CC events but received [%d]", numEvents, numReceived)
			}
		}

		if numReceived == numEvents {
			// Unregister will close the event channel and done will be set to true
			eventClient.Unregister(reg)
		}
	}

	return nil
}

func testConcurrentTxStatusEvents(channelID string, numEvents int, eventClient apifabclient.ChannelEventClient, conn *MockConnection) error {
	var wg sync.WaitGroup
	wg.Add(numEvents)

	var errs []error
	var mutex sync.Mutex

	var receivedEvents int
	for i := 0; i < numEvents; i++ {
		txID := fmt.Sprintf("txid_tx_%d", i)
		go func() {
			defer wg.Done()

			reg, eventch, err := eventClient.RegisterTxStatusEvent(txID)
			if err != nil {
				mutex.Lock()
				errs = append(errs, errors.New("Error registering for TxStatus event"))
				mutex.Unlock()
				return
			}
			defer eventClient.Unregister(reg)

			conn.ProduceEvent(
				newMockFilteredBlock(
					channelID,
					newMockTxEvent(txID, pb.TxValidationCode_VALID),
				),
			)

			select {
			case _, ok := <-eventch:
				mutex.Lock()
				if !ok {
					errs = append(errs, errors.New("unexpected closed channel"))
				} else {
					receivedEvents++
				}
				mutex.Unlock()
			case <-time.After(5 * time.Second):
				mutex.Lock()
				errs = append(errs, errors.New("timed out waiting for TxStatus event"))
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(errs) > 0 {
		return errors.Errorf("Received %d events and %d errors. First error %s\n", receivedEvents, len(errs), errs[0])
	}
	return nil
}

func listenConnection(eventch chan *apifabclient.ConnectionEvent, outcome chan Outcome) {
	state := initialState

	for {
		e, ok := <-eventch
		fmt.Printf("listenConnection - got event [%v] - ok=[%v]\n", e, ok)
		if !ok {
			fmt.Printf("listenConnection - Returning terminated outcome\n")
			outcome <- terminatedOutcome
			break
		}
		if e.Connected {
			if state == State(disconnected) {
				fmt.Printf("listenConnection - Returning reconnected outcome\n")
				outcome <- reconnectedOutcome
			}
			state = State(connected)
		} else {
			state = State(disconnected)
		}
	}
}

func listenEvents(blockch <-chan *apifabclient.BlockEvent, ccch <-chan *apifabclient.CCEvent, waitDuration time.Duration, numEventsCh chan EventsReceived, expectedBlockEvents NumBlockEvents, expectedCCEvents NumCCEvents) {
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

func newClientWithMockConn(channelID string, peerURL string, userName string, connOpts ...MockConnOpt) (*Client, *MockConnection, error) {
	conn := NewMockConnection(connOpts...)

	opts := newMockClientOpts()
	opts.connectionProvider = NewMockConnectionProviderFactory().Provider(conn)

	client, err := newClientWithMockConnAndOpts(channelID, peerURL, userName, opts)
	return client, conn, err
}

func newClientWithMockConnAndOpts(channelID string, peerURL string, userName string, opts *ClientOpts) (*Client, error) {
	fabClient := fab.NewClient(mocks.NewMockConfig())

	if userName != "" {
		user := mocks.NewMockUser(userName)
		fabClient.SetUserContext(user)
	}

	fabClient.SetSigningManager(mocks.NewMockSigningManager())

	return NewClientWithOpts(fabClient, newPeerConfig(peerURL), channelID, opts)
}

func newAdminClient(channelID string, peerURL string, userName string, connOpts ...MockConnOpt) (*AdminClient, *MockConnection, error) {
	return newAdminClientWithOpts(channelID, peerURL, userName, newMockClientOpts(), connOpts...)
}

func newAdminClientWithOpts(channelID string, peerURL string, userName string, opts *ClientOpts, connOpts ...MockConnOpt) (*AdminClient, *MockConnection, error) {
	fabClient := fab.NewClient(mocks.NewMockConfig())

	if userName != "" {
		user := mocks.NewMockUser(userName)
		fabClient.SetUserContext(user)
	}

	fabClient.SetSigningManager(mocks.NewMockSigningManager())

	var conn *MockConnection
	if opts.connectionProvider == nil {
		conn = NewMockConnection(connOpts...)
		opts.connectionProvider = NewMockConnectionProviderFactory().Provider(conn)
	}

	client, err := NewAdminClientWithOpts(fabClient, newPeerConfig(peerURL), channelID, opts)
	return client, conn, err
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

func newMockBlock(channelID string) *pb.ChannelServiceResponse_Event {
	return &pb.ChannelServiceResponse_Event{
		Event: &pb.Event{
			ChannelId: channelID,
			Creator:   []byte("some-id"),
			Timestamp: &timestamp.Timestamp{Seconds: 1000},
			Event: &pb.Event_Block{
				Block: &cb.Block{},
			},
		},
	}
}

func newMockFilteredBlock(channelID string, filteredTx ...*pb.FilteredTransaction) *pb.ChannelServiceResponse_Event {
	return &pb.ChannelServiceResponse_Event{
		Event: &pb.Event{
			ChannelId: channelID,
			Creator:   []byte("some-id"),
			Timestamp: &timestamp.Timestamp{Seconds: 1000},
			Event: &pb.Event_FilteredBlock{
				FilteredBlock: &pb.FilteredBlock{
					ChannelId:  channelID,
					FilteredTx: filteredTx,
				},
			},
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
		FilteredAction: []*pb.FilteredAction{
			&pb.FilteredAction{
				CcEvent: &pb.ChaincodeEvent{
					ChaincodeId: ccID,
					EventName:   event,
					TxId:        txID,
				},
			},
		},
	}
}

func newPeerConfig(peerURL string) *apiconfig.PeerConfig {
	return &apiconfig.PeerConfig{
		URL:         peerURL,
		GRPCOptions: make(map[string]interface{}),
	}
}

func newMockClientOpts() *ClientOpts {
	opts := DefaultClientOpts()
	opts.connectionProvider = nil
	return opts
}
