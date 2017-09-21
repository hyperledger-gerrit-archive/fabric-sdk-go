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

// Client connects to a peer and receives channel events, such as block, chaincode, and transaction status events.
type Client struct {
	opts                ClientOpts
	dispatcher          *eventDispatcher
	connEvent           chan *fab.ConnectionEvent
	connEventSubscriber chan<- *fab.ConnectionEvent
	registerOnce        sync.Once
	stopped             int32
}

// ClientOpts provides options for the events client
type ClientOpts struct {
	// ResponseTimeout is the response timeout when communication with the server.
	ResponseTimeout time.Duration

	// Reconnect indicates whether the client should automatically attempt to reconnect
	// to the serever after a connection has been lost.
	Reconnect bool

	// MaxReconnectAttempts is the maximum number of times that the client will attempt
	// to connect to the server. If set to 0 then the client will try until it is stopped.
	MaxConnectAttempts uint

	// MaxReconnectAttempts is the maximum number of times that the client will attempt
	// to reconnect to the server after a connection has been lost. If set to 0 then the
	// client will try until it is stopped.
	MaxReconnectAttempts uint

	// TimeBetweenConnectAttempts is the time between connection attempts.
	TimeBetweenConnectAttempts time.Duration

	// ConnectEvents is the channel that is to receive connection events, i.e. when the client connects and/or
	// disconnects from the channel event service.
	ConnectEvents chan *fab.ConnectionEvent

	// connectionProvider specifies the connection provider. This is only used for unit testing.
	connectionProvider func(fab.FabricClient, *apiconfig.PeerConfig) (Connection, error)
}

// NewClientOpts returns client options set to default values
func NewClientOpts() *ClientOpts {
	return &ClientOpts{
		ResponseTimeout:            3 * time.Second,
		Reconnect:                  true,
		MaxConnectAttempts:         1,
		MaxReconnectAttempts:       0, // Try forever
		TimeBetweenConnectAttempts: 5 * time.Second,
		connectionProvider:         grpcConnectionProvider,
	}
}

// NewClient returns a new ChannelEventClient
func NewClient(fabclient fab.FabricClient, peerConfig *apiconfig.PeerConfig, channelID string) (*Client, error) {
	return NewClientWithOpts(fabclient, peerConfig, channelID, NewClientOpts())
}

// NewClientWithOpts returns a new ChannelEventClient
func NewClientWithOpts(fabclient fab.FabricClient, peerConfig *apiconfig.PeerConfig, channelID string, opts *ClientOpts) (*Client, error) {
	if peerConfig.URL == "" {
		return nil, errors.New("expecting peer URL")
	}
	if channelID == "" {
		return nil, errors.New("expecting channel ID")
	}

	dispatcher := newEventDispatcher(fabclient, peerConfig, channelID, opts.connectionProvider)

	cc := &Client{
		opts:                *opts,
		dispatcher:          dispatcher,
		connEvent:           make(chan *fab.ConnectionEvent),
		connEventSubscriber: opts.ConnectEvents,
	}

	go dispatcher.start()

	return cc, nil
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

	logger.Debugf("... channel event client is stopped\n")
}

// RegisterBlockEvent registers for block events. If the client is not authorized to receive
// block events then an error is returned.
// - eventch is the Go channel to which events are sent. Note that the events should be processed
//         in a separate Go routine so that the event dispatcher is not blocked.
func (cc *Client) RegisterBlockEvent(eventch chan<- *fab.BlockEvent) (fab.Registration, error) {
	if cc.Stopped() {
		return nil, errors.New("channel event client is closed")
	}

	respch := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(newRegisterBlockEvent(eventch, respch))
	response := <-respch

	return response.Reg, response.Err
}

// RegisterChaincodeEvent registers for chaincode events. If the client is not authorized to receive
// chaincode events then an error is returned.
// - ccID is the chaincode ID for which events are to be received
// - eventFilter is the chaincode event name for which events are to be received
// - eventch is the Go channel to which events are sent. Note that the events should be processed
//         in a separate Go routine so that the event dispatcher is not blocked.
func (cc *Client) RegisterChaincodeEvent(ccID, eventFilter string, eventch chan<- *fab.CCEvent) (fab.Registration, error) {
	if cc.Stopped() {
		return nil, errors.New("channel event client is closed")
	}
	if ccID == "" {
		return nil, errors.New("chaincode ID is required")
	}
	if eventFilter == "" {
		return nil, errors.New("event filter is required")
	}

	regExp, err := regexp.Compile(eventFilter)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid event filter [%s] for chaincode [%s]", eventFilter, ccID)
	}

	respch := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(newRegisterCCEvent(ccID, eventFilter, regExp, eventch, respch))
	response := <-respch

	return response.Reg, response.Err
}

// RegisterTxStatusEvent registers for transaction status events. If the client is not authorized to receive
// transaction status events then an error is returned.
// - txID is the transaction ID for which events are to be received
// - eventch is the Go channel to which events are sent. Note that the events should be processed
//         in a separate Go routine so that the event dispatcher is not blocked.
func (cc *Client) RegisterTxStatusEvent(txID string, eventch chan<- *fab.TxStatusEvent) (fab.Registration, error) {
	if cc.Stopped() {
		return nil, errors.New("channel event client is closed")
	}
	if txID == "" {
		return nil, errors.New("txID must be provided")
	}

	respch := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(newRegisterTxStatusEvent(txID, eventch, respch))
	response := <-respch

	return response.Reg, response.Err
}

// Unregister unregisters the given registration.
// - reg is the registration handle that was returned from one of the RegisterXXX functions
func (cc *Client) Unregister(reg fab.Registration) error {
	if cc.Stopped() {
		return errors.New("channel event client is closed")
	}

	s := <-cc.unregisterAsync(reg)
	return s.Err
}

func (cc *Client) connect() error {
	if cc.Stopped() {
		return errors.New("channel event client is closed")
	}

	respch := make(chan *connectionResponse)
	cc.dispatcher.submit(newConnectEvent(respch))
	r := <-respch
	if r.err != nil {
		return r.err
	}

	if err := cc.registerChannel(); err != nil {
		logger.Errorf("Error registering channel: %s. Disconnecting...\n", err)
		cc.dispatcher.submit(newDisconnectEvent(respch))
		response := <-respch
		if response.err != nil {
			logger.Warnf("Error attempting to disconnect: %s\n", response.err)
		}
		return err
	}

	go cc.dispatcher.submit(&connectedEvent{})

	var err error
	cc.registerOnce.Do(func() {
		_, err = cc.registerConnectionEvent(cc.connEvent)
		if err != nil {
			logger.Errorf("Error registering for connection events: %s\n", err)
			cc.Disconnect()
		}
		go cc.monitorConnection()
	})

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

func (cc *Client) registerConnectionEvent(eventch chan<- *fab.ConnectionEvent) (fab.Registration, error) {
	if cc.Stopped() {
		return nil, errors.New("channel event client is closed")
	}

	respch := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(newRegisterConnectionEvent(eventch, respch))
	response := <-respch
	return response.Reg, response.Err
}

func (cc *Client) unregisterAsync(reg fab.Registration) <-chan *fab.RegistrationResponse {
	respch := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(newUnregisterEvent(reg, respch))
	return respch
}

func (cc *Client) registerChannelAsync() <-chan *fab.RegistrationResponse {
	respch := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(newRegisterChannelEvent(respch))
	return respch
}

func (cc *Client) unregisterChannelAsync() <-chan *fab.RegistrationResponse {
	respch := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(newUnregisterChannelEvent(respch))
	return respch
}

func (cc *Client) registerChannel() error {
	logger.Debugf("registering channel....\n")

	var err error
	select {
	case s := <-cc.registerChannelAsync():
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

func (cc *Client) monitorConnection() {
	logger.Debugf("Monitoring connection\n")
	for {
		event, ok := <-cc.connEvent
		if !ok {
			break
		}

		if cc.connEventSubscriber != nil {
			logger.Debugln("Sending connection event to subscriber.")
			cc.connEventSubscriber <- event
		}

		if event.Connected {
			logger.Infof("channel event client has connected\n")
		} else if cc.opts.Reconnect {
			logger.Warnf("channel event client has disconnected. Details: %s. Attempting to reconnect...\n", event.Err)
			go cc.reconnect()
		} else {
			logger.Warnf("channel event client has disconnected. Terminating: %s\n", event.Err)
			go cc.Disconnect()
			break
		}
	}
	logger.Debugf("Exiting connection monitor\n")
}

func (cc *Client) reconnect() {
	if err := cc.connectWithRetry(cc.opts.MaxReconnectAttempts, cc.opts.TimeBetweenConnectAttempts); err != nil {
		logger.Warnf("Could not reconnect channel event client. Terminating: %s\n", err)
		cc.Disconnect()
		return
	}
}

func newConnectEvent(respch chan<- *connectionResponse) *connectEvent {
	return &connectEvent{respch: respch}
}

func newDisconnectEvent(respch chan<- *connectionResponse) *disconnectEvent {
	return &disconnectEvent{respch: respch}
}

func newRegisterChannelEvent(respch chan<- *fab.RegistrationResponse) *registerChannelEvent {
	return &registerChannelEvent{
		reg:           &channelRegistration{},
		registerEvent: registerEvent{respch: respch},
	}
}

func newRegisterBlockEvent(eventch chan<- *fab.BlockEvent, respch chan<- *fab.RegistrationResponse) *registerBlockEvent {
	return &registerBlockEvent{
		reg:           &blockRegistration{eventch: eventch},
		registerEvent: registerEvent{respch: respch},
	}
}

func newRegisterCCEvent(ccID, eventFilter string, eventRegExp *regexp.Regexp, eventch chan<- *fab.CCEvent, respch chan<- *fab.RegistrationResponse) *registerCCEvent {
	return &registerCCEvent{
		reg: &ccRegistration{
			ccID:        ccID,
			eventFilter: eventFilter,
			eventRegExp: eventRegExp,
			eventch:     eventch,
		},
		registerEvent: registerEvent{respch: respch},
	}
}

func newRegisterTxStatusEvent(txID string, eventch chan<- *fab.TxStatusEvent, respch chan<- *fab.RegistrationResponse) *registerTxStatusEvent {
	return &registerTxStatusEvent{
		reg:           &txRegistration{txID: txID, eventch: eventch},
		registerEvent: registerEvent{respch: respch},
	}
}

func newRegisterConnectionEvent(eventch chan<- *fab.ConnectionEvent, respch chan<- *fab.RegistrationResponse) *registerConnectionEvent {
	return &registerConnectionEvent{
		reg:           &connectionRegistration{eventch: eventch},
		registerEvent: registerEvent{respch: respch},
	}
}

func newUnregisterEvent(reg fab.Registration, respch chan<- *fab.RegistrationResponse) *unregisterEvent {
	return &unregisterEvent{
		reg:           reg,
		registerEvent: registerEvent{respch: respch},
	}
}

func newUnregisterChannelEvent(respch chan<- *fab.RegistrationResponse) *unregisterChannelEvent {
	return &unregisterChannelEvent{
		registerEvent: registerEvent{respch: respch},
	}
}
