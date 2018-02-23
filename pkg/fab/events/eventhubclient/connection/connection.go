/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package connection

import (
	"context"
	"fmt"
	"io"
	"time"

	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/context"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	comm "github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	conn "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	clientdisp "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	logging "github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/options"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// Connection defines the functions for an event hub connection
type Connection interface {
	conn.Connection
	Send(emsg *pb.Event) error
}

type connection struct {
	comm.GRPCConnection
}

// New returns a new Connection to the event hub.
func New(ctx fabcontext.Context, channelID string, url string, opts ...options.Opt) (Connection, error) {
	if channelID == "" {
		return nil, errors.New("channel ID not provided")
	}

	connect, err := comm.NewConnection(
		ctx, channelID,
		func(grpcconn *grpc.ClientConn) (grpc.ClientStream, error) {
			return pb.NewEventsClient(grpcconn).Chat(context.Background())
		},
		url, opts...,
	)
	if err != nil {
		return nil, err
	}

	return &connection{
		GRPCConnection: *connect,
	}, nil
}

func (c *connection) EventHubStream() pb.Events_ChatClient {
	if c.Stream() == nil {
		return nil
	}
	stream, ok := c.Stream().(pb.Events_ChatClient)
	if !ok {
		panic(fmt.Sprintf("invalid events chat client type %T", c.Stream()))
	}
	return stream
}

func (c *connection) Send(emsg *pb.Event) error {
	creator, err := c.Context().Identity()
	if err != nil {
		return errors.WithMessage(err, "error getting creator identity")
	}

	timestamp, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return errors.Wrap(err, "failed to create timestamp")
	}

	event := *emsg
	event.Creator = creator
	event.Timestamp = timestamp
	event.TlsCertHash = c.TLSCertHash()

	evtBytes, err := proto.Marshal(&event)
	if err != nil {
		return err
	}

	signature, err := c.Context().SigningManager().Sign(evtBytes, c.Context().PrivateKey())
	if err != nil {
		return err
	}

	return c.EventHubStream().Send(&pb.SignedEvent{
		EventBytes: evtBytes,
		Signature:  signature,
	})
}

func (c *connection) Receive(eventch chan<- interface{}) {
	for {
		logger.Debugf("Listening for events...")
		if c.EventHubStream() == nil {
			logger.Warnf("The stream has closed. Terminating loop.")
			break
		}

		in, err := c.EventHubStream().Recv()

		if c.Closed() {
			logger.Debugf("The connection has closed. Terminating loop.")
			break
		}

		if err == io.EOF {
			// This signifies that the stream has been terminated at the client-side. No need to send an event.
			logger.Debugf("Received EOF from stream.")
			break
		}

		if err != nil {
			logger.Errorf("Received error from stream: [%s]. Sending disconnected event.", err)
			eventch <- clientdisp.NewDisconnectedEvent(err)
			break
		}

		logger.Debugf("Got event %#v", in)
		eventch <- in
	}
	logger.Debugf("Exiting stream listener")
}
