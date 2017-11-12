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
	fab "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
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
