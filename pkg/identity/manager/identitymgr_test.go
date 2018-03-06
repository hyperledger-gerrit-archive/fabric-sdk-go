/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package manager

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
)

const (
	org1               = "Org1"
	dummyUserStorePath = "/tmp/userstore"
)

var (
	fullConfig              core.Config
	noRegistrarConfig       core.Config
	embeddedRegistrarConfig core.Config
	cryptoSuite             core.CryptoSuite
	userStore               UserStore
)

// TestMain Load testing config
func TestMain(m *testing.M) {

	var err error
	fullConfig, err = config.FromFile("testdata/config_test.yaml")()
	if err != nil {
		panic(fmt.Sprintf("Failed to read full config: %v", err))
	}

	noRegistrarConfig, err = config.FromFile("testdata/config_no_registrar.yaml")()
	if err != nil {
		panic(fmt.Sprintf("Failed to read no-registrar config: %v", err))
	}

	embeddedRegistrarConfig, err = config.FromFile("testdata/config_embedded_registrar.yaml")()
	if err != nil {
		panic(fmt.Sprintf("Failed to read embedded-registrar config: %v", err))
	}

	// Delete all private keys from the crypto suite store
	// and users from the user store
	cleanup(fullConfig.KeyStorePath())
	defer cleanup(fullConfig.KeyStorePath())
	cleanup(fullConfig.CredentialStorePath())
	defer cleanup(fullConfig.CredentialStorePath())

	cryptoSuite, err = sw.GetSuiteByConfig(fullConfig)
	if cryptoSuite == nil {
		panic(fmt.Sprintf("Failed initialize cryptoSuite: %v", err))
	}
	if fullConfig.CredentialStorePath() != "" {
		userStore, err = NewCertFileUserStore(fullConfig.CredentialStorePath())
		if err != nil {
			panic(fmt.Sprintf("creating a user store failed: %v", err))
		}
	}

	os.Exit(m.Run())
}

// TestCreateValidBCCSPOptsForNewFabricClient test newidentityManager Client creation with valid inputs, successful scenario
func TestCreateValidBCCSPOptsForNewFabricClient(t *testing.T) {

	newCryptosuiteProvider, err := sw.GetSuiteByConfig(fullConfig)
	if err != nil {
		t.Fatalf("Expected fabric client ryptosuite to be created with SW BCCS provider, but got %v", err.Error())
	}

	stateStore := stateStoreFromConfig(t, fullConfig)
	_, err = New(stateStore, newCryptosuiteProvider, fullConfig)
	if err != nil {
		t.Fatalf("Expected fabric client to be created with SW BCCS provider, but got %v", err.Error())
	}
}

// readCert Reads a random cert for testing
func readCert(t *testing.T) []byte {
	cert, err := ioutil.ReadFile("testdata/root.pem")
	if err != nil {
		t.Fatalf("Error reading cert: %s", err.Error())
	}
	return cert
}

// TestInterfaces will test if the interface instantiation happens properly, ie no nil returned
func TestInterfaces(t *testing.T) {
	var apiIM core.IdentityManager
	var im IdentityManager

	apiIM = &im
	if apiIM == nil {
		t.Fatalf("this shouldn't happen.")
	}
}

func cleanup(storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to remove dir %s: %v\n", storePath, err))
	}
}

func cleanupTestPath(t *testing.T, storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		t.Fatalf("Cleaning up directory '%s' failed: %v", storePath, err)
	}
}

func mspIDByOrgName(t *testing.T, c core.Config, orgName string) string {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		t.Fatalf("network config retrieval failed: %v", err)
	}

	// viper keys are case insensitive
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		t.Fatalf("org config retrieval failed: %v", err)
	}
	return orgConfig.MspID
}
