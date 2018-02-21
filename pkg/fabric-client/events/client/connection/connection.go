/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package connection

import (
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"

	ccomm "github.com/hyperledger/fabric-sdk-go/pkg/config/comm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var logger = logging.NewLogger("fabric_sdk_go")

// Provider is a function that creates a Connection.
type Provider func(channelID string, context apifabclient.Context, peer apifabclient.Peer) (Connection, error)

// Connection defines the functions for an event service connection
type Connection interface {
	// Receive sends events to the given channel
	Receive(chan<- interface{})
	// Close closes the connection
	Close()
	// Closed return true if the connection is closed
	Closed() bool
}

// StreamProvider creates a GRPC stream
type StreamProvider func(conn *grpc.ClientConn) (grpc.ClientStream, error)

// GRPCConnection manages the GRPC connection and client stream
type GRPCConnection struct {
	channelID   string
	conn        *grpc.ClientConn
	stream      grpc.ClientStream
	context     apifabclient.Context
	tlsCertHash []byte
	done        int32
	grpcOpts    []grpc.DialOption
}

// New creates a new connection
func New(channelID string, context apifabclient.Context, url string, serverHostOverride string, streamProvider StreamProvider) (*GRPCConnection, error) {
	if url == "" {
		return nil, errors.New("server URL not specified")
	}

	opts, err := newDialOpts(context.Config(), url, serverHostOverride)
	if err != nil {
		return nil, err
	}

	grpcconn, err := grpc.Dial(urlutil.ToAddress(url), opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "could not connect to %s", url)
	}

	stream, err := streamProvider(grpcconn)
	if err != nil {
		if err := grpcconn.Close(); err != nil {
			logger.Warnf("error closing GRPC connection: %s", err)
		}
		return nil, errors.Wrapf(err, "could not create stream to %s", url)
	}

	if stream == nil {
		return nil, errors.New("unexpected nil stream received from provider")
	}

	return &GRPCConnection{
		channelID:   channelID,
		conn:        grpcconn,
		stream:      stream,
		context:     context,
		tlsCertHash: ccomm.TLSCertHash(context.Config()),
		grpcOpts:    opts,
	}, nil
}

// ChannelID returns the ID of the channel
func (c *GRPCConnection) ChannelID() string {
	return c.channelID
}

// Close closes the connection
func (c *GRPCConnection) Close() {
	if !c.setClosed() {
		logger.Debugf("Already closed\n")
		return
	}

	logger.Infof("Closing stream....\n")
	if err := c.stream.CloseSend(); err != nil {
		logger.Warnf("error closing GRPC stream: %s", err)
	}

	logger.Debugf("Closing connection....\n")
	if err := c.conn.Close(); err != nil {
		logger.Warnf("error closing GRPC connection: %s", err)
	}
}

// Closed returns true if the connection has been closed
func (c *GRPCConnection) Closed() bool {
	return atomic.LoadInt32(&c.done) == 1
}

func (c *GRPCConnection) setClosed() bool {
	return atomic.CompareAndSwapInt32(&c.done, 0, 1)
}

// Stream returns the GRPC stream
func (c *GRPCConnection) Stream() grpc.Stream {
	return c.stream
}

// TLSCertHash returns the hash of the TLS cert
func (c *GRPCConnection) TLSCertHash() []byte {
	return c.tlsCertHash
}

// Context returns the context of the client establishing the connection
func (c *GRPCConnection) Context() apifabclient.Context {
	return c.context
}

func newDialOpts(config apiconfig.Config, url string, serverHostOverride string) ([]grpc.DialOption, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(config.TimeoutOrDefault(apiconfig.EventHub)))
	if urlutil.IsTLSEnabled(url) {
		tlsCaCertPool, err := config.TLSCACertPool()
		if err != nil {
			return nil, err
		}

		logger.Debugf("Creating a secure connection to [%s] with TLS serverHostOverride [%s]\n", url, serverHostOverride)
		creds := credentials.NewClientTLSFromCert(tlsCaCertPool, serverHostOverride)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		logger.Debugf("Creating an insecure connection [%s]\n", url)
		opts = append(opts, grpc.WithInsecure())
	}
	return opts, nil
}
