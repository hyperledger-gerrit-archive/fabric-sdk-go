/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txnhandler

import (
	"errors"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSignatureValidationHandlerSuccess(t *testing.T) {
	request := chclient.Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, chclient.Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupContextForSignatureValidation(nil, nil, []apifabclient.Peer{mockPeer1}, t)

	handler := NewQueryHandler()
	handler.Handle(requestContext, clientContext)
	assert.Nil(t, requestContext.Error)
}

func TestSignatureValidationCreatorValidateError(t *testing.T) {
	validateErr := errors.New("ValidateErr")
	// Sample request
	request := chclient.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}
	requestContext := prepareRequestContext(request, chclient.Opts{}, t)
	handler := NewQueryHandler()

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupContextForSignatureValidation(nil, validateErr, []apifabclient.Peer{mockPeer1}, t)
	handler.Handle(requestContext, clientContext)
	verifyExpectedError(requestContext, validateErr.Error(), t)
}

func TestSignatureValidationCreatorVerifyError(t *testing.T) {
	verifyErr := errors.New("Verify")

	// Sample request
	request := chclient.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}
	requestContext := prepareRequestContext(request, chclient.Opts{}, t)
	handler := NewQueryHandler()

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupContextForSignatureValidation(verifyErr, nil, []apifabclient.Peer{mockPeer1}, t)
	handler.Handle(requestContext, clientContext)
	verifyExpectedError(requestContext, verifyErr.Error(), t)
}

func verifyExpectedError(requestContext *chclient.RequestContext, expected string, t *testing.T) {
	assert.NotNil(t, requestContext.Error)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), expected) {
		t.Fatal("Expected error: ", expected, ", Received error:", requestContext.Error)
	}
}

func setupContextForSignatureValidation(verifyErr, validateErr error, peers []apifabclient.Peer, t *testing.T) *chclient.ClientContext {
	ctx := setupTestContext()
	memberID := fcmocks.NewMockMemberID()
	memberID.ValidateErr = validateErr
	memberID.VerifyErr = verifyErr

	discoveryService, err := setupTestDiscovery(nil, nil)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	selectionService, err := setupTestSelection(nil, peers)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	transactor := txnmocks.MockTransactor{
		Ctx:       ctx,
		ChannelID: "",
	}

	return &chclient.ClientContext{
		MemberID:   memberID,
		Discovery:  discoveryService,
		Selection:  selectionService,
		Transactor: &transactor,
	}

}

func setupTestContext() apifabclient.Context {
	user := fcmocks.NewMockUser("test")
	ctx := fcmocks.NewMockContext(user)
	return ctx
}
