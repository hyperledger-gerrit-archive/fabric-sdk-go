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

// The CachingConnector provides the ability to cache GRPC connections.
// It provides a GRPC compatible Context Dialer interface via the "DialContext" method.
// Connections provided by this component are monitored for becoming idle or entering shutdown state.
// When connections has its usages closed for longer than "idleTime", the connection is closed and removed
// from the connection cache. Callers must release connections by calling the "ReleaseConn" method.
// The Close method will flush all remaining open connections. This component should be considered
// unusable after calling Close.
//
// This component has been designed to be safe for concurrency.

// TODO: make configurable
const (
	sweepTime = 1 * time.Second
	idleTime  = 10 * time.Second
)

type cachingConnector struct {
	conns       sync.Map
	index       map[*grpc.ClientConn]*cachedConn
	lock        sync.Mutex
	waitgroup   sync.WaitGroup
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

func (cc *cachingConnector) Close() {
	logger.Debug("closing caching GRPC connector")

	close(cc.janitorDone)
	cc.waitgroup.Wait()
}

func (cc *cachingConnector) DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	logger.Debugf("DialContext: %s", target)

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
		logger.Warnf("connection not found: %v", conn)
		return
	}

	if cconn.open > 0 {
		cconn.lastClose = time.Now()
		cconn.open--
	}

	cc.updateJanitor(cconn)
}

func (cc *cachingConnector) loadConn(target string) (*cachedConn, bool) {
	connRaw, ok := cc.conns.Load(target)
	if ok {
		c, ok := connRaw.(*cachedConn)
		if ok {
			if c.conn.GetState() != connectivity.Shutdown {
				logger.Debugf("using cached connection: %v", c)
				return c, true
			}
			cc.shutdownConn(c)
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

	logger.Debugf("creating connection for %s", target)
	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "dialing peer failed")
	}

	logger.Debugf("storing connection for %s", target)
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

	logger.Debugf("connection was opened [%v]", c)
	return true
}

func (cc *cachingConnector) shutdownConn(cconn *cachedConn) {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	logger.Infof("connection was shutdown [%s]", cconn.target)
	cc.conns.Delete(cconn.target)
	delete(cc.index, cconn.conn)

	cconn.open = 0
	cconn.lastClose = time.Time{}

	cc.updateJanitor(cconn)
}

func (cc *cachingConnector) removeConn(target string) {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	logger.Infof("removing connection [%s]", target)
	connRaw, ok := cc.conns.Load(target)
	if ok {
		c, ok := connRaw.(*cachedConn)
		if ok {
			delete(cc.index, c.conn)
			cc.conns.Delete(target)
		}
	}
}

func (cc *cachingConnector) updateJanitor(c *cachedConn) {
	select {
	case <-cc.janitorDone:
		cc.janitorChan = make(chan *cachedConn)
		cc.waitgroup.Add(1)
		go janitor(&cc.waitgroup, cc.janitorChan, cc.janitorDone, cc.removeConn)
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

// The janitor monitors open connections for shutdown state or extended non-usage.
// This component operates by running a sweep with a period determined by "sweepTime".
// When a connection returned the GRPC status connectivity.Shutdown or when the connection
// has its usages closed for longer than "idleTime", the connection is closed and the
// "connRemove" notifier is called.
//
// The caching connector:
//    pushes connection information via the "conn" go channel.
//    notifies the janitor of close by closing the "done" go channel.
//
// The janitor:
//    calls "connRemove" callback when closing a connection.
//    decrements the "wg" waitgroup when exiting.
//    writes to the "done" go channel when closing due to becoming empty.

type connRemoveNotifier func(target string)

func janitor(wg *sync.WaitGroup, conn chan *cachedConn, done chan bool, connRemove connRemoveNotifier) {
	logger.Debugf("starting connection janitor")
	defer wg.Done()

	conns := map[string]*cachedConn{}
	for {
		select {
		case c := <-conn:
			logger.Debugf("updating connection in connection janitor")
			conns[c.target] = c

			if len(conns) == 0 {
				logger.Debugf("closing connection janitor")
				done <- true
				return
			}
		case _, ok := <-done:
			if !ok {
				if len(conns) > 0 {
					logger.Debugf("flushing connection janitor with open connections: %d", len(conns))
				} else {
					logger.Debugf("flushing connection janitor")
				}
				flush(conns)
				return
			}
		case <-time.After(sweepTime):
			rm := sweep(conns)
			for _, target := range rm {
				connRemove(target)
				delete(conns, target)
			}

			if len(conns) == 0 {
				logger.Debugf("closing connection janitor")
				done <- true
				return
			}
		}
	}
}

func flush(conns map[string]*cachedConn) {
	for _, c := range conns {
		logger.Debugf("connection janitor closing connection [%s]", c.target)
		c.conn.Close()
	}
}

func sweep(conns map[string]*cachedConn) []string {
	rm := make([]string, 0, len(conns))
	for _, c := range conns {
		if c.open == 0 && time.Now().After(c.lastClose.Add(idleTime)) {
			logger.Debugf("connection janitor closing connection [%s]", c.target)
			c.conn.Close()
			rm = append(rm, c.target)
		} else if c.conn.GetState() == connectivity.Shutdown {
			logger.Debugf("connection already closed [%s]", c.target)
			rm = append(rm, c.target)
		}
	}
	return rm
}
