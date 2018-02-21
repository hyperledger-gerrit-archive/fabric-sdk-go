/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/client/connection"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/client/lbp"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/service/dispatcher"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// Options provides options for the events client
type Options interface {
	esdispatcher.Options

	// LoadBalancePolicy is the load-balance policy to use when selecting event endpoints
	LoadBalancePolicy() lbp.LoadBalancePolicy
}

// Dispatcher is responsible for handling all events, including connection and registration events originating from the client,
// and events originating from the event server. All events are processed in a single Go routine
// in order to avoid any race conditions and to ensure that events are processed in the order that they are received.
// This avoids the need for synchronization.
type Dispatcher struct {
	esdispatcher.Dispatcher

	opts                   Options
	channelID              string
	context                apifabclient.Context
	discoveryService       apifabclient.DiscoveryService
	signingMgr             apifabclient.SigningManager
	connection             connection.Connection
	connectionRegistration *ConnectionReg
	connectionProvider     connection.Provider
}

type handler func(esdispatcher.Event)

// New creates a new dispatcher
func New(channelID string, context apifabclient.Context, connectionProvider connection.Provider, discoveryService apifabclient.DiscoveryService, opts Options) *Dispatcher {
	return &Dispatcher{
		Dispatcher:         *esdispatcher.New(opts),
		opts:               opts,
		context:            context,
		discoveryService:   discoveryService,
		channelID:          channelID,
		connectionProvider: connectionProvider,
	}
}

// Start starts the dispatcher
func (ed *Dispatcher) Start() error {
	ed.registerHandlers()

	if err := ed.Dispatcher.Start(); err != nil {
		return errors.WithMessage(err, "error starting client event dispatcher")
	}
	return nil
}

// ChannelID returns the channel ID
func (ed *Dispatcher) ChannelID() string {
	return ed.channelID
}

// Connection returns the connection to the event server
func (ed *Dispatcher) Connection() connection.Connection {
	return ed.connection
}

// HandleStopEvent handles a Stop event by clearing all registrations
// and stopping the listener
func (ed *Dispatcher) HandleStopEvent(e esdispatcher.Event) {
	// Remove all registrations and close the associated event channels
	// so that the client is notified that the registration has been removed
	ed.clearConnectionRegistration()

	ed.Dispatcher.HandleStopEvent(e)
}

// HandleConnectEvent initiates a connection to the event server
func (ed *Dispatcher) HandleConnectEvent(e esdispatcher.Event) {
	evt := e.(*ConnectEvent)

	if ed.connection != nil {
		evt.Respch <- &ConnectionResponse{}
		return
	}

	eventch, err := ed.EventCh()
	if err != nil {
		evt.Respch <- &ConnectionResponse{Err: err}
		return
	}

	peers, err := ed.discoveryService.GetPeers()
	if err != nil {
		evt.Respch <- &ConnectionResponse{Err: err}
		return
	}
	if len(peers) == 0 {
		evt.Respch <- &ConnectionResponse{Err: errors.New("no peers to connect to")}
		return
	}

	peer, err := ed.opts.LoadBalancePolicy().Choose(peers)
	if err != nil {
		evt.Respch <- &ConnectionResponse{Err: err}
		return
	}

	conn, err := ed.connectionProvider(ed.channelID, ed.context, peer)
	if err != nil {
		logger.Warnf("error creating connection: %s\n", err)
		evt.Respch <- &ConnectionResponse{Err: errors.WithMessage(err, fmt.Sprintf("could not create client conn"))}
		return
	}

	ed.connection = conn

	go ed.connection.Receive(eventch)

	evt.Respch <- &ConnectionResponse{}
}

// HandleDisconnectEvent disconnects from the event server
func (ed *Dispatcher) HandleDisconnectEvent(e esdispatcher.Event) {
	evt := e.(*DisconnectEvent)

	if ed.connection == nil {
		evt.Respch <- &ConnectionResponse{Err: errors.New("connection already closed")}
		return
	}

	logger.Infof("Closing connection...\n")

	ed.connection.Close()
	ed.connection = nil

	evt.Respch <- &ConnectionResponse{}
}

// HandleRegisterConnectionEvent registers a connection listener
func (ed *Dispatcher) HandleRegisterConnectionEvent(e esdispatcher.Event) {
	evt := e.(*RegisterConnectionEvent)

	if ed.connectionRegistration != nil {
		evt.RespCh <- esdispatcher.ErrorResponse(errors.New("registration already exists for connection event"))
		return
	}

	ed.connectionRegistration = evt.Reg
	evt.RespCh <- esdispatcher.SuccessResponse(evt.Reg)
}

// HandleConnectedEvent sends a 'connected' event to any registered listener
func (ed *Dispatcher) HandleConnectedEvent(e esdispatcher.Event) {
	evt := e.(*ConnectedEvent)

	logger.Debugf("Handling connected event: %v\n", evt)

	if ed.connectionRegistration != nil && ed.connectionRegistration.Eventch != nil {
		select {
		case ed.connectionRegistration.Eventch <- &apifabclient.ConnectionEvent{Connected: true}:
		default:
			logger.Warnf("Unable to send to connection event channel.")
		}
	}
}

// HandleDisconnectedEvent sends a 'disconnected' event to any registered listener
func (ed *Dispatcher) HandleDisconnectedEvent(e esdispatcher.Event) {
	evt := e.(*DisconnectedEvent)

	logger.Infof("Disconnecting from event server: %s\n", evt.Err)

	if ed.connection != nil {
		ed.connection.Close()
		ed.connection = nil
	}

	if ed.connectionRegistration != nil {
		logger.Debugf("Disconnected from event server: %s\n", evt.Err)
		select {
		case ed.connectionRegistration.Eventch <- &apifabclient.ConnectionEvent{Connected: false, Err: evt.Err}:
		default:
			logger.Warnf("Unable to send to connection event channel.")
		}
	} else {
		logger.Warnf("Disconnected from event server: %s\n", evt.Err)
	}
}

func (ed *Dispatcher) registerHandlers() {
	// Override existing handlers
	ed.RegisterHandler(&esdispatcher.StopEvent{}, ed.HandleStopEvent)

	// Register new handlers
	ed.RegisterHandler(&ConnectEvent{}, ed.HandleConnectEvent)
	ed.RegisterHandler(&DisconnectEvent{}, ed.HandleDisconnectEvent)
	ed.RegisterHandler(&ConnectedEvent{}, ed.HandleConnectedEvent)
	ed.RegisterHandler(&DisconnectedEvent{}, ed.HandleDisconnectedEvent)
	ed.RegisterHandler(&RegisterConnectionEvent{}, ed.HandleRegisterConnectionEvent)
}

func (ed *Dispatcher) clearConnectionRegistration() {
	if ed.connectionRegistration != nil {
		logger.Debugf("Closing connection registration event channel.\n")
		close(ed.connectionRegistration.Eventch)
		ed.connectionRegistration = nil
	}
}
