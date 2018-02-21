/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/client/dispatcher"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/client/lbp"
	eventservice "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/service"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// ConnectionState is the state of the client connection
type ConnectionState int32

const (
	// Disconnected indicates that the client is disconnected from the server
	Disconnected ConnectionState = iota
	// Connecting indicates that the client is in the process of establishing a connection
	Connecting
	// Connected indicates that the client is connected to the server
	Connected
)

// Options provides options for the events client
type Options interface {
	eventservice.Options

	// LoadBalancePolicy returns the load-balance policy to use when
	// choosing an event server endpoint from a set of endpoints
	LoadBalancePolicy() lbp.LoadBalancePolicy

	// Reconnect indicates whether the client should automatically attempt to reconnect
	// to another server after a connection has been lost
	// Default: true
	Reconnect() bool

	// MaxConnectAttempts is the maximum number of times that the client will attempt
	// to connect to an event server. If set to 0 then the client will try until it is stopped.
	// Default: 1
	MaxConnectAttempts() uint

	// MaxReconnectAttempts is the maximum number of times that the client will attempt
	// to reconnect to the event server after a connection has been lost. If set to 0 then the
	// client will try until it is stopped.
	// Default: 0 (try forever)
	MaxReconnectAttempts() uint

	// ReconnectInitialDelay is the initial delay before attempting to reconnect.
	// Default: 0 seconds
	ReconnectInitialDelay() time.Duration

	// TimeBetweenConnectAttempts is the time between connection attempts.
	// Default: 5 seconds
	TimeBetweenConnectAttempts() time.Duration

	// ConnectEventCh is the channel that is to receive connection events, i.e. when the client connects and/or
	// disconnects from the channel event service.
	ConnectEventCh() chan *apifabclient.ConnectionEvent

	// ResponseTimeout is the timeout when waiting for a response from the event server
	ResponseTimeout() time.Duration
}

// Client connects to an event server and receives events, such as bock, filtered block,
// chaincode, and transaction status events. Client also monitors the connection to the
// event server and attempts to reconnect if the connection is closed.
type Client struct {
	eventservice.Service
	sync.RWMutex
	opts              Options
	connEvent         chan *apifabclient.ConnectionEvent
	connectionState   int32
	stopped           int32
	registerOnce      sync.Once
	permitBlockEvents bool
	afterConnect      handler
	beforeReconnect   handler
}

type handler func() error

// New returns a new event client
func New(opts Options, permitBlockEvents bool, dispatcher eventservice.Dispatcher) Client {
	return Client{
		Service:           *eventservice.New(dispatcher, opts),
		opts:              opts,
		connEvent:         make(chan *apifabclient.ConnectionEvent),
		connectionState:   int32(Disconnected),
		permitBlockEvents: permitBlockEvents,
	}
}

// SetAfterConnectHandler registers a handler that is called
// after the client connects to the event server. This allows for
// custom code to be executed for a particular
// event client implementation.
func (c *Client) SetAfterConnectHandler(h handler) {
	c.Lock()
	defer c.Unlock()
	c.afterConnect = h
}

func (c *Client) afterConnectHandler() handler {
	c.RLock()
	defer c.RUnlock()
	return c.afterConnect
}

// SetBeforeReconnectHandler registers a handler that will be called
// before retrying to reconnect to the event server. This allows for
// custom code to be executed for a particular event client implementation.
func (c *Client) SetBeforeReconnectHandler(h handler) {
	c.Lock()
	defer c.Unlock()
	c.beforeReconnect = h
}

func (c *Client) beforeReconnectHandler() handler {
	c.RLock()
	defer c.RUnlock()
	return c.beforeReconnect
}

// Connect connects to the peer and registers for events on a particular channel.
func (c *Client) Connect() error {
	if c.opts.MaxConnectAttempts() == 1 {
		return c.connect()
	}
	return c.connectWithRetry(c.opts.MaxConnectAttempts(), c.opts.TimeBetweenConnectAttempts())
}

// Close closes the connection to the event server and unallocates all resources.
// Once this function is invoked the client may no longer be used.
func (c *Client) Close() {
	logger.Debugf("Attempting to close event client...\n")

	if !c.setStoppped() {
		// Already stopped
		logger.Infof("Client already stopped\n")
		return
	}

	logger.Debugf("Stopping client...\n")

	if c.opts.ConnectEventCh() != nil {
		close(c.opts.ConnectEventCh())
	}

	logger.Debugf("Sending disconnect request...\n")

	respch := make(chan *dispatcher.ConnectionResponse)
	c.Submit(dispatcher.NewDisconnectEvent(respch))
	response := <-respch

	if response.Err != nil {
		logger.Warnf("Received error from disconnect request: %s\n", response.Err)
	} else {
		logger.Debugf("Received success from disconnect request\n")
	}

	logger.Debugf("Stopping dispatcher...\n")

	c.Stop()

	c.mustSetConnectionState(Disconnected)

	logger.Debugf("... event client is stopped\n")
}

func (c *Client) connect() error {
	if c.Stopped() {
		return errors.New("event client is closed")
	}

	if !c.setConnectionState(Disconnected, Connecting) {
		return errors.Errorf("unable to connect event client since client is [%s]. Expecting client to be in state [%s]", c.ConnectionState(), Disconnected)
	}

	logger.Debugf("Submitting connection request...\n")

	respch := make(chan *dispatcher.ConnectionResponse)
	c.Submit(dispatcher.NewConnectEvent(respch))

	r := <-respch

	if r.Err != nil {
		c.mustSetConnectionState(Disconnected)
		logger.Debugf("... got error in connection response: %s\n", r.Err)
		return r.Err
	}

	var err error
	c.registerOnce.Do(func() {
		logger.Debugf("Submitting connection event registration...\n")
		_, eventch, err := c.RegisterConnectionEvent()
		if err != nil {
			logger.Errorf("Error registering for connection events: %s\n", err)
			c.Close()
		}
		c.connEvent = eventch
		go c.monitorConnection()
	})

	handler := c.afterConnectHandler()
	if handler != nil {
		if err := handler(); err != nil {
			logger.Warnf("Error invoking afterConnect handler: %s. Disconnecting...", err)

			respch := make(chan *dispatcher.ConnectionResponse)
			c.Submit(dispatcher.NewDisconnectEvent(respch))
			response := <-respch

			if response.Err != nil {
				logger.Warnf("Received error from disconnect request: %s\n", response.Err)
			} else {
				logger.Debugf("Received success from disconnect request\n")
			}

			c.setConnectionState(Connecting, Disconnected)

			return errors.Errorf("unable to register for events: %s", err)
		}
	}

	c.setConnectionState(Connecting, Connected)

	logger.Debugf("Submitting connected event\n")
	c.Submit(dispatcher.NewConnectedEvent())

	return err
}

func (c *Client) connectWithRetry(maxAttempts uint, timeBetweenAttempts time.Duration) error {
	if c.Stopped() {
		return errors.New("event client is closed")
	}
	if timeBetweenAttempts < time.Second {
		timeBetweenAttempts = time.Second
	}

	var attempts uint
	for {
		attempts++
		logger.Infof("Attempt #%d to connect...\n", attempts)
		if err := c.connect(); err != nil {
			logger.Warnf("... connection attempt failed: %s\n", err)
			if maxAttempts > 0 && attempts >= maxAttempts {
				logger.Warnf("maximum connect attempts exceeded\n")
				return errors.New("maximum connect attempts exceeded")
			}
			time.Sleep(timeBetweenAttempts)
		} else {
			logger.Infof("... connect succeeded.\n")
			return nil
		}
	}
}

// RegisterBlockEvent registers for block events. If the client is not authorized to receive
// block events then an error is returned.
func (c *Client) RegisterBlockEvent(filter ...apifabclient.BlockFilter) (apifabclient.Registration, <-chan *apifabclient.BlockEvent, error) {
	if !c.permitBlockEvents {
		return nil, nil, errors.New("block events are not permitted")
	}
	return c.Service.RegisterBlockEvent(filter...)
}

// RegisterConnectionEvent registers a connection event. The returned
// ConnectionEvent channel will be called whenever the client clients or disconnects
// from the event server
func (c *Client) RegisterConnectionEvent() (apifabclient.Registration, chan *apifabclient.ConnectionEvent, error) {
	if c.Stopped() {
		return nil, nil, errors.New("event client is closed")
	}

	eventch := make(chan *apifabclient.ConnectionEvent, c.opts.EventConsumerBufferSize())
	respch := make(chan *apifabclient.RegistrationResponse)
	c.Submit(dispatcher.NewRegisterConnectionEvent(eventch, respch))
	response := <-respch
	return response.Reg, eventch, response.Err
}

// Stopped returns true if the client has been stopped (disconnected)
// and is no longer usable.
func (c *Client) Stopped() bool {
	return atomic.LoadInt32(&c.stopped) == 1
}

func (c *Client) setStoppped() bool {
	return atomic.CompareAndSwapInt32(&c.stopped, 0, 1)
}

// ConnectionState returns the connection state
func (c *Client) ConnectionState() ConnectionState {
	return ConnectionState(atomic.LoadInt32(&c.connectionState))
}

// setConnectionState sets the connection state only if the given currentState
// matches the actual state. True is returned if the connection state was successfully set.
func (c *Client) setConnectionState(currentState, newState ConnectionState) bool {
	return atomic.CompareAndSwapInt32(&c.connectionState, int32(currentState), int32(newState))
}

func (c *Client) mustSetConnectionState(newState ConnectionState) {
	atomic.StoreInt32(&c.connectionState, int32(newState))
}

func (c *Client) monitorConnection() {
	logger.Debugf("Monitoring connection\n")
	for {
		event, ok := <-c.connEvent
		if !ok {
			logger.Debugln("Connection has closed.")
			break
		}

		if c.Stopped() {
			logger.Debugln("Event client has been stopped.")
			break
		}

		if c.opts.ConnectEventCh() != nil {
			logger.Debugln("Sending connection event to subscriber.")
			c.opts.ConnectEventCh() <- event
		}

		if event.Connected {
			logger.Debugf("Event client has connected\n")
		} else if c.opts.Reconnect() {
			logger.Warnf("Event client has disconnected. Details: %s\n", event.Err)
			if c.setConnectionState(Connected, Disconnected) {
				logger.Warnf("Attempting to reconnect...\n")
				go c.reconnect()
			} else if c.setConnectionState(Connecting, Disconnected) {
				logger.Warnf("Reconnect already in progress. Setting state to disconnected\n")
			}
		} else {
			logger.Warnf("Event client has disconnected. Terminating: %s\n", event.Err)
			go c.Close()
			break
		}
	}
	logger.Debugf("Exiting connection monitor\n")
}

func (c *Client) reconnect() {
	logger.Debugf("Waiting %s before attempting to reconnect event client...", c.opts.ReconnectInitialDelay)
	time.Sleep(c.opts.ReconnectInitialDelay())

	logger.Debugf("Attempting to reconnect event client...\n")

	handler := c.beforeReconnectHandler()
	if handler != nil {
		if err := handler(); err != nil {
			logger.Errorf("Error invoking beforeReconnect handler: %s", err)
			return
		}
	}

	if err := c.connectWithRetry(c.opts.MaxReconnectAttempts(), c.opts.TimeBetweenConnectAttempts()); err != nil {
		logger.Warnf("Could not reconnect event client: %s. Closing.\n", err)
		c.Close()
	}
}

func (s ConnectionState) String() string {
	switch s {
	case Disconnected:
		return "Disconnected"
	case Connected:
		return "Connected"
	case Connecting:
		return "Connecting"
	default:
		return "undefined"
	}
}
