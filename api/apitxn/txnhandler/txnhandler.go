/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txnhandler

import (
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
)

//Handler for chaining transaction executions
type Handler interface {
	Handle(context *RequestContext, clientContext *ClientContext)
}

//ClientContext contains context parameters for handler execution
type ClientContext struct {
	Channel   apifabclient.Channel
	Discovery apifabclient.DiscoveryService
	Selection apifabclient.SelectionService
	EventHub  apifabclient.EventHub
}

//RequestContext contains request, response parameters for handler execution
type RequestContext struct {
	Request  TxRequestContext
	Response TxResponseContext
}

//TxRequestContext contains request parameters for handler execution
type TxRequestContext struct {
	ChaincodeID  string
	Fcn          string
	Args         [][]byte
	TransientMap map[string][]byte
	Opts         apitxn.TxOpts
}

//TxResponseContext contains request parameters for handler execution
type TxResponseContext struct {
	Responses []*apitxn.TransactionProposalResponse
	TxnID     apitxn.TransactionID
	Payload   chan []byte
}
