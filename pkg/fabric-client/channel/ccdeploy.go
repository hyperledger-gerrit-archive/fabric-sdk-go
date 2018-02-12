/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
)

// ChaincodeProposalType reflects transitions in the chaincode lifecycle
type ChaincodeProposalType int

// Define chaincode proposal types
const (
	InstantiateChaincode ChaincodeProposalType = iota
	UpgradeChaincode
)

// SendInstantiateProposal sends an instantiate proposal to one or more endorsing peers.
// chaincodeName: required - The name of the chain.
// args: optional - string Array arguments specific to the chaincode being instantiated
// chaincodePath: required - string of the path to the location of the source code of the chaincode
// chaincodeVersion: required - string of the version of the chaincode
// chaincodePolicy: required - chaincode signature policy
// collConfig: optional - private data collection configuration
func (c *Channel) SendInstantiateProposal(chaincodeName string,
	args [][]byte, chaincodePath string, chaincodeVersion string,
	chaincodePolicy *common.SignaturePolicyEnvelope,
	collConfig []*common.CollectionConfig, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, fab.TransactionID, error) {

	cp := ChaincodeDeployProposal{
		Name:       chaincodeName,
		Args:       args,
		Path:       chaincodePath,
		Version:    chaincodeVersion,
		Policy:     chaincodePolicy,
		CollConfig: collConfig,
	}

	tp, err := CreateChaincodeDeployProposal(c.clientContext, InstantiateChaincode, c.name, cp)
	if err != nil {
		return nil, fab.TransactionID{}, errors.WithMessage(err, "creation of chaincode proposal failed")
	}

	tpr, err := txn.SendProposal(c.clientContext, tp, targets)
	return tpr, tp.TxnID, err
}

// SendUpgradeProposal sends an upgrade proposal to one or more endorsing peers.
// chaincodeName: required - The name of the chain.
// args: optional - string Array arguments specific to the chaincode being upgraded
// chaincodePath: required - string of the path to the location of the source code of the chaincode
// chaincodeVersion: required - string of the version of the chaincode
func (c *Channel) SendUpgradeProposal(chaincodeName string,
	args [][]byte, chaincodePath string, chaincodeVersion string,
	chaincodePolicy *common.SignaturePolicyEnvelope, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, fab.TransactionID, error) {

	cp := ChaincodeDeployProposal{
		Name:    chaincodeName,
		Args:    args,
		Path:    chaincodePath,
		Version: chaincodeVersion,
		Policy:  chaincodePolicy,
	}

	tp, err := CreateChaincodeDeployProposal(c.clientContext, UpgradeChaincode, c.name, cp)
	if err != nil {
		return nil, fab.TransactionID{}, errors.WithMessage(err, "creation of chaincode proposal failed")
	}

	tpr, err := txn.SendProposal(c.clientContext, tp, targets)
	return tpr, tp.TxnID, err
}

func validateChaincodeInvokeRequest(request fab.ChaincodeInvokeRequest) error {
	if request.ChaincodeID == "" {
		return errors.New("ChaincodeID is required")
	}

	if request.Fcn == "" {
		return errors.New("Fcn is required")
	}
	return nil
}

func (c *Channel) chaincodeInvokeRequestAddDefaultPeers(targets []fab.ProposalProcessor) ([]fab.ProposalProcessor, error) {
	// Use default peers if targets are not specified.
	if targets == nil || len(targets) == 0 {
		if c.peers == nil || len(c.peers) == 0 {
			return nil, status.New(status.ClientStatus, status.NoPeersFound.ToInt32(),
				"targets were not specified and no peers have been configured", nil)
		}

		return peersToTxnProcessors(c.Peers()), nil
	}
	return targets, nil
}

// ChaincodeDeployProposal holds parameters for creating an instantiate or upgrade chaincode proposal.
type ChaincodeDeployProposal struct {
	Name       string
	Path       string
	Version    string
	Args       [][]byte
	Policy     *common.SignaturePolicyEnvelope
	CollConfig []*common.CollectionConfig
}

// CreateChaincodeDeployProposal creates an instantiate or upgrade chaincode proposal.
func CreateChaincodeDeployProposal(ctx fab.Context, deploy ChaincodeProposalType, channelID string, chaincode ChaincodeDeployProposal) (*fab.TransactionProposal, error) {

	if chaincode.Name == "" {
		return nil, errors.New("chaincodeName is required")
	}
	if chaincode.Path == "" {
		return nil, errors.New("chaincodePath is required")
	}
	if chaincode.Version == "" {
		return nil, errors.New("chaincodeVersion is required")
	}
	if chaincode.Policy == nil {
		return nil, errors.New("chaincodePolicy is required")
	}

	ccds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincode.Name, Path: chaincode.Path, Version: chaincode.Version},
		Input: &pb.ChaincodeInput{Args: chaincode.Args}}}

	creator, err := ctx.Identity()
	if err != nil {
		return nil, errors.Wrap(err, "getting user context's identity failed")
	}
	chaincodePolicyBytes, err := protos_utils.Marshal(chaincode.Policy)
	if err != nil {
		return nil, err
	}
	var collConfigBytes []byte
	if chaincode.CollConfig != nil {
		var err error
		collConfigBytes, err = proto.Marshal(&common.CollectionConfigPackage{Config: chaincode.CollConfig})
		if err != nil {
			return nil, err
		}
	}

	var proposal *pb.Proposal
	var txID string

	switch deploy {

	case InstantiateChaincode:
		proposal, txID, err = protos_utils.CreateDeployProposalFromCDS(channelID, ccds, creator, chaincodePolicyBytes, []byte("escc"), []byte("vscc"), collConfigBytes)
		if err != nil {
			return nil, errors.Wrap(err, "create instantiate chaincode proposal failed")
		}
	case UpgradeChaincode:
		proposal, txID, err = protos_utils.CreateUpgradeProposalFromCDS(channelID, ccds, creator, chaincodePolicyBytes, []byte("escc"), []byte("vscc"))
		if err != nil {
			return nil, errors.Wrap(err, "create  upgrade chaincode proposal failed")
		}
	default:
		return nil, errors.Errorf("chaincode proposal type %d not supported", deploy)
	}

	txnID := fab.TransactionID{ID: txID} // Nonce is missing
	tp := fab.TransactionProposal{
		Proposal: proposal,
		TxnID:    txnID,
	}

	return &tp, err
}

// peersToTxnProcessors converts a slice of Peers to a slice of ProposalProcessors
func peersToTxnProcessors(peers []fab.Peer) []fab.ProposalProcessor {
	tpp := make([]fab.ProposalProcessor, len(peers))

	for i := range peers {
		tpp[i] = peers[i]
	}
	return tpp
}
