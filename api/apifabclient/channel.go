/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	txn "github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// Channel ...
/**
 * Channel representing a Channel with which the client SDK interacts.
 *
 * The Channel object captures settings for a channel, which is created by
 * the orderers to isolate transactions delivery to peers participating on channel.
 * A channel must be initialized after it has been configured with the list of peers
 * and orderers. The initialization sends a get configuration block request to the
 * primary orderer to retrieve the configuration settings for this channel.
 */
type Channel interface {
	txn.Sender
	txn.ProposalSender
	ChannelRequest
	ChannelLedger
	ChannelConfig

	Name() string
	IsInitialized() bool

	// Network
	// TODO: Use PeerEndorser
	AddPeer(peer Peer) error
	RemovePeer(peer Peer)
	SetPrimaryPeer(peer Peer) error
	PrimaryPeer() Peer
	AddOrderer(orderer Orderer) error
	RemoveOrderer(orderer Orderer)
	SetMSPManager(mspManager msp.MSPManager)
	OrganizationUnits() ([]string, error)

	// Channel Info
	UpdateChannel() bool
	IsReadonly() bool

	// Query
	QueryBySystemChaincode(request txn.ChaincodeInvokeRequest) ([][]byte, error)
}

// ChannelRequest ...
type ChannelRequest interface {
	JoinChannel(request *JoinChannelRequest) error
	SendInstantiateProposal(chaincodeName string, args [][]byte, chaincodePath string, chaincodeVersion string, chaincodePolicy *common.SignaturePolicyEnvelope,
		collConfig []*common.CollectionConfig, targets []txn.ProposalProcessor) ([]*txn.TransactionProposalResponse, txn.TransactionID, error)
	SendUpgradeProposal(chaincodeName string, args [][]byte, chaincodePath string, chaincodeVersion string, chaincodePolicy *common.SignaturePolicyEnvelope, targets []txn.ProposalProcessor) ([]*txn.TransactionProposalResponse, txn.TransactionID, error)
	QueryByChaincode(txn.ChaincodeInvokeRequest) ([][]byte, error)
}

// ChannelLedger ...
type ChannelLedger interface {
	GenesisBlock() (*common.Block, error)
	QueryInfo() (*common.BlockchainInfo, error)
	QueryBlock(blockNumber int) (*common.Block, error)
	QueryBlockByHash(blockHash []byte) (*common.Block, error)
	QueryTransaction(transactionID string) (*pb.ProcessedTransaction, error)
	QueryInstantiatedChaincodes() (*pb.ChaincodeQueryResponse, error)
}

// ChannelConfig ...
type ChannelConfig interface {
	Initialize(data []byte) error
	LoadConfigUpdateEnvelope(data []byte) error
	ChannelConfig() (*common.ConfigEnvelope, error)

	MSPManager() msp.MSPManager
	Peers() []Peer
	AnchorPeers() []OrgAnchorPeer
	Orderers() []Orderer
}

// OrgAnchorPeer contains information about an anchor peer on this channel
type OrgAnchorPeer struct {
	Org  string
	Host string
	Port int32
}

// JoinChannelRequest allows a set of peers to transact on a channel on the network
type JoinChannelRequest struct {
	Targets      []Peer
	GenesisBlock *common.Block
}
