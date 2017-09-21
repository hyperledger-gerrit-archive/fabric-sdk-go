// +build channelevents

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"fmt"
	"reflect"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

const (
	BLOCKEVENT         = "BLOCKEVENT"
	FILTEREDBLOCKEVENT = "FILTEREDBLOCKEVENT"
)

// Connection defines the functions for a channel event service connection
type Connection interface {
	Close()
	Send(emsg *pb.Event) error
	Receive(events chan<- interface{})
}

// ConnectionProvider is a function that creates a Connection.
type ConnectionProvider func(fab.FabricClient, *apiconfig.PeerConfig) (Connection, error)

// eventDispatcher is responsible for handling all events, including connection and registration events originating from the client,
// and block and filtered block events originating from the channel event service. All events are processed in a single Go routine
// in order to avoid any race conditions. This avoids the need for synchronization blocks.
type eventDispatcher struct {
	handlers               map[reflect.Type]handler
	fabclient              fab.FabricClient
	peerConfig             apiconfig.PeerConfig
	channelID              string
	event                  chan interface{}
	connection             Connection
	authorized             map[string]bool
	chRegistration         *channelRegistration
	connectionRegistration *connectionRegistration
	blockRegistration      *blockRegistration
	txRegistrations        map[string]*txRegistration
	ccRegistrations        map[string]*ccRegistration
	connectionProvider     ConnectionProvider
}

type handler func(event)

// grpcConnectionProvider is the default connection provider used for creating a GRPC connection to the channel event service
var grpcConnectionProvider = func(fabClient fab.FabricClient, peerConfig *apiconfig.PeerConfig) (Connection, error) {
	return newConnection(fabClient, peerConfig)
}

func newEventDispatcher(fabclient fab.FabricClient, peerConfig *apiconfig.PeerConfig, channelID string, connectionProvider ConnectionProvider) *eventDispatcher {
	dispatcher := &eventDispatcher{
		handlers:           make(map[reflect.Type]handler),
		fabclient:          fabclient,
		peerConfig:         *peerConfig,
		channelID:          channelID,
		event:              make(chan interface{}),
		txRegistrations:    make(map[string]*txRegistration),
		ccRegistrations:    make(map[string]*ccRegistration),
		connectionProvider: connectionProvider,
	}
	dispatcher.registerHandlers()
	return dispatcher
}

func (ed *eventDispatcher) submit(event interface{}) {
	ed.event <- event
}

// start starts dispatching events as they arrive. All events are processed in
// a single Go routine in order to avoid any race conditions
func (ed *eventDispatcher) start() {
	logger.Debugf("Starting event dispatcher\n")
	for {
		logger.Debugf("Listening for events...\n")

		e, ok := <-ed.event
		if !ok {
			break
		}

		logger.Debugf("Received event: %v\n", reflect.TypeOf(e))

		if f, ok := ed.handlers[reflect.TypeOf(e)]; ok {
			logger.Debugf("Dispatching event: %v\n", reflect.TypeOf(e))
			f(e)
		} else {
			logger.Errorf("Unsupported event type: %v", reflect.TypeOf(e))
		}
	}
	logger.Debugf("Exiting event dispatcher\n")
}

func (ed *eventDispatcher) stop() {
	// Remove all registrations and close the associated event channels
	// so that the client is notified that the registration has been removed
	ed.removeBlockRegistration()
	ed.removeAllTxRegistrations()
	ed.removeAllChaincodeRegistrations()
	ed.removeConnectionEvents()

	logger.Debugf("Closing dispatcher event channel.\n")
	close(ed.event)
}

func (ed *eventDispatcher) removeAllChaincodeRegistrations() {
	for _, reg := range ed.ccRegistrations {
		logger.Debugf("Closing chaincode registration event channel for CC ID [%s] and event filter [%s].\n", reg.ccID, reg.eventFilter)
		close(reg.eventch)
	}
	ed.ccRegistrations = make(map[string]*ccRegistration)
}

func (ed *eventDispatcher) removeAllTxRegistrations() {
	for _, reg := range ed.txRegistrations {
		logger.Debugf("Closing TX registration event channel for TxID [%s].\n", reg.txID)
		close(reg.eventch)
	}
	ed.txRegistrations = make(map[string]*txRegistration)
}

func (ed *eventDispatcher) removeBlockRegistration() {
	if ed.blockRegistration != nil {
		logger.Debugf("Closing block registration event channel.\n")
		close(ed.blockRegistration.eventch)
		ed.blockRegistration = nil
	}
}

func (ed *eventDispatcher) removeConnectionEvents() {
	if ed.connectionRegistration != nil {
		logger.Debugf("Closing connection registration event channel.\n")
		close(ed.connectionRegistration.eventch)
		ed.connectionRegistration = nil
	}
}

func (ed *eventDispatcher) handleConnectEvent(e event) {
	event := e.(*connectEvent)

	if ed.connection != nil {
		event.respch <- &connectionResponse{err: errors.New("already connected")}
		return
	}

	conn, err := ed.connectionProvider(ed.fabclient, &ed.peerConfig)
	if err != nil {
		logger.Warnf("error creating connection: %s\n", err)
		event.respch <- &connectionResponse{err: errors.WithMessage(err, fmt.Sprintf("could not create client conn to [%s]", ed.peerConfig.URL))}
		return
	}

	ed.connection = conn

	go ed.connection.Receive(ed.event)

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

	if ed.chRegistration != nil {
		event.respch <- errorResponse(errors.Errorf("registration already exists for channel [%s]", ed.channelID))
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

	if err := ed.connection.Send(ed.newRegisterChannelEvent(identity)); err != nil {
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

func (ed *eventDispatcher) handleRegisterConnectionEvent(e event) {
	event := e.(*registerConnectionEvent)

	if ed.connectionRegistration != nil {
		event.respch <- errorResponse(errors.New("registration already exists for connection event"))
		return
	}

	ed.connectionRegistration = event.reg
	event.respch <- successResponse(event.reg)
}

func (ed *eventDispatcher) handleRegisterBlockEvent(e event) {
	event := e.(*registerBlockEvent)

	if !ed.authorized[BLOCKEVENT] {
		event.respch <- errorResponse(errors.New("client not authorized to receive block events"))
	} else if ed.blockRegistration != nil {
		event.respch <- errorResponse(errors.New("registration already exists for block event"))
	} else {
		ed.blockRegistration = event.reg
		event.respch <- successResponse(event.reg)
	}
}

func (ed *eventDispatcher) handleRegisterCCEvent(e event) {
	event := e.(*registerCCEvent)

	key := getKey(event.reg.ccID, event.reg.eventFilter)
	if !ed.authorized[FILTEREDBLOCKEVENT] {
		event.respch <- errorResponse(errors.New("client not authorized to receive chaincode events"))
	} else if _, exists := ed.ccRegistrations[key]; exists {
		event.respch <- errorResponse(errors.Errorf("registration already exists for chaincode [%s] and event [%s]", event.reg.ccID, event.reg.eventFilter))
	} else {
		ed.ccRegistrations[key] = event.reg
		event.respch <- successResponse(event.reg)
	}
}

func (ed *eventDispatcher) handleRegisterTxStatusEvent(e event) {
	event := e.(*registerTxStatusEvent)

	if !ed.authorized[FILTEREDBLOCKEVENT] {
		event.respch <- errorResponse(errors.New("client not authorized to receive TX events"))
	} else if _, exists := ed.txRegistrations[event.reg.txID]; exists {
		event.respch <- errorResponse(errors.Errorf("registration already exists for TX ID [%s]", event.reg.txID))
	} else {
		ed.txRegistrations[event.reg.txID] = event.reg
		event.respch <- successResponse(event.reg)
	}
}

func (ed *eventDispatcher) handleUnregisterEvent(e event) {
	event := e.(*unregisterEvent)

	switch registration := event.reg.(type) {
	case *blockRegistration:
		event.respch <- unregResponse(ed.unregisterBlockEvents(registration))
	case *ccRegistration:
		event.respch <- unregResponse(ed.unregisterCCEvents(registration))
	case *txRegistration:
		event.respch <- unregResponse(ed.unregisterTXEvents(registration))
	default:
		event.respch <- errorResponse(errors.Errorf("unsupported registration type: %v", reflect.TypeOf(registration)))
	}
}

func (ed *eventDispatcher) handleServiceResponse(e event) {
	event := e.(*pb.Event_ChannelServiceResponse)

	resp := event.ChannelServiceResponse

	if resp.Action == "RegisterChannel" {
		ed.handleRegisterChannelResponse(resp)
	} else if resp.Action == "DeregisterChannel" {
		ed.handleUnregisterChannelResponse(resp)
	} else {
		logger.Warnf("unsupported ChannelServiceResponse action [%s]\n", resp.Action)
	}
}

func (ed *eventDispatcher) handleRegisterChannelResponse(resp *pb.ChannelServiceResponse) {
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

	ed.authorized = make(map[string]bool)
	if !resp.Success {
		if ed.chRegistration.respch != nil {
			ed.chRegistration.respch <- errorResponse(errors.Errorf("failed to register for channel events. ErrorMsg [%s]", result.ErrorMsg))
		}
		ed.chRegistration = nil
		return
	}

	for _, authEvent := range result.AuthorizedEvents {
		logger.Debugf("Authorized Event [%s]\n", authEvent)
		ed.authorized[authEvent] = true
	}

	// Check existing registrations to ensure they're valid, otherwise remove the registrations
	if ed.blockRegistration != nil && !ed.authorized[BLOCKEVENT] {
		logger.Warnf("Client not authorized to receive block events. The block registration will be removed.")
		ed.removeBlockRegistration()
	}

	if len(ed.ccRegistrations) > 0 && !ed.authorized[FILTEREDBLOCKEVENT] {
		logger.Warnf("Client not authorized to receive filtered block events. All chaincode and transaction events will be removed.")
		ed.removeAllChaincodeRegistrations()
		ed.removeAllTxRegistrations()
	}

	if ed.chRegistration.respch != nil {
		ed.chRegistration.respch <- successResponse(nil)

		// Set the event to nil since we've already responded
		ed.chRegistration.respch = nil
	}
}

func (ed *eventDispatcher) getChannelServiceResult(resp *pb.ChannelServiceResponse) *pb.ChannelServiceResult {
	for _, result := range resp.ChannelServiceResults {
		if result.ChannelId == ed.channelID {
			return result
		}
	}
	return nil
}

func (ed *eventDispatcher) handleUnregisterChannelResponse(resp *pb.ChannelServiceResponse) {
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

func (ed *eventDispatcher) handleFilteredBlockEvent(e event) {
	event := e.(*pb.Event_FilteredBlock)

	logger.Debugf("Handling event: %v\n", event)

	if event.FilteredBlock == nil || event.FilteredBlock.FilteredTx == nil {
		logger.Errorf("Received invalid filtered block event: %s", event)
		return
	}

	for _, tx := range event.FilteredBlock.FilteredTx {
		ed.triggerTxStatusEvent(tx)

		// Only send a chaincode event if the transaction has committed
		if tx.TxValidationCode == pb.TxValidationCode_VALID && tx.CcEvent != nil {
			ed.triggerCCEvent(tx.CcEvent)
		}
	}
}

func (ed *eventDispatcher) handleBlockEvent(e event) {
	event := e.(*pb.Event_Block)

	logger.Debugf("Handling event: %v\n", event)

	if ed.blockRegistration != nil {
		ed.blockRegistration.eventch <- &fab.BlockEvent{Block: event.Block}
	}
}

func (ed *eventDispatcher) handleConnectedEvent(e event) {
	event := e.(*connectedEvent)

	logger.Debugf("Handling event: %v\n", event)

	if ed.connectionRegistration != nil && ed.connectionRegistration.eventch != nil {
		ed.connectionRegistration.eventch <- &apifabclient.ConnectionEvent{Connected: true}
	}
}

func (ed *eventDispatcher) handleDisconnectedEvent(e event) {
	event := e.(*disconnectedEvent)

	if ed.connection != nil {
		ed.connection.Close()
		ed.connection = nil
	}

	if ed.chRegistration != nil && ed.chRegistration.respch != nil {
		// We're in the middle of a channel registration. Send an error response to the caller.
		ed.chRegistration.respch <- errorResponse(errors.New("connection terminated"))
	}

	ed.chRegistration = nil
	ed.authorized = nil

	if ed.connectionRegistration != nil {
		logger.Debugf("Disconnected from channel service: %s\n", event.err)
		ed.connectionRegistration.eventch <- &fab.ConnectionEvent{
			Connected: false,
			Err:       event.err,
		}
	} else {
		logger.Warnf("Disconnected from channel service: %s\n", event.err)
	}
}

func (ed *eventDispatcher) unregisterBlockEvents(registration *blockRegistration) error {
	if ed.blockRegistration != registration {
		return errors.New("the provided registration is invalid")
	}
	ed.blockRegistration = nil
	return nil
}

func (ed *eventDispatcher) unregisterCCEvents(registration *ccRegistration) error {
	key := getKey(registration.ccID, registration.eventFilter)
	if _, ok := ed.ccRegistrations[key]; !ok {
		return errors.New("the provided registration is invalid")
	}
	delete(ed.ccRegistrations, key)
	return nil
}

func (ed *eventDispatcher) unregisterTXEvents(registration *txRegistration) error {
	if _, ok := ed.txRegistrations[registration.txID]; !ok {
		return errors.New("the provided registration is invalid")
	}

	logger.Debugf("Unregistering Tx Status event for TxID [%s]...\n", registration.txID)
	delete(ed.txRegistrations, registration.txID)
	return nil
}

func (ed *eventDispatcher) triggerTxStatusEvent(tx *pb.FilteredTransaction) {
	logger.Debugf("Triggering Tx Status event for TxID [%s]...\n", tx.Txid)
	if reg, ok := ed.txRegistrations[tx.Txid]; ok {
		logger.Debugf("Sending Tx Status event for TxID [%s] to registrant...\n", tx.Txid)
		reg.eventch <- newTxStatusEvent(tx.Txid, tx.TxValidationCode)
	}
}

func (ed *eventDispatcher) triggerCCEvent(ccEvent *pb.ChaincodeEvent) {
	for _, reg := range ed.ccRegistrations {
		logger.Debugf("Matching CCEvent[%s,%s] against Reg[%s,%s] ...\n", ccEvent.ChaincodeId, ccEvent.EventName, reg.ccID, reg.eventFilter)
		if reg.ccID == ccEvent.ChaincodeId && reg.eventRegExp.MatchString(ccEvent.EventName) {
			logger.Debugf("... matched CCEvent[%s,%s] against Reg[%s,%s]\n", ccEvent.ChaincodeId, ccEvent.EventName, reg.ccID, reg.eventFilter)
			reg.eventch <- newCCEvent(ccEvent.ChaincodeId, ccEvent.EventName, ccEvent.TxId)
		}
	}
}

func (ed *eventDispatcher) registerHandlers() {
	ed.registerHandler(&connectEvent{}, ed.handleConnectEvent)
	ed.registerHandler(&disconnectEvent{}, ed.handleDisconnectEvent)
	ed.registerHandler(&connectedEvent{}, ed.handleConnectedEvent)
	ed.registerHandler(&disconnectedEvent{}, ed.handleDisconnectedEvent)
	ed.registerHandler(&registerChannelEvent{}, ed.handleRegisterChannelEvent)
	ed.registerHandler(&unregisterChannelEvent{}, ed.handleUnregisterChannelEvent)
	ed.registerHandler(&registerConnectionEvent{}, ed.handleRegisterConnectionEvent)
	ed.registerHandler(&registerCCEvent{}, ed.handleRegisterCCEvent)
	ed.registerHandler(&registerTxStatusEvent{}, ed.handleRegisterTxStatusEvent)
	ed.registerHandler(&registerBlockEvent{}, ed.handleRegisterBlockEvent)
	ed.registerHandler(&unregisterEvent{}, ed.handleUnregisterEvent)
	ed.registerHandler(&pb.Event_ChannelServiceResponse{}, ed.handleServiceResponse)
	ed.registerHandler(&pb.Event_FilteredBlock{}, ed.handleFilteredBlockEvent)
	ed.registerHandler(&pb.Event_Block{}, ed.handleBlockEvent)
}

func (ed *eventDispatcher) registerHandler(t interface{}, h handler) {
	ed.handlers[reflect.TypeOf(t)] = h
}

func (ed *eventDispatcher) newRegisterChannelEvent(identity []byte) *pb.Event {
	return &pb.Event{
		Event: &pb.Event_RegisterChannel{
			RegisterChannel: &pb.RegisterChannel{
				ChannelIds: []string{ed.channelID},
			},
		},
		Creator: identity,
	}
}

func (ed *eventDispatcher) newUnregisterChannelEvent(identity []byte) *pb.Event {
	return &pb.Event{
		Event: &pb.Event_DeregisterChannel{
			DeregisterChannel: &pb.DeregisterChannel{
				ChannelIds: []string{ed.channelID},
			},
		},
		Creator: identity,
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

func newTxStatusEvent(txID string, txValidationCode pb.TxValidationCode) *fab.TxStatusEvent {
	return &fab.TxStatusEvent{
		TxID:             txID,
		TxValidationCode: txValidationCode,
	}
}

func newCCEvent(chaincodeID, eventName, txID string) *fab.CCEvent {
	return &fab.CCEvent{
		ChaincodeID: chaincodeID,
		EventName:   eventName,
		TxID:        txID,
	}
}
