// +build experimental

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"fmt"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

// Client connects to a peer and receives channel events, such as block, chaincode, and transaction status events.
type Client struct {
	opts         ClientOpts
	dispatcher   *eventDispatcher
	connEvent    chan *fab.ConnectionEvent
	registerOnce sync.Once
	stopped      int32
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
}

// NewClientOpts returns client options set to default values
func NewClientOpts() ClientOpts {
	return ClientOpts{
		ResponseTimeout:            3 * time.Second,
		Reconnect:                  true,
		MaxConnectAttempts:         1,
		MaxReconnectAttempts:       0, // Try forever
		TimeBetweenConnectAttempts: 5 * time.Second,
	}
}

// NewClient returns a new ChannelEventClient
func NewClient(fabclient fab.FabricClient, peerConfig apiconfig.PeerConfig, channelID string) (*Client, error) {
	return NewClientWithOpts(fabclient, peerConfig, channelID, NewClientOpts())
}

// NewClientWithOpts returns a new ChannelEventClient
func NewClientWithOpts(fabclient fab.FabricClient, peerConfig apiconfig.PeerConfig, channelID string, opts ClientOpts) (*Client, error) {
	if peerConfig.Url == "" {
		return nil, fmt.Errorf("expecting peer URL")
	}
	if channelID == "" {
		return nil, fmt.Errorf("expecting channel ID")
	}

	dispatcher := newEventDispatcher(fabclient, peerConfig, channelID)

	cc := &Client{
		opts:       opts,
		dispatcher: dispatcher,
		connEvent:  make(chan *fab.ConnectionEvent),
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
		logger.Debugf("client already stopped\n")
		return
	}

	logger.Debugf("stopping client...\n")

	select {
	case resp := <-cc.unregisterChannelAsync():
		if resp.Err != nil {
			logger.Warningf("error in unregister channel: %s", resp.Err)
		} else {
			logger.Debugf("channel unregistered\n")
		}
	case <-time.After(cc.opts.ResponseTimeout):
		logger.Warningf("timed out waiting for unregister channel response")
	}

	logger.Debugf("Sending disconnect request...\n")

	response := make(chan *connectionResponse)
	cc.dispatcher.submit(&disconnectEvent{response: response})
	r := <-response

	if r.err != nil {
		logger.Warningf("Received error from disconnect request: %s\n", r.err)
	} else {
		logger.Debugf("Received success from disconnect request\n")
	}

	logger.Debugf("Stopping dispatcher...\n")

	cc.dispatcher.stop()

	logger.Debugf("... client is stopped\n")
}

// RegisterBlockEvent registers for block events. If the client is not authorized to receive
// block events then an error is returned.
// - eventChan is the Go channel to which events are sent. Note that the events should be processed
//         in a separate Go routine so that the event dispatcher is not blocked.
func (cc *Client) RegisterBlockEvent(blockChan chan<- *fab.BlockEvent) (fab.Registration, error) {
	if cc.Stopped() {
		return nil, fmt.Errorf("channel event client is closed")
	}

	response := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(&registerBlockEvent{
		reg:           &blockRegistration{event: blockChan},
		registerEvent: registerEvent{response: response},
	})

	s := <-response
	return s.Reg, s.Err
}

// RegisterChaincodeEvent registers for chaincode events. If the client is not authorized to receive
// chaincode events then an error is returned.
// - ccID is the chaincode ID for which events are to be received
// - eventFilter is the chaincode event name for which events are to be received
// - eventChan is the Go channel to which events are sent. Note that the events should be processed
//         in a separate Go routine so that the event dispatcher is not blocked.
func (cc *Client) RegisterChaincodeEvent(ccID, eventFilter string, event chan<- *fab.CCEvent) (fab.Registration, error) {
	if cc.Stopped() {
		return nil, fmt.Errorf("channel event client is closed")
	}

	if ccID == "" {
		return nil, fmt.Errorf("chaincode ID must be provided")
	}

	if eventFilter == "" {
		return nil, fmt.Errorf("event filter must be provided")
	}

	regExp, err := regexp.Compile(eventFilter)
	if err != nil {
		return nil, fmt.Errorf("invalid event filter [%s] for chaincode [%s]: %s", eventFilter, ccID, err)
	}

	response := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(&registerCCEvent{
		reg: &ccRegistration{
			ccID:        ccID,
			eventFilter: eventFilter,
			eventRegExp: regExp,
			event:       event,
		},
		registerEvent: registerEvent{response: response},
	})

	s := <-response
	return s.Reg, s.Err
}

// RegisterTxStatusEvent registers for transaction status events. If the client is not authorized to receive
// transaction status events then an error is returned.
// - txID is the transaction ID for which events are to be received
// - eventChan is the Go channel to which events are sent. Note that the events should be processed
//         in a separate Go routine so that the event dispatcher is not blocked.
func (cc *Client) RegisterTxStatusEvent(txID string, event chan<- *fab.TxStatusEvent) (fab.Registration, error) {
	if cc.Stopped() {
		return nil, fmt.Errorf("channel event client is closed")
	}

	if txID == "" {
		return nil, fmt.Errorf("txID must be provided")
	}

	response := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(&registerTxStatusEvent{
		reg:           &txRegistration{txID: txID, event: event},
		registerEvent: registerEvent{response: response},
	})

	s := <-response
	return s.Reg, s.Err
}

// Unregister unregisters the given registration.
// - reg is the registration handle that was returned from one of the RegisterXXX functions
func (cc *Client) Unregister(reg fab.Registration) error {
	if cc.Stopped() {
		return fmt.Errorf("channel event client is closed")
	}

	s := <-cc.unregisterAsync(reg)
	return s.Err
}

func (cc *Client) connect() error {
	if cc.Stopped() {
		return fmt.Errorf("channel event client is closed")
	}

	response := make(chan *connectionResponse)
	cc.dispatcher.submit(&connectEvent{response: response})
	r := <-response
	if r.err != nil {
		return r.err
	}

	if err := cc.registerChannel(); err != nil {
		logger.Errorf("Error registering channel: %s. Disconnecting...\n", err)
		cc.dispatcher.submit(&disconnectEvent{response: response})
		resp := <-response
		if resp.err != nil {
			logger.Warningf("Error attempting to disconnect: %s\n", resp.err)
		}
		return err
	}

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
		return fmt.Errorf("channel event client is closed")
	}

	if timeBetweenAttempts < time.Second {
		timeBetweenAttempts = time.Second
	}

	var attempts uint
	for {
		attempts++
		logger.Debugf("Attempt #%d to connect...\n", attempts)
		if err := cc.connect(); err != nil {
			logger.Warningf("... connection attempt failed: %s\n", err)
			if maxAttempts > 0 && attempts > maxAttempts {
				logger.Warningf("maximum connect attempts exceeded\n")
				return fmt.Errorf("maximum connect attempts exceeded")
			}
			time.Sleep(timeBetweenAttempts)
		} else {
			logger.Debugf("... connect succeeded.\n")
			return nil
		}
	}
}

func (cc *Client) registerConnectionEvent(event chan<- *fab.ConnectionEvent) (fab.Registration, error) {
	if cc.Stopped() {
		return nil, fmt.Errorf("channel event client is closed")
	}

	response := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(&registerConnectionEvent{
		reg:           &connectionRegistration{event: event},
		registerEvent: registerEvent{response: response},
	})

	s := <-response
	return s.Reg, s.Err
}

func (cc *Client) unregisterAsync(reg fab.Registration) <-chan *fab.RegistrationResponse {
	response := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(&unregisterEvent{reg: reg, registerEvent: registerEvent{response: response}})
	return response
}

func (cc *Client) registerChannelAsync() <-chan *fab.RegistrationResponse {
	response := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(&registerChannelEvent{
		reg:           &channelRegistration{event: response},
		registerEvent: registerEvent{response: response},
	})
	return response
}

func (cc *Client) unregisterChannelAsync() <-chan *fab.RegistrationResponse {
	response := make(chan *fab.RegistrationResponse)
	cc.dispatcher.submit(&unregisterChannelEvent{registerEvent: registerEvent{response: response}})
	return response
}

func (cc *Client) registerChannel() error {
	logger.Debugf("registering channel....\n")

	var err error
	select {
	case s := <-cc.registerChannelAsync():
		err = s.Err
	case <-time.After(cc.opts.ResponseTimeout):
		err = fmt.Errorf("timeout waiting for channel registration response")
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
			logger.Debugf("Exiting connection monitor\n")
			return
		}

		if event.Connected {
			logger.Infof("Channel event client has reconnected\n")
		} else if cc.opts.Reconnect {
			logger.Warningf("Channel event client has disconnected. Details: %s. Attempting to reconnect...\n", event.Err)
			if err := cc.connectWithRetry(cc.opts.MaxReconnectAttempts, cc.opts.TimeBetweenConnectAttempts); err != nil {
				logger.Warningf("Could not reconnect channel event client. Terminating: %s\n", err)
				cc.Disconnect()
				return
			}
		} else {
			logger.Warningf("Channel event client has disconnected. Terminating: %s\n", event.Err)
			cc.Disconnect()
			return
		}
	}
}
