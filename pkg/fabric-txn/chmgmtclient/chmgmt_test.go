/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chmgmtclient

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

const channelConfig = "./testdata/test.tx"
const networkCfg = "../../../test/fixtures/config/config_test.yaml"

func TestSaveChannel(t *testing.T) {

	cc := setupChannelMgmtClient(t)

	// Test empty channel request
	err := cc.SaveChannel(chmgmtclient.SaveChannelRequest{})
	if err == nil {
		t.Fatalf("Should have failed for empty channel request")
	}

	// Test empty channel name
	err = cc.SaveChannel(chmgmtclient.SaveChannelRequest{ChannelID: "", ChannelConfig: channelConfig})
	if err == nil {
		t.Fatalf("Should have failed for empty channel id")
	}

	// Test empty channel config
	err = cc.SaveChannel(chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: ""})
	if err == nil {
		t.Fatalf("Should have failed for empty channel config")
	}

	// Test extract configuration error
	err = cc.SaveChannel(chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: "./testdata/extractcherr.tx"})
	if err == nil {
		t.Fatalf("Should have failed to extract configuration")
	}

	// Test sign channel error
	err = cc.SaveChannel(chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: "./testdata/signcherr.tx"})
	if err == nil {
		t.Fatalf("Should have failed to sign configuration")
	}

	// Test valid Save Channel request (success)
	err = cc.SaveChannel(chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: channelConfig})
	if err != nil {
		t.Fatal(err)
	}

}

func TestSaveChannelFailure(t *testing.T) {

	// Set up context with error in create channel
	user := fcmocks.NewMockUser("test")
	errCtx := fcmocks.NewMockContext(user)
	network := getNetworkConfig(t)
	errCtx.SetConfig(network)
	resource := fcmocks.NewMockInvalidResource()

	ctx := Context{
		ProviderContext: errCtx,
		IdentityContext: user,
		Resource:        resource,
	}
	cc, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create new channel management client: %s", err)
	}

	// Test create channel failure
	err = cc.SaveChannel(chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: channelConfig})
	if err == nil {
		t.Fatal("Should have failed with create channel error")
	}

}

func TestSaveChannelWithOpts(t *testing.T) {

	cc := setupChannelMgmtClient(t)

	// Valid request (same for all options)
	req := chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: channelConfig}

	// Test empty option (default order is random orderer from config)
	opts := chmgmtclient.WithOrdererID("")
	err := cc.SaveChannel(req, opts)
	if err != nil {
		t.Fatal(err)
	}

	// Test valid orderer ID
	opts = chmgmtclient.WithOrdererID("orderer.example.com")
	err = cc.SaveChannel(req, opts)
	if err != nil {
		t.Fatal(err)
	}

	// Test invalid orderer ID
	opts = chmgmtclient.WithOrdererID("Invalid")
	err = cc.SaveChannel(req, opts)
	if err == nil {
		t.Fatal("Should have failed for invalid orderer ID")
	}
}

func TestSaveChannelWithMultipleIdenities(t *testing.T) {
	cc := setupChannelMgmtClient(t)

	// empty list of signing identities (defaults to context user)
	req := chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: channelConfig, SigningIdentity: []apifabclient.IdentityContext{}}

	err := cc.SaveChannel(req, chmgmtclient.WithOrdererID(""))
	if err != nil {
		t.Fatal(err)
	}

	// multiple signing identities
	secondUser := fcmocks.NewMockUser("test2")
	secondCtx := fcmocks.NewMockContext(secondUser)
	req = chmgmtclient.SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: channelConfig, SigningIdentity: []apifabclient.IdentityContext{cc.identity, secondCtx}}

	err = cc.SaveChannel(req, chmgmtclient.WithOrdererID(""))
	if err != nil {
		t.Fatal(err)
	}
}

func setupTestContext() *fcmocks.MockContext {
	user := fcmocks.NewMockUser("test")
	ctx := fcmocks.NewMockContext(user)
	return ctx
}

func getNetworkConfig(t *testing.T) apiconfig.Config {
	config, err := config.FromFile(networkCfg)()
	if err != nil {
		t.Fatal(err)
	}

	return config
}

func setupChannelMgmtClient(t *testing.T) *ChannelMgmtClient {

	fabCtx := setupTestContext()
	network := getNetworkConfig(t)
	fabCtx.SetConfig(network)
	resource := fcmocks.NewMockResource()

	ctx := Context{
		ProviderContext: fabCtx,
		IdentityContext: fabCtx,
		Resource:        resource,
	}
	consClient, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create new channel management client: %s", err)
	}

	return consClient
}
