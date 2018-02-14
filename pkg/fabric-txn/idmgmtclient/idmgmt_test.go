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

const networkCfg = "../../../test/fixtures/config/registrar.yaml"

func TestClient(t *testing.T) {

	_ = setupIdentityManager(t)

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

func setupIdentityManager(t *testing.T) *IdentityManager {

	fabCtx := setupTestContext()
	config := getNetworkConfig(t)
	fabCtx.SetConfig(config)

	// Without provider context
	_, err := New(Context{
		ProviderContext: nil,
		MspID:           "Org1",
	})
	if err == nil {
		t.Fatalf("should fail for missing provider context")
	}

	// Without org
	_, err = New(Context{
		ProviderContext: fcmocks.NewMockProviderContext(),
		MspID:           "",
	})
	if err == nil {
		t.Fatalf("should fail for missing org")
	}

	// With valid context
	ctx := Context{
		ProviderContext: fcmocks.NewMockProviderContext(),
		MspID:           "Org1",
	}

	client, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create new identity management client: %s", err)
	}

	return client
}
