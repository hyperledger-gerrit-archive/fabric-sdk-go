/*
Copyright Hyperledger and its contributors.

SPDX-License-Identifier: Apache-2.0
*/

package vault_test

import (
	"testing"

	"reflect"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/vault"
	"github.com/stretchr/testify/assert"
)

func TestVaultCryptoSuiteImplementsInterface(t *testing.T) {
	cspi := reflect.TypeOf((*core.CryptoSuite)(nil)).Elem()

	ok := reflect.PtrTo(reflect.TypeOf(vault.CryptoSuite{})).Implements(cspi)

	if !ok {
		t.Fatalf("vault.CryptoSuite does not implement core.CryptoSuite")
	}
}

func TestKeyGenECDSAP256(t *testing.T) {
	csp, closer := testVaultCryptoSuite(t)
	defer closer()

	key, err := csp.KeyGen(cryptosuite.GetECDSAP256KeyGenOpts(false))
	assert.NoError(t, err)

	testVerificationFlow(t, csp, key.SKI())
}

func TestKeyGenRSA2048(t *testing.T) {
	csp, closer := testVaultCryptoSuite(t)
	defer closer()

	key, err := csp.KeyGen(cryptosuite.GetRSA2048KeyGenOpts(false))
	assert.NoError(t, err)

	testVerificationFlow(t, csp, key.SKI())
}

func TestKeyGenRSA4096(t *testing.T) {
	csp, closer := testVaultCryptoSuite(t)
	defer closer()

	key, err := csp.KeyGen(cryptosuite.GetRSA4096KeyGenOpts(false))
	assert.NoError(t, err)

	testVerificationFlow(t, csp, key.SKI())
}
