/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msppvdr

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
)

func TestCreateMSPProvider(t *testing.T) {

	coreFactory := defcore.NewProviderFactory()

	configBackend, err := config.FromFile("../../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}

	cryptoSuiteConfig, endpointConfig, _, err := config.FromBackend(configBackend)()
	if err != nil {
		t.Fatalf(err.Error())
	}

	cryptosuite, err := coreFactory.CreateCryptoSuiteProvider(cryptoSuiteConfig)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	userStore := &mockmsp.MockUserStore{}

	provider, err := New(endpointConfig, cryptosuite, userStore)
	assert.Nil(t, err, "New should not have failed")

	if provider.UserStore() != userStore {
		t.Fatalf("Invalid user store returned")
	}

	mgr, ok := provider.IdentityManager("Org1")
	if !ok {
		t.Fatalf("Expected to return idnetity manager")
	}

	_, ok = mgr.(*msp.IdentityManager)
	if !ok {
		t.Fatalf("Unexpected signing manager created")
	}
}
