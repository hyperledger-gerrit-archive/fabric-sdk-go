// +build channelevents

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"context"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// connection manages the GRPC connection to the Chat client
type connection struct {
	conn      *grpc.ClientConn
	stream    pb.Channel_ChatClient
	fabclient fab.FabricClient
}

func newConnection(fabclient fab.FabricClient, peerConfig apiconfig.PeerConfig) (*connection, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(fabclient.Config().TimeoutOrDefault(apiconfig.EventHub)))
	if fabclient.Config().IsTLSEnabled() {
		tlsCaCertPool, err := fabclient.Config().TLSCACertPool(peerConfig.TLSCACerts.Path)
		if err != nil {
			return nil, err
		}
		serverHostOverride := ""
		if str, ok := peerConfig.GRPCOptions["ssl-target-name-override"].(string); ok {
			serverHostOverride = str
		}
		creds := credentials.NewClientTLSFromCert(tlsCaCertPool, serverHostOverride)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(peerConfig.URL, opts...)

	stream, err := pb.NewChannelClient(conn).Chat(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not create client conn to %s:%s", peerConfig.URL, err)
	}

	return &connection{conn: conn, stream: stream, fabclient: fabclient}, nil
}

func (c *connection) close() error {
	logger.Debugf("Closing stream....\n")
	err := c.stream.CloseSend()
	return err
}

func (c *connection) send(emsg *pb.Event) error {
	user, err := c.fabclient.LoadUserFromStateStore("")
	if err != nil {
		return fmt.Errorf("LoadUserFromStateStore returned error: %s", err)
	}
	payload, err := proto.Marshal(emsg)
	if err != nil {
		return fmt.Errorf("Error marshaling message: %s", err)
	}

	signingMgr := c.fabclient.SigningManager()
	if signingMgr == nil {
		return fmt.Errorf("Signing Manager is nil")
	}

	signature, err := signingMgr.Sign(payload, user.PrivateKey())
	if err != nil {
		return fmt.Errorf("Error signing message: %s", err)
	}
	signedEvt := &pb.SignedEvent{EventBytes: payload, Signature: signature}

	return c.stream.Send(signedEvt)
}

func (c *connection) receive(events chan<- interface{}) {
	for {
		in, err := c.stream.Recv()
		if err == io.EOF {
			// This signifies that the stream has been terminated at the client-side. No need to send an event.
			logger.Debugf("Received EOF from stream.\n")
			return
		}

		if err != nil {
			logger.Debugf("Received error from stream: [%s]. Sending disconnected event.\n", err)
			events <- &disconnectedEvent{err: err}
			return
		}

		events <- in.Event
	}
}
