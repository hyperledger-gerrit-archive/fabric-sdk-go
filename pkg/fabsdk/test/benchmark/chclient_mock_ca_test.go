/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package benchmark

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	mspImpl "github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

const (
	caServerURLListen = "http://localhost:0"
	configPath        = "../../../core/config/testdata/config_test.yaml"
)

var caServerURL string

type testFixture struct {
	cryptoSuiteConfig core.CryptoSuiteConfig
	identityConfig    msp.IdentityConfig
}

var caServer = &mockmsp.MockFabricCAServer{}

func (f *testFixture) setup() *fabsdk.FabricSDK {

	var lis net.Listener
	var err error
	if !caServer.Running() {
		lis, err = net.Listen("tcp", strings.TrimPrefix(caServerURLListen, "http://"))
		if err != nil {
			panic(fmt.Sprintf("Error starting CA Server %s", err))
		}

		caServerURL = "http://" + lis.Addr().String()
	}

	backend, err := config.FromFile(configPath)()
	if err != nil {
		panic(err)
	}

	//Override ca matchers for this test
	customBackend := getCustomBackend(backend...)

	configProvider := func() ([]core.ConfigBackend, error) {
		return customBackend, nil
	}

	// Instantiate the SDK
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		panic(fmt.Sprintf("SDK init failed: %s", err))
	}

	configBackend, err := sdk.Config()
	if err != nil {
		panic(fmt.Sprintf("Failed to get config: %s", err))
	}

	f.cryptoSuiteConfig = cryptosuite.ConfigFromBackend(configBackend)
	f.identityConfig, err = mspImpl.ConfigFromBackend(configBackend)
	if err != nil {
		panic(fmt.Sprintf("Failed to get identity config: %s", err))
	}

	// Delete all private keys from the crypto suite store
	// and users from the user store
	cleanup(f.cryptoSuiteConfig.KeyStorePath())
	cleanup(f.identityConfig.CredentialStorePath())

	ctxProvider := sdk.Context()
	ctx, err := ctxProvider()
	if err != nil {
		panic(fmt.Sprintf("Failed to init context: %s", err))
	}

	// Start Http Server if it's not running
	if !caServer.Running() {
		caServer.Start(lis, ctx.CryptoSuite())
	}

	return sdk
}

func (f *testFixture) close() {
	cleanup(f.identityConfig.CredentialStorePath())
	cleanup(f.cryptoSuiteConfig.KeyStorePath())
}

func cleanup(storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to remove dir %s: %s\n", storePath, err))
	}
}

func getCustomBackend(currentBackends ...core.ConfigBackend) []core.ConfigBackend {
	backendMap := make(map[string]interface{})

	//Custom URLs for ca configs
	networkConfig := fab.NetworkConfig{}
	configLookup := lookup.New(currentBackends...)
	configLookup.UnmarshalKey("certificateAuthorities", &networkConfig.CertificateAuthorities)

	ca1Config := networkConfig.CertificateAuthorities["ca.org1.example.com"]
	ca1Config.URL = caServerURL
	ca2Config := networkConfig.CertificateAuthorities["ca.org2.example.com"]
	ca2Config.URL = caServerURL

	networkConfig.CertificateAuthorities["ca.org1.example.com"] = ca1Config
	networkConfig.CertificateAuthorities["ca.org2.example.com"] = ca2Config
	backendMap["certificateAuthorities"] = networkConfig.CertificateAuthorities

	backends := append([]core.ConfigBackend{}, &mocks.MockConfigBackend{KeyValueMap: backendMap})
	return append(backends, currentBackends...)

}
