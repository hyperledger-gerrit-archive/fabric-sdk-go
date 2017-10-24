/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ccpolicy

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

func TestCCPolicyProvider(t *testing.T) {

	config, err := config.InitConfig("../../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	client := setupTestClient()

	ccPolicyProvider, err := NewCCPolicyProvider(config)
	if err != nil {
		t.Fatalf("Failed to setup cc policy provider: %s", err)
	}

	ccPolicyService, err := ccPolicyProvider.NewCCPolicyService(nil)
	if err == nil {
		t.Fatalf("Should have failed since policy service requires client")
	}

	ccPolicyService, err = ccPolicyProvider.NewCCPolicyService(client)
	if err != nil {
		t.Fatalf("Failed to setup cc policy service: %s", err)
	}

	_, err = ccPolicyService.GetChaincodePolicy("", "ccID")
	if err == nil {
		t.Fatalf("Should have failed to retrieve chaincode policy for empty channel")
	}

	_, err = ccPolicyService.GetChaincodePolicy("testChannel", "")
	if err == nil {
		t.Fatalf("Should have failed to retrieve chaincode policy for empty caincode id")
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
