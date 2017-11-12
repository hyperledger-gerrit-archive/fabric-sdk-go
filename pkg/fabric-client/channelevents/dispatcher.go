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

const (
	BLOCKEVENT         eventType = "BLOCKEVENT"
	FILTEREDBLOCKEVENT eventType = "FILTEREDBLOCKEVENT"

	REGISTER_CHANNEL_ACTION   = "RegisterChannel"
	DEREGISTER_CHANNEL_ACTION = "DeregisterChannel"
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
	handlers                  map[reflect.Type]handler
	fabclient                 fab.FabricClient
	peerConfig                apiconfig.PeerConfig
	channelID                 string
	eventch                   chan interface{}
	connection                Connection
	authorized                map[eventType]bool
	eventTypes                []eventType
	chRegistration            *channelRegistration
	filteredBlockRegistration *filteredBlockRegistration
	connectionProvider        ConnectionProvider
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
	// Remove all registrations and close the associated event channels
	// so that the client is notified that the registration has been removed
	ed.clearFilteredBlockRegistration()

	logger.Debugf("Closing dispatcher event channel.\n")
	close(ed.eventch)
}

func (ed *eventDispatcher) clearFilteredBlockRegistration() {
	if ed.filteredBlockRegistration != nil {
		logger.Debugf("Closing filtered block registration event channel.\n")
		close(ed.filteredBlockRegistration.eventch)
		ed.filteredBlockRegistration = nil
	}
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

func (ed *eventDispatcher) handleRegisterChannelEvent(e event) {
	event := e.(*registerChannelEvent)

	if ed.connection == nil {
		logger.Warnf("Unable to register channel since no connection was established.")
		return
	}

	if ed.fabclient.UserContext() == nil {
		event.respch <- errorResponse(errors.New("user context not set"))
		return
	}

	identity, err := ed.fabclient.UserContext().Identity()
	if err != nil {
		event.respch <- errorResponse(errors.Wrap(err, "error getting signing identity"))
		return
	}

	if ed.chRegistration != nil {
		logger.Debugf("Already register events for channel [%s] and event types %v\n", ed.channelID, ed.eventTypes)
		event.respch <- successResponse(ed.chRegistration)
		return
	} else if event.eventTypes != nil {
		// First time registering
		ed.eventTypes = event.eventTypes
		logger.Debugf("Sending register events for channel [%s] and event types %v\n", ed.channelID, ed.eventTypes)
	} else {
		// No events to register for
		logger.Debugf("No events to register for on channel [%s]\n", ed.channelID)
		event.respch <- errorResponse(errors.Wrapf(err, "no events to register on channel [%s]", ed.channelID))
		return
	}

	if err := ed.connection.Send(ed.newRegisterChannelEvent(ed.eventTypes, identity)); err != nil {
		event.respch <- errorResponse(errors.Wrapf(err, "error sending register event for channel [%s]", ed.channelID))
	} else {
		ed.chRegistration = event.reg
		ed.chRegistration.respch = event.respch
	}
}

func (ed *eventDispatcher) handleUnregisterChannelEvent(e event) {
	event := e.(*unregisterChannelEvent)

	if ed.connection == nil {
		event.respch <- errorResponse(errors.New("not connected"))
		return
	}

	if ed.chRegistration == nil {
		event.respch <- errorResponse(errors.Errorf("invalid registration for channel [%s]", ed.channelID))
		return
	}

	identity, err := ed.fabclient.UserContext().Identity()
	if err != nil {
		event.respch <- errorResponse(errors.Wrap(err, "error getting signing identity"))
		return
	}

	if err := ed.connection.Send(ed.newUnregisterChannelEvent(identity)); err != nil {
		event.respch <- errorResponse(errors.Wrapf(err, "error sending deregister event for channel [%s]", ed.channelID))
	} else {
		ed.chRegistration.respch = event.respch
	}
}

func (ed *eventDispatcher) handleRegisterFilteredBlockEvent(e event) {
	event := e.(*registerFilteredBlockEvent)

	if !ed.authorized[FILTEREDBLOCKEVENT] {
		event.respch <- errorResponse(errors.New("client not authorized to receive filtered block events"))
	} else if ed.filteredBlockRegistration != nil {
		event.respch <- errorResponse(errors.New("registration already exists for filtered block event"))
	} else {
		ed.filteredBlockRegistration = event.reg
		event.respch <- successResponse(event.reg)
	}
}

func (ed *eventDispatcher) handleUnregisterEvent(e event) {
	event := e.(*unregisterEvent)

	var err error
	switch registration := event.reg.(type) {
	case *filteredBlockRegistration:
		err = ed.unregisterFilteredBlockEvents(registration)
	default:
		err = errors.Errorf("Unsupported registration type: %v", reflect.TypeOf(registration))
	}
	if err != nil {
		logger.Warnf("Error in unregister: %s\n", err)
	}
}

func (ed *eventDispatcher) handleServiceResponse(e event) {
	event := e.(*pb.ChannelServiceResponse_Result)

	resp := event.Result

	if resp.Action == REGISTER_CHANNEL_ACTION {
		ed.handleRegisterChannelResponse(resp)
	} else if resp.Action == DEREGISTER_CHANNEL_ACTION {
		ed.handleUnregisterChannelResponse(resp)
	} else {
		logger.Warnf("unsupported ChannelServiceResponse action [%s]\n", resp.Action)
	}
}

func (ed *eventDispatcher) handleRegisterChannelResponse(resp *pb.ChannelServiceResult) {
	logger.Debugf("Handling response: %v\n", resp)

	if ed.chRegistration == nil {
		logger.Warnf("Unexpected service response for nil channel registration\n")
		return
	}

	result := ed.getChannelServiceResult(resp)
	if result == nil {
		if ed.chRegistration.respch != nil {
			ed.chRegistration.respch <- errorResponse(errors.Errorf("no channel registration result found for channel [%s] in channel service response", ed.channelID))
		}
		ed.chRegistration = nil
		return
	}

	ed.authorized = make(map[eventType]bool)
	if !resp.Success {
		if ed.chRegistration.respch != nil {
			ed.chRegistration.respch <- errorResponse(errors.Errorf("failed to register for channel events. ErrorMsg [%s]", result.ErrorMsg))
		}
		ed.chRegistration = nil
		return
	}

	for _, authEvent := range result.RegisteredEvents {
		logger.Debugf("Registered Event [%s]\n", authEvent)
		ed.authorized[eventType(authEvent)] = true
	}

	// Check existing registrations to ensure they're still valid, otherwise remove the registrations
	if ed.filteredBlockRegistration != nil && !ed.authorized[FILTEREDBLOCKEVENT] {
		logger.Warnf("Client not authorized to receive filtered block events. Filtered block registration will be removed.")
		ed.clearFilteredBlockRegistration()
	}

	if ed.chRegistration.respch != nil {
		ed.chRegistration.respch <- successResponse(nil)

		// Set the event to nil since we've already responded
		ed.chRegistration.respch = nil
	}
}

func (ed *eventDispatcher) getChannelServiceResult(resp *pb.ChannelServiceResult) *pb.ChannelResult {
	for _, result := range resp.ChannelResults {
		if result.ChannelId == ed.channelID {
			return result
		}
	}
	return nil
}

func (ed *eventDispatcher) handleUnregisterChannelResponse(resp *pb.ChannelServiceResult) {
	logger.Debugf("Handling response: %v\n", resp)

	if ed.chRegistration == nil || ed.chRegistration.respch == nil {
		logger.Warnf("Unexpected service response for nil channel registration\n")
		return
	}

	result := ed.getChannelServiceResult(resp)
	if result == nil {
		ed.chRegistration.respch <- errorResponse(errors.Errorf("no channel unregistration result found for channel [%s] in channel service response", ed.channelID))
	} else if !resp.Success {
		ed.chRegistration.respch <- errorResponse(errors.Errorf("failed to unregister for channel events. ErrorMsg [%s]", result.ErrorMsg))
	} else {
		ed.chRegistration.respch <- successResponse(nil)
	}

	ed.chRegistration = nil
}

func (ed *eventDispatcher) handleChannelEvent(e event) {
	event := e.(*pb.ChannelServiceResponse_Event)

	switch evt := event.Event.Event.(type) {
	case *pb.Event_FilteredBlock:
		ed.handleFilteredBlockEvent(evt)
	default:
		logger.Warnf("Unsupported event type: %v", reflect.TypeOf(event.Event.Event))
	}
}

func (ed *eventDispatcher) handleFilteredBlockEvent(event *pb.Event_FilteredBlock) {
	logger.Debugf("Handling filtered block event: %v\n", event)

	if event.FilteredBlock == nil || event.FilteredBlock.FilteredTx == nil {
		logger.Errorf("Received invalid filtered block event: %s", event)
		return
	}

	if ed.filteredBlockRegistration != nil {
		select {
		case ed.filteredBlockRegistration.eventch <- &fab.FilteredBlockEvent{FilteredBlock: event.FilteredBlock}:
		default:
			logger.Warnf("Unable to send to filtered block event channel.")
		}
	}
}

func (ed *eventDispatcher) unregisterFilteredBlockEvents(registration *filteredBlockRegistration) error {
	if ed.filteredBlockRegistration == nil {
		return errors.New("no filtered block registration found")
	}
	if ed.filteredBlockRegistration != registration {
		return errors.New("the provided registration is invalid")
	}
	close(ed.filteredBlockRegistration.eventch)
	ed.filteredBlockRegistration = nil
	return nil
}

func (ed *eventDispatcher) registerHandlers() {
	ed.registerHandler(&connectEvent{}, ed.handleConnectEvent)
	ed.registerHandler(&disconnectEvent{}, ed.handleDisconnectEvent)
	ed.registerHandler(&registerChannelEvent{}, ed.handleRegisterChannelEvent)
	ed.registerHandler(&unregisterChannelEvent{}, ed.handleUnregisterChannelEvent)
	ed.registerHandler(&registerFilteredBlockEvent{}, ed.handleRegisterFilteredBlockEvent)
	ed.registerHandler(&unregisterEvent{}, ed.handleUnregisterEvent)
	ed.registerHandler(&pb.ChannelServiceResponse_Result{}, ed.handleServiceResponse)
	ed.registerHandler(&pb.ChannelServiceResponse_Event{}, ed.handleChannelEvent)
}

func (ed *eventDispatcher) registerHandler(t interface{}, h handler) {
	ed.handlers[reflect.TypeOf(t)] = h
}

func (ed *eventDispatcher) newRegisterChannelEvent(eventTypes []eventType, identity []byte) *pb.ChannelServiceRequest {
	var interestedEvents []*pb.Interest
	for _, eventType := range eventTypes {
		if eventType == BLOCKEVENT {
			interestedEvents = append(interestedEvents, &pb.Interest{EventType: pb.EventType_BLOCK})
		} else if eventType == FILTEREDBLOCKEVENT {
			interestedEvents = append(interestedEvents, &pb.Interest{EventType: pb.EventType_FILTEREDBLOCK})
		}
	}

	return &pb.ChannelServiceRequest{
		Request: &pb.ChannelServiceRequest_RegisterChannel{
			RegisterChannel: &pb.RegisterChannel{
				ChannelIds: []string{ed.channelID},
				Events:     interestedEvents,
			},
		},
	}
}

func (ed *eventDispatcher) newUnregisterChannelEvent(identity []byte) *pb.ChannelServiceRequest {
	return &pb.ChannelServiceRequest{
		Request: &pb.ChannelServiceRequest_DeregisterChannel{
			DeregisterChannel: &pb.DeregisterChannel{
				ChannelIds: []string{ed.channelID},
			},
		},
	}
}

func successResponse(reg fab.Registration) *fab.RegistrationResponse {
	return &fab.RegistrationResponse{Reg: reg}
}

func errorResponse(err error) *fab.RegistrationResponse {
	return &fab.RegistrationResponse{Err: err}
}

func unregResponse(err error) *fab.RegistrationResponse {
	return &fab.RegistrationResponse{Err: err}
}

func getKey(ccID, eventFilter string) string {
	return ccID + "/" + eventFilter
}
