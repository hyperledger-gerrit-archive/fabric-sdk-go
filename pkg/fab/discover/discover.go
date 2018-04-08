/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discover

import (
	"context"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	corecomm "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"google.golang.org/grpc"
)

var logger = logging.NewLogger("fabsdk/fab")

// Client implements a Discover client
type Client struct {
	discClient discclient.Client
	connRef    *lazyref.Reference
	ctx        fabcontext.Client
}

// New returns a new Discover client
func New(ctx fabcontext.Client, chConfig fab.ChannelCfg, url string, idleTimeout time.Duration, opts ...options.Opt) (*Client, error) {
	authInfo, err := newAuthInfo(ctx)
	if err != nil {
		return nil, err
	}

	c := &Client{
		ctx: ctx,
		connRef: lazyref.New(
			initializer(ctx, chConfig, url),
			lazyref.WithIdleExpiration(idleTimeout),
			lazyref.WithFinalizer(finalizer),
		),
	}
	c.discClient = discclient.NewClient(c.dialer, authInfo, c.signer)

	return c, nil
}

// Send retrieves information about channel peers, endorsers, and MSP config
func (c *Client) Send(ctx context.Context, req *discclient.Request) (discclient.Response, error) {
	return c.discClient.Send(ctx, req)
}

// Close closes the connection
func (c *Client) Close() {
	c.connRef.Close()
}

func (c *Client) dialer() (*grpc.ClientConn, error) {
	conn, err := c.connRef.Get()
	if err != nil {
		return nil, err
	}
	return conn.(*comm.GRPCConnection).ClientConn(), nil
}

func (c *Client) signer(msg []byte) ([]byte, error) {
	return c.ctx.SigningManager().Sign(msg, c.ctx.PrivateKey())
}

func initializer(ctx fabcontext.Client, chConfig fab.ChannelCfg, url string) lazyref.Initializer {
	return func() (interface{}, error) {
		return comm.NewConnection(ctx, chConfig, url)
	}
}

func finalizer(value interface{}) {
	if value != nil {
		logger.Infof("Closing GRPC connection")
		value.(*comm.GRPCConnection).Close()
	}
}

func newAuthInfo(ctx fabcontext.Client) (*discovery.AuthInfo, error) {
	identity, err := ctx.Serialize()
	if err != nil {
		return nil, err
	}

	return &discovery.AuthInfo{
		ClientIdentity:    identity,
		ClientTlsCertHash: corecomm.TLSCertHash(ctx.EndpointConfig()),
	}, nil
}
