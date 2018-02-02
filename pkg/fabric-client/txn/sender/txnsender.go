/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sender

import (
	"math/rand"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn/env"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"
)

var logger = logging.NewLogger("fabric_sdk_go")

// CCProposalType reflects transitions in the chaincode lifecycle
type CCProposalType int

// Define chaincode proposal types
const (
	Instantiate CCProposalType = iota
	Upgrade
)

type context interface {
	SigningManager() fab.SigningManager
	Config() apiconfig.Config
	fab.IdentityContext
}

// CreateTransaction create a transaction with proposal response, following the endorsement policy.
func CreateTransaction(resps []*apitxn.TransactionProposalResponse) (*apitxn.Transaction, error) {
	if len(resps) == 0 {
		return nil, errors.New("at least one proposal response is necessary")
	}

	proposal := &resps[0].Proposal

	// the original header
	hdr, err := protos_utils.GetHeader(proposal.Proposal.Header)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal proposal header failed")
	}

	// the original payload
	pPayl, err := protos_utils.GetChaincodeProposalPayload(proposal.Proposal.Payload)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal proposal payload failed")
	}

	// get header extensions so we have the visibility field
	hdrExt, err := protos_utils.GetChaincodeHeaderExtension(hdr)
	if err != nil {
		return nil, err
	}

	for _, r := range resps {
		if r.ProposalResponse.Response.Status != 200 {
			return nil, errors.Errorf("proposal response was not successful, error code %d, msg %s", r.ProposalResponse.Response.Status, r.ProposalResponse.Response.Message)
		}
	}

	// fill endorsements
	endorsements := make([]*pb.Endorsement, len(resps))
	for n, r := range resps {
		endorsements[n] = r.ProposalResponse.Endorsement
	}
	// create ChaincodeEndorsedAction
	cea := &pb.ChaincodeEndorsedAction{ProposalResponsePayload: resps[0].ProposalResponse.Payload, Endorsements: endorsements}

	// obtain the bytes of the proposal payload that will go to the transaction
	propPayloadBytes, err := protos_utils.GetBytesProposalPayloadForTx(pPayl, hdrExt.PayloadVisibility)
	if err != nil {
		return nil, err
	}

	// serialize the chaincode action payload
	cap := &pb.ChaincodeActionPayload{ChaincodeProposalPayload: propPayloadBytes, Action: cea}
	capBytes, err := protos_utils.GetBytesChaincodeActionPayload(cap)
	if err != nil {
		return nil, err
	}

	// create a transaction
	taa := &pb.TransactionAction{Header: hdr.SignatureHeader, Payload: capBytes}
	taas := make([]*pb.TransactionAction, 1)
	taas[0] = taa

	return &apitxn.Transaction{
		Transaction: &pb.Transaction{Actions: taas},
		Proposal:    proposal,
	}, nil
}

// SendTransaction send a transaction to the chain’s orderer service (one or more orderer endpoints) for consensus and committing to the ledger.
func SendTransaction(ctx context, tx *apitxn.Transaction, orderers []fab.Orderer) (*apitxn.TransactionResponse, error) {
	if orderers == nil || len(orderers) == 0 {
		return nil, errors.New("orderers is nil")
	}
	if tx == nil {
		return nil, errors.New("transaction is nil")
	}
	if tx.Proposal == nil || tx.Proposal.Proposal == nil {
		return nil, errors.New("proposal is nil")
	}

	// the original header
	hdr, err := protos_utils.GetHeader(tx.Proposal.Proposal.Header)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal proposal header failed")
	}
	// serialize the tx
	txBytes, err := protos_utils.GetBytesTransaction(tx.Transaction)
	if err != nil {
		return nil, err
	}

	// create the payload
	payl := &common.Payload{Header: hdr, Data: txBytes}
	paylBytes, err := protos_utils.GetBytesPayload(payl)
	if err != nil {
		return nil, err
	}

	// here's the envelope
	envelope, err := env.SignPayload(ctx, paylBytes)
	if err != nil {
		return nil, err
	}

	transactionResponse, err := BroadcastEnvelope(envelope, orderers)
	if err != nil {
		return nil, err
	}

	return transactionResponse, nil
}

// SendInstantiateProposal sends an instantiate proposal to one or more endorsing peers.
// chaincodeName: required - The name of the chain.
// args: optional - string Array arguments specific to the chaincode being instantiated
// chaincodePath: required - string of the path to the location of the source code of the chaincode
// chaincodeVersion: required - string of the version of the chaincode
// chaincodePolicy: required - chaincode signature policy
// collConfig: optional - private data collection configuration
func SendInstantiateProposal(ctx context, channelID string, chaincodeName string,
	args [][]byte, chaincodePath string, chaincodeVersion string,
	chaincodePolicy *common.SignaturePolicyEnvelope,
	collConfig []*common.CollectionConfig, targets []apitxn.ProposalProcessor) ([]*apitxn.TransactionProposalResponse, apitxn.TransactionID, error) {

	return sendCCProposal(ctx, Instantiate, channelID, chaincodeName, args, chaincodePath, chaincodeVersion, chaincodePolicy, collConfig, targets)

}

// SendUpgradeProposal sends an upgrade proposal to one or more endorsing peers.
// chaincodeName: required - The name of the chain.
// args: optional - string Array arguments specific to the chaincode being upgraded
// chaincodePath: required - string of the path to the location of the source code of the chaincode
// chaincodeVersion: required - string of the version of the chaincode
func SendUpgradeProposal(ctx context, channelID string, chaincodeName string,
	args [][]byte, chaincodePath string, chaincodeVersion string,
	chaincodePolicy *common.SignaturePolicyEnvelope, targets []apitxn.ProposalProcessor) ([]*apitxn.TransactionProposalResponse, apitxn.TransactionID, error) {

	return sendCCProposal(ctx, Upgrade, channelID, chaincodeName, args, chaincodePath, chaincodeVersion, chaincodePolicy, nil, targets)

}

// helper function that sends an instantiate or upgrade chaincode proposal to one or more endorsing peers
func sendCCProposal(ctx context, ccProposalType CCProposalType,
	channelID string, chaincodeName string,
	args [][]byte, chaincodePath string, chaincodeVersion string,
	chaincodePolicy *common.SignaturePolicyEnvelope,
	collConfig []*common.CollectionConfig,
	targets []apitxn.ProposalProcessor) ([]*apitxn.TransactionProposalResponse, apitxn.TransactionID, error) {

	if chaincodeName == "" {
		return nil, apitxn.TransactionID{}, errors.New("chaincodeName is required")
	}
	if chaincodePath == "" {
		return nil, apitxn.TransactionID{}, errors.New("chaincodePath is required")
	}
	if chaincodeVersion == "" {
		return nil, apitxn.TransactionID{}, errors.New("chaincodeVersion is required")
	}
	if chaincodePolicy == nil {
		return nil, apitxn.TransactionID{}, errors.New("chaincodePolicy is required")
	}

	if targets == nil || len(targets) < 1 {
		return nil, apitxn.TransactionID{}, errors.New("missing peer objects for chaincode proposal")
	}

	ccds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincodeName, Path: chaincodePath, Version: chaincodeVersion},
		Input: &pb.ChaincodeInput{Args: args}}}

	creator, err := ctx.Identity()
	if err != nil {
		return nil, apitxn.TransactionID{}, errors.Wrap(err, "getting user context's identity failed")
	}
	chaincodePolicyBytes, err := protos_utils.Marshal(chaincodePolicy)
	if err != nil {
		return nil, apitxn.TransactionID{}, err
	}
	var collConfigBytes []byte
	if collConfig != nil {
		var err error
		collConfigBytes, err = proto.Marshal(&common.CollectionConfigPackage{Config: collConfig})
		if err != nil {
			return nil, apitxn.TransactionID{}, err
		}
	}

	var proposal *pb.Proposal
	var txID string

	switch ccProposalType {

	case Instantiate:
		proposal, txID, err = protos_utils.CreateDeployProposalFromCDS(channelID, ccds, creator, chaincodePolicyBytes, []byte("escc"), []byte("vscc"), collConfigBytes)
		if err != nil {
			return nil, apitxn.TransactionID{}, errors.Wrap(err, "create instantiate chaincode proposal failed")
		}
	case Upgrade:
		proposal, txID, err = protos_utils.CreateUpgradeProposalFromCDS(channelID, ccds, creator, chaincodePolicyBytes, []byte("escc"), []byte("vscc"))
		if err != nil {
			return nil, apitxn.TransactionID{}, errors.Wrap(err, "create  upgrade chaincode proposal failed")
		}
	default:
		return nil, apitxn.TransactionID{}, errors.Errorf("chaincode proposal type %d not supported", ccProposalType)
	}

	signedProposal, err := signProposal(ctx, proposal)
	if err != nil {
		return nil, apitxn.TransactionID{}, err
	}

	txnID := apitxn.TransactionID{ID: txID} // Nonce is missing

	transactionProposalResponse, err := env.SendTransactionProposalToProcessors(&apitxn.TransactionProposal{
		SignedProposal: signedProposal,
		Proposal:       proposal,
		TxnID:          txnID,
	}, targets)

	return transactionProposalResponse, txnID, err
}

// BroadcastEnvelope will send the given envelope to some orderer, picking random endpoints
// until all are exhausted
func BroadcastEnvelope(envelope *fab.SignedEnvelope, orderers []fab.Orderer) (*apitxn.TransactionResponse, error) {
	// Check if orderers are defined
	if len(orderers) == 0 {
		return nil, errors.New("orderers not set")
	}

	// Copy aside the ordering service endpoints
	randOrderers := []fab.Orderer{}
	for _, o := range orderers {
		randOrderers = append(randOrderers, o)
	}

	// Iterate them in a random order and try broadcasting 1 by 1
	var errResp *apitxn.TransactionResponse
	for _, i := range rand.Perm(len(randOrderers)) {
		resp := sendBroadcast(envelope, randOrderers[i])
		if resp.Err != nil {
			errResp = resp
		} else {
			return resp, nil
		}
	}
	return errResp, nil
}

func sendBroadcast(envelope *fab.SignedEnvelope, orderer fab.Orderer) *apitxn.TransactionResponse {
	logger.Debugf("Broadcasting envelope to orderer :%s\n", orderer.URL())
	if _, err := orderer.SendBroadcast(envelope); err != nil {
		logger.Debugf("Receive Error Response from orderer :%v\n", err)
		return &apitxn.TransactionResponse{Orderer: orderer.URL(),
			Err: errors.Wrapf(err, "calling orderer '%s' failed", orderer.URL())}
	}

	logger.Debugf("Receive Success Response from orderer\n")
	return &apitxn.TransactionResponse{Orderer: orderer.URL(), Err: nil}
}

// SendEnvelope sends the given envelope to each orderer and returns a block response
func SendEnvelope(ctx context, envelope *fab.SignedEnvelope, orderers []fab.Orderer) (*common.Block, error) {
	if orderers == nil || len(orderers) == 0 {
		return nil, errors.New("orderers not set")
	}

	var blockResponse *common.Block
	var errorResponse error
	var mutex sync.Mutex
	outstandingRequests := len(orderers)
	done := make(chan bool)

	// Send the request to all orderers and return as soon as one responds with a block.
	for _, o := range orderers {

		go func(orderer fab.Orderer) {
			logger.Debugf("Broadcasting envelope to orderer :%s\n", orderer.URL())

			blocks, errs := orderer.SendDeliver(envelope)
			select {
			case block := <-blocks:
				mutex.Lock()
				if blockResponse == nil {
					blockResponse = block
					done <- true
				}
				mutex.Unlock()

			case err := <-errs:
				mutex.Lock()
				if errorResponse == nil {
					errorResponse = err
				}
				outstandingRequests--
				if outstandingRequests == 0 {
					done <- true
				}
				mutex.Unlock()

			case <-time.After(ctx.Config().TimeoutOrDefault(apiconfig.OrdererResponse)):
				mutex.Lock()
				if errorResponse == nil {
					errorResponse = errors.New("timeout waiting for response from orderer")
				}
				outstandingRequests--
				if outstandingRequests == 0 {
					done <- true
				}
				mutex.Unlock()
			}
		}(o)
	}

	<-done

	if blockResponse != nil {
		return blockResponse, nil
	}

	// There must be an error
	if errorResponse != nil {
		return nil, errors.Wrap(errorResponse, "error returned from orderer service")
	}

	return nil, errors.New("unexpected: didn't receive a block from any of the orderer servces and didn't receive any error")
}

func signProposal(ctx context, proposal *pb.Proposal) (*pb.SignedProposal, error) {
	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "mashal proposal failed")
	}

	signingMgr := ctx.SigningManager()
	if signingMgr == nil {
		return nil, errors.New("signing manager is nil")
	}

	signature, err := signingMgr.Sign(proposalBytes, ctx.PrivateKey())
	if err != nil {
		return nil, errors.WithMessage(err, "signing proposal failed")
	}

	return &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}, nil
}
