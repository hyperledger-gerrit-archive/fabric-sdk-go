// +build channelevents

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"testing"
	"time"

	fab "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/grpc"
)

const (
	peerAddress = "localhost:9999"
	peerURL     = "grpc://" + peerAddress
)

func TestInvalidConnectionOpts(t *testing.T) {
	config := mocks.NewMockConfig()

	fabClient := fab.NewClient(config)
	user := mocks.NewMockUser("admin")
	fabClient.SetUserContext(user)
	fabClient.SetSigningManager(mocks.NewMockSigningManager())

	_, err := newConnection("", fabClient, newPeerConfig(peerURL))
	if err == nil {
		t.Fatalf("expecting error creating new connection without channel but got none")
	}
	_, err = newConnection("channelid", nil, newPeerConfig(peerURL))
	if err == nil {
		t.Fatalf("expecting error creating new connection without fab client but got none")
	}
	_, err = newConnection("channelid", fabClient, nil)
	if err == nil {
		t.Fatalf("expecting error creating new connection without peer config but got none")
	}
	_, err = newConnection("channelid", fabClient, newPeerConfig("grpc://invalidhost:7051"))
	if err == nil {
		t.Fatalf("expecting error creating new connection with invaid address but got none")
	}
}

func TestConnection(t *testing.T) {
	fabClient := fab.NewClient(mocks.NewMockConfig())

	channelID := "mychannel"
	conn, err := newConnection(channelID, fabClient, newPeerConfig(peerURL))
	if err != nil {
		t.Fatalf("error creating new connection: %s", err)
	}

	eventch := make(chan interface{})

	go conn.Receive(eventch)

	time.Sleep(1 * time.Second)

	resp := &pb.ChannelServiceResponse{
		Response: &pb.ChannelServiceResponse_Event{
			Event: &pb.Event{
				Event: &pb.Event_Block{
					Block: &cb.Block{},
				},
			},
		},
	}

	mockChannelServer.send(resp)

	select {
	case e, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed connection")
		}
		fmt.Printf("Received event: %v\n", reflect.TypeOf(e))
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for event")
	}

	conn.Close()

	// Calling close again should be ignored
	conn.Close()
}

func TestSend(t *testing.T) {
	fabClient := fab.NewClient(mocks.NewMockConfig())
	user := mocks.NewMockUser("admin")
	fabClient.SetUserContext(user)
	fabClient.SetSigningManager(mocks.NewMockSigningManager())

	channelID := "mychannel"
	conn, err := newConnection(channelID, fabClient, newPeerConfig(peerURL))
	if err != nil {
		t.Fatalf("error creating new connection: %s", err)
	}

	eventch := make(chan interface{})

	go conn.Receive(eventch)

	time.Sleep(1 * time.Second)

	if err := conn.Send(newRegisterChannelRequest(channelID)); err != nil {
		t.Fatalf("error sending register event for channel [%s]: err", err)
	}

	time.Sleep(1 * time.Second)

	conn.Close()
}

var mockChannelServer *MockChannelServer

func TestMain(m *testing.M) {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	lis, err := net.Listen("tcp", peerAddress)
	if err != nil {
		panic(fmt.Sprintf("Error starting events listener %s", err))
		return
	}

	mockChannelServer = newMockChannelServer()
	pb.RegisterChannelServer(grpcServer, mockChannelServer)

	go grpcServer.Serve(lis)

	time.Sleep(2 * time.Second)
	os.Exit(m.Run())
}

func newRegisterChannelRequest(channelID string) *pb.ChannelServiceRequest {
	return &pb.ChannelServiceRequest{
		Request: &pb.ChannelServiceRequest_RegisterChannel{
			RegisterChannel: &pb.RegisterChannel{
				ChannelIds: []string{channelID},
				// Events:     interestedEvents,
			},
		},
	}
}
