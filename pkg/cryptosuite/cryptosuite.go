/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory"
	cryptosuiteimpl "github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp"
)

var defaultCryptoSuite = cryptosuiteimpl.GetSuite(factory.GetDefault())

//GetDefault returns apicryptosuite from factory default
func GetDefault() apicryptosuite.CryptoSuite {
	return defaultCryptoSuite
}

//GetSHA256Opts returns options relating to SHA-256.
func GetSHA256Opts() apicryptosuite.HashOpts {
	return &bccsp.SHA256Opts{}
}

//GetSHAOpts returns options for computing SHA.
func GetSHAOpts() apicryptosuite.HashOpts {
	return &bccsp.SHAOpts{}
}

//GetECDSAP256KeyGenOpts returns options for ECDSA key generation with curve P-256.
func GetECDSAP256KeyGenOpts(ephemeral bool) apicryptosuite.KeyGenOpts {
	return &bccsp.ECDSAP256KeyGenOpts{Temporary: ephemeral}
}
