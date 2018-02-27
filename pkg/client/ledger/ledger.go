/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package ledger enables ability to query ledger in a Fabric network.
package ledger

import (
	"math/rand"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/greylist"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

const (
	defaultHandlerTimeout = time.Second * 10
)

// Client enables ledger queries on a Fabric network.
//
// A ledger client instance provides a handler to query various info on specified channel.
// An application that requires interaction with multiple channels should create a separate
// instance of the ledger client for each channel. Ledger client supports specific queries only.
type Client struct {
	provider  context.ProviderContext
	identity  context.IdentityContext
	discovery fab.DiscoveryService
	ledger    *channel.Ledger
	greylist  *greylist.Filter
	filter    TargetFilter
}

// Context holds the providers and services needed to create a Client.
type Context struct {
	context.ProviderContext
	context.IdentityContext
	DiscoveryService fab.DiscoveryService
	ChannelService   fab.ChannelService
}

// MSPFilter is default filter
type MSPFilter struct {
	mspID string
}

// Accept returns true if this peer is to be included in the target list
func (f *MSPFilter) Accept(peer fab.Peer) bool {
	return peer.MSPID() == f.mspID
}

// New returns a Client instance.
func New(c Context, chName string, opts ...ClientOption) (*Client, error) {

	greylistProvider := greylist.New(c.Config().TimeoutOrDefault(core.DiscoveryGreylistExpiry))

	l, err := channel.NewLedger(c, chName)
	if err != nil {
		return nil, err
	}

	ledgerClient := Client{
		greylist:  greylistProvider,
		provider:  c,
		identity:  c,
		discovery: discovery.NewDiscoveryFilterService(c.DiscoveryService, greylistProvider),
		ledger:    l,
	}

	for _, opt := range opts {
		err := opt(&ledgerClient)
		if err != nil {
			return nil, err
		}
	}

	// check if target filter was set - if not set the default
	if ledgerClient.filter == nil {
		// Default target filter is based on user msp
		if c.MspID() == "" {
			return nil, errors.New("mspID not available in user context")
		}
		filter := &MSPFilter{mspID: c.MspID()}
		ledgerClient.filter = filter
	}

	return &ledgerClient, nil
}

// QueryInfo queries for various useful information on the state of the channel
// (height, known peers).
func (c *Client) QueryInfo(options ...RequestOption) (*common.BlockchainInfo, error) {

	opts, err := c.prepareRequestOpts(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get opts for QueryBlockByHash")
	}

	// Determine targets
	targets, err := c.calculateTargets(opts)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to determine target peers for QueryBlockByHash")
	}

	responses, err := c.ledger.QueryInfo(peersToTxnProcessors(targets))
	if err != nil && len(responses) == 0 {
		return nil, errors.WithMessage(err, "Failed to QueryBlockByHash")
	}

	if len(responses) < opts.MinMatches {
		return nil, errors.Errorf("Number of responses %d is less than MinMatches %d", len(responses), opts.MinMatches)
	}

	response := responses[0]
	maxHeight := response.Height
	for i, r := range responses {
		if i == 0 {
			continue
		}

		// Match one with highest block height
		if r.Height > maxHeight {
			response = r
			maxHeight = r.Height
		}
		// TODO: Wrap with peer info
	}

	return response, err
}

// QueryBlockByHash queries the ledger for Block by block hash.
// This query will be made to specified targets.
// Returns the block.
func (c *Client) QueryBlockByHash(blockHash []byte, options ...RequestOption) (*common.Block, error) {

	opts, err := c.prepareRequestOpts(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get opts for QueryBlockByHash")
	}

	// Determine targets
	targets, err := c.calculateTargets(opts)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to determine target peers for QueryBlockByHash")
	}

	responses, err := c.ledger.QueryBlockByHash(blockHash, peersToTxnProcessors(targets))
	if err != nil && len(responses) == 0 {
		return nil, errors.WithMessage(err, "Failed to QueryBlockByHash")
	}

	if len(responses) < opts.MinMatches {
		return nil, errors.Errorf("Number of responses %d is less than MinMatches %d", len(responses), opts.MinMatches)
	}

	response := responses[0]
	for i, r := range responses {
		if i == 0 {
			continue
		}

		// All payloads have to match
		if !proto.Equal(response.Data, r.Data) {
			return nil, errors.New("Payloads for QueryBlockByHash do not match")
		}
	}

	return response, err
}

// QueryBlock queries the ledger for Block by block number.
// This query will be made to specified targets.
// blockNumber: The number which is the ID of the Block.
// It returns the block.
func (c *Client) QueryBlock(blockNumber int, options ...RequestOption) (*common.Block, error) {
	// TODO: Retrieve and match responses
	return nil, nil
}

// QueryTransaction queries the ledger for Transaction by number.
// This query will be made to specified targets.
// Returns the ProcessedTransaction information containing the transaction.
func (c *Client) QueryTransaction(transactionID fab.TransactionID, options ...RequestOption) (*pb.ProcessedTransaction, error) {
	// TODO: Retrieve and match responses
	return nil, nil
}

//prepareRequestOpts Reads Opts from Option array
func (c *Client) prepareRequestOpts(options ...RequestOption) (Opts, error) {
	opts := Opts{}
	for _, option := range options {
		err := option(&opts)
		if err != nil {
			return opts, errors.WithMessage(err, "Failed to read request opts")
		}
	}

	// Set defaults for max targets
	if opts.MaxTargets == 0 {
		opts.MaxTargets = maxTargets
	}

	// Set defaults for min matches
	if opts.MinMatches == 0 {
		opts.MinMatches = minMatches
	}

	return opts, nil
}

// calculateTargets calculates targets based on targets and filter
func (c *Client) calculateTargets(opts Opts) ([]fab.Peer, error) {

	if opts.Targets != nil && opts.TargetFilter != nil {
		return nil, errors.New("If targets are provided, filter cannot be provided")
	}

	targets := opts.Targets
	targetFilter := opts.TargetFilter

	var err error
	if targets == nil {
		// Retrieve targets from discovery
		targets, err = c.discovery.GetPeers()
		if err != nil {
			return nil, err
		}

		if targetFilter == nil {
			targetFilter = c.filter
		}
	}

	if targetFilter != nil {
		targets = filterTargets(targets, targetFilter)
	}

	if len(targets) == 0 {
		return nil, errors.New("No targets available")
	}

	// TODO: Shuffle targets to randomize and pick number of targets between MaxTargets, len(targets) and MinMatches

	return targets, nil
}

// filterTargets is helper method to filter peers
func filterTargets(peers []fab.Peer, filter TargetFilter) []fab.Peer {

	if filter == nil {
		return peers
	}

	filteredPeers := []fab.Peer{}
	for _, peer := range peers {
		if filter.Accept(peer) {
			filteredPeers = append(filteredPeers, peer)
		}
	}

	return filteredPeers
}

// peersToTxnProcessors converts a slice of Peers to a slice of ProposalProcessors
func peersToTxnProcessors(peers []fab.Peer) []fab.ProposalProcessor {
	tpp := make([]fab.ProposalProcessor, len(peers))

	for i := range peers {
		tpp[i] = peers[i]
	}
	return tpp
}

func shuffle(a []fab.Peer) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}
