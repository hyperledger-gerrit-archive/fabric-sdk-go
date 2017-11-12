// +build channelevents

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"sync/atomic"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

const (
	disconnected int32 = iota
	connecting
	connected
)

// Client connects to a peer and receives channel events, such as filtered block, chaincode, and transaction status events.
type Client struct {
	eventTypes       []eventType
	opts             ClientOpts
	dispatcher       *eventDispatcher
	connectionState  int32
	stopped          int32
	eventChannelSize int
}

// ClientOpts provides options for the events client
type ClientOpts struct {
	// ResponseTimeout is the response timeout when communicating with the server.
	// Default: 3 seconds
	ResponseTimeout time.Duration

	// EventQueueSize specifies the maximum number of events that can be queued on an event channel,
	// after which further events are rejected.
	// Default: 1000
	EventQueueSize int

	// connectionProvider specifies the connection provider. This is only used for unit testing.
	connectionProvider func(string, fab.FabricClient, *apiconfig.PeerConfig) (Connection, error)
}

// DefaultClientOpts returns client options set to default values
func DefaultClientOpts() *ClientOpts {
	return &ClientOpts{
		ResponseTimeout:    3 * time.Second,
		EventQueueSize:     1000,
		connectionProvider: grpcConnectionProvider,
	}
}

// NewClient returns a new ChannelEventClient
func NewClient(fabclient fab.FabricClient, peerConfig *apiconfig.PeerConfig, channelID string) (*Client, error) {
	return NewClientWithOpts(fabclient, peerConfig, channelID, DefaultClientOpts())
}

// NewClientWithOpts returns a new ChannelEventClient initialized with the given options
func NewClientWithOpts(fabclient fab.FabricClient, peerConfig *apiconfig.PeerConfig, channelID string, opts *ClientOpts) (*Client, error) {
	return newClient(fabclient, peerConfig, channelID, opts, []eventType{FILTEREDBLOCKEVENT})
}

// Connect connects to the peer and registers for channel events on a particular channel.
func (cc *Client) Connect() error {
	if cc.Stopped() {
		return errors.New("channel event client is closed")
	}

	if !cc.setConnectionState(disconnected, connecting) {
		return errors.Errorf("unable to connect channel event client since client is [%d]", cc.ConnectionState())
	}

	logger.Debugf("Submitting connection request...\n")

	respch := make(chan *connectionResponse)
	cc.dispatcher.submit(newConnectEvent(respch))

	r := <-respch

	if r.err != nil {
		cc.mustSetConnectionState(disconnected)
		logger.Debugf("... got error in connection response: %s\n", r.err)
		return r.err
	}

	if err := cc.registerChannel(cc.eventTypes); err != nil {
		logger.Warnf("Unable to register for events: %s. Disconnecting...", err)

		respch := make(chan *connectionResponse)
		cc.dispatcher.submit(newDisconnectEvent(respch))
		response := <-respch

		if response.err != nil {
			logger.Warnf("Received error from disconnect request: %s\n", response.err)
		} else {
			logger.Debugf("Received success from disconnect request\n")
		}

		cc.setConnectionState(connecting, disconnected)

		return errors.Errorf("unable to register for events: %s", err)
	}

	cc.setConnectionState(connecting, connected)

	return nil
}

// Disconnect disconnects from the peer
func (cc *Client) Disconnect() {
	logger.Debugf(" Attempting to disconnect channel event client...\n")

	if !cc.setStoppped() {
		// Already stopped
		logger.Debugf("Client already stopped\n")
		return
	}

	logger.Debugf("Stopping client...\n")

	select {
	case resp := <-cc.unregisterChannelAsync():
		if resp.Err != nil {
			logger.Warnf("Error in unregister channel: %s", resp.Err)
		} else {
			logger.Debugf("Channel unregistered\n")
		}
	case <-time.After(cc.opts.ResponseTimeout):
		logger.Warnf("Timed out waiting for unregister channel response")
	}

	logger.Debugf("Sending disconnect request...\n")

	respch := make(chan *connectionResponse)
	cc.dispatcher.submit(newDisconnectEvent(respch))
	response := <-respch

	if response.err != nil {
		logger.Warnf("Received error from disconnect request: %s\n", response.err)
	} else {
		logger.Debugf("Received success from disconnect request\n")
	}

	logger.Debugf("Stopping dispatcher...\n")

	cc.dispatcher.stop()

	cc.mustSetConnectionState(disconnected)

	logger.Debugf("... channel event client is stopped\n")
}

func newClient(fabclient fab.FabricClient, peerConfig *apiconfig.PeerConfig, channelID string, opts *ClientOpts, eventTypes []eventType) (*Client, error) {
	if peerConfig.URL == "" {
		return nil, errors.New("expecting peer URL")
	}
	if channelID == "" {
		return nil, errors.New("expecting channel ID")
	}

	dispatcher := newEventDispatcher(fabclient, peerConfig, channelID, opts.connectionProvider, opts.EventQueueSize)

	cc := &Client{
		eventTypes:       eventTypes,
		opts:             *opts,
		dispatcher:       dispatcher,
		connectionState:  disconnected,
		eventChannelSize: opts.EventQueueSize,
	}

	go dispatcher.start()

	return cc, nil
}

func (cc *Client) registerChannelAsync(eventTypes []eventType) <-chan *fab.RegistrationResponse {
	respch := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(newRegisterChannelEvent(eventTypes, respch))
	return respch
}

func (cc *Client) unregisterChannelAsync() <-chan *fab.RegistrationResponse {
	respch := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(newUnregisterChannelEvent(respch))
	return respch
}

func (cc *Client) registerChannel(eventTypes []eventType) error {
	logger.Debugf("registering channel....\n")

	var err error
	select {
	case s := <-cc.registerChannelAsync(eventTypes):
		err = s.Err
	case <-time.After(cc.opts.ResponseTimeout):
		err = errors.New("timeout waiting for channel registration response")
	}

	if err != nil {
		logger.Errorf("unable to register for channel events: %s\n", err)
		return err
	}

	logger.Debugf("successfully registered for channel events\n")
	return nil
}

// Stopped returns true if the client has been stopped (disconnected)
// and is no longer usable.
func (cc *Client) Stopped() bool {
	return atomic.LoadInt32(&cc.stopped) == 1
}

func (cc *Client) setStoppped() bool {
	return atomic.CompareAndSwapInt32(&cc.stopped, 0, 1)
}

// ConnectionState returns the connection state
func (cc *Client) ConnectionState() int32 {
	return atomic.LoadInt32(&cc.connectionState)
}

// setConnectionState sets the connection state only if the given currentState
// matches the actual state. True is returned if the connection state was successfully set.
func (cc *Client) setConnectionState(currentState, newState int32) bool {
	return atomic.CompareAndSwapInt32(&cc.connectionState, currentState, newState)
}

func (cc *Client) mustSetConnectionState(newState int32) {
	atomic.StoreInt32(&cc.connectionState, newState)
}
