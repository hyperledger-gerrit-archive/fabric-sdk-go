/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// ProposalProcessor simulates transaction proposal, so that a client can submit the result for ordering.
type ProposalProcessor interface {
	ProcessTransactionProposal(ProcessProposalRequest) (*TransactionProposalResponse, error)
}

// ProposalSender provides the ability for a transaction proposal to be created and sent.
type ProposalSender interface {
	//CreateTxnID() (TransactionID, error)
	CreateTransactionProposal(TransactionProposalRequest) (*TransactionProposal, error)
	SendTransactionProposal(*TransactionProposal, []ProposalProcessor) ([]*TransactionProposalResponse, error)
}

// TransactionID contains the ID of a Fabric Transaction Proposal
type TransactionID struct {
	ID    string
	Nonce []byte
}

// TransactionProposalRequest contains the parameters for sending a transaction proposal.
type TransactionProposalRequest struct {
	ChaincodeID  string
	TransientMap map[string][]byte
	Fcn          string
	Args         [][]byte
}

// ChaincodeInvokeRequest contains the parameters for sending a transaction proposal.
//
// Deprecated: this struct has been replaced by TransactionProposalRequest.
type ChaincodeInvokeRequest struct {
	Targets      []ProposalProcessor // TODO: remove
	ChaincodeID  string
	TransientMap map[string][]byte
	Fcn          string
	Args         [][]byte
}

// TransactionProposal contains a marashalled transaction proposal.
type TransactionProposal struct {
	TxnID TransactionID // TODO: remove?
	*pb.Proposal
}

// ProcessProposalRequest requests simulation of a proposed transaction from transaction processors.
type ProcessProposalRequest struct {
	SignedProposal *pb.SignedProposal
}

// TransactionProposalResponse respresents the result of transaction proposal processing.
type TransactionProposalResponse struct {
	Endorser string
	Status   int32
	*pb.ProposalResponse
}
