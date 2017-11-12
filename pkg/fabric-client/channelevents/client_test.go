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

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

func TestInvalidOptionsInNewClient(t *testing.T) {
	config := mocks.NewMockConfig()

	fabClient := fab.NewClient(config)
	user := mocks.NewMockUser("admin")
	fabClient.SetUserContext(user)
	fabClient.SetSigningManager(mocks.NewMockSigningManager())

	if _, err := NewClient(fabClient, newPeerConfig("grpc://localhost:7051"), ""); err == nil {
		t.Fatalf("expecting error with no channel ID but got none")
	}
	if _, err := NewClient(fabClient, newPeerConfig(""), "channelid"); err == nil {
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

	if _, _, err := eventClient.RegisterFilteredBlockEvent(); err == nil {
		t.Fatalf("expecting error registering for block events on closed channel event client but got none")
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
