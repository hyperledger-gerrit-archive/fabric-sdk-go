/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package invoke

import (
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

const (
	testTimeOut              = 20 * time.Second
	discoveryServiceError    = "Discovery service error"
	selectionServiceError    = "Selection service error"
	endorsementMisMatchError = "ProposalResponsePayloads do not match"

	filterTxError = "Filter Tx error"
)

func TestQueryHandlerSuccess(t *testing.T) {

	//Sample request
	request := Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupChannelClientContext(nil, nil, []fab.Peer{mockPeer1, mockPeer2}, t)

	//Get query handler
	queryHandler := NewQueryHandler()

	//Perform action through handler
	queryHandler.Handle(requestContext, clientContext)
	if requestContext.Error != nil {
		t.Fatal("Query handler failed", requestContext.Error)
	}
}

func TestExecuteTxHandlerSuccess(t *testing.T) {
	//Sample request
	request := Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupChannelClientContext(nil, nil, []fab.Peer{mockPeer1, mockPeer2}, t)

	//Prepare mock eventhub
	mockEventService := fcmocks.NewMockEventService()
	clientContext.EventService = mockEventService

	go func() {
		select {
		case txStatusReg := <-mockEventService.TxStatusRegCh:
			txStatusReg.Eventch <- &fab.TxStatusEvent{TxID: txStatusReg.TxID, TxValidationCode: pb.TxValidationCode_VALID}
		case <-time.After(requestContext.Opts.Timeouts[core.Execute]):
			t.Fatal("Execute handler : time out not expected")
		}
	}()

	//Get query handler
	executeHandler := NewExecuteHandler()
	//Perform action through handler
	executeHandler.Handle(requestContext, clientContext)
	assert.Nil(t, requestContext.Error)
}

func TestQueryHandlerErrors(t *testing.T) {

	//Error Scenario 1
	request := Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, Opts{}, t)
	clientContext := setupChannelClientContext(errors.New(discoveryServiceError), nil, nil, t)

	//Get query handler
	queryHandler := NewQueryHandler()

	//Perform action through handler
	queryHandler.Handle(requestContext, clientContext)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), discoveryServiceError) {
		t.Fatal("Expected error: ", discoveryServiceError, ", Received error:", requestContext.Error.Error())
	}

	//Error Scenario 2
	clientContext = setupChannelClientContext(nil, errors.New(selectionServiceError), nil, t)

	//Perform action through handler
	queryHandler.Handle(requestContext, clientContext)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), selectionServiceError) {
		t.Fatal("Expected error: ", selectionServiceError, ", Received error:", requestContext.Error.Error())
	}

	//Error Scenario 3 different payload return
	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200,
		Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200,
		Payload: []byte("value1")}

	clientContext = setupChannelClientContext(nil, nil, []fab.Peer{mockPeer1, mockPeer2}, t)

	//Perform action through handler
	queryHandler.Handle(requestContext, clientContext)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), endorsementMisMatchError) {
		t.Fatal("Expected error: ", endorsementMisMatchError, ", Received error:", requestContext.Error.Error())
	}
}

func TestExecuteTxHandlerErrors(t *testing.T) {

	//Sample request
	request := Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP",
		Status: 200, Payload: []byte("value")}
	mockPeer2 := &fcmocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP",
		Status: 200, Payload: []byte("value1")}

	clientContext := setupChannelClientContext(nil, nil, []fab.Peer{mockPeer1, mockPeer2}, t)

	//Get query handler
	executeHandler := NewExecuteHandler()
	//Perform action through handler
	executeHandler.Handle(requestContext, clientContext)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), endorsementMisMatchError) {
		t.Fatal("Expected error: ", endorsementMisMatchError, ", Received error:", requestContext.Error.Error())
	}
}

func TestEndorsementHandler(t *testing.T) {
	request := Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	requestContext := prepareRequestContext(request, Opts{Targets: []fab.Peer{fcmocks.NewMockPeer("p2", "")}}, t)
	clientContext := setupChannelClientContext(nil, nil, nil, t)

	handler := NewEndorsementHandler()
	handler.Handle(requestContext, clientContext)
	assert.Nil(t, requestContext.Error)
}

func TestProposalProcessorHandler(t *testing.T) {
	peer1 := fcmocks.NewMockPeer("p1", "")
	peer2 := fcmocks.NewMockPeer("p2", "")
	discoveryPeers := []fab.Peer{peer1, peer2}

	//Get query handler
	handler := NewProposalProcessorHandler()

	request := Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}

	selectionErr := errors.New("Some selection error")
	requestContext := prepareRequestContext(request, Opts{}, t)
	handler.Handle(requestContext, setupChannelClientContext(nil, selectionErr, discoveryPeers, t))
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), selectionErr.Error()) {
		t.Fatal("Expected error: ", selectionErr, ", Received error:", requestContext.Error)
	}

	requestContext = prepareRequestContext(request, Opts{}, t)
	handler.Handle(requestContext, setupChannelClientContext(nil, nil, discoveryPeers, t))
	if requestContext.Error != nil {
		t.Fatalf("Got error: %s", requestContext.Error)
	}
	if len(requestContext.Opts.Targets) != len(discoveryPeers) {
		t.Fatalf("Expecting %d proposal processors but got %d", len(discoveryPeers), len(requestContext.Opts.Targets))
	}
	if requestContext.Opts.Targets[0] != peer1 || requestContext.Opts.Targets[1] != peer2 {
		t.Fatalf("Didn't get expected peers")
	}

	// Directly pass in the proposal processors. In this case it should use those directly
	requestContext = prepareRequestContext(request, Opts{Targets: []fab.Peer{peer2}}, t)
	handler.Handle(requestContext, setupChannelClientContext(nil, nil, discoveryPeers, t))
	if requestContext.Error != nil {
		t.Fatalf("Got error: %s", requestContext.Error)
	}
	if len(requestContext.Opts.Targets) != 1 {
		t.Fatalf("Expecting 1 proposal processor but got %d", len(requestContext.Opts.Targets))
	}
	if requestContext.Opts.Targets[0] != peer2 {
		t.Fatalf("Didn't get expected peers")
	}
}

//prepareHandlerContexts prepares context objects for handlers
func prepareRequestContext(request Request, opts Opts, t *testing.T) *RequestContext {
	requestContext := &RequestContext{Request: request,
		Opts:     opts,
		Response: Response{},
	}

	requestContext.Opts.Timeouts = make(map[core.TimeoutType]time.Duration)
	requestContext.Opts.Timeouts[core.Execute] = testTimeOut

	return requestContext
}

func setupChannelClientContext(discErr error, selectionErr error, peers []fab.Peer, t *testing.T) *ClientContext {
	membership := fcmocks.NewMockMembership()

	discoveryService, err := setupTestDiscovery(discErr, nil)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	selectionService, err := setupTestSelection(selectionErr, peers)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	ctx := setupTestContext()
	orderer := fcmocks.NewMockOrderer("", nil)
	transactor := txnmocks.MockTransactor{
		Ctx:       ctx,
		ChannelID: "testChannel",
		Orderers:  []fab.Orderer{orderer},
	}

	return &ClientContext{
		Membership: membership,
		Discovery:  discoveryService,
		Selection:  selectionService,
		Transactor: &transactor,
	}

}

func setupTestContext() context.Client {
	user := fcmocks.NewMockUser("test")
	ctx := fcmocks.NewMockContext(user)
	return ctx
}

func setupTestDiscovery(discErr error, peers []fab.Peer) (fab.DiscoveryService, error) {

	mockDiscovery, err := txnmocks.NewMockDiscoveryProvider(discErr, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockDiscoveryProvider failed")
	}

	return mockDiscovery.CreateDiscoveryService("mychannel")
}

func setupTestSelection(discErr error, peers []fab.Peer) (*txnmocks.MockSelectionService, error) {

	mockSelection, err := txnmocks.NewMockSelectionProvider(discErr, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockSelectinProvider failed")
	}

	return mockSelection.CreateSelectionService("mychannel")
}
