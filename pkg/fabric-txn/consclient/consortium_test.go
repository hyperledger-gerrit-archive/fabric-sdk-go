/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package consclient

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn/consclient"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

const channelConfig = "./testdata/test.tx"
const networkCfg = "../../../test/fixtures/config/config_test.yaml"

func TestSaveChannel(t *testing.T) {

	cc := setupConsortiumClient(t)

	// Test empty channel request
	err := cc.SaveChannel(consclient.SaveChannelRequest{})
	if err == nil {
		t.Fatalf("Should have failed for empty channel request")
	}

	// Test empty channel name
	err = cc.SaveChannel(consclient.SaveChannelRequest{ChannelID: "", ChannelConfig: channelConfig})
	if err == nil {
		t.Fatalf("Should have failed for empty channel id")
	}

	// Test empty channel config
	err = cc.SaveChannel(consclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: ""})
	if err == nil {
		t.Fatalf("Should have failed for empty channel config")
	}

	// Test extract configuration error
	err = cc.SaveChannel(consclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: "./testdata/extractcherr.tx"})
	if err == nil {
		t.Fatalf("Should have failed to extract configuration")
	}

	// Test sign channel error
	err = cc.SaveChannel(consclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: "./testdata/signcherr.tx"})
	if err == nil {
		t.Fatalf("Should have failed to sign configuration")
	}

	// Test valid Save Channel request (success)
	err = cc.SaveChannel(consclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: channelConfig})
	if err != nil {
		t.Fatal(err)
	}

}

func TestSaveChannelFailure(t *testing.T) {

	// Set up client with error in create channel
	errClient := fcmocks.NewMockInvalidClient()
	user := fcmocks.NewMockUser("test")
	errClient.SetUserContext(user)
	network := getNetworkConfig(t)

	cc, err := NewConsortiumClient(errClient, network)
	if err != nil {
		t.Fatalf("Failed to create new consortium client: %s", err)
	}

	// Test create channel failure
	err = cc.SaveChannel(consclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: channelConfig})
	if err == nil {
		t.Fatal("Should have failed with create channel error")
	}

}

func TestNoSigningUserFailure(t *testing.T) {

	// Setup client without user context
	client := fcmocks.NewMockClient()
	network := getNetworkConfig(t)

	cc, err := NewConsortiumClient(client, network)
	if err != nil {
		t.Fatalf("Failed to create new consortium client: %s", err)
	}

	// Test save channel without signing user set (and no default context user)
	err = cc.SaveChannel(consclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: channelConfig})
	if err == nil {
		t.Fatal("Should have failed due to missing signing user")
	}

}

func TestSaveChannelWithOpts(t *testing.T) {

	cc := setupConsortiumClient(t)

	// Valid request (same for all options)
	req := consclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: channelConfig}

	// Test empty option (default order is random orderer from config)
	opts := consclient.SaveChannelOpts{}
	err := cc.SaveChannelWithOpts(req, opts)
	if err != nil {
		t.Fatal(err)
	}

	// Test valid orderer ID
	opts.OrdererID = "orderer.example.com"
	err = cc.SaveChannelWithOpts(req, opts)
	if err != nil {
		t.Fatal(err)
	}

	// Test invalid orderer ID
	opts.OrdererID = "Invalid"
	err = cc.SaveChannelWithOpts(req, opts)
	if err == nil {
		t.Fatal("Should have failed for invalid orderer ID")
	}
}

func setupTestClient() *fcmocks.MockClient {
	client := fcmocks.NewMockClient()
	user := fcmocks.NewMockUser("test")
	cryptoSuite := &fcmocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetUserContext(user)
	client.SetCryptoSuite(cryptoSuite)
	return client
}

func getNetworkConfig(t *testing.T) *config.Config {
	config, err := config.InitConfig(networkCfg)
	if err != nil {
		t.Fatal(err)
	}

	return config
}

func setupConsortiumClient(t *testing.T) *ConsortiumClient {

	fcClient := setupTestClient()
	network := getNetworkConfig(t)

	consClient, err := NewConsortiumClient(fcClient, network)
	if err != nil {
		t.Fatalf("Failed to create new consortium client: %s", err)
	}

	return consClient
}
