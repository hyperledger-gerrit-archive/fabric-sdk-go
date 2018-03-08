// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
)

func TestNewDefaultSDK(t *testing.T) {
	// Test New SDK with valid config file
	sdk, err := New(configImpl.FromFile(sdkConfigFile))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	verifySDK(t, sdk)
}

func verifySDK(t *testing.T, sdk *FabricSDK) {

	// Mock channel provider cache
	sdk.provider.ChannelProvider().(*chpvdr.ChannelProvider).SetChannelConfig(mocks.NewMockChannelCfg("mychannel"))
	sdk.provider.ChannelProvider().(*chpvdr.ChannelProvider).SetChannelConfig(mocks.NewMockChannelCfg("orgchannel"))

	// Get a common client context for the following tests
	chCtx1 := sdk.ChannelContext("mychannel", WithUser(sdkValidClientUser), WithOrgName(sdkValidClientOrg2))
	chCtx2 := sdk.ChannelContext("orgchannel", WithUser(sdkValidClientUser), WithOrgName(sdkValidClientOrg2))

	// Test configuration failure for channel client (mychannel does't have event source configured for Org2)
	_, err := channel.New(chCtx1)
	if err == nil {
		t.Fatalf("Should have failed to create channel client since event source not configured for Org2")
	}

	// Test new channel client with options
	_, err = channel.New(chCtx2)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}
}

func TestWithConfigOpt(t *testing.T) {
	// Test New SDK with valid config file
	c, err := configImpl.FromFile(sdkConfigFile)()
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(WithConfig(c))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	verifySDK(t, sdk)
}

func TestNewDefaultTwoValidSDK(t *testing.T) {
	sdk1, err := New(configImpl.FromFile(sdkConfigFile))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Mock channel provider cache
	sdk1.provider.ChannelProvider().(*chpvdr.ChannelProvider).SetChannelConfig(mocks.NewMockChannelCfg("mychannel"))
	sdk1.provider.ChannelProvider().(*chpvdr.ChannelProvider).SetChannelConfig(mocks.NewMockChannelCfg("orgchannel"))

	sdk2, err := New(configImpl.FromFile("./testdata/test.yaml"))
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	// Mock channel provider cache
	sdk2.provider.ChannelProvider().(*chpvdr.ChannelProvider).SetChannelConfig(mocks.NewMockChannelCfg("orgchannel"))

	// Default sdk with two channels
	client1, err := sdk1.Config().Client()
	if err != nil {
		t.Fatalf("Error getting client from config: %s", err)
	}

	if client1.Organization != sdkValidClientOrg1 {
		t.Fatalf("Unexpected org in config: %s", client1.Organization)
	}

	client2, err := sdk2.Config().Client()
	if err != nil {
		t.Fatalf("Error getting client from config: %s", err)
	}

	if client2.Organization != sdkValidClientOrg2 {
		t.Fatalf("Unexpected org in config: %s", client1.Organization)
	}

	// Get a common client context for the following tests
	//cc1 := sdk1.NewClient(WithUser(sdkValidClientUser))

	cc1CtxC1 := sdk1.ChannelContext("mychannel", WithUser(sdkValidClientUser), WithOrgName(sdkValidClientOrg1))
	cc1CtxC2 := sdk1.ChannelContext("orgchannel", WithUser(sdkValidClientUser), WithOrgName(sdkValidClientOrg1))

	// Test SDK1 channel clients ('mychannel', 'orgchannel')
	_, err = channel.New(cc1CtxC1)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	_, err = channel.New(cc1CtxC2)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Get a common client context for the following tests
	cc2CtxC1 := sdk1.ChannelContext("mychannel", WithUser(sdkValidClientUser), WithOrgName(sdkValidClientOrg2))
	cc2CtxC2 := sdk1.ChannelContext("orgchannel", WithUser(sdkValidClientUser), WithOrgName(sdkValidClientOrg2))

	// SDK 2 doesn't have 'mychannel' configured
	_, err = channel.New(cc2CtxC1)
	if err == nil {
		t.Fatalf("Should have failed to create channel that is not configured")
	}

	// SDK 2 has 'orgchannel' configured
	_, err = channel.New(cc2CtxC2)
	if err != nil {
		t.Fatalf("Failed to create new 'orgchannel' channel client: %s", err)
	}
}
