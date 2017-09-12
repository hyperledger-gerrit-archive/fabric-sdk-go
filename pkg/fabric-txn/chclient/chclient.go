/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package chclient enables channel client
package chclient

import (
	"fmt"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
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
func NewChannelClient(sdk context.SDK, session context.Session, channelName string) (apitxn.ChannelClient, error) {

	client := clientImpl.NewClient(sdk.ConfigProvider())
	client.SetCryptoSuite(sdk.CryptoSuiteProvider())
	client.SetStateStore(sdk.StateStoreProvider())
	client.SetUserContext(session.Identity())
	client.SetSigningManager(sdk.SigningManager())

	channel, err := GetChannel(client, channelName)
	if err != nil {
		return nil, fmt.Errorf("Unable to create channel:%v", err)
	}

	discovery, err := sdk.DiscoveryProvider().NewDiscoveryService(channel)
	if err != nil {
		return nil, fmt.Errorf("Unable to create discovery service:%v", err)
	}

	channelClient := ChannelClient{client: client, channel: channel, discovery: discovery}

	return &channelClient, nil
}

// Query chaincode
func (cc *ChannelClient) Query(request apitxn.QueryRequest) (string, error) {

	if request.ChaincodeID == "" || request.Fcn == "" {
		return "", fmt.Errorf("Chaincode name and function name must be provided")
	}

	peers, err := cc.discovery.GetPeers(request.ChaincodeID)
	if err != nil {
		return "", fmt.Errorf("Unable to get peers: %v", err)
	}

	transactionProposalResponses, _, err := internal.CreateAndSendTransactionProposal(cc.channel,
		request.ChaincodeID, request.Fcn, request.Args, peer.PeersToTxnProcessors(peers), nil)

	if err != nil {
		return "", fmt.Errorf("CreateAndSendTransactionProposal returned error: %v", err)
	}

	return string(transactionProposalResponses[0].ProposalResponse.GetResponse().Payload), nil

}

// QueryWithOpts allows the user to provide options for query (sync vs async, etc.)
func (cc *ChannelClient) QueryWithOpts(request apitxn.QueryRequest, opt apitxn.QueryOpts) error {
	// TODO
	return nil
}

// GetChannel is helper method to initializes and returns a channel based on config
func GetChannel(client fab.FabricClient, channelID string) (fab.Channel, error) {

	channel, err := client.NewChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("NewChannel return error: %v", err)
	}

	// TODO: Read orderer config based on channel
	ordererConfig, err := client.Config().RandomOrdererConfig()
	if err != nil {
		return nil, fmt.Errorf("RandomOrdererConfig() return error: %s", err)
	}

	orderer, err := orderer.NewOrderer(fmt.Sprintf("%s:%d", ordererConfig.Host,
		ordererConfig.Port), ordererConfig.TLS.Certificate,
		ordererConfig.TLS.ServerHostOverride, client.Config())
	if err != nil {
		return nil, fmt.Errorf("NewOrderer return error: %v", err)
	}
	err = channel.AddOrderer(orderer)
	if err != nil {
		return nil, fmt.Errorf("Error adding orderer: %v", err)
	}

	return channel, nil
}
