/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package invoke

import (
	"errors"
	"strings"
	"testing"

	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSignatureValidationHandlerSuccess(t *testing.T) {
	request := Request{ChaincodeID: "test", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}}

	//Prepare context objects for handler
	requestContext := prepareRequestContext(request, Opts{}, t)

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupContextForSignatureValidation(nil, nil, []fab.Peer{mockPeer1}, t)

	handler := NewQueryHandler()
	handler.Handle(requestContext, clientContext)
	assert.Nil(t, requestContext.Error)
}

func TestSignatureValidationCreatorValidateError(t *testing.T) {
	validateErr := errors.New("ValidateErr")
	// Sample request
	request := Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}
	requestContext := prepareRequestContext(request, Opts{}, t)
	handler := NewQueryHandler()

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupContextForSignatureValidation(nil, validateErr, []fab.Peer{mockPeer1}, t)
	handler.Handle(requestContext, clientContext)
	verifyExpectedError(requestContext, validateErr.Error(), t)
}

func TestSignatureValidationCreatorVerifyError(t *testing.T) {
	verifyErr := errors.New("Verify")

	// Sample request
	request := Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}
	requestContext := prepareRequestContext(request, Opts{}, t)
	handler := NewQueryHandler()

	mockPeer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: 200, Payload: []byte("value")}

	clientContext := setupContextForSignatureValidation(verifyErr, nil, []fab.Peer{mockPeer1}, t)
	handler.Handle(requestContext, clientContext)
	verifyExpectedError(requestContext, verifyErr.Error(), t)
}

func verifyExpectedError(requestContext *RequestContext, expected string, t *testing.T) {
	assert.NotNil(t, requestContext.Error)
	if requestContext.Error == nil || !strings.Contains(requestContext.Error.Error(), expected) {
		t.Fatal("Expected error: ", expected, ", Received error:", requestContext.Error)
	}
}

func setupContextForSignatureValidation(verifyErr, validateErr error, peers []fab.Peer, t *testing.T) *ClientContext {
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

	return &ClientContext{
		MemberID:   memberID,
		Discovery:  discoveryService,
		Selection:  selectionService,
		Transactor: &transactor,
	}

}
