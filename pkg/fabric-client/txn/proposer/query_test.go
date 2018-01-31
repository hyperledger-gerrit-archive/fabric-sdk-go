/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package proposer

import (
	"reflect"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

func TestQueryMissingParams(t *testing.T) {
	channel, _ := setupTestChannel()

	request := apitxn.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	_, err := channel.QueryByChaincode(request)
	if err == nil {
		t.Fatalf("Expected error")
	}
	_, err = queryByChaincode("", request, channel.clientContext)
	if err == nil {
		t.Fatalf("Expected error")
	}

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A")}
	channel.AddPeer(&peer)

	request = apitxn.ChaincodeInvokeRequest{
		Fcn: "Hello",
	}
	_, err = channel.QueryByChaincode(request)
	if err == nil {
		t.Fatalf("Expected error")
	}

	request = apitxn.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
	}
	_, err = channel.QueryByChaincode(request)
	if err == nil {
		t.Fatalf("Expected error")
	}

	request = apitxn.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	_, err = channel.QueryByChaincode(request)
	if err != nil {
		t.Fatalf("Expected success")
	}
}

func TestQueryBySystemChaincode(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A")}
	channel.AddPeer(&peer)

	request := apitxn.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	resp, err := channel.QueryBySystemChaincode(request)
	if err != nil {
		t.Fatalf("Failed to query: %s", err)
	}
	expectedResp := []byte("A")

	if !reflect.DeepEqual(resp[0], expectedResp) {
		t.Fatalf("Unexpected transaction proposal response: %v", resp)
	}
}
