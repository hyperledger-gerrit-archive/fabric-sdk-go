// +build channelevents

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"fmt"
	"reflect"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

type eventType string

const (
	BLOCKEVENT         eventType = "BLOCKEVENT"
	FILTEREDBLOCKEVENT eventType = "FILTEREDBLOCKEVENT"
)

// Connection defines the functions for a channel event service connection
type Connection interface {
	Close()
	Send(emsg *pb.ChannelServiceRequest) error
	Receive(events chan<- interface{})
}

// ConnectionProvider is a function that creates a Connection.
type ConnectionProvider func(string, fab.FabricClient, *apiconfig.PeerConfig) (Connection, error)

// eventDispatcher is responsible for handling all events, including connection and registration events originating from the client,
// and events originating from the channel event service. All events are processed in a single Go routine
// in order to avoid any race conditions. This avoids the need for synchronization.
type eventDispatcher struct {
	handlers           map[reflect.Type]handler
	fabclient          fab.FabricClient
	peerConfig         apiconfig.PeerConfig
	channelID          string
	eventch            chan interface{}
	connection         Connection
	connectionProvider ConnectionProvider
}

type handler func(event)

// grpcConnectionProvider is the default connection provider used for creating a GRPC connection to the channel event service
var grpcConnectionProvider = func(channelID string, fabClient fab.FabricClient, peerConfig *apiconfig.PeerConfig) (Connection, error) {
	return newConnection(channelID, fabClient, peerConfig)
}

func newEventDispatcher(fabclient fab.FabricClient, peerConfig *apiconfig.PeerConfig, channelID string, connectionProvider ConnectionProvider, dispatcherChannelSize int) *eventDispatcher {
	dispatcher := &eventDispatcher{
		handlers:           make(map[reflect.Type]handler),
		fabclient:          fabclient,
		peerConfig:         *peerConfig,
		channelID:          channelID,
		eventch:            make(chan interface{}, dispatcherChannelSize),
		connectionProvider: connectionProvider,
	}
	dispatcher.registerHandlers()
	return dispatcher
}

func (ed *eventDispatcher) submit(event interface{}) {
	defer func() {
		// During shutdown, events may still be produced and we may
		// get a 'send on closed channel' panic. Just log and ignore the error.
		if p := recover(); p != nil {
			logger.Warnf("panic while submitting event: %s", p)
		}
	}()

	ed.eventch <- event
}

// start starts dispatching events as they arrive. All events are processed in
// a single Go routine in order to avoid any race conditions
func (ed *eventDispatcher) start() {
	logger.Debugf("Starting event dispatcher\n")
	for {
		logger.Debugf("Listening for events...\n")

		e, ok := <-ed.eventch
		if !ok {
			break
		}

		logger.Debugf("Received event: %v\n", reflect.TypeOf(e))

		if handler, ok := ed.handlers[reflect.TypeOf(e)]; ok {
			logger.Debugf("Dispatching event: %v\n", reflect.TypeOf(e))
			handler(e)
		} else {
			logger.Errorf("Unsupported event type: %v", reflect.TypeOf(e))
		}
	}
	logger.Debugf("Exiting event dispatcher\n")
}

func (ed *eventDispatcher) stop() {
	logger.Debugf("Closing dispatcher event channel.\n")
	close(ed.eventch)
}

func (ed *eventDispatcher) handleConnectEvent(e event) {
	event := e.(*connectEvent)

	if ed.connection != nil {
		event.respch <- &connectionResponse{}
		return
	}

	conn, err := ed.connectionProvider(ed.channelID, ed.fabclient, &ed.peerConfig)
	if err != nil {
		logger.Warnf("error creating connection: %s\n", err)
		event.respch <- &connectionResponse{err: errors.WithMessage(err, fmt.Sprintf("could not create client conn to [%s]", ed.peerConfig.URL))}
		return
	}

	ed.connection = conn

	go ed.connection.Receive(ed.eventch)

	event.respch <- &connectionResponse{}
}

func (ed *eventDispatcher) handleDisconnectEvent(e event) {
	event := e.(*disconnectEvent)

	if ed.connection == nil {
		event.respch <- &connectionResponse{err: errors.New("connection already closed")}
		return
	}

	logger.Debugf("Closing connection...\n")

	ed.connection.Close()
	ed.connection = nil

	event.respch <- &connectionResponse{}
}

func (ed *eventDispatcher) registerHandlers() {
	ed.registerHandler(&connectEvent{}, ed.handleConnectEvent)
	ed.registerHandler(&disconnectEvent{}, ed.handleDisconnectEvent)
}

func (ed *eventDispatcher) registerHandler(t interface{}, h handler) {
	ed.handlers[reflect.TypeOf(t)] = h
}
