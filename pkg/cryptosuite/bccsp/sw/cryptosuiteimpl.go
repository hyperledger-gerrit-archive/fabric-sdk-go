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
	bccspSw "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp/internal"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

//GetSuiteByConfig returns cryptosuite adaptor for bccsp loaded according to given config
func GetSuiteByConfig(config apiconfig.Config) (apicryptosuite.CryptoSuite, error) {
	opts := getOptsByConfig(config)
	bccsp, err := getBCCSPFromOpts(opts)

	if err != nil {
		return nil, err
	}
	return &internal.CryptoSuite{BCCSP: bccsp}, nil
}

//GetSuiteWithDefaultEphemeral returns cryptosuite adaptor for bccsp with default ephemeral options (intended to aid testing)
func GetSuiteWithDefaultEphemeral() (apicryptosuite.CryptoSuite, error) {
	opts := getEphemeralOpts()
	bccsp, err := getBCCSPFromOpts(opts)

	if err != nil {
		return nil, err
	}
	return &internal.CryptoSuite{BCCSP: bccsp}, nil
}

func getBCCSPFromOpts(config *bccspSw.SwOpts) (bccsp.BCCSP, error) {
	f := &bccspSw.SWFactory{}

	csp, err := f.Get(config)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not initialize BCCSP %s", f.Name())
	}
	return csp, nil
}

//GetOptsByConfig Returns Factory opts for given SDK config
func getOptsByConfig(c apiconfig.Config) *bccspSw.SwOpts {
	// TODO: delete this check?
	if c.SecurityProvider() != "SW" {
		panic(fmt.Sprintf("Unsupported BCCSP Provider: %s", c.SecurityProvider()))
	}

	opts := &bccspSw.SwOpts{
		HashFamily: c.SecurityAlgorithm(),
		SecLevel:   c.SecurityLevel(),
		FileKeystore: &bccspSw.FileKeystoreOpts{
			KeyStorePath: c.KeyStorePath(),
		},
		Ephemeral: c.Ephemeral(),
	}
	logger.Debug("Initialized SW cryptosuite")

	return opts
}

func getEphemeralOpts() *bccspSw.SwOpts {
	opts := &bccspSw.SwOpts{
		HashFamily: "SHA2",
		SecLevel:   256,
		Ephemeral:  true,
	}
	logger.Debug("Initialized ephemeral SW cryptosuite with default opts")

	return opts
}

//GetKey returns implementation of of cryptosuite.Key
func GetKey(newkey bccsp.Key) apicryptosuite.Key {
	return internal.GetKey(newkey)
}
