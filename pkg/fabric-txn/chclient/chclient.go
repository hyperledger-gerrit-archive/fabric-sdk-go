/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package chclient enables channel client
package chclient

import (
	"fmt"
	"time"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/internal"
)

// ChannelClient enables access to a Fabric network.
type ChannelClient struct {
	client    fab.FabricClient
	channel   fab.Channel
	discovery fab.DiscoveryService
}

// NewChannelClient returns a ChannelClient instance.
func NewChannelClient(client fab.FabricClient, channel fab.Channel, discovery fab.DiscoveryService) (*ChannelClient, error) {

	channelClient := ChannelClient{client: client, channel: channel, discovery: discovery}

	return &channelClient, nil
}

// Query chaincode
func (cc *ChannelClient) Query(request apitxn.QueryRequest) (string, error) {

	return cc.QueryWithOpts(request, apitxn.QueryOpts{})

}

// QueryWithOpts allows the user to provide options for query (sync vs async, etc.)
func (cc *ChannelClient) QueryWithOpts(request apitxn.QueryRequest, opt apitxn.QueryOpts) (string, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return "", fmt.Errorf("Chaincode name and function name must be provided")
	}

	notifier := opt.Notifier
	if notifier == nil {
		notifier = make(chan apitxn.QueryResponse)
	}

	peers, err := cc.discovery.GetPeers(request.ChaincodeID)
	if err != nil {
		return "", fmt.Errorf("Unable to get peers: %v", err)
	}

	go sendTransactionProposal(request, cc.channel, peers, notifier)

	if opt.Notifier != nil {
		return "", nil
	}

	select {
	case response := <-notifier:
		return response.Response, response.Error
	case <-time.After(time.Second * 20):
		return "", fmt.Errorf("Request timed out") // TODO: configurable timeout or wait forever
	}

}

func sendTransactionProposal(request apitxn.QueryRequest, channel fab.Channel, peers []fab.Peer, notifier chan apitxn.QueryResponse) {

	transactionProposalResponses, _, err := internal.CreateAndSendTransactionProposal(channel,
		request.ChaincodeID, request.Fcn, request.Args, peer.PeersToTxnProcessors(peers), nil)

	if err != nil {
		notifier <- apitxn.QueryResponse{Response: "", Error: err}
	}

	response := string(transactionProposalResponses[0].ProposalResponse.GetResponse().Payload)

	notifier <- apitxn.QueryResponse{Response: response, Error: nil}
}
