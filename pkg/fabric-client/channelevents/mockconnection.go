// +build channelevents

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// MockConnection is a fake connection used for unit testing
type MockConnection struct {
	AuthorizedEvents []eventType
	rcvch            chan interface{}
}

// NewMockConnection returns a new MockConnection using the given options
func NewMockConnection(opts ...MockConnOpt) *MockConnection {
	conn := &MockConnection{
		rcvch: make(chan interface{}),
	}
	for _, opt := range opts {
		opt.Apply(conn)
	}
	return conn
}

// Close implements the Connection interface
func (c *MockConnection) Close() {
	if c.rcvch != nil {
		close(c.rcvch)
		c.rcvch = nil
	}
}

// Send implements the Connection interface
func (c *MockConnection) Send(emsg *pb.ChannelServiceRequest) error {
	panic("not implemented")
}

// Disconnect implements the Connection interface
func (c *MockConnection) Disconnect(err error) {
}

// Receive implements the Connection interface
func (c *MockConnection) Receive(eventch chan<- interface{}) {
	for {
		e, ok := <-c.rcvch
		if !ok {
			break
		}

		eventch <- e
	}
}

// MockConnectionProviderFactory creates various mock Connection Providers
type MockConnectionProviderFactory struct {
	connection *MockConnection
	mtx        sync.RWMutex
}

func NewMockConnectionProviderFactory() *MockConnectionProviderFactory {
	return &MockConnectionProviderFactory{}
}

// Provider returns a connection provider that always returns the given connection
func (cp *MockConnectionProviderFactory) Provider(conn *MockConnection) ConnectionProvider {
	return func(string, apifabclient.FabricClient, *apiconfig.PeerConfig) (Connection, error) {
		return conn, nil
	}
}

// MockConnOpt applies an option to a MockConnection
type MockConnOpt interface {
	// Apply applies the option to the MockConnection
	Apply(conn *MockConnection)
}

// AuthorizedEventsOpt is a connection option that applies authorized events to the MockConnection
type AuthorizedEventsOpt struct {
	AuthorizedEvents []eventType
}

// Apply applies the option to the MockConnection
func (o *AuthorizedEventsOpt) Apply(conn *MockConnection) {
	conn.AuthorizedEvents = o.AuthorizedEvents
}

// NewAuthorizedEventsOpt returns a new AuthorizedEventsOpt
func NewAuthorizedEventsOpt(authorizedEvents ...eventType) *AuthorizedEventsOpt {
	return &AuthorizedEventsOpt{AuthorizedEvents: authorizedEvents}
}
