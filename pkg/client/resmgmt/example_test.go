/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package resmgmt

import (
	"fmt"
	"os"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	sdkCtx "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/api"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

func Example() {

	// Create new resource management client
	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	// Read channel configuration
	r, err := os.Open(channelConfig)
	if err != nil {
		fmt.Printf("failed to open channel config: %v", err)
	}
	defer r.Close()

	// Create new channel 'mychannel'
	_, err = c.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r})
	if err != nil {
		fmt.Printf("failed to save channel: %v", err)
	}

	peer := mockPeer()

	// Peer joins channel 'mychannel'
	err = c.JoinChannel("mychannel", WithTargets(peer))
	if err != nil {
		fmt.Printf("failed to join channel: %v", err)
	}

	// Install example chaincode to peer
	installReq := InstallCCRequest{Name: "ExampleCC", Version: "v0", Path: "path", Package: &api.CCPackage{Type: 1, Code: []byte("bytes")}}
	_, err = c.InstallCC(installReq, WithTargets(peer))
	if err != nil {
		fmt.Printf("failed to install chaincode: %v", err)
	}

	// Instantiate example chaincode on channel 'mychannel'
	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	instantiateReq := InstantiateCCRequest{Name: "ExampleCC", Version: "v0", Path: "path", Policy: ccPolicy}
	_, err = c.InstantiateCC("mychannel", instantiateReq, WithTargets(peer))
	if err != nil {
		fmt.Printf("failed to install chaincode: %v", err)
	}

	fmt.Println("Network setup completed")

	// Output: Network setup completed
}

func ExampleNew() {

	ctx := mockClientProvider()

	c, err := New(ctx)
	if err != nil {
		fmt.Println("failed to create client")
	}

	if c != nil {
		fmt.Println("resource management client created")
	}

	// Output: resource management client created
}

func ExampleWithDefaultTargetFilter() {

	ctx := mockClientProvider()

	c, err := New(ctx, WithDefaultTargetFilter(&urlTargetFilter{url: "example.com"}))
	if err != nil {
		fmt.Println("failed to create client")
	}

	if c != nil {
		fmt.Println("resource management client created with url target filter")
	}

	// Output: resource management client created with url target filter
}

// urlTargetFilter filters targets based on url
type urlTargetFilter struct {
	url string
}

// Accept returns true if this peer is to be included in the target list
func (f *urlTargetFilter) Accept(peer fab.Peer) bool {
	return peer.URL() == f.url
}

func ExampleWithParentContext() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	clientContext, err := mockClientProvider()()
	if err != nil {
		fmt.Println("failed to return client context")
		return
	}

	// get parent context and cancel
	parentContext, cancel := sdkCtx.NewRequest(clientContext, sdkCtx.WithTimeout(20*time.Second))
	defer cancel()

	channels, err := c.QueryChannels(WithParentContext(parentContext), WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to query for blockchain info: %v", err)
	}

	if channels != nil {
		fmt.Println("Retrieved channels that peer belongs to")
	}

	// Output: Retrieved channels that peer belongs to
}

func ExampleWithTargets() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	response, err := c.QueryChannels(WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to query channels: %v", err)
	}

	if response != nil {
		fmt.Println("Retrieved channels")
	}

	// Output: Retrieved channels
}

func ExampleWithTargetFilter() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := InstantiateCCRequest{Name: "ExampleCC", Version: "v0", Path: "path", Policy: ccPolicy}

	resp, err := c.InstantiateCC("mychannel", req, WithTargetFilter(&urlTargetFilter{url: "http://peer1.com"}))
	if err != nil {
		fmt.Printf("failed to install chaincode: %v", err)
	}

	if resp.TransactionID == "" {
		fmt.Println("Failed to instantiate chaincode")
	}

	fmt.Println("Chaincode instantiated")

	// Output: Chaincode instantiated

}

func ExampleClient_SaveChannel() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Printf("failed to create client: %v", err)
	}

	r, err := os.Open(channelConfig)
	if err != nil {
		fmt.Printf("failed to open channel config: %v", err)
	}
	defer r.Close()

	resp, err := c.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r})
	if err != nil {
		fmt.Printf("failed to save channel: %v", err)
	}

	if resp.TransactionID == "" {
		fmt.Println("Failed to save channel")
	}

	fmt.Println("Saved channel")

	// Output: Saved channel
}

func ExampleClient_SaveChannel_withNetworkOrderer() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Printf("failed to create client: %v", err)
	}

	r, err := os.Open(channelConfig)
	if err != nil {
		fmt.Printf("failed to open channel config: %v", err)
	}
	defer r.Close()

	resp, err := c.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r}, WithOrdererEndpoint("example.com"))
	if err != nil {
		fmt.Printf("failed to save channel: %v", err)
	}

	if resp.TransactionID == "" {
		fmt.Println("Failed to save channel")
	}

	fmt.Println("Saved channel")

	// Output: Saved channel

}

func ExampleClient_JoinChannel() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	err = c.JoinChannel("mychannel", WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to join channel: %v", err)
	}

	fmt.Println("Joined channel")

	// Output: Joined channel
}

func ExampleClient_InstallCC() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	req := InstallCCRequest{Name: "ExampleCC", Version: "v0", Path: "path", Package: &api.CCPackage{Type: 1, Code: []byte("bytes")}}
	responses, err := c.InstallCC(req, WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to install chaincode: %v", err)
	}

	if len(responses) > 0 {
		fmt.Println("Chaincode installed")
	}

	// Output: Chaincode installed
}

func ExampleClient_InstantiateCC() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := InstantiateCCRequest{Name: "ExampleCC", Version: "v0", Path: "path", Policy: ccPolicy}

	resp, err := c.InstantiateCC("mychannel", req)
	if err != nil {
		fmt.Printf("failed to install chaincode: %v", err)
	}

	if resp.TransactionID == "" {
		fmt.Println("Failed to instantiate chaincode")
	}

	fmt.Println("Chaincode instantiated")

	// Output: Chaincode instantiated
}

func ExampleClient_UpgradeCC() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := UpgradeCCRequest{Name: "ExampleCC", Version: "v1", Path: "path", Policy: ccPolicy}

	resp, err := c.UpgradeCC("mychannel", req, WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to upgrade chaincode: %v", err)
	}

	if resp.TransactionID == "" {
		fmt.Println("Failed to upgrade chaincode")
	}

	fmt.Println("Chaincode upgraded")

	// Output: Chaincode upgraded
}

func ExampleClient_QueryChannels() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	response, err := c.QueryChannels(WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to query channels: %v", err)
	}

	if response != nil {
		fmt.Println("Retrieved channels")
	}

	// Output: Retrieved channels
}

func ExampleClient_QueryInstalledChaincodes() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	response, err := c.QueryInstalledChaincodes(WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to query installed chaincodes: %v", err)
	}

	if response != nil {
		fmt.Println("Retrieved installed chaincodes")
	}

	// Output: Retrieved installed chaincodes
}

func ExampleClient_QueryInstantiatedChaincodes() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	response, err := c.QueryInstantiatedChaincodes("mychannel", WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to query instantiated chaincodes: %v", err)
	}

	if response != nil {
		fmt.Println("Retrieved instantiated chaincodes")
	}

	// Output: Retrieved instantiated chaincodes
}

func mockClientProvider() context.ClientProvider {

	ctx := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", "Org1MSP"))

	// Create mock orderer with simple mock block
	orderer := mocks.NewMockOrderer("", nil)
	defer orderer.Close()

	orderer.EnqueueForSendDeliver(mocks.NewSimpleMockBlock())
	orderer.EnqueueForSendDeliver(common.Status_SUCCESS)

	setupCustomOrderer(ctx, orderer)

	clientProvider := func() (context.Client, error) {
		return ctx, nil
	}

	return clientProvider
}

func mockPeer() fab.Peer {
	return &mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", Status: 200}
}
