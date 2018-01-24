/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txnhandler

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	txnhandlerApi "github.com/hyperledger/fabric-sdk-go/api/apitxn/txnhandler"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/status"

	"strings"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/mocks"
)

const (
	testTimeOut           = 20 * time.Second
	discoveryServiceError = "Discovery service error"
	selectionServiceError = "Selection service error"
	filterTxError         = "Filter Tx error"
)

func TestQueryHandlerSuccess(t *testing.T) {

	//Sample request
	request := apitxn.QueryRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, apitxn.TxOpts{}, t)
	clientContext := setupChannelClientContext(nil, nil, nil, t)

	//Get query handler
	queryHandler := GetQueryHandler()

	//Perform action through handler
	go queryHandler.Handle(requestContext, clientContext)

	select {
	case response := <-requestContext.Request.Opts.Notifier:
		if response.Error != nil {
			t.Fatal("Query handler failed", response.Error)
		}
	case <-time.After(requestContext.Request.Opts.Timeout):
		t.Fatal("Query handler : time out not expected")
	}
}

func TestExecuteTxHandlerSuccess(t *testing.T) {

	//Sample request
	request := apitxn.ExecuteTxRequest{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, apitxn.TxOpts{}, t)
	clientContext := setupChannelClientContext(nil, nil, nil, t)

	//Prepare mock eventhub
	mockEventHub := fcmocks.NewMockEventHub()
	clientContext.EventHub = mockEventHub

	//Get query handler
	queryHandler := GetExecuteTxHandler()

	//Perform action through handler
	go queryHandler.Handle(requestContext, clientContext)
	payloadReceived := false
	for {

		select {
		case callback := <-mockEventHub.RegisteredTxCallbacks:
			callback("txid", 0,
				status.New(status.EventServerStatus, 0, "test", nil))
		case <-requestContext.Response.Payload:
			payloadReceived = true
		case <-requestContext.Request.Opts.Notifier:
			if !payloadReceived {
				t.Fatal("Not supposed to get response before payload being delivered")
			}
			return
		case <-time.After(requestContext.Request.Opts.Timeout):
			t.Fatal("Execute handler : time out not expected")
		}
	}
}

func TestQueryHandlerErrors(t *testing.T) {

	//Error Scenario 1
	request := apitxn.QueryRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, apitxn.TxOpts{}, t)
	clientContext := setupChannelClientContext(errors.New(discoveryServiceError), nil, nil, t)

	//Get query handler
	queryHandler := GetQueryHandler()

	//Perform action through handler
	go queryHandler.Handle(requestContext, clientContext)

	select {
	case response := <-requestContext.Request.Opts.Notifier:
		if response.Error == nil || !strings.Contains(response.Error.Error(), discoveryServiceError) {
			t.Fatal("Expected error: ", discoveryServiceError, ", Received error:", response.Error.Error())
		}
	case <-time.After(requestContext.Request.Opts.Timeout):
		t.Fatal("Query handler : time out not expected")
	}

	//Error Scenario 2
	clientContext = setupChannelClientContext(nil, errors.New(selectionServiceError), nil, t)

	//Perform action through handler
	go queryHandler.Handle(requestContext, clientContext)

	select {
	case response := <-requestContext.Request.Opts.Notifier:
		if response.Error == nil || !strings.Contains(response.Error.Error(), selectionServiceError) {
			t.Fatal("Expected error: ", selectionServiceError, ", Received error:", response.Error.Error())
		}
	case <-time.After(requestContext.Request.Opts.Timeout):
		t.Fatal("Query handler : time out not expected")

	}

	//Error Scenario 3
	requestContext.Request.Opts.TxFilter = &mockTxProposalResponseFilter{false}
	clientContext = setupChannelClientContext(nil, nil, nil, t)

	//Perform action through handler
	go queryHandler.Handle(requestContext, clientContext)

	select {
	case response := <-requestContext.Request.Opts.Notifier:
		if response.Error == nil || !strings.Contains(response.Error.Error(), filterTxError) {
			t.Fatal("Expected error: ", filterTxError, ", Received error:", response.Error.Error())
		}
	case <-time.After(requestContext.Request.Opts.Timeout):
		t.Fatal("Query handler : time out not expected")

	}
}

//prepareHandlerContexts prepares context objects for handlers
func prepareRequestContext(request interface{}, opts apitxn.TxOpts, t *testing.T) *txnhandlerApi.RequestContext {

	var requestContext *txnhandlerApi.RequestContext
	switch request.(type) {
	case apitxn.QueryRequest:
		requestContext = &txnhandlerApi.RequestContext{Request: txnhandlerApi.TxRequestContext{
			ChaincodeID: request.(apitxn.QueryRequest).ChaincodeID,
			Fcn:         request.(apitxn.QueryRequest).Fcn,
			Args:        request.(apitxn.QueryRequest).Args,
			Opts:        opts,
		}, Response: txnhandlerApi.TxResponseContext{}}

		break

	case apitxn.ExecuteTxRequest:
		requestContext = &txnhandlerApi.RequestContext{Request: txnhandlerApi.TxRequestContext{
			ChaincodeID:  request.(apitxn.ExecuteTxRequest).ChaincodeID,
			Fcn:          request.(apitxn.ExecuteTxRequest).Fcn,
			Args:         request.(apitxn.ExecuteTxRequest).Args,
			TransientMap: request.(apitxn.ExecuteTxRequest).TransientMap,
			Opts:         opts,
		}, Response: txnhandlerApi.TxResponseContext{}}

		if requestContext.Request.Opts.Timeout == 0 {

		}
		requestContext.Response.Payload = make(chan []byte)
		break

	default:
		requestContext = &txnhandlerApi.RequestContext{Request: txnhandlerApi.TxRequestContext{
			Opts: apitxn.TxOpts{},
		}, Response: txnhandlerApi.TxResponseContext{}}
	}

	requestContext.Request.Opts.Timeout = testTimeOut

	requestContext.Request.Opts.TxFilter = &mockTxProposalResponseFilter{true}

	requestContext.Request.Opts.Notifier = make(chan apitxn.Response)

	return requestContext

}

func setupTestChannel() (*channel.Channel, error) {
	client := setupTestClient()
	return channel.NewChannel("testChannel", client)
}

func setupTestClient() *fcmocks.MockClient {
	client := fcmocks.NewMockClient()
	user := fcmocks.NewMockUser("test")
	cryptoSuite := &fcmocks.MockCryptoSuite{}
	client.SetIdentityContext(user)
	client.SetCryptoSuite(cryptoSuite)
	return client
}

func setupChannelClientContext(discErr error, selectionErr error, peers []apifabclient.Peer, t *testing.T) *txnhandlerApi.ClientContext {

	testChannel, err := setupTestChannel()
	if err != nil {
		t.Fatalf("Failed to setup test channel: %s", err)
	}

	orderer := fcmocks.NewMockOrderer("", nil)
	testChannel.AddOrderer(orderer)

	discoveryService, err := setupTestDiscovery(discErr, nil)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	selectionService, err := setupTestSelection(selectionErr, peers)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	return &txnhandlerApi.ClientContext{
		Channel:   testChannel,
		Discovery: discoveryService,
		Selection: selectionService,
	}

}

func setupTestDiscovery(discErr error, peers []apifabclient.Peer) (apifabclient.DiscoveryService, error) {

	mockDiscovery, err := txnmocks.NewMockDiscoveryProvider(discErr, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockDiscoveryProvider failed")
	}

	return mockDiscovery.NewDiscoveryService("mychannel")
}

func setupTestSelection(discErr error, peers []apifabclient.Peer) (apifabclient.SelectionService, error) {

	mockSelection, err := txnmocks.NewMockSelectionProvider(discErr, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockSelectinProvider failed")
	}

	return mockSelection.NewSelectionService("mychannel")
}

type mockTxProposalResponseFilter struct {
	success bool
}

func (m *mockTxProposalResponseFilter) ProcessTxProposalResponse(txProposalResponse []*apitxn.TransactionProposalResponse) ([]*apitxn.TransactionProposalResponse, error) {
	if m.success {
		return txProposalResponse, nil
	}
	return nil, errors.New(filterTxError)
}
