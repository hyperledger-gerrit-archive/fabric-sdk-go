// +build channelevents

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// connection manages the GRPC connection to the Chat client
type connection struct {
	channelID string
	conn      *grpc.ClientConn
	stream    pb.Channel_ChatClient
	fabclient fab.FabricClient
	done      int32
}

func newConnection(channelID string, fabclient fab.FabricClient, peerConfig *apiconfig.PeerConfig) (*connection, error) {
	if channelID == "" {
		return nil, errors.New("channel ID not specified")
	}
	if fabclient == nil {
		return nil, errors.New("fabric client is nil")
	}
	if peerConfig == nil {
		return nil, errors.New("peer config is nil")
	}

	opts, err := newDialOpts(fabclient.Config(), peerConfig)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(urlutil.ToAddress(peerConfig.URL), opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "could not connect to %s", peerConfig.URL)
	}

	stream, err := pb.NewChannelClient(conn).Chat(context.Background())
	if err != nil {
		if err := conn.Close(); err != nil {
			logger.Warnf("error closing GRPC connection: %s", err)
		}
		return nil, errors.Wrapf(err, "could not create client conn to %s", peerConfig.URL)
	}

	return &connection{
		conn:      conn,
		stream:    stream,
		fabclient: fabclient,
	}, nil
}

func newDialOpts(config apiconfig.Config, peerConfig *apiconfig.PeerConfig) ([]grpc.DialOption, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(config.TimeoutOrDefault(apiconfig.EventHub)))
	if urlutil.IsTLSEnabled(peerConfig.URL) {
		tlsCaCertPool, err := config.TLSCACertPool(peerConfig.TLSCACerts.Path)
		if err != nil {
			return nil, err
		}
		serverHostOverride := ""
		if str, ok := peerConfig.GRPCOptions["ssl-target-name-override"].(string); ok {
			serverHostOverride = str
		}

		logger.Debugf("Creating a secure connection to [%s] with TLS serverHostOverride [%s] and cert [%s]\n", peerConfig.URL, serverHostOverride, peerConfig.TLSCACerts.Path)

		creds := credentials.NewClientTLSFromCert(tlsCaCertPool, serverHostOverride)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		logger.Debugf("Creating an insecure connection [%s]\n", peerConfig.URL)
		opts = append(opts, grpc.WithInsecure())
	}
	return opts, nil
}

func (c *connection) Close() {
	if !c.setClosed() {
		logger.Debugf("Already closed\n")
		return
	}

	logger.Debugf("Closing stream....\n")
	if err := c.stream.CloseSend(); err != nil {
		logger.Warnf("error closing GRPC stream: %s", err)
	}

	logger.Debugf("Closing connection....\n")
	if err := c.conn.Close(); err != nil {
		logger.Warnf("error closing GRPC connection: %s", err)
	}

	c.stream = nil
	c.conn = nil
}

func (c *connection) Send(emsg *pb.ChannelServiceRequest) error {
	panic("not implemented")
}

func (c *connection) Receive(eventch chan<- interface{}) {
	for {
		if c.stream == nil {
			logger.Warnf("The stream has closed. Terminating loop.\n")
			break
		}

		in, err := c.stream.Recv()

		if c.closed() {
			logger.Debugf("The connection has closed. Terminating loop.\n")
			break
		}

		if err == io.EOF {
			// This signifies that the stream has been terminated at the client-side. No need to send an event.
			logger.Debugf("Received EOF from stream.\n")
			break
		}

		eventch <- in.Response
	}
	logger.Debugf("Exiting stream listener\n")
}

func (c *connection) closed() bool {
	return atomic.LoadInt32(&c.done) == 1
}

func (c *connection) setClosed() bool {
	return atomic.CompareAndSwapInt32(&c.done, 0, 1)
}
