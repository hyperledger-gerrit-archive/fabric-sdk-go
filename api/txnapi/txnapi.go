/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package txnapi allows SDK users to plugin their own implementations of transaction processing.
package txnapi

import (
	pb "github.com/hyperledger/fabric/protos/peer"
)

// TxnSender ...
type TxnSender interface {
	CreateTransaction(resps []*TransactionProposalResponse) (*Transaction, error)
	SendTransaction(tx *Transaction) ([]*TransactionResponse, error)
}

// TxnProposalSender ...
type TxnProposalSender interface {
	CreateTransactionProposal(chaincodeName string, channelID string, args []string, sign bool, transientData map[string][]byte) (*TransactionProposal, error)
	SendTransactionProposal(proposal *TransactionProposal, retry int, targets []TxnProposalProcessor) ([]*TransactionProposalResponse, error)
}

// The Transaction object created from an endorsed proposal
type Transaction struct {
	Proposal    *TransactionProposal
	Transaction *pb.Transaction
}

// TransactionResponse ...
/**
 * The TransactionProposalResponse result object returned from orderers.
 */
type TransactionResponse struct {
	Orderer string
	Err     error
}
