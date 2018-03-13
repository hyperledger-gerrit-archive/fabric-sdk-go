/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"os"
	"testing"

	cryptosuite "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
)

const (
	org1Name = "Org1"
	org2Name = "Org2"
)

func TestEnrollOrg2(t *testing.T) {
	// Using shared SDK instance to increase test speed.
	sdk := mainSDK

	cryptoSuiteProvider, err := cryptosuite.GetSuiteByConfig(sdk.Config())
	if err != nil {
		t.Fatalf("Failed getting cryptosuite from config : %s", err)
	}

	stateStore, err := kvs.New(&kvs.FileKeyValueStoreOptions{Path: sdk.Config().CredentialStorePath()})
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore failed: %v", err)
	}

	userStore, err := msp.NewCertFileUserStore1(stateStore)
	if err != nil {
		t.Fatalf("msp.NewCertFileUserStore1 failed: %v", err)
	}

	identityManager, err := msp.NewIdentityManager(org2Name, userStore, cryptoSuiteProvider, sdk.Config())
	if err != nil {
		t.Fatalf("msp.NewIdentityManager failed: %v", err)
	}

	caClient, err := msp.NewCAClient(org2Name, identityManager, userStore, cryptoSuiteProvider, sdk.Config())
	if err != nil {
		t.Fatalf("msp.NewCAClient failed: %v", err)
	}

	err = caClient.Enroll("admin", "adminpw")
	if err != nil {
		t.Fatalf("Enroll returned error: %v", err)
	}

	//clean up the Keystore file, as its affecting other tests
	err = os.RemoveAll(sdk.Config().CredentialStorePath())
	if err != nil {
		t.Fatalf("Error deleting keyvalue store file: %v", err)
	}
}
