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
	eventHub  fab.EventHub
}

// NewChannelClient returns a ChannelClient instance.
func NewChannelClient(client fab.FabricClient, channel fab.Channel, discovery fab.DiscoveryService, eventHub fab.EventHub) (*ChannelClient, error) {

	channelClient := ChannelClient{client: client, channel: channel, discovery: discovery, eventHub: eventHub}

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

	txProcessors := peer.PeersToTxnProcessors(peers)

	go sendTransactionProposal(request, cc.channel, txProcessors, notifier)

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

func sendTransactionProposal(request apitxn.QueryRequest, channel fab.Channel, proposalProcessors []apitxn.ProposalProcessor, notifier chan apitxn.QueryResponse) {

	transactionProposalResponses, _, err := internal.CreateAndSendTransactionProposal(channel,
		request.ChaincodeID, request.Fcn, request.Args, proposalProcessors, nil)

	if err != nil {
		notifier <- apitxn.QueryResponse{Response: "", Error: err}
	}

	response := string(transactionProposalResponses[0].ProposalResponse.GetResponse().Payload)

	notifier <- apitxn.QueryResponse{Response: response, Error: nil}
}

// ExecuteTxWithOpts ...
func (cc *ChannelClient) ExecuteTxWithOpts(request apitxn.ExecuteTxRequest, opts apitxn.ExecuteTxOpts) (string, error) {
	// TODO: Implement
	return "", nil
}

// ExecuteTx ...
func (cc *ChannelClient) ExecuteTx(request apitxn.ExecuteTxRequest) (string, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return "", fmt.Errorf("Chaincode name and function name must be provided")
	}

	peers, err := cc.discovery.GetPeers(request.ChaincodeID)
	if err != nil {
		return "", fmt.Errorf("Unable to get peers: %v", err)
	}

	if cc.eventHub.IsConnected() == false {
		err = cc.eventHub.Connect()
		if err != nil {
			return "", fmt.Errorf("Error connecting to eventhub: %v", err)
		}
		defer cc.eventHub.Disconnect()
	}

	transactionProposalResponses, txID, err := internal.CreateAndSendTransactionProposal(cc.channel,
		request.ChaincodeID, request.Fcn, request.Args, peer.PeersToTxnProcessors(peers), request.TransientMap)

	if err != nil {
		return "", fmt.Errorf("CreateAndSendTransactionProposal returned error: %v", err)
	}

	done, fail := internal.RegisterTxEvent(txID, cc.eventHub)

	_, err = internal.CreateAndSendTransaction(cc.channel, transactionProposalResponses)
	if err != nil {
		return txID.ID, fmt.Errorf("CreateAndSendTransaction returned error: %v", err)
	}

	select {
	case <-done:
	case err := <-fail:
		return txID.ID, fmt.Errorf("invoke Error received from eventhub for txid(%s), error(%v)", txID, err)
	case <-time.After(time.Second * 30):
		return txID.ID, fmt.Errorf("invoke Didn't receive block event for txid(%s)", txID)
	}

	return txID.ID, nil
}
