/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn"
	"github.com/pkg/errors"
)

// CreateTransaction create a transaction with proposal response, following the endorsement policy.
func (c *Channel) CreateTransaction(request fab.TransactionRequest) (*fab.Transaction, error) {
	return txn.New(request)
}

// SendTransaction send a transaction to the chain’s orderer service (one or more orderer endpoints) for consensus and committing to the ledger.
func (c *Channel) SendTransaction(tx *fab.Transaction) (*fab.TransactionResponse, error) {
	return txn.Send(c.clientContext, tx, c.Orderers())
}

// SendTransactionProposal sends the created proposal to peer for endorsement.
// TODO: return the entire request or just the txn ID?
func (c *Channel) SendTransactionProposal(request fab.ChaincodeInvokeRequest) ([]*fab.TransactionProposalResponse, fab.TransactionID, error) {
	targets, err := c.chaincodeInvokeRequestAddDefaultPeers(request.Targets)
	if err != nil {
		return nil, fab.TransactionID{}, err
	}

	tpRequest := fab.TransactionProposalRequest{
		ChaincodeID:  request.ChaincodeID,
		TransientMap: request.TransientMap,
		Fcn:          request.Fcn,
		Args:         request.Args,
	}

	tp, err := txn.NewProposal(c.clientContext, c.name, tpRequest)
	if err != nil {
		return nil, fab.TransactionID{}, errors.WithMessage(err, "new transaction proposal failed")
	}

	tpr, err := txn.SendProposal(c.clientContext, tp, targets)
	if err != nil {
		return nil, fab.TransactionID{}, errors.WithMessage(err, "send transaction proposal failed")
	}

	return tpr, tp.TxnID, nil
}
