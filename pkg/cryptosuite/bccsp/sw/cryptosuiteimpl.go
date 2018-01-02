/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sw

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	bccspFactory "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp/internal"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

//GetSuite returns cryptosuite adaptor for given bccsp.BCCSP implementation
func GetSuite(bccsp bccsp.BCCSP) apicryptosuite.CryptoSuite {
	return &internal.CryptoSuite{BCCSP: bccsp}
}

//GetSuiteByConfig returns cryptosuite adaptor for bccsp loaded according to given config
func GetSuiteByConfig(config apiconfig.Config) (apicryptosuite.CryptoSuite, error) {
	opts := GetOptsByConfig(config)
	bccsp, err := bccspFactory.GetBCCSPFromOpts(opts)

	if err != nil {
		return nil, err
	}
	return &internal.CryptoSuite{BCCSP: bccsp}, nil
}

//GetOptsByConfig Returns Factory opts for given SDK config
func GetOptsByConfig(c apiconfig.Config) *bccspFactory.FactoryOpts {
	var opts *bccspFactory.FactoryOpts

	switch c.SecurityProvider() {
	case "SW":
		opts = &bccspFactory.FactoryOpts{
			ProviderName: "SW",
			SwOpts: &bccspFactory.SwOpts{
				HashFamily: c.SecurityAlgorithm(),
				SecLevel:   c.SecurityLevel(),
				FileKeystore: &bccspFactory.FileKeystoreOpts{
					KeyStorePath: c.KeyStorePath(),
				},
				Ephemeral: c.Ephemeral(),
			},
		}
		logger.Debug("Initialized SW ")
		bccspFactory.InitFactories(opts)
		return opts

	default:
		panic(fmt.Sprintf("Unsupported BCCSP Provider: %s", c.SecurityProvider()))
	}
}

//GetKey returns implementation of of cryptosuite.Key
func GetKey(newkey bccsp.Key) apicryptosuite.Key {
	return internal.GetKey(newkey)
}
