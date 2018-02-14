/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package idmgmtclient

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

const channelConfig = "./testdata/test.tx"
const networkCfg = "../../../test/fixtures/config/config_test.yaml"

func TestClient(t *testing.T) {

	_ = setupIdentityMgmtClient(t)

	// TODO
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

func setupIdentityMgmtClient(t *testing.T) *IdentityMgmtClient {

	fabCtx := setupTestContext()
	network := getNetworkConfig(t)
	fabCtx.SetConfig(network)
	resource := fcmocks.NewMockResource()

	ctx := Context{
		ProviderContext: fabCtx,
		Context:         fabCtx,
		Resource:        resource,
	}
	consClient, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create new channel management client: %s", err)
	}

	return consClient
}
