/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"context"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var logger = logging.NewLogger("fabric_sdk_go")

//
//
// WIP - This is just a sketch!!
//
//

type cachingConnector struct {
	conns sync.Map
}

func newCachingConnector() *cachingConnector {
	cc := cachingConnector{
		conns: sync.Map{},
	}
	return &cc
}

func (cc *cachingConnector) DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	logger.Infof("Caching DialContext: %s", target)

	connRaw, ok := cc.conns.Load(target)
	if ok {
		conn, ok := connRaw.(*grpc.ClientConn)
		if ok {
			if conn.GetState() != connectivity.Shutdown {
				logger.Infof("Peer using cached connection: %s", target)
				return conn, nil
			}
			logger.Infof("Connection was shutdown - removing %s", target)
			cc.conns.Delete(target)
		}
	}
	logger.Infof("Creating connection", target)
	conn, err = grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "dialing peer failed")
	}

	logger.Infof("Storing connection", target)
	cc.conns.Store(target, conn)

	return conn, nil
}

func (cc *cachingConnector) ReleaseConn(conn *grpc.ClientConn) {
	logger.Infof("Caching ReleaseConn: %v", conn)
	//conn.Close()
}
