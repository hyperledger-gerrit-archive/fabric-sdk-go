/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialmgr

import (
	"testing"

	factory "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/sdkpatch/cryptosuitebridge"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

func TestCredentialManager(t *testing.T) {

	config, err := config.FromFile("../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}

	credentialMgr, err := NewCredentialManager("Org1", config, factory.GetDefault())
	if err != nil {
		t.Fatalf("Failed to setup credential manager: %s", err)
	}

	_, err = credentialMgr.GetSigningIdentity("")
	if err == nil {
		t.Fatalf("Should have failed to retrieve signing identity for empty user name")
	}

	_, err = credentialMgr.GetSigningIdentity("Non-Existent")
	if err == nil {
		t.Fatalf("Should have failed to retrieve signing identity for non-existent user")
	}

	_, err = credentialMgr.GetSigningIdentity("User1")
	if err != nil {
		t.Fatalf("Failed to retrieve signing identity: %s", err)
	}

}

func TestInvalidOrgCredentialManager(t *testing.T) {

	config, err := config.FromFile("../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Invalid Org
	_, err = NewCredentialManager("invalidOrg", config, &fcmocks.MockCryptoSuite{})
	if err == nil {
		t.Fatalf("Should have failed to setup manager for invalid org")
	}

}

func TestCredentialManagerFromEmbeddedCryptoConfig(t *testing.T) {
	config, err := config.FromFile("../../../test/fixtures/config/config_test_embedded_pems.yaml")()

	if err != nil {
		t.Fatalf(err.Error())
	}

	credentialMgr, err := NewCredentialManager("Org1", config, factory.GetDefault())
	if err != nil {
		t.Fatalf("Failed to setup credential manager: %s", err)
	}

	_, err = credentialMgr.GetSigningIdentity("")
	if err == nil {
		t.Fatalf("Should have failed to retrieve signing identity for empty user name")
	}

	_, err = credentialMgr.GetSigningIdentity("Non-Existent")
	if err == nil {
		t.Fatalf("Should have failed to retrieve signing identity for non-existent user")
	}

	_, err = credentialMgr.GetSigningIdentity("EmbeddedUser")
	if err != nil {
		t.Fatalf("Failed to retrieve signing identity: %+v", err)
	}

	_, err = credentialMgr.GetSigningIdentity("EmbeddedUserWithPaths")
	if err != nil {
		t.Fatalf("Failed to retrieve signing identity: %+v", err)
	}

	_, err = credentialMgr.GetSigningIdentity("EmbeddedUserMixed")
	if err != nil {
		t.Fatalf("Failed to retrieve signing identity: %+v", err)
	}

	_, err = credentialMgr.GetSigningIdentity("EmbeddedUserMixed2")
	if err != nil {
		t.Fatalf("Failed to retrieve signing identity: %+v", err)
	}
}
