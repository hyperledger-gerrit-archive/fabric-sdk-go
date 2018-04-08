/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"google.golang.org/grpc"
)

// StreamProvider creates a GRPC stream
type StreamProvider func(conn *grpc.ClientConn) (grpc.ClientStream, error)

// StreamConnection manages the GRPC connection and client stream
type StreamConnection struct {
	*GRPCConnection
	stream grpc.ClientStream
}

// NewStreamConnection creates a new connection with stream
func NewStreamConnection(ctx fabcontext.Client, chConfig fab.ChannelCfg, streamProvider StreamProvider, url string, opts ...options.Opt) (*StreamConnection, error) {
	conn, err := NewConnection(ctx, chConfig, url, opts...)
	if err != nil {
		return nil, err
	}

	stream, err := streamProvider(conn.conn)
	if err != nil {
		conn.commManager.ReleaseConn(conn.conn)
		return nil, errors.Wrapf(err, "could not create stream to %s", url)
	}

	if stream == nil {
		return nil, errors.New("unexpected nil stream received from provider")
	}

	return &StreamConnection{
		GRPCConnection: conn,
		stream:         stream,
	}, nil
}

// Close closes the connection
func (c *StreamConnection) Close() {
	if c.Closed() {
		return
	}

	logger.Debug("Closing stream....")
	if err := c.stream.CloseSend(); err != nil {
		logger.Warnf("error closing GRPC stream: %s", err)
	}

	c.GRPCConnection.Close()
}

// Stream returns the GRPC stream
func (c *StreamConnection) Stream() grpc.Stream {
	return c.stream
}
