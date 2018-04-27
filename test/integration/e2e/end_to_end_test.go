/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package e2e

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func TestE2E(t *testing.T) {
	configPath := "../../fixtures/config/config_test.yaml"
	//End to End testing
	Run(t, true, config.FromFile(configPath))

	//Using setup done set above by end to end test, run below test with new config which has no orderer config inside
	runWithNoOrdererConfig(t, config.FromFile("../../fixtures/config/config_test_no_orderer.yaml"))

	// use a config that does not haven any EndpointConfig entries to test sub interfaces override
	configPath = "../../fixtures/config/config_test_crypto_bccsp.yaml"

	//Using same Run call as above but with programmatically overriding interfaces
	Run(t, false, config.FromFile(configPath),
		fabsdk.WithConfigEndpoint(endpointConfigImpls...))

	// TODO test with below line once IdentityConfig and CryptoConfig are split into
	// TODO sub interfaces like EndpointConfig and pass them in like WithConfigEndpoint,
	// TODO this will allow to test overriding all config interfaces without the need of a config file
	// TODO maybe add config.BareBone() to get a configProvider without a config file instead of passing in an empty file
	// use an empty config file to fully depend on injected EndpointConfig interfaces
	//configPath = "../../../pkg/core/config/testdata/viper-test.yaml"
}
