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

const (
	sweepTime = 1 * time.Second
	idleTime  = 10 * time.Second // TODO: make configurable for tests
)

type cachingConnector struct {
	conns       sync.Map
	index       map[*grpc.ClientConn]*cachedConn
	lock        sync.Mutex
	janitorChan chan *cachedConn
	janitorDone chan bool
}

type cachedConn struct {
	target    string
	conn      *grpc.ClientConn
	open      int
	lastOpen  time.Time
	lastClose time.Time
}

func newCachingConnector() *cachingConnector {
	cc := cachingConnector{
		conns:       sync.Map{},
		index:       map[*grpc.ClientConn]*cachedConn{},
		janitorChan: make(chan *cachedConn),
		janitorDone: make(chan bool, 1),
	}

	cc.janitorDone <- true
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

	cc.updateJanitor(cconn)
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
			cc.failConn(c)
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
	cc.updateJanitor(c)

	logger.Infof("connection was opened [%v]", c)
	return true
}

func (cc *cachingConnector) failConn(cconn *cachedConn) {
	logger.Infof("connection was shutdown - removing %s", cconn.target)

	cc.lock.Lock()
	defer cc.lock.Unlock()
	cc.conns.Delete(cconn.target)

	cconn.open = 0
	cconn.lastClose = time.Time{}

	cc.updateJanitor(cconn)
}

func (cc *cachingConnector) updateJanitor(c *cachedConn) {
	select {
	case <-cc.janitorDone:
		cc.janitorChan = make(chan *cachedConn)
		go janitor(cc.janitorChan, cc.janitorDone)
	default:
		cClone := cachedConn{
			target:    string([]byte(c.target)),
			conn:      c.conn,
			open:      c.open,
			lastOpen:  c.lastOpen,
			lastClose: c.lastClose,
		}

		cc.janitorChan <- &cClone
	}
}

func janitor(conn chan *cachedConn, done chan bool) {
	logger.Infof("starting connection janitor")
	idle := map[string]*cachedConn{}
	for {
		select {
		case c := <-conn:
			logger.Infof("updating connection in connection janitor")
			idle[c.target] = c
		case <-time.After(sweepTime):
			//logger.Infof("connection janitor sweeping connections")
			idle = sweep(idle)
			if len(idle) == 0 {
				logger.Infof("closing connection janitor")
				done <- true
				return
			}
		}
	}
}

func sweep(idle map[string]*cachedConn) map[string]*cachedConn {
	keep := map[string]*cachedConn{}
	for _, c := range idle {
		if c.open == 0 && time.Now().After(c.lastClose.Add(idleTime)) {
			logger.Infof("connection janitor closing connection")
			c.conn.Close()
		} else {
			keep[c.target] = c
		}
	}
	return keep
}
