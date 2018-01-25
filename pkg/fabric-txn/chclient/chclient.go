/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package chclient enables channel client
package chclient

import (
	"reflect"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn/txnhandler"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	txnHandlerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/txnhandler"
)

const (
	defaultHandlerTimeout = time.Second * 10
)

// ChannelClient enables access to a Fabric network.
type ChannelClient struct {
	client    fab.Resource
	channel   fab.Channel
	discovery fab.DiscoveryService
	selection fab.SelectionService
	eventHub  fab.EventHub
}

// NewChannelClient returns a ChannelClient instance.
func NewChannelClient(client fab.Resource, channel fab.Channel, discovery fab.DiscoveryService, selection fab.SelectionService, eventHub fab.EventHub) (*ChannelClient, error) {

	channelClient := ChannelClient{client: client, channel: channel, discovery: discovery, selection: selection, eventHub: eventHub}

	return &channelClient, nil
}

// Query chaincode using request and optional opts provided
func (cc *ChannelClient) Query(request apitxn.Request, opts ...apitxn.Option) ([]byte, error) {

	response := cc.InvokeHandler(txnHandlerImpl.GetQueryHandler(), request, cc.addDefaultTimeout(apiconfig.Query, opts...)...)

	return response.Payload, response.Error
}

// ExecuteTx prepares and executes transaction using request and optional opts provided
func (cc *ChannelClient) ExecuteTx(request apitxn.Request, opts ...apitxn.Option) ([]byte, apitxn.TransactionID, error) {

	response := cc.InvokeHandler(txnHandlerImpl.GetExecuteTxHandler(), request, cc.addDefaultTimeout(apiconfig.ExecuteTx, opts...)...)

	return response.Payload, response.TransactionID, response.Error
}

//InvokeHandler invokes handler using request and opts provided
func (cc *ChannelClient) InvokeHandler(handler txnhandler.Handler, request apitxn.Request, opts ...apitxn.Option) apitxn.Response {
	//TODO: this function going to be exposed through ChannelClient interface
	//Read execute tx opts
	exOpts, err := cc.prepareOptsFromOptions(opts...)
	if err != nil {
		return apitxn.Response{Error: err}
	}

	//Prepare context objects for handler
	requestContext, clientContext, err := cc.prepareHandlerContexts(request, exOpts)
	if err != nil {
		return apitxn.Response{Error: err}
	}

	//Perform action through handler
	go handler.Handle(requestContext, clientContext)

	//notifier in opts will handle response if provided
	if exOpts.Notifier != nil {
		return apitxn.Response{}
	}

	select {
	case response := <-requestContext.Request.Opts.Notifier:
		return response
	case <-time.After(requestContext.Request.Opts.Timeout):
		return apitxn.Response{Error: errors.New("handler timed out while performing operation")}
	}
}

//prepareHandlerContexts prepares context objects for handlers
func (cc *ChannelClient) prepareHandlerContexts(request apitxn.Request, opts apitxn.Opts) (*txnhandler.RequestContext, *txnhandler.ClientContext, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return nil, nil, errors.New("ChaincodeID and Fcn are required")
	}

	clientContext := &txnhandler.ClientContext{
		Channel:   cc.channel,
		Selection: cc.selection,
		Discovery: cc.discovery,
		EventHub:  cc.eventHub,
	}

	requestContext := &txnhandler.RequestContext{Request: txnhandler.TxRequestContext{
		ChaincodeID:  request.ChaincodeID,
		Fcn:          request.Fcn,
		Args:         request.Args,
		TransientMap: request.TransientMap,
		Opts:         opts,
	}, Response: apitxn.Response{}}

	if requestContext.Request.Opts.Timeout == 0 {
		requestContext.Request.Opts.Timeout = defaultHandlerTimeout
	}

	if requestContext.Request.Opts.Notifier == nil {
		requestContext.Request.Opts.Notifier = make(chan apitxn.Response)
	}

	return requestContext, clientContext, nil

}

//prepareOptsFromOptions Reads apitxn.Opts from apitxn.Option array
func (cc *ChannelClient) prepareOptsFromOptions(opts ...apitxn.Option) (apitxn.Opts, error) {
	txnOpts := apitxn.Opts{}
	for _, option := range opts {
		err := option(&txnOpts)
		if err != nil {
			return txnOpts, errors.WithMessage(err, "Failed to read opts")
		}
	}
	return txnOpts, nil
}

//addDefaultTimeout adds given default timeout if it is missing in options
func (cc *ChannelClient) addDefaultTimeout(timeOutType apiconfig.TimeoutType, opts ...apitxn.Option) []apitxn.Option {
	txnOpts := apitxn.Opts{}
	for _, option := range opts {
		option(&txnOpts)
	}

	if txnOpts.Timeout == 0 {
		timeout := cc.client.Config().TimeoutOrDefault(timeOutType)
		timeoutOpt := func(opts *apitxn.Opts) error {
			opts.Timeout = timeout
			return nil
		}
		return append(opts, timeoutOpt)
	}
	return opts
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
