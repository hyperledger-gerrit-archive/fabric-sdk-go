/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defclient

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/identitymgr"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
)

/*
func TestCreateMSPClient(t *testing.T) {
	factory := NewOrgClientFactory()

	config := mocks.NewMockConfig()

	coreFactory := defcore.CreateProviderFactory()
	cryptosuite, err := coreFactory.CreateCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	mspClient, err := factory.CreateMSPClient("org1", config, cryptosuite)
	if err != nil {
		t.Fatalf("Unexpected error creating MSP client %v", err)
	}

	_, ok := mspClient.(*fabricCAClient.FabricCA)
	if !ok {
		t.Fatalf("Unexpected selection provider created")
	}
}
*/

func TestNew(t *testing.T) {
	factory := NewOrgClientFactory()

	config, err := config.FromFile("../../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}

	coreFactory := defcore.NewProviderFactory()
	cryptosuite, err := coreFactory.CreateCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	mspClient, err := factory.New("org1", config, cryptosuite)
	if err != nil {
		t.Fatalf("Unexpected error creating credential manager %v", err)
	}

	_, ok := mspClient.(*identitymgr.IdentityManager)
	if !ok {
		t.Fatalf("Unexpected credential manager created")
	}
}
