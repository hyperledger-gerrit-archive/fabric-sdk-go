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

	// Without org
	_, err := New(Context{
		MspID:       "",
		Config:      &fcmocks.MockConfig{},
		CryptoSuite: &fcmocks.MockCryptoSuite{},
	})
	if err == nil {
		t.Fatalf("should fail for missing org")
	}

	// Without config
	_, err = New(Context{
		MspID:       "Org1",
		Config:      nil,
		CryptoSuite: &fcmocks.MockCryptoSuite{},
	})
	if err == nil {
		t.Fatalf("should fail for missing config")
	}

	// Without crypto suite
	_, err = New(Context{
		MspID:       "Org1",
		Config:      &fcmocks.MockConfig{},
		CryptoSuite: nil,
	})
	if err == nil {
		t.Fatalf("should fail for missing crypto suite")
	}

	// With valid context
	ctx := Context{
		MspID:       "Org1",
		Config:      &fcmocks.MockConfig{},
		CryptoSuite: &fcmocks.MockCryptoSuite{},
	}

	client, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create new identity management client: %s", err)
	}

	return client
}
