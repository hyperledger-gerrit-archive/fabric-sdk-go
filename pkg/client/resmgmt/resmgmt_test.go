/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

const (
	channelConfig = "../../../test/fixtures/fabric/v1.0/channel/mychannel.tx"
	networkCfg    = "../../../test/fixtures/config/config_test.yaml"
	configPath    = "../../core/config/testdata/config_test.yaml"
)

func withLocalContextProvider(provider context.LocalProvider) ClientOption {
	return func(rmc *Client) error {
		rmc.localCtxProvider = provider
		return nil
	}
}

func TestJoinChannelFail(t *testing.T) {

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	endorserServer, addr := startEndorserServer(t, grpcServer)
	ctx := setupTestContext("test", "Org1MSP")

	// Create mock orderer with simple mock block
	orderer := fcmocks.NewMockOrderer("", nil)
	defer orderer.Close()
	orderer.EnqueueForSendDeliver(fcmocks.NewSimpleMockBlock())
	orderer.EnqueueForSendDeliver(common.Status_SUCCESS)

	setupCustomOrderer(ctx, orderer)

	rc := setupResMgmtClient(t, ctx)

	// Test nil target
	err := rc.JoinChannel("mychannel", WithTargets(nil))
	if err == nil || !strings.Contains(err.Error(), "target is nil") {
		t.Fatalf("Should have failed due to nil target")
	}

	// Setup target peers
	peer1, _ := peer.New(fcmocks.NewMockEndpointConfig(), peer.WithURL("grpc://"+addr))

	// Test fail with send proposal error
	endorserServer.ProposalError = errors.New("Test Error")
	err = rc.JoinChannel("mychannel", WithTargets(peer1))

	if err == nil || !strings.Contains(err.Error(), "Test Error") {
		t.Fatal("Should have failed to get genesis block")
	}

}

func TestJoinChannelSuccess(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	_, addr := startEndorserServer(t, grpcServer)
	ctx := setupTestContext("test", "Org1MSP")

	// Create mock orderer with simple mock block
	orderer := fcmocks.NewMockOrderer("", nil)
	defer orderer.Close()
	orderer.EnqueueForSendDeliver(fcmocks.NewSimpleMockBlock())
	orderer.EnqueueForSendDeliver(common.Status_SUCCESS)

	setupCustomOrderer(ctx, orderer)

	rc := setupResMgmtClient(t, ctx)

	// Setup target peers
	peer1, _ := peer.New(fcmocks.NewMockEndpointConfig(), peer.WithURL("grpc://"+addr))

	// Test valid join channel request (success)
	err := rc.JoinChannel("mychannel", WithTargets(peer1))
	if err != nil {
		t.Fatal(err)
	}

}

func TestWithFilterOption(t *testing.T) {
	ctx := setupTestContext("test", "Org1MSP")
	rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
	if rc == nil {
		t.Fatal("Expected Resource Management Client to be set")
	}
}

func TestJoinChannelWithFilter(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	_, addr := startEndorserServer(t, grpcServer)
	ctx := setupTestContext("test", "Org1MSP")

	// Create mock orderer with simple mock block
	orderer := fcmocks.NewMockOrderer("", nil)
	defer orderer.Close()
	orderer.EnqueueForSendDeliver(fcmocks.NewSimpleMockBlock())
	orderer.EnqueueForSendDeliver(common.Status_SUCCESS)
	setupCustomOrderer(ctx, orderer)

	//the target filter ( client option) will be set
	rc := setupResMgmtClient(t, ctx)

	// Setup target peers
	peer1, _ := peer.New(fcmocks.NewMockEndpointConfig(), peer.WithURL("grpc://"+addr))

	// Test valid join channel request (success)
	err := rc.JoinChannel("mychannel", WithTargets(peer1))
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoSigningUserFailure(t *testing.T) {

	// Setup client without MSP
	clientCtx := fcmocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", ""))

	_, err := New(createClientContext(clientCtx))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mspID not available in user context")

}

func TestContextFailure(t *testing.T) {

	clientCtx := fcmocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", "Org1MSP"))

	_, err := New(createClientContextWithError(clientCtx))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "due to context error")
}

func TestClientOptionFailure(t *testing.T) {

	clientCtx := fcmocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", "Org1MSP"))

	_, err := New(createClientContext(clientCtx), withOptionError())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Option Error")
}

func withOptionError() ClientOption {
	return func(rmc *Client) error {
		return errors.New("Option Error")
	}
}

func TestJoinChannelRequiredParameters(t *testing.T) {
	rc := setupDefaultResMgmtClient(t)

	// Test empty channel name
	err := rc.JoinChannel("")
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	// Setup test client with different msp (default targets cannot be calculated)
	ctx := setupTestContext("test", "otherMSP")

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(t, ctx)

	// Test missing default targets
	err = rc.JoinChannel("mychannel")

	assert.NotNil(t, err, "error should have been returned")
	s, ok := status.FromError(err)
	assert.True(t, ok, "status code should be available")
	assert.Equal(t, status.NoPeersFound.ToInt32(), s.Code, "code should be no peers found")
}

func TestJoinChannelWithOptsRequiredParameters(t *testing.T) {

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	_, addr := startEndorserServer(t, grpcServer)

	ctx := setupTestContext("test", "Org1MSP")
	network := getNetworkConfig(t)
	ctx.SetEndpointConfig(network)

	// Create mock orderer with simple mock block
	orderer := fcmocks.NewMockOrderer("", nil)
	defer orderer.Close()
	orderer.EnqueueForSendDeliver(fcmocks.NewSimpleMockBlock())
	orderer.EnqueueForSendDeliver(common.Status_SUCCESS)
	setupCustomOrderer(ctx, orderer)

	rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())

	// Test empty channel name for request with no opts
	err := rc.JoinChannel("")
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	var peers []fab.Peer
	peer1, _ := peer.New(fcmocks.NewMockEndpointConfig(), peer.WithURL("grpc://"+addr), peer.WithMSPID("Org1MSP"))
	peers = append(peers, peer1)

	// Test both targets and filter provided (error condition)
	err = rc.JoinChannel("mychannel", WithTargets(peers...), WithTargetFilter(&mspFilter{mspID: "MSPID"}))
	if err == nil || !strings.Contains(err.Error(), "If targets are provided, filter cannot be provided") {
		t.Fatalf("Should have failed if both target and filter provided")
	}

	// Test targets only
	err = rc.JoinChannel("mychannel", WithTargets(peers...), WithOrdererEndpoint("orderer.example.com"))
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Test filter only (filter has no match)
	err = rc.JoinChannel("mychannel", WithTargetFilter(&mspFilter{mspID: "MSPID"}))
	assert.NotNil(t, err, "error should have been returned")
	s, ok := status.FromError(err)
	assert.True(t, ok, "status code should be available")
	assert.Equal(t, status.NoPeersFound.ToInt32(), s.Code, "code should be no peers found")

	//Some cleanup before further test
	orderer = fcmocks.NewMockOrderer("", nil)
	defer orderer.Close()

	ctx = setupTestContext("test", "Org1MSP")
	setupCustomOrderer(ctx, orderer)
	orderer.EnqueueForSendDeliver(fcmocks.NewSimpleMockBlock())
	orderer.EnqueueForSendDeliver(common.Status_SUCCESS)

	rc = setupResMgmtClientWithLocalPeers(t, ctx, peers, getDefaultTargetFilterOption())

	err = rc.JoinChannel("mychannel")
	if err != nil {
		t.Fatalf(err.Error())
	}

}

func TestJoinChannelDiscoveryError(t *testing.T) {

	// Setup test client and config
	ctx := setupTestContext("test", "Org1MSP")
	config := getNetworkConfig(t)
	ctx.SetEndpointConfig(config)

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClientWithDiscoveryError(t, ctx, errors.New("Test Error"))

	err := rc.JoinChannel("mychannel")
	if err == nil || !strings.Contains(err.Error(), "Test Error") {
		t.Fatalf("Should have failed to join channel with discovery error: %s", err)
	}
}

func TestOrdererConfigFail(t *testing.T) {

	ctx := setupTestContext("test", "Org1MSP")

	backend, err := configImpl.FromFile(configPath)()
	assert.Nil(t, err)

	// remove channel orderer and global orderers from config backend
	configBackend := getNoOrdererBackend(backend...)

	noOrdererConfig, err := fabImpl.ConfigFromBackend(configBackend)
	assert.Nil(t, err)

	ctx.SetEndpointConfig(noOrdererConfig)
	rc := setupResMgmtClient(t, ctx)

	orderer, err := rc.ordererConfig("mychannel")
	assert.Nil(t, orderer)
	assert.NotNil(t, err, "should fail since no orderer has been configured")
}

func TestJoinChannelNoOrdererConfig(t *testing.T) {

	ctx := setupTestContext("test", "Org1MSP")

	// No channel orderer, no global orderer
	backend, err := configImpl.FromFile(configPath)()
	assert.Nil(t, err)

	// remove channel orderer and global orderers from config backend
	configBackend := getNoOrdererBackend(backend...)
	noOrdererConfig, err := fabImpl.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetEndpointConfig(noOrdererConfig)
	rc := setupResMgmtClient(t, ctx)

	peer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "grpc://peer1.com", MockMSP: "Org1MSP"}

	err = rc.JoinChannel("mychannel", WithTargets(peer1))
	assert.NotNil(t, err, "Should have failed to join channel since no orderer has been configured")
	assert.Contains(t, err.Error(), "orderer not found: orderers lookup failed")

	// Misconfigured channel orderer
	configBackend = getInvalidChannelOrdererBackend(backend...)
	invalidChOrdererConfig, err := fabImpl.ConfigFromBackend(configBackend)

	if err != nil {
		t.Fatal(err)
	}
	ctx.SetEndpointConfig(invalidChOrdererConfig)

	rc = setupResMgmtClient(t, ctx)

	err = rc.JoinChannel("mychannel", WithTargets(peer1))
	if err == nil || !strings.Contains(err.Error(), "orderer not found: orderers lookup failed") {
		t.Fatalf("Should have failed to join channel since channel orderer has been misconfigured")
	}

	// Misconfigured global orderer (cert cannot be loaded)
	configBackend = getInvalidOrdererBackend(backend...)
	invalidOrdererConfig, err := fabImpl.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetEndpointConfig(invalidOrdererConfig)
	customFabProvider := fabpvdr.New(ctx.EndpointConfig())
	customFabProvider.Initialize(ctx)
	ctx.SetCustomInfraProvider(customFabProvider)

	rc = setupResMgmtClient(t, ctx)

	err = rc.JoinChannel("mychannel", WithTargets(peer1))
	if err == nil || !strings.Contains(err.Error(), "CONNECTION_FAILED") {
		t.Fatalf("Should have failed to join channel since global orderer certs are not configured properly: %s", err)
	}
}

func TestIsChaincodeInstalled(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	//prepare sample response
	response := new(pb.ChaincodeQueryResponse)
	chaincodes := make([]*pb.ChaincodeInfo, 1)
	chaincodes[0] = &pb.ChaincodeInfo{Name: "test-name", Path: "test-path", Version: "test-version"}
	response.Chaincodes = chaincodes
	responseBytes, err := proto.Marshal(response)
	if err != nil {
		t.Fatal("failed to marshal sample response")
	}

	peer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "grpc://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: http.StatusOK, Payload: responseBytes}

	// Chaincode found request
	req := InstallCCRequest{Name: "test-name", Path: "test-path", Version: "test-version"}

	reqCtx, cancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	// Test chaincode installed (valid peer)
	installed, err := rc.isChaincodeInstalled(reqCtx, req, peer1, retry.Opts{})
	if err != nil {
		t.Fatal(err)
	}
	if !installed {
		t.Fatalf("CC should have been installed: %s", req)
	}

	// Chaincode not found request
	req = InstallCCRequest{Name: "ID", Version: "v0", Path: "path"}

	// Test chaincode installed
	installed, err = rc.isChaincodeInstalled(reqCtx, req, peer1, retry.Opts{})
	if err != nil {
		t.Fatal(err)
	}
	if installed {
		t.Fatalf("CC should NOT have been installed: %s", req)
	}

	// Test error retrieving installed cc info (peer is nil)
	_, err = rc.isChaincodeInstalled(reqCtx, req, nil, retry.Opts{})
	if err == nil || !strings.Contains(err.Error(), "peer required") {
		t.Fatalf("Should have failed with error in get installed chaincodes since peer is required: %s", err)
	}

}

func TestQueryInstalledChaincodes(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test error
	_, err := rc.QueryInstalledChaincodes()
	if err == nil {
		t.Fatalf("QueryInstalledChaincodes: peer cannot be nil")
	}

	peer := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: http.StatusOK}

	// Test success (valid peer)
	_, err = rc.QueryInstalledChaincodes(WithTargets(peer))
	if err != nil {
		t.Fatal(err)
	}

}

func TestQueryInstantiatedChaincodes(t *testing.T) {
	rc := setupDefaultResMgmtClient(t)

	// Test error
	_, err := rc.QueryInstantiatedChaincodes("mychannel")
	if err == nil {
		t.Fatalf("QueryInstalledChaincodes: peer cannot be nil")
	}

	peer := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: http.StatusOK}

	// Test success (valid peer)
	_, err = rc.QueryInstantiatedChaincodes("mychannel", WithTargets(peer))
	if err != nil {
		t.Fatal(err)
	}
}

func TestQueryChannels(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	//prepare sample response
	response := new(pb.ChannelQueryResponse)
	channels := make([]*pb.ChannelInfo, 1)
	channels[0] = &pb.ChannelInfo{ChannelId: "test"}
	response.Channels = channels

	responseBytes, err := proto.Marshal(response)
	if err != nil {
		t.Fatal("failed to marshal sample response")
	}

	// Test error
	_, err = rc.QueryChannels()
	if err == nil {
		t.Fatalf("QueryChannels: peer cannot be nil")
	}

	peer := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: http.StatusOK, Payload: responseBytes}

	// Test success (valid peer)
	found := false
	response, err = rc.QueryChannels(WithTargets(peer))
	if err != nil {
		t.Fatalf("failed to query channel for peer: %s", err)
	}
	for _, responseChannel := range response.Channels {
		if responseChannel.ChannelId == "test" {
			found = true
		}
	}

	if !found {
		t.Fatal("Peer has not joined 'test' channel")
	}

}

func TestInstallCCWithOpts(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	//prepare sample response
	response := new(pb.ChaincodeQueryResponse)
	chaincodes := make([]*pb.ChaincodeInfo, 1)
	chaincodes[0] = &pb.ChaincodeInfo{Name: "name", Path: "path", Version: "version"}
	response.Chaincodes = chaincodes
	responseBytes, err := proto.Marshal(response)
	assert.Nil(t, err, "marshal should not have failed")

	// Setup targets
	peer1 := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com",
		Status: http.StatusOK, MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Payload: responseBytes}

	// Already installed chaincode request
	req := InstallCCRequest{Name: "name", Version: "version", Path: "path", Package: &resource.CCPackage{Type: 1, Code: []byte("code")}}
	responses, err := rc.InstallCC(req, WithTargets(&peer1))
	if err != nil {
		t.Fatal(err)
	}

	if responses == nil || len(responses) != 1 {
		t.Fatal("Should have one 'already installed' response")
	}

	if !strings.Contains(responses[0].Info, "already installed") {
		t.Fatal("Should have 'already installed' info set")
	}

	if responses[0].Target != peer1.MockURL {
		t.Fatalf("Expecting %s target URL, got %s", peer1.MockURL, responses[0].Target)
	}

	// Chaincode not found request (it will be installed)
	req = InstallCCRequest{Name: "ID", Version: "v0", Path: "path", Package: &resource.CCPackage{Type: 1, Code: []byte("code")}}
	responses, err = rc.InstallCC(req, WithTargets(&peer1))
	if err != nil {
		t.Fatal(err)
	}

	if responses[0].Target != peer1.MockURL {
		t.Fatal("Wrong target URL set")
	}

	if strings.Contains(responses[0].Info, "already installed") {
		t.Fatal("Should not have 'already installed' info set since it was not previously installed")
	}

}

func TestInstallCCWithOptsError(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Setup targets
	peer1 := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com",
		Status: http.StatusOK, MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}

	req := InstallCCRequest{Name: "error", Version: "v0", Path: "path", Package: &resource.CCPackage{Type: 1, Code: []byte("code")}}

	// Test both targets and filter provided (error condition)
	_, err := rc.InstallCC(req, WithTargets(&peer1), WithTargetFilter(&mspFilter{mspID: "Org1MSP"}))
	if err == nil || !strings.Contains(err.Error(), "If targets are provided, filter cannot be provided") {
		t.Fatalf("Should have failed if both target and filter provided: %s", err)
	}
}

func TestInstallError(t *testing.T) {
	rc := setupDefaultResMgmtClient(t)

	testErr := fmt.Errorf("Test error message")
	peer1 := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com",
		Status: http.StatusInternalServerError, MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Error: testErr}

	peer2 := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com",
		Status: http.StatusOK, MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}

	req := InstallCCRequest{Name: "ID", Version: "v0", Path: "path", Package: &resource.CCPackage{Type: 1, Code: []byte("code")}}
	_, err := rc.InstallCC(req, WithTargets(&peer1, &peer2))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), testErr.Error())
}

func TestInstallCC(t *testing.T) {
	rc := setupDefaultResMgmtClient(t)

	peer2 := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com",
		Status: http.StatusOK, MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}

	// Chaincode that is not installed already (it will be installed)
	req := InstallCCRequest{Name: "ID", Version: "v0", Path: "path", Package: &resource.CCPackage{Type: 1, Code: []byte("code")}}
	responses, err := rc.InstallCC(req, WithTargets(&peer2))
	if err != nil {
		t.Fatal(err)
	}
	if responses == nil || len(responses) != 1 {
		t.Fatal("Should have one successful response")
	}

	expected := "http://peer1.com"
	if responses[0].Target != expected {
		t.Fatalf("Expecting %s target URL, got %s", expected, responses[0].Target)
	}

	if responses[0].Status != http.StatusOK {
		t.Fatalf("Expecting %d status, got %d", 0, responses[0].Status)
	}
}

func TestInstallCCRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test missing required parameters
	req := InstallCCRequest{}
	_, err := rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed for empty install cc request")
	}

	// Test missing chaincode ID
	req = InstallCCRequest{Name: "", Version: "v0", Path: "path"}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc ID")
	}

	// Test missing chaincode version
	req = InstallCCRequest{Name: "ID", Version: "", Path: "path"}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = InstallCCRequest{Name: "ID", Version: "v0", Path: ""}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("InstallCC should have failed for empty cc path")
	}

	// Test missing chaincode package
	req = InstallCCRequest{Name: "ID", Version: "v0", Path: "path"}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("InstallCC should have failed for nil chaincode package")
	}

}

func TestInstallCCWithDifferentMSP(t *testing.T) {

	// Setup test client with different msp (default targets cannot be calculated)
	ctx := setupTestContext("test", "otherMSP")
	rc := setupResMgmtClient(t, ctx)

	// Valid request
	req := InstallCCRequest{Name: "name", Version: "version", Path: "path", Package: &resource.CCPackage{Type: 1, Code: []byte("code")}}

	// No targets and no filter -- default filter msp doesn't match discovery service peer msp
	_, err := rc.InstallCC(req)
	if err == nil || !strings.Contains(err.Error(), "no targets") {
		t.Fatalf("Should have failed with no targets error: %s", err)
	}

	ctx = setupTestContext("test", "Org1MSP")
	// Setup targets
	var peers []fab.Peer

	peers = append(peers, &fcmocks.MockPeer{MockName: "Peer1", MockURL: "grpc://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: http.StatusOK})

	// Create new resource management client with client level target filter ("otherMSP")
	targetFilter := &mspFilter{mspID: "otherMSP"}
	rc = setupResMgmtClientWithLocalPeers(t, ctx, peers, WithDefaultTargetFilter(targetFilter))

	// No targets and no filter -- default filter msp doesn't match discovery service peer msp
	_, err = rc.InstallCC(req)
	if err == nil || !strings.Contains(err.Error(), "no targets") {
		t.Fatalf("Should have failed with no targets error: %s", err)
	}

	rc = setupResMgmtClientWithLocalPeers(t, ctx, peers)

	// Test filter only provided at request level (filter rejects discovery service peer msp)
	_, err = rc.InstallCC(req, WithTargetFilter(&mspFilter{mspID: "Org2MSP"}))
	if err == nil || !strings.Contains(err.Error(), "no targets") {
		t.Fatalf("Should have failed with no targets since filter rejected all discovery targets: %s", err)
	}

	_, err = rc.InstallCC(req)
	if err != nil {
		t.Fatalf("Failed to install CC: %s", err)
	}

}

func TestInstallCCDiscoveryError(t *testing.T) {

	// Setup test client and config
	ctx := setupTestContext("test", "Org1MSP")

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClientWithDiscoveryError(t, ctx, errors.New("Test Error"))

	// Test InstallCC discovery service error
	req := InstallCCRequest{Name: "ID", Version: "v0", Path: "path", Package: &resource.CCPackage{Type: 1, Code: []byte("code")}}

	// Test InstallCC discovery service error
	// if targets are not provided discovery service is used
	_, err := rc.InstallCC(req)
	if err == nil || !strings.Contains(err.Error(), "Test Error") {
		t.Fatalf("Should have failed to install cc with opts with discovery error")
	}
}

func TestInstantiateCCRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test missing required parameters
	req := InstantiateCCRequest{}

	// Test empty channel name
	_, err := rc.InstantiateCC("", req)
	if err == nil {
		t.Fatalf("Should have failed for empty request")
	}

	// Test empty request
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty request")
	}

	// Test missing chaincode ID
	req = InstantiateCCRequest{Name: "", Version: "v0", Path: "path"}
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc name")
	}

	// Test missing chaincode version
	req = InstantiateCCRequest{Name: "ID", Version: "", Path: "path"}
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = InstantiateCCRequest{Name: "ID", Version: "v0", Path: ""}
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc path")
	}

	// Test missing chaincode policy
	req = InstantiateCCRequest{Name: "ID", Version: "v0", Path: "path"}
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for nil chaincode policy")
	}

}

func TestInstantiateCCWithDifferentMSP(t *testing.T) {

	// Setup test client with different msp (default targets cannot be calculated)
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetEndpointConfig(config)

	// Create new resource management client ("otherMSP")
	rc := setupResMgmtClient(t, ctx)

	// Valid request
	ccPolicy := cauthdsl.SignedByMspMember("otherMSP")
	req := InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test filter only provided (filter rejects discovery service peer msp)
	_, err := rc.InstantiateCC("mychannel", req, WithTargetFilter(&mspFilter{mspID: "Org2MSP"}))
	if err == nil || !strings.Contains(err.Error(), "no targets") {
		t.Fatalf("Should have failed with no targets since filter rejected all discovery targets")
	}

	// Channel discovery service will return peer that belongs Org1MSP (valid for instantiate)
	_, err = rc.InstantiateCC("mychannel", req)
	if err != nil {
		t.Fatalf("InstallCC error: %s", err)
	}
}

func TestInstantiateCCWithOpts(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Valid request
	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Setup targets
	var peers []fab.Peer
	peers = append(peers, &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"})

	// Test both targets and filter provided (error condition)
	_, err := rc.InstantiateCC("mychannel", req, WithTargets(peers...), WithTargetFilter(&mspFilter{mspID: "Org1MSP"}))

	if err == nil || !strings.Contains(err.Error(), "If targets are provided, filter cannot be provided") {
		t.Fatalf("Should have failed if both target and filter provided: %s", err)
	}
}

func TestInstantiateCCDiscoveryError(t *testing.T) {

	// Setup test client and config
	ctx := setupTestContext("test", "Org1MSP")
	rc := setupResMgmtClient(t, ctx)

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test InstantiateCCWithOpts create new discovery service per channel error
	_, err := rc.InstantiateCC("error", req)
	if err == nil || !strings.Contains(err.Error(), "failed to get discovery service") {
		t.Fatalf("Should have failed to instantiate cc with opts with get discovery service error: %s", err)
	}
}

func TestInstantiateCCTransactorError(t *testing.T) {

	// Setup test client
	ctx := setupTestContext("test", "Org1MSP")
	rc := setupResMgmtClient(t, ctx)

	transactor := txnmocks.MockTransactor{
		Ctx:       ctx,
		ChannelID: "mychannel",
	}
	rc.ctx.InfraProvider().(*fcmocks.MockInfraProvider).SetCustomTransactor(&transactor)

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test InstantiateCC for transactor error
	_, err := rc.InstantiateCC("mychannel", req)
	if err == nil || !strings.Contains(err.Error(), "Failed to verify signature") {
		t.Fatalf("Should have failed to instantiate cc with opts with get discovery service error: %s", err)
	}
}

func TestUpgradeCCRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test missing required parameters
	req := UpgradeCCRequest{}

	// Test empty channel name
	_, err := rc.UpgradeCC("", req)
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	// Test empty request
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty upgrade cc request")
	}

	// Test missing chaincode ID
	req = UpgradeCCRequest{Name: "", Version: "v0", Path: "path"}
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc name")
	}

	// Test missing chaincode version
	req = UpgradeCCRequest{Name: "ID", Version: "", Path: "path"}
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = UpgradeCCRequest{Name: "ID", Version: "v0", Path: ""}
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc path")
	}

	// Test missing chaincode policy
	req = UpgradeCCRequest{Name: "ID", Version: "v0", Path: "path"}
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for nil chaincode policy")
	}
}

func TestUpgradeCCWithDifferentMSP(t *testing.T) {

	// Setup test client with different msp (default targets cannot be calculated)
	ctx := setupTestContext("test", "otherMSP")

	// Create new resource management client ("otherMSP")
	rc := setupResMgmtClient(t, ctx)

	// Valid request
	ccPolicy := cauthdsl.SignedByMspMember("otherMSP")
	req := UpgradeCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test filter only provided (filter rejects discovery service peer msp)
	_, err := rc.UpgradeCC("mychannel", req, WithTargetFilter(&mspFilter{mspID: "Org2MSP"}))
	if err == nil || !strings.Contains(err.Error(), "no targets") {
		t.Fatalf("Should have failed with no targets since filter rejected all discovery targets")
	}

	// Channel discovery service will return peer that belongs Org1MSP (valid for upgrade)
	_, err = rc.UpgradeCC("mychannel", req)
	if err != nil {
		t.Fatalf("UpgradeCC error: %s", err)
	}
}

func TestUpgradeCCWithOpts(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Valid request
	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := UpgradeCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Setup targets
	var peers []fab.Peer
	peers = append(peers, &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"})

	// Test both targets and filter provided (error condition)
	_, err := rc.UpgradeCC("mychannel", req, WithTargets(peers...), WithTargetFilter(&mspFilter{mspID: "Org1MSP"}))
	if err == nil || !strings.Contains(err.Error(), "If targets are provided, filter cannot be provided") {
		t.Fatalf("Should have failed if both target and filter provided")
	}
}

func TestUpgradeCCDiscoveryError(t *testing.T) {

	// Setup test client and config
	ctx := setupTestContext("test", "Org1MSP")

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClient(t, ctx)

	// Test UpgradeCC discovery service error
	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := UpgradeCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test error while creating discovery service for channel "error"
	_, err := rc.UpgradeCC("error", req)
	if err == nil {
		t.Fatalf("Should have failed to upgrade cc with discovery error")
	}
}

func TestCCProposal(t *testing.T) {

	ctx := setupTestContext("Admin", "Org1MSP")

	// Setup resource management client
	configBackend, err := configImpl.FromFile(configPath)()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := fabImpl.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetEndpointConfig(cfg)

	// Setup target peers
	var peers []fab.Peer
	peer1, _ := peer.New(fcmocks.NewMockEndpointConfig(), peer.WithURL("127.0.0.1:0"))
	peers = append(peers, peer1)

	rc := setupResMgmtClient(t, ctx)

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	instantiateReq := InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test invalid function (only 'instatiate' and 'upgrade' are supported)
	reqCtx, cancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	opts := requestOptions{Targets: peers}
	_, err = rc.sendCCProposal(reqCtx, 3, "mychannel", instantiateReq, opts)
	if err == nil || !strings.Contains(err.Error(), "chaincode deployment type unknown") {
		t.Fatalf("Should have failed for invalid chaincode deployment type: %s", err)
	}

	// Create mock orderer
	orderer := fcmocks.NewMockOrderer("", nil)

	transactor := txnmocks.MockTransactor{
		Ctx:       ctx,
		ChannelID: "mychannel",
		Orderers:  []fab.Orderer{orderer},
	}
	rc.ctx.InfraProvider().(*fcmocks.MockInfraProvider).SetCustomTransactor(&transactor)

	// Test error in transactor
	_, err = rc.InstantiateCC("mychannel", instantiateReq)
	if err == nil || !strings.Contains(err.Error(), "Failed to verify signature") {
		t.Fatalf("Should have failed due to error in commit: %s", err)
	}

	// Test no event source in config
	backends, err := configImpl.FromFile(configPath)()
	if err != nil {
		t.Fatal(err)
	}
	backend := getNoEventSourceBackend(backends...)
	cfg, err = fabImpl.ConfigFromBackend(backend)
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetEndpointConfig(cfg)
	rc = setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
	_, err = rc.InstantiateCC("mychannel", instantiateReq)
	// TODO: Add verification
	if err == nil {
		t.Fatalf("Should have failed since no event source has been configured")
	}
}

func getDefaultTargetFilterOption() ClientOption {
	targetFilter := &mspFilter{mspID: "Org1MSP"}
	return WithDefaultTargetFilter(targetFilter)
}

func setupDefaultResMgmtClient(t *testing.T) *Client {
	ctx := setupTestContext("test", "Org1MSP")
	network := getNetworkConfig(t)
	ctx.SetEndpointConfig(network)
	return setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
}

func setupResMgmtClient(t *testing.T, fabCtx *fcmocks.MockContext, opts ...ClientOption) *Client {
	return setupResMgmtClientWithLocalPeers(t, fabCtx, []fab.Peer{}, opts...)
}

func setupResMgmtClientWithDiscoveryError(t *testing.T, fabCtx *fcmocks.MockContext, discErr error, opts ...ClientOption) *Client {
	return setupResMgmtClientWithLocalPeersAndError(t, fabCtx, []fab.Peer{}, discErr, opts...)
}

func setupResMgmtClientWithLocalPeers(t *testing.T, fabCtx *fcmocks.MockContext, peers []fab.Peer, opts ...ClientOption) *Client {
	return setupResMgmtClientWithLocalPeersAndError(t, fabCtx, peers, nil, opts...)
}

func setupResMgmtClientWithLocalPeersAndError(t *testing.T, fabCtx *fcmocks.MockContext, peers []fab.Peer, discErr error, opts ...ClientOption) *Client {
	localProvider := func() (context.Local, error) {
		localDiscoveryProvider := fcmocks.NewMockDiscoveryProvider(discErr, peers)
		return fcmocks.NewMockLocalContext(fabCtx, localDiscoveryProvider), nil
	}

	opts = append(opts, withLocalContextProvider(localProvider))
	resClient, err := New(createClientContext(fabCtx), opts...)
	if err != nil {
		t.Fatalf("Failed to create new client with options: %s %v", err, opts)
	}

	return resClient
}

func setupTestContext(username string, mspID string) *fcmocks.MockContext {
	user := mspmocks.NewMockSigningIdentity(username, mspID)
	return fcmocks.NewMockContext(user)
}

func setupCustomOrderer(ctx *fcmocks.MockContext, mockOrderer fab.Orderer) *fcmocks.MockContext {
	mockInfraProvider := &fcmocks.MockInfraProvider{}
	mockInfraProvider.SetCustomOrderer(mockOrderer)
	ctx.SetCustomInfraProvider(mockInfraProvider)
	return ctx
}

func startEndorserServer(t *testing.T, grpcServer *grpc.Server) (*fcmocks.MockEndorserServer, string) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	addr := lis.Addr().String()

	endorserServer := &fcmocks.MockEndorserServer{}
	pb.RegisterEndorserServer(grpcServer, endorserServer)
	if err != nil {
		t.Logf("Error starting test server %s", err)
		t.FailNow()
	}
	t.Logf("Starting test server on %s\n", addr)
	go grpcServer.Serve(lis)
	return endorserServer, addr
}

func getNetworkConfig(t *testing.T) fab.EndpointConfig {
	configBackend, err := configImpl.FromFile(networkCfg)()
	if err != nil {
		t.Fatal(err)
	}

	config, err := fabImpl.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatal(err)
	}

	return config
}

func TestSaveChannelSuccess(t *testing.T) {

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	_, addr := fcmocks.StartMockBroadcastServer("127.0.0.1:0", grpcServer)

	ctx := setupTestContext("test", "Org1MSP")

	mockConfig := &fcmocks.MockConfig{}
	grpcOpts := make(map[string]interface{})
	grpcOpts["allow-insecure"] = true

	oConfig := &fab.OrdererConfig{
		URL:         addr,
		GRPCOptions: grpcOpts,
	}
	mockConfig.SetCustomOrdererCfg(oConfig)
	ctx.SetEndpointConfig(mockConfig)

	cc := setupResMgmtClient(t, ctx)

	// Test empty channel request
	_, err := cc.SaveChannel(SaveChannelRequest{})
	assert.NotNil(t, err, "Should have failed for empty channel request")
	assert.Contains(t, err.Error(), "must provide channel ID and channel config")

	r, err := os.Open(channelConfig)
	assert.Nil(t, err, "opening channel config file failed")
	defer r.Close()

	// Test empty channel name
	_, err = cc.SaveChannel(SaveChannelRequest{ChannelID: "", ChannelConfig: r})
	assert.NotNil(t, err, "Should have failed for empty channel id")
	assert.Contains(t, err.Error(), "must provide channel ID and channel config")

	// Test empty channel config
	_, err = cc.SaveChannel(SaveChannelRequest{ChannelID: "mychannel"})
	assert.NotNil(t, err, "Should have failed for empty channel config")
	assert.Contains(t, err.Error(), "must provide channel ID and channel config")

	// Test extract configuration error
	r1, err := os.Open("./testdata/extractcherr.tx")
	assert.Nil(t, err, "opening channel config file failed")
	defer r1.Close()

	_, err = cc.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r1})
	assert.NotNil(t, err, "Should have failed to extract configuration")
	assert.Contains(t, err.Error(), "unmarshal config envelope failed")

	// Test sign channel error
	r2, err := os.Open("./testdata/signcherr.tx")
	assert.Nil(t, err, "opening channel config file failed")
	defer r2.Close()

	_, err = cc.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r2})
	assert.NotNil(t, err, "Should have failed to sign configuration")
	// TODO: Msg bellow should be 'signing configuration failed' ?
	assert.Contains(t, err.Error(), "unmarshal config envelope failed")

	// Test sign channel error
	r3, err := os.Open("./testdata/non-existent.tx")
	assert.NotNil(t, err, "opening channel config should have file failed")
	defer r3.Close()

	_, err = cc.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r3})
	assert.NotNil(t, err, "Should have failed to sign configuration")
	assert.Contains(t, err.Error(), "reading channel config file failed")

	// Test valid Save Channel request (success)
	resp, err := cc.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r}, WithOrdererEndpoint("example.com"))
	assert.Nil(t, err, "error should be nil")
	assert.NotEmpty(t, resp.TransactionID, "transaction ID should be populated")

	// Test valid Save Channel request (success / filename)
	resp, err = cc.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfigPath: channelConfig}, WithOrdererEndpoint("example.com"))
	assert.Nil(t, err, "error should be nil")
	assert.NotEmpty(t, resp.TransactionID, "transaction ID should be populated")
}

func TestSaveChannelFailure(t *testing.T) {

	// Set up context with error in create channel
	user := mspmocks.NewMockSigningIdentity("test", "test")
	errCtx := fcmocks.NewMockContext(user)
	network := getNetworkConfig(t)
	errCtx.SetEndpointConfig(network)
	fabCtx := setupTestContext("user", "Org1Msp1")

	cc, err := New(createClientContext(fabCtx))
	if err != nil {
		t.Fatalf("Failed to create new channel management client: %s", err)
	}

	// Test create channel failure
	r, err := os.Open(channelConfig)
	assert.Nil(t, err, "opening channel config file failed")
	defer r.Close()

	_, err = cc.SaveChannel(SaveChannelRequest{ChannelID: "Invalid", ChannelConfig: r})
	assert.NotNil(t, err, "Should have failed with create channel error")
	assert.Contains(t, err.Error(), "failed to find orderer for request")
}

func TestSaveChannelWithOpts(t *testing.T) {

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	_, addr := fcmocks.StartMockBroadcastServer("127.0.0.1:0", grpcServer)

	ctx := setupTestContext("test", "Org1MSP")

	mockConfig := &fcmocks.MockConfig{}
	grpcOpts := make(map[string]interface{})
	grpcOpts["allow-insecure"] = true

	oConfig := &fab.OrdererConfig{
		URL:         addr,
		GRPCOptions: grpcOpts,
	}
	mockConfig.SetCustomOrdererCfg(oConfig)
	mockConfig.SetCustomRandomOrdererCfg(oConfig)
	ctx.SetEndpointConfig(mockConfig)

	cc := setupResMgmtClient(t, ctx)

	// Valid request (same for all options)
	r1, err := os.Open(channelConfig)
	assert.Nil(t, err, "opening channel config file failed")
	defer r1.Close()

	req := SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r1}

	// Test empty option (default order is random orderer from config)
	opts := WithOrdererEndpoint("")
	resp, err := cc.SaveChannel(req, opts)
	assert.Nil(t, err, "error should be nil")
	assert.NotEmpty(t, resp.TransactionID, "transaction ID should be populated")

	// Test valid orderer ID
	r2, err := os.Open(channelConfig)
	assert.Nil(t, err, "opening channel config file failed")
	defer r2.Close()

	req = SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r2}

	opts = WithOrdererEndpoint("orderer.example.com")
	resp, err = cc.SaveChannel(req, opts)
	assert.Nil(t, err, "error should be nil")
	assert.NotEmpty(t, resp.TransactionID, "transaction ID should be populated")

	// Test invalid orderer ID
	r3, err := os.Open(channelConfig)
	assert.Nil(t, err, "opening channel config file failed")
	defer r3.Close()

	req = SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r3}

	mockConfig = &fcmocks.MockConfig{}
	ctx.SetEndpointConfig(mockConfig)

	cc = setupResMgmtClient(t, ctx)

	opts = WithOrdererEndpoint("Invalid")
	_, err = cc.SaveChannel(req, opts)
	assert.NotNil(t, err, "Should have failed for invalid orderer ID")
	assert.Contains(t, err.Error(), "failed to read opts in resmgmt: orderer not found for url")
}

func TestJoinChannelWithInvalidOpts(t *testing.T) {

	cc := setupDefaultResMgmtClient(t)
	opts := WithOrdererEndpoint("Invalid")
	err := cc.JoinChannel("mychannel", opts)
	assert.NotNil(t, err, "Should have failed for invalid orderer ID")
	assert.Contains(t, err.Error(), "failed to read opts in resmgmt: orderer not found for url")
}

func TestSaveChannelWithMultipleSigningIdenities(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	_, addr := fcmocks.StartMockBroadcastServer("127.0.0.1:0", grpcServer)
	ctx := setupTestContext("test", "Org1MSP")

	mockConfig := &fcmocks.MockConfig{}
	grpcOpts := make(map[string]interface{})
	grpcOpts["allow-insecure"] = true

	oConfig := &fab.OrdererConfig{
		URL:         addr,
		GRPCOptions: grpcOpts,
	}
	mockConfig.SetCustomRandomOrdererCfg(oConfig)
	mockConfig.SetCustomOrdererCfg(oConfig)
	ctx.SetEndpointConfig(mockConfig)

	cc := setupResMgmtClient(t, ctx)

	// empty list of signing identities (defaults to context user)
	r1, err := os.Open(channelConfig)
	assert.Nil(t, err, "opening channel config file failed")
	defer r1.Close()

	req := SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r1, SigningIdentities: []msp.SigningIdentity{}}
	resp, err := cc.SaveChannel(req)
	assert.Nil(t, err, "Failed to save channel with default signing identity: %s", err)
	assert.NotEmpty(t, resp.TransactionID, "transaction ID should be populated")

	// multiple signing identities
	r2, err := os.Open(channelConfig)
	assert.Nil(t, err, "opening channel config file failed")
	defer r2.Close()

	secondCtx := fcmocks.NewMockContext(mspmocks.NewMockSigningIdentity("second", "second"))
	req = SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r2, SigningIdentities: []msp.SigningIdentity{cc.ctx, secondCtx}}
	resp, err = cc.SaveChannel(req, WithOrdererEndpoint(""))
	assert.Nil(t, err, "Failed to save channel with multiple signing identities: %s", err)
	assert.NotEmpty(t, resp.TransactionID, "transaction ID should be populated")
}

func createClientContext(fabCtx context.Client) context.ClientProvider {
	return func() (context.Client, error) {
		return fabCtx, nil
	}
}

func createClientContextWithError(fabCtx context.Client) context.ClientProvider {
	return func() (context.Client, error) {
		return nil, errors.New("Test Error")
	}
}

func getCustomBackend(backend ...core.ConfigBackend) *mocks.MockConfigBackend {

	backendMap := make(map[string]interface{})
	backendMap["client"], _ = backend[0].Lookup("client")
	backendMap["certificateAuthorities"], _ = backend[0].Lookup("certificateAuthorities")
	backendMap["entityMatchers"], _ = backend[0].Lookup("entityMatchers")
	backendMap["peers"], _ = backend[0].Lookup("peers")
	backendMap["organizations"], _ = backend[0].Lookup("organizations")
	backendMap["orderers"], _ = backend[0].Lookup("orderers")
	backendMap["channels"], _ = backend[0].Lookup("channels")

	return &mocks.MockConfigBackend{KeyValueMap: backendMap}
}

func getNoOrdererBackend(backend ...core.ConfigBackend) *mocks.MockConfigBackend {

	mockConfigBackend := getCustomBackend(backend...)
	mockConfigBackend.KeyValueMap["channels"] = nil
	mockConfigBackend.KeyValueMap["orderers"] = nil

	return mockConfigBackend
}

func getInvalidChannelOrdererBackend(backend ...core.ConfigBackend) *mocks.MockConfigBackend {

	//Create an invalid channel
	channels := make(map[string]fab.ChannelNetworkConfig)
	mychannel := fab.ChannelNetworkConfig{
		Orderers: []string{"invalid.orderer.com"},
	}
	channels["mychannel"] = mychannel

	mockConfigBackend := getCustomBackend(backend...)
	mockConfigBackend.KeyValueMap["channels"] = channels

	return mockConfigBackend
}

func getInvalidOrdererBackend(backend ...core.ConfigBackend) *mocks.MockConfigBackend {

	//Create invalid orderer
	networkConfig := fab.NetworkConfig{}
	err := lookup.New(backend...).UnmarshalKey("orderers", &networkConfig.Orderers)
	if err != nil {
		panic(err)
	}
	exampleOrderer := networkConfig.Orderers["orderer.example.com"]
	exampleOrderer.TLSCACerts.Path = "/some/invalid/path"
	exampleOrderer.TLSCACerts.Pem = validRootCA
	networkConfig.Orderers["orderer.example.com"] = exampleOrderer

	mockConfigBackend := getCustomBackend(backend...)
	mockConfigBackend.KeyValueMap["orderers"] = networkConfig.Orderers

	return mockConfigBackend
}

func getNoEventSourceBackend(backend ...core.ConfigBackend) *mocks.MockConfigBackend {

	//Create no event source channels
	networkConfig := fab.NetworkConfig{}
	err := lookup.New(backend...).UnmarshalKey("channels", &networkConfig.Channels)
	if err != nil {
		panic(err)
	}
	mychannel := networkConfig.Channels["mychannel"]
	chPeer := mychannel.Peers["peer0.org1.example.com"]
	chPeer.EventSource = false
	mychannel.Peers["peer0.org1.example.com"] = chPeer
	networkConfig.Channels["mychannel"] = mychannel

	mockConfigBackend := getCustomBackend(backend...)
	mockConfigBackend.KeyValueMap["channels"] = networkConfig.Channels

	return mockConfigBackend
}

var validRootCA = `-----BEGIN CERTIFICATE-----
MIICYjCCAgmgAwIBAgIUB3CTDOU47sUC5K4kn/Caqnh114YwCgYIKoZIzj0EAwIw
fzELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xHzAdBgNVBAoTFkludGVybmV0IFdpZGdldHMsIEluYy4xDDAK
BgNVBAsTA1dXVzEUMBIGA1UEAxMLZXhhbXBsZS5jb20wHhcNMTYxMDEyMTkzMTAw
WhcNMjExMDExMTkzMTAwWjB/MQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZv
cm5pYTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEfMB0GA1UEChMWSW50ZXJuZXQg
V2lkZ2V0cywgSW5jLjEMMAoGA1UECxMDV1dXMRQwEgYDVQQDEwtleGFtcGxlLmNv
bTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABKIH5b2JaSmqiQXHyqC+cmknICcF
i5AddVjsQizDV6uZ4v6s+PWiJyzfA/rTtMvYAPq/yeEHpBUB1j053mxnpMujYzBh
MA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBQXZ0I9
qp6CP8TFHZ9bw5nRtZxIEDAfBgNVHSMEGDAWgBQXZ0I9qp6CP8TFHZ9bw5nRtZxI
EDAKBggqhkjOPQQDAgNHADBEAiAHp5Rbp9Em1G/UmKn8WsCbqDfWecVbZPQj3RK4
oG5kQQIgQAe4OOKYhJdh3f7URaKfGTf492/nmRmtK+ySKjpHSrU=
-----END CERTIFICATE-----
`
