/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defclient

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	credentialMgr "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/credentialmgr"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/idmgmtclient"
	mockapisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
)

/*
func TestNewMSPClient(t *testing.T) {
	factory := NewOrgClientFactory()

	config := mocks.NewMockConfig()

	coreFactory := defcore.NewProviderFactory()
	cryptosuite, err := coreFactory.NewCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	mspClient, err := factory.NewMSPClient("org1", config, cryptosuite)
	if err != nil {
		t.Fatalf("Unexpected error creating MSP client %v", err)
	}

	_, ok := mspClient.(*fabricCAClient.FabricCA)
	if !ok {
		t.Fatalf("Unexpected selection provider created")
	}
}
*/

func TestNewCredentialManager(t *testing.T) {
	factory := NewOrgClientFactory()

	config, err := config.FromFile("../../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}

	coreFactory := defcore.NewProviderFactory()
	cryptosuite, err := coreFactory.NewCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	cm, err := factory.NewCredentialManager("org1", config, cryptosuite)
	if err != nil {
		t.Fatalf("Unexpected error creating credential manager %v", err)
	}
	_, ok := cm.(*credentialMgr.CredentialManager)
	if !ok {
		t.Fatalf("Unexpected credential manager created")
	}
}

func TestNewIdentityManager(t *testing.T) {
	config, err := config.FromFile("../../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}

	coreFactory := defcore.NewProviderFactory()
	cryptosuite, err := coreFactory.NewCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockSDK := mockapisdk.NewMockProviders(mockCtrl)

	mockSDK.EXPECT().Config().Return(config)
	mockSDK.EXPECT().CryptoSuite().Return(cryptosuite)

	factory := NewOrgClientFactory()
	imc, err := factory.NewIdentityManager(mockSDK, "Org1")
	if err != nil {
		t.Fatalf("Unexpected error creating credential manager %v", err)
	}
	_, ok := imc.(*idmgmtclient.IdentityManager)
	if !ok {
		t.Fatalf("Unexpected credential manager created")
	}
}
