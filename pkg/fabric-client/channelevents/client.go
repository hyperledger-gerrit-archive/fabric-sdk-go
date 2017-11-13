// +build channelevents

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"regexp"
	"sync"
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
	eventTypes          []eventType
	opts                ClientOpts
	dispatcher          *eventDispatcher
	connEvent           chan *fab.ConnectionEvent
	connEventSubscriber chan<- *fab.ConnectionEvent
	registerOnce        sync.Once
	connectionState     int32
	stopped             int32
	eventChannelSize    int
}

// AdminClient extends Client with ability to receive block events.
type AdminClient struct {
	Client
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

	// Reconnect indicates whether the client should automatically attempt to reconnect
	// to the serever after a connection has been lost
	// Default: true
	Reconnect bool

	// MaxReconnectAttempts is the maximum number of times that the client will attempt
	// to connect to the server. If set to 0 then the client will try until it is stopped.
	// Default: 1
	MaxConnectAttempts uint

	// MaxReconnectAttempts is the maximum number of times that the client will attempt
	// to reconnect to the server after a connection has been lost. If set to 0 then the
	// client will try until it is stopped.
	// Default: 0 (try forever)
	MaxReconnectAttempts uint

	// TimeBetweenConnectAttempts is the time between connection attempts.
	// Default: 5 seconds
	TimeBetweenConnectAttempts time.Duration

	// ConnectEventCh is the channel that is to receive connection events, i.e. when the client connects and/or
	// disconnects from the channel event service.
	ConnectEventCh chan *fab.ConnectionEvent

	// connectionProvider specifies the connection provider. This is only used for unit testing.
	connectionProvider func(string, fab.FabricClient, *apiconfig.PeerConfig) (Connection, error)
}

// DefaultClientOpts returns client options set to default values
func DefaultClientOpts() *ClientOpts {
	return &ClientOpts{
		ResponseTimeout:            3 * time.Second,
		Reconnect:                  true,
		MaxConnectAttempts:         1,
		MaxReconnectAttempts:       0, // Try forever
		TimeBetweenConnectAttempts: 5 * time.Second,
		EventQueueSize:             1000,
		connectionProvider:         grpcConnectionProvider,
	}
}

// NewClient returns a new channel event Client
func NewClient(fabclient fab.FabricClient, peerConfig *apiconfig.PeerConfig, channelID string) (*Client, error) {
	return NewClientWithOpts(fabclient, peerConfig, channelID, DefaultClientOpts())
}

// NewClientWithOpts returns a new channel event Client initialized with the given options
func NewClientWithOpts(fabclient fab.FabricClient, peerConfig *apiconfig.PeerConfig, channelID string, opts *ClientOpts) (*Client, error) {
	return newClient(fabclient, peerConfig, channelID, opts, []eventType{FILTEREDBLOCKEVENT})
}

// NewAdminClient returns a new channel event Admin Client
func NewAdminClient(fabclient fab.FabricClient, peerConfig *apiconfig.PeerConfig, channelID string) (*AdminClient, error) {
	return NewAdminClientWithOpts(fabclient, peerConfig, channelID, DefaultClientOpts())
}

// NewAdminClientWithOpts returns a new channel event Admin Client
func NewAdminClientWithOpts(fabclient fab.FabricClient, peerConfig *apiconfig.PeerConfig, channelID string, opts *ClientOpts) (*AdminClient, error) {
	client, err := newClient(fabclient, peerConfig, channelID, opts, []eventType{BLOCKEVENT, FILTEREDBLOCKEVENT})
	if err != nil {
		return nil, err
	}
	return &AdminClient{Client: *client}, nil
}

// Connect connects to the peer and registers for channel events on a particular channel.
func (cc *Client) Connect() error {
	if cc.opts.MaxConnectAttempts == 1 {
		return cc.connect()
	}
	return cc.connectWithRetry(cc.opts.MaxConnectAttempts, cc.opts.TimeBetweenConnectAttempts)
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

	if cc.connEventSubscriber != nil {
		close(cc.connEventSubscriber)
	}

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

// RegisterFilteredBlockEvent registers for filtered block events. If the client is not authorized to receive
// filtered block events then an error is returned.
func (cc *Client) RegisterFilteredBlockEvent() (fab.Registration, <-chan *fab.FilteredBlockEvent, error) {
	if cc.Stopped() {
		return nil, nil, errors.New("channel event client is closed")
	}

	eventch := make(chan *fab.FilteredBlockEvent, cc.eventChannelSize)
	respch := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(newRegisterFilteredBlockEvent(eventch, respch))
	response := <-respch

	return response.Reg, eventch, response.Err
}

// RegisterChaincodeEvent registers for chaincode events. If the client is not authorized to receive
// chaincode events then an error is returned.
// - ccID is the chaincode ID for which events are to be received
// - eventFilter is the chaincode event name for which events are to be received
func (cc *Client) RegisterChaincodeEvent(ccID, eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error) {
	if cc.Stopped() {
		return nil, nil, errors.New("channel event client is closed")
	}
	if ccID == "" {
		return nil, nil, errors.New("chaincode ID is required")
	}
	if eventFilter == "" {
		return nil, nil, errors.New("event filter is required")
	}

	regExp, err := regexp.Compile(eventFilter)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "invalid event filter [%s] for chaincode [%s]", eventFilter, ccID)
	}

	eventch := make(chan *fab.CCEvent, cc.eventChannelSize)
	respch := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(newRegisterCCEvent(ccID, eventFilter, regExp, eventch, respch))
	response := <-respch

	return response.Reg, eventch, response.Err
}

// RegisterTxStatusEvent registers for transaction status events. If the client is not authorized to receive
// transaction status events then an error is returned.
// - txID is the transaction ID for which events are to be received
func (cc *Client) RegisterTxStatusEvent(txID string) (fab.Registration, <-chan *fab.TxStatusEvent, error) {
	if cc.Stopped() {
		return nil, nil, errors.New("channel event client is closed")
	}
	if txID == "" {
		return nil, nil, errors.New("txID must be provided")
	}

	eventch := make(chan *fab.TxStatusEvent, cc.eventChannelSize)
	respch := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(newRegisterTxStatusEvent(txID, eventch, respch))
	response := <-respch

	return response.Reg, eventch, response.Err
}

// Unregister unregisters the given registration.
// - reg is the registration handle that was returned from one of the RegisterXXX functions
func (cc *Client) Unregister(reg fab.Registration) {
	if cc.Stopped() {
		// Client is already closed. Do nothing.
		return
	}

	cc.dispatcher.submit(newUnregisterEvent(reg))
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
		eventTypes:          eventTypes,
		opts:                *opts,
		dispatcher:          dispatcher,
		connEvent:           make(chan *fab.ConnectionEvent),
		connEventSubscriber: opts.ConnectEventCh,
		connectionState:     disconnected,
		eventChannelSize:    opts.EventQueueSize,
	}

	go dispatcher.start()

	return cc, nil
}

func (cc *Client) connect() error {
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

	var err error
	cc.registerOnce.Do(func() {
		logger.Debugf("Submitting connection event registration...\n")
		_, eventch, err := cc.registerConnectionEvent()
		if err != nil {
			logger.Errorf("Error registering for connection events: %s\n", err)
			cc.Disconnect()
		}
		cc.connEvent = eventch
		go cc.monitorConnection()
	})

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

	logger.Debugf("Submitting connected event\n")
	cc.dispatcher.submit(&connectedEvent{})

	return err
}

func (cc *Client) connectWithRetry(maxAttempts uint, timeBetweenAttempts time.Duration) error {
	if cc.Stopped() {
		return errors.New("channel event client is closed")
	}
	if timeBetweenAttempts < time.Second {
		timeBetweenAttempts = time.Second
	}

	var attempts uint
	for {
		attempts++
		logger.Debugf("Attempt #%d to connect...\n", attempts)
		if err := cc.connect(); err != nil {
			logger.Warnf("... connection attempt failed: %s\n", err)
			if maxAttempts > 0 && attempts >= maxAttempts {
				logger.Warnf("maximum connect attempts exceeded\n")
				return errors.New("maximum connect attempts exceeded")
			}
			time.Sleep(timeBetweenAttempts)
		} else {
			logger.Debugf("... connect succeeded.\n")
			return nil
		}
	}
}

func (cc *Client) registerConnectionEvent() (fab.Registration, chan *fab.ConnectionEvent, error) {
	if cc.Stopped() {
		return nil, nil, errors.New("channel event client is closed")
	}

	eventch := make(chan *fab.ConnectionEvent, cc.eventChannelSize)
	respch := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(newRegisterConnectionEvent(eventch, respch))
	response := <-respch
	return response.Reg, eventch, response.Err
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

// RegisterBlockEvent registers for block events. If the client is not authorized to receive
// block events then an error is returned.
func (cc *AdminClient) RegisterBlockEvent() (fab.Registration, <-chan *fab.BlockEvent, error) {
	if cc.Stopped() {
		return nil, nil, errors.New("channel event client is closed")
	}

	eventch := make(chan *fab.BlockEvent, cc.eventChannelSize)
	respch := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(newRegisterBlockEvent(eventch, respch))
	response := <-respch

	return response.Reg, eventch, response.Err
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

func (cc *Client) monitorConnection() {
	logger.Debugf("Monitoring connection\n")
	for {
		event, ok := <-cc.connEvent
		if !ok {
			logger.Debugln("Connection has closed.")
			break
		}

		if cc.Stopped() {
			logger.Debugln("Channel event client has been stopped.")
			break
		}

		if cc.connEventSubscriber != nil {
			logger.Debugln("Sending connection event to subscriber.")
			cc.connEventSubscriber <- event
		}

		if event.Connected {
			logger.Infof("Channel event client has connected\n")
		} else if cc.opts.Reconnect {
			logger.Warnf("Channel event client has disconnected. Details: %s\n", event.Err)
			if cc.setConnectionState(connected, disconnected) {
				logger.Warnf("Attempting to reconnect...\n")
				go cc.reconnect()
			} else if cc.setConnectionState(connecting, disconnected) {
				logger.Warnf("Reconnect already in progress. Setting state to disconnected\n")
			}
		} else {
			logger.Warnf("Channel event client has disconnected. Terminating: %s\n", event.Err)
			go cc.Disconnect()
			break
		}
	}
	logger.Debugf("Exiting connection monitor\n")
}

func (cc *Client) reconnect() {
	logger.Debugf("Attempting to reconnect channel event client...\n")
	if err := cc.connectWithRetry(cc.opts.MaxReconnectAttempts, cc.opts.TimeBetweenConnectAttempts); err != nil {
		logger.Warnf("Could not reconnect channel event client: %s\n", err)
		cc.Disconnect()
	}
}
