/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package chclient enables channel client
package chclient

import (
	"bytes"
	"reflect"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn/txnhandler"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	txnHandlerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/txnhandler"
	"github.com/hyperledger/fabric-sdk-go/pkg/status"
)

// ChannelClient enables access to a Fabric network.
type ChannelClient struct {
	client    fab.Resource
	channel   fab.Channel
	discovery fab.DiscoveryService
	selection fab.SelectionService
	eventHub  fab.EventHub
}

// txProposalResponseFilter process transaction proposal response
type txProposalResponseFilter struct {
}

// ProcessTxProposalResponse process transaction proposal response
func (txProposalResponseFilter *txProposalResponseFilter) ProcessTxProposalResponse(txProposalResponse []*apitxn.TransactionProposalResponse) ([]*apitxn.TransactionProposalResponse, error) {
	var a1 []byte
	for n, r := range txProposalResponse {
		if r.ProposalResponse.GetResponse().Status != int32(common.Status_SUCCESS) {
			return nil, status.NewFromProposalResponse(r.ProposalResponse, r.Endorser)
		}
		if n == 0 {
			a1 = r.ProposalResponse.GetResponse().Payload
			continue
		}

		if bytes.Compare(a1, r.ProposalResponse.GetResponse().Payload) != 0 {
			return nil, status.New(status.EndorserClientStatus, status.EndorsementMismatch.ToInt32(),
				"ProposalResponsePayloads do not match", nil)
		}
	}

	return txProposalResponse, nil
}

// NewChannelClient returns a ChannelClient instance.
func NewChannelClient(client fab.Resource, channel fab.Channel, discovery fab.DiscoveryService, selection fab.SelectionService, eventHub fab.EventHub) (*ChannelClient, error) {

	channelClient := ChannelClient{client: client, channel: channel, discovery: discovery, selection: selection, eventHub: eventHub}

	return &channelClient, nil
}

// Query chaincode
func (cc *ChannelClient) Query(request apitxn.QueryRequest) ([]byte, error) {

	return cc.QueryWithOpts(request, apitxn.TxOpts{})

}

// QueryWithOpts  allows the user to provide options for query (sync vs async, etc.)
func (cc *ChannelClient) QueryWithOpts(request apitxn.QueryRequest, opts apitxn.TxOpts) ([]byte, error) {

	//Basic Validation
	if request.ChaincodeID == "" || request.Fcn == "" {
		return nil, errors.New("ChaincodeID and Fcn are required")
	}

	isAsync := opts.Notifier != nil

	//Prepare context objects for handler
	requestContext, clientContext := cc.prepareHandlerContexts(request, opts)

	//Get query handler
	queryHandler := txnHandlerImpl.GetQueryHandler()

	//Perform action through handler
	go queryHandler.Handle(requestContext, clientContext)

	//If async return
	if isAsync {
		return nil, nil
	}

	//If Sync
	select {
	case response := <-requestContext.Request.Opts.Notifier:
		return response.Payload, response.Error
	case <-time.After(requestContext.Request.Opts.Timeout):
		return nil, errors.New("query request timed out")
	}

}

// ExecuteTxWithOpts allows the user to provide options for execute transaction:
// sync vs async, filter to inspect proposal response before commit etc)
func (cc *ChannelClient) ExecuteTxWithOpts(request apitxn.ExecuteTxRequest, opts apitxn.TxOpts) ([]byte, apitxn.TransactionID, error) {

	//Basic Validation
	if request.ChaincodeID == "" || request.Fcn == "" {
		return nil, apitxn.TransactionID{}, errors.New("chaincode name and function name are required")
	}

	isAsync := opts.Notifier != nil

	//Prepare context objects for handler
	requestContext, clientContext := cc.prepareHandlerContexts(request, opts)

	//Get query handler
	executeTxHandler := txnHandlerImpl.GetExecuteTxHandler()

	//Perform action through handler
	go executeTxHandler.Handle(requestContext, clientContext)

	var payload []byte
	for {
		select {
		case response := <-requestContext.Response.Payload:
			payload = response
			if isAsync {
				//If async return with payload, commit tx will occur asynchronously
				return payload, requestContext.Response.TxnID, nil
			}
		case response := <-requestContext.Request.Opts.Notifier:
			return payload, response.TransactionID, response.Error
		case <-time.After(requestContext.Request.Opts.Timeout): // This should never happen since there's timeout in sendTransaction
			return payload, requestContext.Response.TxnID, errors.New("ExecuteTx request timed out")
		}
	}

}

//prepareHandlerContexts prepares context objects for handlers
func (cc *ChannelClient) prepareHandlerContexts(request interface{}, opts apitxn.TxOpts) (*txnhandler.RequestContext, *txnhandler.ClientContext) {

	clientContext := &txnhandler.ClientContext{
		Channel:   cc.channel,
		Selection: cc.selection,
		Discovery: cc.discovery,
		EventHub:  cc.eventHub,
	}

	var requestContext *txnhandler.RequestContext
	switch request.(type) {
	case apitxn.QueryRequest:
		requestContext = &txnhandler.RequestContext{Request: txnhandler.TxRequestContext{
			ChaincodeID: request.(apitxn.QueryRequest).ChaincodeID,
			Fcn:         request.(apitxn.QueryRequest).Fcn,
			Args:        request.(apitxn.QueryRequest).Args,
			Opts:        opts,
		}, Response: txnhandler.TxResponseContext{}}

		if requestContext.Request.Opts.Timeout == 0 {
			requestContext.Request.Opts.Timeout = cc.client.Config().TimeoutOrDefault(apiconfig.Query)
		}
		break

	case apitxn.ExecuteTxRequest:
		requestContext = &txnhandler.RequestContext{Request: txnhandler.TxRequestContext{
			ChaincodeID:  request.(apitxn.ExecuteTxRequest).ChaincodeID,
			Fcn:          request.(apitxn.ExecuteTxRequest).Fcn,
			Args:         request.(apitxn.ExecuteTxRequest).Args,
			TransientMap: request.(apitxn.ExecuteTxRequest).TransientMap,
			Opts:         opts,
		}, Response: txnhandler.TxResponseContext{}}

		if requestContext.Request.Opts.Timeout == 0 {
			requestContext.Request.Opts.Timeout = cc.client.Config().TimeoutOrDefault(apiconfig.ExecuteTx)
		}
		requestContext.Response.Payload = make(chan []byte)
		break

	default:
		requestContext = &txnhandler.RequestContext{Request: txnhandler.TxRequestContext{
			Opts: apitxn.TxOpts{},
		}, Response: txnhandler.TxResponseContext{}}
	}

	if requestContext.Request.Opts.TxFilter == nil {
		requestContext.Request.Opts.TxFilter = &txProposalResponseFilter{}
	}

	if requestContext.Request.Opts.Notifier == nil {
		requestContext.Request.Opts.Notifier = make(chan apitxn.Response)
	}

	return requestContext, clientContext

}

// ExecuteTx prepares and executes transaction
func (cc *ChannelClient) ExecuteTx(request apitxn.ExecuteTxRequest) ([]byte, apitxn.TransactionID, error) {

	return cc.ExecuteTxWithOpts(request, apitxn.TxOpts{})
}

// Close releases channel client resources (disconnects event hub etc.)
func (cc *ChannelClient) Close() error {
	if cc.eventHub.IsConnected() == true {
		return cc.eventHub.Disconnect()
	}

	return nil
}

// RegisterChaincodeEvent registers chain code event
// @param {chan bool} channel which receives event details when the event is complete
// @returns {object} object handle that should be used to unregister
func (cc *ChannelClient) RegisterChaincodeEvent(notify chan<- *apitxn.CCEvent, chainCodeID string, eventID string) apitxn.Registration {

	// Register callback for CE
	rce := cc.eventHub.RegisterChaincodeEvent(chainCodeID, eventID, func(ce *fab.ChaincodeEvent) {
		notify <- &apitxn.CCEvent{ChaincodeID: ce.ChaincodeID, EventName: ce.EventName, TxID: ce.TxID, Payload: ce.Payload}
	})

	return rce
}

// UnregisterChaincodeEvent removes chain code event registration
func (cc *ChannelClient) UnregisterChaincodeEvent(registration apitxn.Registration) error {

	switch regType := registration.(type) {

	case *fab.ChainCodeCBE:
		cc.eventHub.UnregisterChaincodeEvent(regType)
	default:
		return errors.Errorf("Unsupported registration type: %v", reflect.TypeOf(registration))
	}

	return nil

}
