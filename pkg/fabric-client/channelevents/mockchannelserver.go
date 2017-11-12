// +build channelevents

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"fmt"
	"io"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

type MockChannelServer struct {
	stream pb.Channel_ChatServer
}

func newMockChannelServer() *MockChannelServer {
	return new(MockChannelServer)
}

func (cc *MockChannelServer) Chat(stream pb.Channel_ChatServer) error {
	cc.stream = stream

	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
	}
	return nil
}

func (cc *MockChannelServer) send(resp *pb.ChannelServiceResponse) {
	go func() {
		if err := cc.stream.Send(resp); err != nil {
			fmt.Printf("error sending event: %s\n", err)
		}
	}()
}
