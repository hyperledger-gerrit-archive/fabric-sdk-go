// +build !testpkcs11

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	cryptosuite "github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp/sw"
)

//getSuiteByConfig returns cryptosuite adaptor for bccsp loaded according to given config
func getSuiteByConfig(config apiconfig.Config) (apicryptosuite.CryptoSuite, error) {
	return cryptosuite.GetSuiteByConfig(config)
}

//getSuite returns cryptosuite adaptor for given bccsp.BCCSP implementation
func getSuite(bccsp bccsp.BCCSP) apicryptosuite.CryptoSuite {
	return cryptosuite.GetSuite(bccsp)
}
