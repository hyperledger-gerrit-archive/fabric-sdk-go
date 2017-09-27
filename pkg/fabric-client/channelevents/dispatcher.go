// +build channelevents

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"fmt"
	"reflect"

	"github.com/cloudflare/cfssl/log"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

const (
	BLOCKEVENT         = "BLOCKEVENT"
	FILTEREDBLOCKEVENT = "FILTEREDBLOCKEVENT"
)

type handler func(event)

type eventDispatcher struct {
	handlers               map[reflect.Type]handler
	fabclient              fab.FabricClient
	peerConfig             apiconfig.PeerConfig
	channelID              string
	event                  chan interface{}
	connection             *connection
	authorized             map[string]bool
	chRegistration         *channelRegistration
	connectionRegistration *connectionRegistration
	blockRegistration      *blockRegistration
	txRegistrations        map[string]*txRegistration
	ccRegistrations        map[string]*ccRegistration
}

func newEventDispatcher(fabclient fab.FabricClient, peerConfig apiconfig.PeerConfig, channelID string) *eventDispatcher {
	dispatcher := &eventDispatcher{
		handlers:        make(map[reflect.Type]handler),
		fabclient:       fabclient,
		peerConfig:      peerConfig,
		channelID:       channelID,
		event:           make(chan interface{}),
		txRegistrations: make(map[string]*txRegistration),
		ccRegistrations: make(map[string]*ccRegistration),
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
	for {
		logger.Debugf("Listening for events...\n")

		e, ok := <-ed.event
		if !ok {
			logger.Debugf("Exiting dispatch loop\n")
			return
		}

		logger.Debugf("Received event: %v\n", reflect.TypeOf(e))

		if f, ok := ed.handlers[reflect.TypeOf(e)]; ok {
			logger.Debugf("Dispatching event: %v\n", reflect.TypeOf(e))
			f(e)
		} else {
			logger.Debugf("Unsupported event type: %v", reflect.TypeOf(e))
		}
	}
}

func (ed *eventDispatcher) stop() {
	// Close the channels in all registrations

	if ed.blockRegistration != nil {
		logger.Debugf("Closing block registration event channel.\n")
		close(ed.blockRegistration.event)
	}
	for _, reg := range ed.txRegistrations {
		logger.Debugf("Closing TX registration event channel for TxID %s.\n", reg.txID)
		close(reg.event)
	}
	for _, reg := range ed.ccRegistrations {
		logger.Debugf("Closing chaincode registration event channel for CC ID [%s] and event filter '%s'.\n", reg.ccID, reg.eventFilter)
		close(reg.event)
	}

	if ed.connectionRegistration != nil {
		logger.Debugf("Closing connection registration event channel.\n")
		close(ed.connectionRegistration.event)
	}

	logger.Debugf("Closing dispatcher event channel.\n")
	close(ed.event)
}

func (ed *eventDispatcher) handleConnectEvent(e event) {
	event := e.(*connectEvent)

	if ed.connection != nil {
		event.response <- &connectionResponse{err: fmt.Errorf("already connected")}
		return
	}

	conn, err := newConnection(ed.fabclient, ed.peerConfig)
	if err != nil {
		logger.Warningf("error creating connection: %s\n", err)
		event.response <- &connectionResponse{err: fmt.Errorf("could not create client conn to %s:%s", ed.peerConfig.Url, err)}
		return
	}

	ed.connection = conn

	go conn.receive(ed.event)

	event.response <- &connectionResponse{}
}

func (ed *eventDispatcher) handleDisconnectEvent(e event) {
	event := e.(*disconnectEvent)

	if ed.connection == nil {
		event.response <- &connectionResponse{err: fmt.Errorf("connection already closed")}
		return
	}

	log.Debugf("Closing connection...\n")

	if err := ed.connection.close(); err != nil {
		// Ignore
		logger.Warningf("Error closing connection: %s", err)
	}

	ed.connection = nil
	event.response <- &connectionResponse{}
}

func (ed *eventDispatcher) handleRegisterChannelEvent(e event) {
	event := e.(*registerChannelEvent)

	if ed.chRegistration != nil {
		event.response <- errorResponse(fmt.Errorf("registration already exists for channel [%s]", ed.channelID))
	} else if ed.fabclient.UserContext() == nil {
		event.response <- errorResponse(fmt.Errorf("no user context"))
	} else {
		ed.chRegistration = event.reg
		creator, err := ed.fabclient.UserContext().Identity()
		if err != nil {
			event.response <- errorResponse(fmt.Errorf("error getting signing identity: %s", err))
		} else {
			msg := &pb.Event{
				Event: &pb.Event_RegisterChannel{
					RegisterChannel: &pb.RegisterChannel{
						ChannelIds: []string{ed.channelID},
					},
				},
				Creator: creator,
			}
			if err = ed.connection.send(msg); err != nil {
				event.response <- errorResponse(fmt.Errorf("error sending register event for channel %s: %s", ed.channelID, err))
			}
		}
	}
	return
}

func (ed *eventDispatcher) handleUnregisterChannelEvent(e event) {
	event := e.(*unregisterChannelEvent)

	if ed.connection == nil {
		event.response <- errorResponse(fmt.Errorf("not connected"))
		return
	}

	if ed.chRegistration == nil {
		event.response <- errorResponse(fmt.Errorf("invalid registration for channel [%s]", ed.channelID))
	} else {
		creator, err := ed.fabclient.UserContext().Identity()
		if err != nil {
			event.response <- errorResponse(fmt.Errorf("error getting signing identity: %s", err))
		} else {
			msg := &pb.Event{
				Event: &pb.Event_DeregisterChannel{
					DeregisterChannel: &pb.DeregisterChannel{
						ChannelIds: []string{ed.channelID},
					},
				},
				Creator: creator,
			}
			if err = ed.connection.send(msg); err != nil {
				event.response <- errorResponse(fmt.Errorf("error sending deregister event for channel [%s]: %s", ed.channelID, err))
			} else {
				ed.chRegistration.event = event.response
			}
		}
	}
}

func (ed *eventDispatcher) handleRegisterConnectionEvent(e event) {
	event := e.(*registerConnectionEvent)

	if ed.connectionRegistration != nil {
		event.response <- errorResponse(fmt.Errorf("registration already exists for connection event"))
	} else {
		ed.connectionRegistration = event.reg
		event.response <- successResponse(event.reg)
	}
}

func (ed *eventDispatcher) handleRegisterBlockEvent(e event) {
	event := e.(*registerBlockEvent)

	if !ed.authorized[BLOCKEVENT] {
		event.response <- errorResponse(fmt.Errorf("client not authorized to receive block events"))
	} else if ed.blockRegistration != nil {
		event.response <- errorResponse(fmt.Errorf("registration already exists for block event"))
	} else {
		ed.blockRegistration = event.reg
		event.response <- successResponse(event.reg)
	}
}

func (ed *eventDispatcher) handleRegisterCCEvent(e event) {
	event := e.(*registerCCEvent)

	key := getKey(event.reg.ccID, event.reg.eventFilter)
	if !ed.authorized[FILTEREDBLOCKEVENT] {
		event.response <- errorResponse(fmt.Errorf("client not authorized to receive chaincode events"))
	} else if _, exists := ed.ccRegistrations[key]; exists {
		event.response <- errorResponse(fmt.Errorf("registration already exists for CC [%s] and event [%s]", event.reg.ccID, event.reg.eventFilter))
	} else {
		ed.ccRegistrations[key] = event.reg
		event.response <- successResponse(event.reg)
	}
}

func (ed *eventDispatcher) handleRegisterTxStatusEvent(e event) {
	event := e.(*registerTxStatusEvent)

	if !ed.authorized[FILTEREDBLOCKEVENT] {
		event.response <- errorResponse(fmt.Errorf("client not authorized to receive TX events"))
	} else if _, exists := ed.txRegistrations[event.reg.txID]; exists {
		event.response <- errorResponse(fmt.Errorf("registration already exists for TX ID [%s]", event.reg.txID))
	} else {
		ed.txRegistrations[event.reg.txID] = event.reg
		event.response <- successResponse(event.reg)
	}
}

func (ed *eventDispatcher) handleUnregisterEvent(e event) {
	event := e.(*unregisterEvent)

	switch registration := event.reg.(type) {
	case *blockRegistration:
		event.response <- unregResponse(ed.unregisterBlockEvents(registration))
	case *ccRegistration:
		event.response <- unregResponse(ed.unregisterCCEvents(registration))
	case *txRegistration:
		event.response <- unregResponse(ed.unregisterTXEvents(registration))
	default:
		event.response <- errorResponse(fmt.Errorf("unsupported registration type: %v", reflect.TypeOf(registration)))
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
		logger.Warningf("unsupported ChannelServiceResponse action: %s\n", resp.Action)
	}
}

func (ed *eventDispatcher) handleRegisterChannelResponse(resp *pb.ChannelServiceResponse) {
	logger.Debugf("Handling response: %v\n", resp)

	if ed.chRegistration == nil {
		logger.Warningf("Unexpected service response for nil channel registration\n")
		return
	}

	result := ed.getChannelServiceResult(resp)
	if result == nil {
		ed.chRegistration.event <- errorResponse(fmt.Errorf("no channel registration result found for channel %s in channel service response", ed.channelID))
		ed.chRegistration = nil
	} else {
		ed.authorized = make(map[string]bool)
		if resp.Success {
			for _, authEvent := range result.AuthorizedEvents {
				logger.Debugf("Authorized Event: %s\n", authEvent)
				ed.authorized[authEvent] = true
			}

			// Check existing registrations to ensure they're valid.
			if ed.blockRegistration != nil && !ed.authorized[BLOCKEVENT] {
				ed.chRegistration.event <- errorResponse(fmt.Errorf("client not authorized to receive block events"))
			} else if len(ed.ccRegistrations) > 0 && !ed.authorized[FILTEREDBLOCKEVENT] {
				ed.chRegistration.event <- errorResponse(fmt.Errorf("client not authorized to receive filtered block events"))
			} else {
				ed.chRegistration.event <- successResponse(nil)
			}

			// Set the event to nil since we've already responded
			ed.chRegistration.event = nil
		} else {
			ed.chRegistration.event <- errorResponse(fmt.Errorf("failed to register for channel events: %s", result.ErrorMsg))
			ed.chRegistration = nil
		}
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

	if ed.chRegistration == nil {
		logger.Warningf("Unexpected service response for nil channel registration\n")
		return
	}

	result := ed.getChannelServiceResult(resp)
	if result == nil {
		ed.chRegistration.event <- errorResponse(fmt.Errorf("no channel unregistration result found for channel %s in channel service response", ed.channelID))
	} else {
		if resp.Success {
			ed.chRegistration.event <- successResponse(nil)
		} else {
			ed.chRegistration.event <- errorResponse(fmt.Errorf("failed to unregister for channel events: %s", result.ErrorMsg))
		}
	}

	ed.chRegistration = nil
}

func (ed *eventDispatcher) handleFilteredBlockEvent(e event) {
	event := e.(*pb.Event_FilteredBlock)

	logger.Debugf("Handling event: %v\n", event)

	for _, tx := range event.FilteredBlock.FilteredTx {
		ed.triggerTxStatusEvent(tx)

		// Only send a chaincode event if the transaction has committed
		if tx.TxValidationCode == pb.TxValidationCode_VALID && tx.CcEvent != nil {
			ed.triggerCCEvent(tx.Txid, tx.CcEvent)
		}
	}
}

func (ed *eventDispatcher) handleBlockEvent(e event) {
	event := e.(*pb.Event_Block)

	logger.Debugf("Handling event: %v\n", event)

	if ed.blockRegistration != nil {
		ed.blockRegistration.event <- &fab.BlockEvent{Block: event.Block}
	}
}

func (ed *eventDispatcher) handleDisconnectedEvent(e event) {
	event := e.(*disconnectedEvent)

	ed.connection = nil
	ed.chRegistration = nil
	ed.authorized = nil

	if ed.connectionRegistration != nil {
		ed.connectionRegistration.event <- &fab.ConnectionEvent{
			Connected: false,
			Err:       event.err,
		}
	}
}

func (ed *eventDispatcher) unregisterBlockEvents(registration *blockRegistration) error {
	if ed.blockRegistration != registration {
		return fmt.Errorf("the provided registration is invalid")
	}
	ed.blockRegistration = nil
	return nil
}

func (ed *eventDispatcher) unregisterCCEvents(registration *ccRegistration) error {
	key := getKey(registration.ccID, registration.eventFilter)
	if _, ok := ed.ccRegistrations[key]; !ok {
		return fmt.Errorf("the provided registration is invalid")
	}
	delete(ed.ccRegistrations, key)
	return nil
}

func (ed *eventDispatcher) unregisterTXEvents(registration *txRegistration) error {
	if _, ok := ed.txRegistrations[registration.txID]; !ok {
		return fmt.Errorf("the provided registration is invalid")
	}

	logger.Debugf("Unregistering Tx Status event for TxID [%s]...\n", registration.txID)
	delete(ed.txRegistrations, registration.txID)
	return nil
}

func (ed *eventDispatcher) triggerTxStatusEvent(tx *pb.FilteredTransaction) {
	logger.Debugf("Triggering Tx Status event for TxID [%s]...\n", tx.Txid)
	if reg, ok := ed.txRegistrations[tx.Txid]; ok {
		logger.Debugf("Sending Tx Status event for TxID [%s] to registrant...\n", tx.Txid)
		reg.event <- &fab.TxStatusEvent{
			TxID:             tx.Txid,
			TxValidationCode: tx.TxValidationCode,
		}
	}
}

func (ed *eventDispatcher) triggerCCEvent(txID string, ccEvent *pb.ChaincodeEvent) {
	for _, reg := range ed.ccRegistrations {
		logger.Debugf("Matching [%s] against pattern [%s]...\n", ccEvent.EventName, reg.eventFilter)
		if reg.eventRegExp.MatchString(ccEvent.EventName) {
			logger.Debugf("... matched [%s] against pattern [%s]\n", ccEvent.EventName, reg.eventFilter)
			reg.event <- &fab.CCEvent{
				ChaincodeID: ccEvent.ChaincodeId,
				EventName:   ccEvent.EventName,
				TxID:        txID,
			}
		}
	}
}

func (ed *eventDispatcher) registerHandlers() {
	ed.registerHandler(&connectEvent{}, ed.handleConnectEvent)
	ed.registerHandler(&registerChannelEvent{}, ed.handleRegisterChannelEvent)
	ed.registerHandler(&registerConnectionEvent{}, ed.handleRegisterConnectionEvent)
	ed.registerHandler(&registerCCEvent{}, ed.handleRegisterCCEvent)
	ed.registerHandler(&registerTxStatusEvent{}, ed.handleRegisterTxStatusEvent)
	ed.registerHandler(&registerBlockEvent{}, ed.handleRegisterBlockEvent)
	ed.registerHandler(&unregisterEvent{}, ed.handleUnregisterEvent)
	ed.registerHandler(&unregisterChannelEvent{}, ed.handleUnregisterChannelEvent)
	ed.registerHandler(&pb.Event_ChannelServiceResponse{}, ed.handleServiceResponse)
	ed.registerHandler(&pb.Event_FilteredBlock{}, ed.handleFilteredBlockEvent)
	ed.registerHandler(&pb.Event_Block{}, ed.handleBlockEvent)
	ed.registerHandler(&disconnectedEvent{}, ed.handleDisconnectedEvent)
	ed.registerHandler(&disconnectEvent{}, ed.handleDisconnectEvent)
}

func (ed *eventDispatcher) registerHandler(t interface{}, h handler) {
	ed.handlers[reflect.TypeOf(t)] = h
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

func getKey(ccID, eventName string) string {
	return ccID + "/" + eventName
}
