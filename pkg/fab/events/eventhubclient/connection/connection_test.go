/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package connection

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	clientdisp "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	eventmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const (
	eventAddressListen = "localhost:0"
)

var eventAddress string
var eventURL string

func TestInvalidConnectionOpts(t *testing.T) {
	_, err := New(newMockContext(), fabmocks.NewMockChannelCfg("channelid"), "grpcs://invalidhost:7053")
	assert.Error(t, err, "expecting error creating new connection with invaid address but got none")
}

func TestConnection(t *testing.T) {
	channelID := "mychannel"
	conn, err := New(newMockContext(), fabmocks.NewMockChannelCfg(channelID), eventURL)
	assert.NoError(t, err, "error creating new connection")

	conn.Close()

	// Calling close again should be ignored
	conn.Close()
}

func TestSend(t *testing.T) {
	channelID := "mychannel"
	conn, err := New(newMockContext(), fabmocks.NewMockChannelCfg(channelID), eventURL)
	assert.NoError(t, err, "error creating new connection")

	eventch := make(chan interface{})

	go conn.Receive(eventch)

	emsg := &pb.Event{
		Event: &pb.Event_Register{
			Register: &pb.Register{
				Events: []*pb.Interest{
					{EventType: pb.EventType_FILTEREDBLOCK},
				},
			},
		},
	}

	t.Log("Sending register event...")
	err = conn.Send(emsg)
	assert.NoError(t, err, "Error sending register interest event")

	select {
	case e, ok := <-eventch:
		assert.True(t, ok, "unexpected closed connection")
		t.Logf("Got response: %#v", e)

		eventHubEvent, ok := e.(*Event)
		assert.True(t, ok, "expected EventHubEvent but got %T", e)

		evt, ok := eventHubEvent.Event.(*pb.Event)
		assert.True(t, ok, "expected Event but got %T", eventHubEvent.Event)

		_, ok = evt.Event.(*pb.Event_Register)
		assert.True(t, ok, "expected register response but got %T", evt.Event)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
	}

	emsg = &pb.Event{
		Event: &pb.Event_Unregister{
			Unregister: &pb.Unregister{
				Events: []*pb.Interest{
					{EventType: pb.EventType_FILTEREDBLOCK},
				},
			},
		},
	}

	t.Log("Sending unregister event...")
	err = conn.Send(emsg)
	assert.NoError(t, err, "Error sending unregister interest event")

	checkEvent(eventch, t)

	conn.Close()
}

func checkEvent(eventch chan interface{}, t *testing.T) {
	select {
	case e, ok := <-eventch:
		assert.True(t, ok, "unexpected closed connection")
		t.Logf("Got response: %#v", e)

		eventHubEvent, ok := e.(*Event)
		assert.True(t, ok, "expected EventHubEvent but got %T", e)

		evt, ok := eventHubEvent.Event.(*pb.Event)
		assert.True(t, ok, "expected Event but got %T", eventHubEvent.Event)

		_, ok = evt.Event.(*pb.Event_Unregister)
		assert.True(t, ok, "expected unregister response but got %T", evt.Event)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestDisconnected(t *testing.T) {
	channelID := "mychannel"
	conn, err := New(newMockContext(), fabmocks.NewMockChannelCfg(channelID), eventURL)
	assert.NoError(t, err, "error creating new connection")

	eventch := make(chan interface{})

	go conn.Receive(eventch)

	emsg := &pb.Event{
		Event: &pb.Event_Register{
			Register: &pb.Register{
				Events: []*pb.Interest{
					{EventType: pb.EventType_FILTEREDBLOCK},
				},
			},
		},
	}

	err = conn.Send(emsg)
	assert.NoError(t, err, "Error sending register interest event")

	ehServer.Disconnect(errors.New("simulating disconnect"))

	select {
	case e, ok := <-eventch:
		assert.True(t, ok, "unexpected closed connection")

		_, ok = e.(*clientdisp.DisconnectedEvent)
		assert.True(t, ok, "expected DisconnectedEvent but got %T", e)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
	}

	conn.Close()
}

var ehServer *eventmocks.MockEventhubServer

func TestMain(m *testing.M) {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	lis, err := net.Listen("tcp", eventAddressListen)
	if err != nil {
		panic(fmt.Sprintf("Error starting events listener %s", err))
	}

	eventAddress = lis.Addr().String()
	eventURL = "grpc://" + eventAddress

	ehServer = eventmocks.NewMockEventhubServer()

	pb.RegisterEventsServer(grpcServer, ehServer)

	go grpcServer.Serve(lis)
	os.Exit(m.Run())
}

func newMockContext() *fabmocks.MockContext {
	context := fabmocks.NewMockContext(mspmocks.NewMockSigningIdentity("user1", "Org1MSP"))
	context.SetCustomInfraProvider(comm.NewMockInfraProvider())
	return context
}
