/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"os"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/utils"
)

const (
	shaHashOptsAlgorithm    = "SHA"
	sha256HashOptsAlgorithm = "SHA256"
	ecdsap256KeyGenOpts     = "ECDSAP256"
)

// TestMain Load testing config
func TestMain(m *testing.M) {
	if !logging.IsLoggerInitialized() {
		logging.InitLogger(deflogger.GetLoggingProvider())
	}

	os.Exit(m.Run())
}

func TestGetDefault(t *testing.T) {

	defSuite := GetDefault()
	utils.VerifyNotEmpty(t, defSuite, "Not supposed to be nil defaultCryptSuite")

	hashbytes, err := defSuite.Hash([]byte("Sample message"), GetSHAOpts())
	utils.VerifyEmpty(t, err, "Not supposed to get error on defaultCryptSuite.Hash() call : %s", err)
	utils.VerifyNotEmpty(t, hashbytes, "Supposed to get valid hash from defaultCryptSuite.Hash()")

	defaultCryptoSuite = nil
	defSuite = GetDefault()
	utils.VerifyEmpty(t, defSuite, "Supposed to be nil defaultCryptSuite")
}

func TestHashOpts(t *testing.T) {

	//Get CryptoSuite SHA Opts
	hashOpts := GetSHAOpts()
	utils.VerifyNotEmpty(t, hashOpts, "Not supposed to be empty shaHashOpts")
	utils.VerifyTrue(t, hashOpts.Algorithm() == shaHashOptsAlgorithm, "Unexpected SHA hash opts, expected [%s], got [%s]", shaHashOptsAlgorithm, hashOpts.Algorithm())

	//Get CryptoSuite SHA256 Opts
	hashOpts = GetSHA256Opts()
	utils.VerifyNotEmpty(t, hashOpts, "Not supposed to be empty sha256HashOpts")
	utils.VerifyTrue(t, hashOpts.Algorithm() == sha256HashOptsAlgorithm, "Unexpected SHA hash opts, expected [%v], got [%v]", sha256HashOptsAlgorithm, hashOpts.Algorithm())

}

func TestKeyGenOpts(t *testing.T) {

	keygenOpts := GetECDSAP256KeyGenOpts(true)
	utils.VerifyNotEmpty(t, keygenOpts, "Not supposed to be empty ECDSAP256KeyGenOpts")
	utils.VerifyTrue(t, keygenOpts.Ephemeral(), "Expected keygenOpts.Ephemeral() ==> true")
	utils.VerifyTrue(t, keygenOpts.Algorithm() == ecdsap256KeyGenOpts, "Unexpected SHA hash opts, expected [%v], got [%v]", ecdsap256KeyGenOpts, keygenOpts.Algorithm())

	keygenOpts = GetECDSAP256KeyGenOpts(false)
	utils.VerifyNotEmpty(t, keygenOpts, "Not supposed to be empty ECDSAP256KeyGenOpts")
	utils.VerifyFalse(t, keygenOpts.Ephemeral(), "Expected keygenOpts.Ephemeral() ==> false")
	utils.VerifyTrue(t, keygenOpts.Algorithm() == ecdsap256KeyGenOpts, "Unexpected SHA hash opts, expected [%v], got [%v]", ecdsap256KeyGenOpts, keygenOpts.Algorithm())

}
