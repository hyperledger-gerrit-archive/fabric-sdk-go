/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkcs11

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defcore"
	cryptosuite "github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp/pkcs11"
	"github.com/hyperledger/fabric-sdk-go/test/integration/e2e"
)

func TestE2E(t *testing.T) {
	// Create SDK setup for the integration tests
	sdkOptions := fabapi.Options{
		ConfigFile:  "../" + ConfigTestFile,
		CoreFactory: &CustomCryptoSuiteProviderFactory{},
	}

	e2e.Run(t, sdkOptions)
}

// CustomCryptoSuiteProviderFactory is will provide custom cryptosuite (bccsp.BCCSP)
type CustomCryptoSuiteProviderFactory struct {
	defcore.ProviderFactory
}

// NewCryptoSuiteProvider returns a new default implementation of BCCSP
func (f *CustomCryptoSuiteProviderFactory) NewCryptoSuiteProvider(config apiconfig.Config) (apicryptosuite.CryptoSuite, error) {
	return cryptosuite.GetSuiteByConfig(config)
}
