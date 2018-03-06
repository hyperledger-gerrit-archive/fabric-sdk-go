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
	"github.com/hyperledger/fabric-sdk-go/pkg/identity/caclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/identity/manager"
)

func TestEnrollOrg2(t *testing.T) {

	cryptoSuiteProvider, err := cryptosuite.GetSuiteByConfig(testFabricConfig)
	if err != nil {
		t.Fatalf("Failed getting cryptosuite from config : %s", err)
	}

	stateStore, err := kvs.New(&kvs.FileKeyValueStoreOptions{Path: testFabricConfig.CredentialStorePath()})
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore failed: %v", err)
	}

	identityManager, err := manager.New(stateStore, cryptoSuiteProvider, testFabricConfig)
	if err != nil {
		t.Fatalf("manager.New failed: %v", err)
	}

	caClient, err := caclient.New(org2Name, identityManager, stateStore, cryptoSuiteProvider, testFabricConfig)
	if err != nil {
		t.Fatalf("caclient.New failed: %v", err)
	}

	err = caClient.Enroll("admin", "adminpw")
	if err != nil {
		t.Fatalf("Enroll returned error: %v", err)
	}

	//clean up the Keystore file, as its affecting other tests
	err = os.RemoveAll(testFabricConfig.CredentialStorePath())
	if err != nil {
		t.Fatalf("Error deleting keyvalue store file: %v", err)
	}
}
