/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txnhandler

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/txnhandler"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/internal"
)

//EndorseTxHandler for handling endorse transactions
type EndorseTxHandler struct {
	next txnhandler.Handler
}

//Handle for endorsing transactions
func (e *EndorseTxHandler) Handle(requestContext *txnhandler.RequestContext, clientContext *txnhandler.ClientContext) {

	//Get proposal processor, if not supplied then use discovery service to get available peers as endorser
	//If selection service available then get endorser peers for this chaincode
	txProcessors := requestContext.Request.Opts.ProposalProcessors
	if len(txProcessors) == 0 {
		// Use discovery service to figure out proposal processors
		peers, err := clientContext.Discovery.GetPeers()
		if err != nil {
			requestContext.Request.Opts.Notifier <- apitxn.Response{Payload: nil, Error: errors.WithMessage(err, "GetPeers failed")}
			return
		}
		endorsers := peers
		if clientContext.Selection != nil {
			endorsers, err = clientContext.Selection.GetEndorsersForChaincode(peers, requestContext.Request.ChaincodeID)
			if err != nil {
				requestContext.Request.Opts.Notifier <- apitxn.Response{Payload: nil, Error: errors.WithMessage(err, "Failed to get endorsing peers")}
				return
			}
		}
		txProcessors = peer.PeersToTxnProcessors(endorsers)
	}

	// Endorse Tx
	transactionProposalResponses, txnID, err := internal.CreateAndSendTransactionProposal(clientContext.Channel,
		requestContext.Request.ChaincodeID, requestContext.Request.Fcn, requestContext.Request.Args, txProcessors, requestContext.Request.TransientMap)

	if err != nil {
		requestContext.Request.Opts.Notifier <- apitxn.Response{Payload: nil, TransactionID: txnID, Error: err}
		return
	}

	requestContext.Response.Responses = transactionProposalResponses
	requestContext.Response.TxnID = txnID

	//Delegate to next step if any
	if e.next != nil {
		e.next.Handle(requestContext, clientContext)
	}
}

//FilterTxHandler for transaction proposal response filtering
type FilterTxHandler struct {
	next txnhandler.Handler
}

//Handle for Filtering proposal response
func (f *FilterTxHandler) Handle(requestContext *txnhandler.RequestContext, clientContext *txnhandler.ClientContext) {

	//Step 5: Filter tx proposal response from Step3 through filter found in previous step (Step 4)
	var err error
	requestContext.Response.Responses, err = requestContext.Request.Opts.TxFilter.ProcessTxProposalResponse(requestContext.Response.Responses)
	if err != nil {
		requestContext.Request.Opts.Notifier <- apitxn.Response{Payload: nil, TransactionID: requestContext.Response.TxnID, Error: errors.WithMessage(err, "TxFilter failed")}
		return
	}

	var response []byte
	if len(requestContext.Response.Responses) > 0 {
		response = requestContext.Response.Responses[0].ProposalResponse.GetResponse().Payload
	}

	//Delegate to next step if any
	if f.next != nil {
		requestContext.Response.Payload <- response
		f.next.Handle(requestContext, clientContext)
	} else {
		requestContext.Request.Opts.Notifier <- apitxn.Response{Payload: response, Error: nil}
	}
}

//CommitTxHandler for committing transactions
type CommitTxHandler struct {
	next txnhandler.Handler
}

//Handle handles commit tx
func (c *CommitTxHandler) Handle(requestContext *txnhandler.RequestContext, clientContext *txnhandler.ClientContext) {

	//Connect to Event hub if not yet connected
	if clientContext.EventHub.IsConnected() == false {
		err := clientContext.EventHub.Connect()
		if err != nil {
			requestContext.Request.Opts.Notifier <- apitxn.Response{TransactionID: apitxn.TransactionID{}, Error: err}
		}
	}

	txnID := requestContext.Response.TxnID

	//Register Tx event
	statusNotifier := internal.RegisterTxEvent(txnID, clientContext.EventHub)
	_, err := internal.CreateAndSendTransaction(clientContext.Channel, requestContext.Response.Responses)
	if err != nil {
		requestContext.Request.Opts.Notifier <- apitxn.Response{TransactionID: apitxn.TransactionID{}, Error: errors.Wrap(err, "CreateAndSendTransaction failed")}
		return
	}

	select {
	case result := <-statusNotifier:
		if result.Error == nil {
			requestContext.Request.Opts.Notifier <- apitxn.Response{TransactionID: txnID, TxValidationCode: result.Code}
		} else {
			requestContext.Request.Opts.Notifier <- apitxn.Response{TransactionID: txnID, TxValidationCode: result.Code, Error: result.Error}
		}
	case <-time.After(requestContext.Request.Opts.Timeout):
		requestContext.Request.Opts.Notifier <- apitxn.Response{TransactionID: txnID, Error: errors.New("ExecuteTx didn't receive block event")}
	}

	//Delegate to next step if any
	if c.next != nil {
		c.next.Handle(requestContext, clientContext)
	}
}

//GetQueryHandler returns query handler with EndorseTxHandler & FilterTxHandler Chained
func GetQueryHandler() txnhandler.Handler {
	return &EndorseTxHandler{&FilterTxHandler{}}
}

//GetExecuteTxHandler returns query handler with EndorseTxHandler, FilterTxHandler & CommitTxHandler Chained
func GetExecuteTxHandler() txnhandler.Handler {
	return &EndorseTxHandler{&FilterTxHandler{&CommitTxHandler{}}}
}
