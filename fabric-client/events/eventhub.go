/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package events

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
	consumer "github.com/hyperledger/fabric-sdk-go/fabric-client/events/consumer"
	"github.com/hyperledger/fabric/core/ledger/util"
	common "github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

// EventHub ...
type EventHub interface {
	SetPeerAddr(peerURL string, certificate string, serverhostoverride string)
	IsConnected() bool
	Connect() error
	RegisterChaincodeEvent(ccid string, eventname string, callback func(*pb.ChaincodeEvent)) *ChainCodeCBE
	UnregisterChaincodeEvent(cbe *ChainCodeCBE)
	RegisterTxEvent(txID string, callback func(string, error))
	UnregisterTxEvent(txID string)
}

type eventHub struct {
	// Protects chaincodeRegistrants, blockRegistrants and txRegistrants
	mtx sync.RWMutex
	// Map of clients registered for chaincode events
	chaincodeRegistrants map[string][]*ChainCodeCBE
	// Map of clients registered for block events
	blockRegistrants []func(*common.Block)
	// Map of clients registered for transactional events
	txRegistrants map[string]func(string, error)
	// peer addr to connect to
	peerAddr string
	// peer tls certificate
	peerTLSCertificate string
	// peer tls serverhostoverride
	peerTLSServerhostoverride string
	// grpc event client interface
	client consumer.EventsClient
	// fabric connection state of this eventhub
	connected bool
	// List of events client is interested in
	interestedEvents []*pb.Interest
}

// ChainCodeCBE ...
/**
 * The ChainCodeCBE is used internal to the EventHub to hold chaincode
 * event registration callbacks.
 */
type ChainCodeCBE struct {
	// chaincode id
	CCID string
	// event name regex filter
	EventNameFilter string
	// callback function to invoke on successful filter match
	CallbackFunc func(*pb.ChaincodeEvent)
}

// NewEventHub ...
func NewEventHub() EventHub {
	chaincodeRegistrants := make(map[string][]*ChainCodeCBE)
	blockRegistrants := make([]func(*common.Block), 0)
	txRegistrants := make(map[string]func(string, error))

	// default interested events
	// TODO: set interestedEvents based on handler registration
	interestedEvents := []*pb.Interest{{EventType: pb.EventType_BLOCK}}

	eventHub := &eventHub{chaincodeRegistrants: chaincodeRegistrants, blockRegistrants: blockRegistrants, txRegistrants: txRegistrants, interestedEvents: interestedEvents}

	return eventHub
}

// SetPeerAddr ...
/**
 * Set peer url for event source<p>
 * Note: Only use this if creating your own EventHub. The chain
 * creates a default eventHub that most Node clients can
 * use (see eventHubConnect, eventHubDisconnect and getEventHub).
 * @param {string} peeraddr peer url
 * @param {string} peerTLSCertificate peer tls certificate
 * @param {string} peerTLSServerhostoverride tls serverhostoverride
 */
func (eventHub *eventHub) SetPeerAddr(peerURL string, peerTLSCertificate string, peerTLSServerhostoverride string) {
	eventHub.peerAddr = peerURL
	eventHub.peerTLSCertificate = peerTLSCertificate
	eventHub.peerTLSServerhostoverride = peerTLSServerhostoverride

}

// Isconnected ...
/**
 * Get connected state of eventhub
 * @returns true if connected to event source, false otherwise
 */
func (eventHub *eventHub) IsConnected() bool {
	return eventHub.connected
}

// Connect ...
/**
 * Establishes connection with peer event source<p>
 */
func (eventHub *eventHub) Connect() error {
	if eventHub.peerAddr == "" {
		return fmt.Errorf("eventHub.peerAddr is empty")
	}

	eventHub.mtx.Lock()
	defer eventHub.mtx.Unlock()

	eventHub.blockRegistrants = make([]func(*common.Block), 0)
	eventHub.blockRegistrants = append(eventHub.blockRegistrants, eventHub.txCallback)

	eventsClient, _ := consumer.NewEventsClient(eventHub.peerAddr, eventHub.peerTLSCertificate, eventHub.peerTLSServerhostoverride, 5, eventHub)
	if err := eventsClient.Start(); err != nil {
		eventsClient.Stop()
		return fmt.Errorf("Error from eventsClient.Start (%s)", err.Error())

	}
	eventHub.connected = true
	eventHub.client = eventsClient
	return nil
}

//GetInterestedEvents implements consumer.EventAdapter interface for registering interested events
func (eventHub *eventHub) GetInterestedEvents() ([]*pb.Interest, error) {
	return eventHub.interestedEvents, nil
}

//Recv implements consumer.EventAdapter interface for receiving events
func (eventHub *eventHub) Recv(msg *pb.Event) (bool, error) {
	eventHub.mtx.RLock()
	defer eventHub.mtx.RUnlock()

	switch msg.Event.(type) {
	case *pb.Event_Block:
		blockEvent := msg.Event.(*pb.Event_Block)
		logger.Debugf("Recv blockEvent:%v\n", blockEvent)
		for _, v := range eventHub.blockRegistrants {
			v(blockEvent.Block)
		}
		return true, nil
	case *pb.Event_ChaincodeEvent:
		ccEvent := msg.Event.(*pb.Event_ChaincodeEvent)
		logger.Debugf("Recv ccEvent:%v\n", ccEvent)

		cbeArray := eventHub.chaincodeRegistrants[ccEvent.ChaincodeEvent.ChaincodeId]
		if len(cbeArray) <= 0 {
			logger.Debugf("No event registration for ccid %s \n", ccEvent.ChaincodeEvent.ChaincodeId)
		}

		for _, v := range cbeArray {
			if v.EventNameFilter == ccEvent.ChaincodeEvent.EventName {
				callback := v.CallbackFunc
				if callback != nil {
					callback(ccEvent.ChaincodeEvent)
				}
			}
		}
		return true, nil
	default:
		return true, nil
	}
}

// Disconnected implements consumer.EventAdapter interface for receiving events
/**
 * Disconnects peer event source<p>
 * Note: Only use this if creating your own EventHub. The chain
 * class creates a default eventHub that most Node clients can
 * use (see eventHubConnect, eventHubDisconnect and getEventHub).
 */
func (eventHub *eventHub) Disconnected(err error) {
	if !eventHub.connected {
		return
	}
	eventHub.client.Stop()
	eventHub.connected = false

}

// RegisterChaincodeEvent ...
/**
 * Register a callback function to receive chaincode events.
 * @param {string} ccid string chaincode id
 * @param {string} eventname string The regex string used to filter events
 * @param {function} callback Function Callback function for filter matches
 * that takes a single parameter which is a json object representation
 * of type "message ChaincodeEvent"
 * @returns {object} ChainCodeCBE object that should be treated as an opaque
 * handle used to unregister (see unregisterChaincodeEvent)
 */
func (eventHub *eventHub) RegisterChaincodeEvent(ccid string, eventname string, callback func(*pb.ChaincodeEvent)) *ChainCodeCBE {
	if !eventHub.connected {
		return nil
	}

	eventHub.mtx.Lock()
	defer eventHub.mtx.Unlock()

	cbe := ChainCodeCBE{CCID: ccid, EventNameFilter: eventname, CallbackFunc: callback}
	cbeArray := eventHub.chaincodeRegistrants[ccid]
	if cbeArray == nil && len(cbeArray) <= 0 {
		cbeArray = make([]*ChainCodeCBE, 0)
		cbeArray = append(cbeArray, &cbe)
		eventHub.chaincodeRegistrants[ccid] = cbeArray
	} else {
		cbeArray = append(cbeArray, &cbe)
		eventHub.chaincodeRegistrants[ccid] = cbeArray
	}
	return &cbe
}

// UnregisterChaincodeEvent ...
/**
 * Unregister chaincode event registration
 * @param {object} ChainCodeCBE handle returned from call to
 * registerChaincodeEvent.
 */
func (eventHub *eventHub) UnregisterChaincodeEvent(cbe *ChainCodeCBE) {
	if !eventHub.connected {
		return
	}

	eventHub.mtx.Lock()
	defer eventHub.mtx.Unlock()

	cbeArray := eventHub.chaincodeRegistrants[cbe.CCID]
	if len(cbeArray) <= 0 {
		logger.Debugf("No event registration for ccid %s \n", cbe.CCID)
		return
	}
	for i, v := range cbeArray {
		if v.EventNameFilter == cbe.EventNameFilter {

			cbeArray = append(cbeArray[:i], cbeArray[i+1:]...)

		}
	}
	if len(cbeArray) <= 0 {
		delete(eventHub.chaincodeRegistrants, cbe.CCID)
	}

}

// RegisterTxEvent ...
/**
 * Register a callback function to receive transactional events.<p>
 * Note: transactional event registration is primarily used by
 * the sdk to track deploy and invoke completion events. Nodejs
 * clients generally should not need to call directly.
 * @param {string} txid string transaction id
 * @param {function} callback Function that takes a single parameter which
 * is a json object representation of type "message Transaction"
 */
func (eventHub *eventHub) RegisterTxEvent(txID string, callback func(string, error)) {
	logger.Debugf("reg txid %s\n", txID)

	eventHub.mtx.Lock()
	eventHub.txRegistrants[txID] = callback
	eventHub.mtx.Unlock()
}

// UnregisterTxEvent ...
/**
 * Unregister transactional event registration.
 * @param txid string transaction id
 */
func (eventHub *eventHub) UnregisterTxEvent(txID string) {
	eventHub.mtx.Lock()
	delete(eventHub.txRegistrants, txID)
	eventHub.mtx.Unlock()
}

/**
 * private internal callback for processing tx events
 * @param {object} block json object representing block of tx
 * from the fabric
 */
func (eventHub *eventHub) txCallback(block *common.Block) {
	logger.Debugf("txCallback block=%v\n", block)

	eventHub.mtx.RLock()
	defer eventHub.mtx.RUnlock()
	txsFltr := util.TxValidationFlags(block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER])
	for i, v := range block.Data.Data {
		if env, err := utils.GetEnvelopeFromBlock(v); err != nil {
			return
		} else if env != nil {
			// get the payload from the envelope
			payload, err := utils.GetPayload(env)
			if err != nil {
				return
			}

			channelHeaderBytes := payload.Header.ChannelHeader
			channelHeader := &common.ChannelHeader{}
			err = proto.Unmarshal(channelHeaderBytes, channelHeader)
			if err != nil {
				return
			}

			callback := eventHub.txRegistrants[channelHeader.TxId]
			if callback != nil {
				if txsFltr.IsInvalid(i) {
					callback(channelHeader.TxId, fmt.Errorf("Received invalid transaction from channel %s\n", channelHeader.ChannelId))

				} else {
					callback(channelHeader.TxId, nil)
				}
			}
		}
	}

}
