/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"context"
	"sync"
	"time"

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
	index map[*grpc.ClientConn]*cachedConn
	lock  sync.Mutex
}

type cachedConn struct {
	target    string
	conn      *grpc.ClientConn
	open      uint64
	lastOpen  time.Time
	lastClose time.Time
}

func newCachingConnector() *cachingConnector {
	cc := cachingConnector{
		conns: sync.Map{},
		index: map[*grpc.ClientConn]*cachedConn{},
	}
	return &cc
}

func (cc *cachingConnector) DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	logger.Infof("DialContext: %s", target)

	c, ok := cc.loadConn(target)
	if ok {
		if !cc.openConn(ctx, c) {
			return nil, errors.Errorf("dialing connection timed out [%s]", target)
		}
		return c.conn, nil
	}

	c, err := cc.createConn(ctx, target, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "connection creation failed")
	}

	if !cc.openConn(ctx, c) {
		return nil, errors.Errorf("dialing connection timed out [%s]", target)
	}
	return c.conn, nil
}

func (cc *cachingConnector) ReleaseConn(conn *grpc.ClientConn) {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	cconn, ok := cc.index[conn]
	if !ok {
		logger.Errorf("connection not found: %v", conn)
		return
	}

	if cconn.open > 0 {
		cconn.lastClose = time.Now()
		cconn.open--
	}
	logger.Infof("ReleaseConn: %v", cconn)

	// TODO set (or reset) a timer to close the connection or janitor sweep
	//conn.Close()
}

func (cc *cachingConnector) loadConn(target string) (*cachedConn, bool) {
	connRaw, ok := cc.conns.Load(target)
	if ok {
		c, ok := connRaw.(*cachedConn)
		if ok {
			if c.conn.GetState() != connectivity.Shutdown {
				logger.Infof("using cached connection: %v", c)
				return c, true
			}
			cc.failConn(target)
		}
	}
	return nil, false
}

func (cc *cachingConnector) createConn(ctx context.Context, target string, opts ...grpc.DialOption) (*cachedConn, error) {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	cconn, ok := cc.loadConn(target)
	if ok {
		return cconn, nil
	}

	logger.Infof("creating connection for %s", target)
	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "dialing peer failed")
	}

	logger.Infof("storing connection for %s", target)

	cconn = &cachedConn{
		target: target,
		conn:   conn,
	}
	cc.conns.Store(target, cconn)
	cc.index[conn] = cconn

	return cconn, nil
}

func (cc *cachingConnector) openConn(ctx context.Context, c *cachedConn) bool {
	if !c.conn.WaitForStateChange(ctx, connectivity.Connecting) {
		return false
	}

	cc.lock.Lock()
	defer cc.lock.Unlock()
	c.open++
	c.lastOpen = time.Now()

	logger.Infof("connection was opened [%v]", c)
	return true
}

func (cc *cachingConnector) failConn(target string) {
	logger.Infof("connection was shutdown - removing %s", target)

	// TBD on the locks.
	cc.lock.Lock()
	defer cc.lock.Unlock()
	cc.conns.Delete(target)
}
