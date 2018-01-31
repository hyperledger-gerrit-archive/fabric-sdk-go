/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chledger

import (
	"bytes"
	"strconv"

	"github.com/golang/protobuf/proto"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	txn "github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

const systemChannel = ""

// Ledger holds the context for interacting with a Channel's ledger via the Orderer.
type Ledger struct {
	ctx fab.Context
	cfg fab.ChannelConfig
}

// New allows interaction with a Channel's ledger based on the context and channel config.
func New(ctx fab.Context, cfg fab.ChannelCfg) *Ledger {
	l := Ledger{ctx, cfg}
	return &l
}

// QueryInfo queries for various useful information on the state of the channel
// (height, known peers).
// This query will be made to the primary peer.
func QueryInfo() (*common.BlockchainInfo, error) {
	logger.Debug("queryInfo - start")

	// prepare arguments to call qscc GetChainInfo function
	var args [][]byte
	args = append(args, []byte(c.Name()))

	payload, err := queryBySystemChaincodeByTarget("qscc", "GetChainInfo", args, c.PrimaryPeer())
	if err != nil {
		return nil, errors.WithMessage(err, "qscc.GetChainInfo failed")
	}

	bci := &common.BlockchainInfo{}
	err = proto.Unmarshal(payload, bci)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal of BlockchainInfo failed")
	}

	return bci, nil
}

// QueryBlockByHash queries the ledger for Block by block hash.
// This query will be made to the primary peer.
// Returns the block.
func QueryBlockByHash(blockHash []byte) (*common.Block, error) {

	if blockHash == nil {
		return nil, errors.New("blockHash is required")
	}

	// prepare arguments to call qscc GetBlockByNumber function
	var args [][]byte
	args = append(args, []byte(c.Name()))
	args = append(args, blockHash[:len(blockHash)])

	payload, err := c.queryBySystemChaincodeByTarget("qscc", "GetBlockByHash", args, c.PrimaryPeer())
	if err != nil {
		return nil, errors.WithMessage(err, "qscc.GetBlockByHash failed")
	}

	block := &common.Block{}
	err = proto.Unmarshal(payload, block)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal of BlockchainInfo failed")
	}

	return block, nil
}

// QueryBlock queries the ledger for Block by block number.
// This query will be made to the primary peer.
// blockNumber: The number which is the ID of the Block.
// It returns the block.
func QueryBlock(blockNumber int) (*common.Block, error) {

	if blockNumber < 0 {
		return nil, errors.New("blockNumber must be a positive integer")
	}

	// prepare arguments to call qscc GetBlockByNumber function
	var args [][]byte
	args = append(args, []byte(c.Name()))
	args = append(args, []byte(strconv.Itoa(blockNumber)))

	payload, err := c.queryBySystemChaincodeByTarget("qscc", "GetBlockByNumber", args, c.PrimaryPeer())
	if err != nil {
		return nil, errors.WithMessage(err, "qscc.GetBlockByNumber failed")
	}

	block := &common.Block{}
	err = proto.Unmarshal(payload, block)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal of BlockchainInfo failed")
	}

	return block, nil
}

// QueryTransaction queries the ledger for Transaction by number.
// This query will be made to the primary peer.
// Returns the ProcessedTransaction information containing the transaction.
// TODO: add optional target
func QueryTransaction(transactionID string) (*pb.ProcessedTransaction, error) {

	// prepare arguments to call qscc GetTransactionByID function
	var args [][]byte
	args = append(args, []byte(c.Name()))
	args = append(args, []byte(transactionID))

	payload, err := c.queryBySystemChaincodeByTarget("qscc", "GetTransactionByID", args, c.PrimaryPeer())
	if err != nil {
		return nil, errors.WithMessage(err, "qscc.GetTransactionByID failed")
	}

	transaction := new(pb.ProcessedTransaction)
	err = proto.Unmarshal(payload, transaction)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal of ProcessedTransaction failed")
	}

	return transaction, nil
}

// QueryInstantiatedChaincodes queries the instantiated chaincodes on this channel.
// This query will be made to the primary peer.
func QueryInstantiatedChaincodes() (*pb.ChaincodeQueryResponse, error) {

	targets := []txn.ProposalProcessor{c.PrimaryPeer()}
	request := txn.ChaincodeInvokeRequest{
		Targets:     targets,
		ChaincodeID: "lscc",
		Fcn:         "getchaincodes",
	}

	payload, err := c.QueryByChaincode(request)
	if err != nil {
		return nil, errors.WithMessage(err, "lscc.getchaincodes failed")
	}

	response := new(pb.ChaincodeQueryResponse)
	err = proto.Unmarshal(payload[0], response)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal of ChaincodeQueryResponse failed")
	}

	return response, nil
}

// QueryConfigBlock returns the current configuration block for the specified channel. If the
// peer doesn't belong to the channel, return error
func QueryConfigBlock(peers []fab.Peer, minResponses int) (*common.ConfigEnvelope, error) {

	if len(peers) == 0 {
		return nil, errors.New("peer(s) required")
	}

	if minResponses <= 0 {
		return nil, errors.New("Minimum endorser has to be greater than zero")
	}

	request := txn.ChaincodeInvokeRequest{
		ChaincodeID: "cscc",
		Fcn:         "GetConfigBlock",
		Args:        [][]byte{[]byte(c.Name())},
		Targets:     peersToTxnProcessors(peers),
	}

	// we are using system channel here (query to system cc)
	transactionProposalResponses, _, err := proposer.SendTransactionProposalWithChannelID(systemChannel, request, c.clientContext)
	if err != nil {
		return nil, errors.WithMessage(err, "SendTransactionProposalWithChannelID failed")
	}

	responses, err := filterProposalResponses(transactionProposalResponses)
	if err != nil {
		return nil, err
	}

	if len(responses) < minResponses {
		return nil, errors.Errorf("Required minimum %d endorsments got %d", minResponses, len(responses))
	}

	r := responses[0]
	for _, p := range responses {
		if bytes.Compare(r, p) != 0 {
			return nil, errors.New("Payloads for config block do not match")
		}
	}

	block := &common.Block{}
	err = proto.Unmarshal(responses[0], block)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal block failed")
	}

	if block.Data == nil || block.Data.Data == nil {
		return nil, errors.New("config block data is nil")
	}

	if len(block.Data.Data) != 1 {
		return nil, errors.New("config block must contain one transaction")
	}

	return createConfigEnvelope(block.Data.Data[0])

}

// queryBySystemChaincodeByTarget is an internal helper function that queries system chaincode.
// This function is not exported to keep the external interface of this package to only expose
// request structs.
func queryBySystemChaincodeByTarget(chaincodeID string, fcn string, args [][]byte, target txn.ProposalProcessor) ([]byte, error) {
	targets := []txn.ProposalProcessor{target}
	request := txn.ChaincodeInvokeRequest{
		ChaincodeID: chaincodeID,
		Fcn:         fcn,
		Args:        args,
		Targets:     targets,
	}
	responses, err := proposer.QueryBySystemChaincode(request)

	// we are only querying one peer hence one result
	if err != nil || len(responses) != 1 {
		return nil, errors.Errorf("QueryBySystemChaincode should have one result only, actual result is %d", len(responses))
	}

	return responses[0], nil
}
