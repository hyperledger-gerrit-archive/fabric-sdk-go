/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"testing"

	"hash"

	"errors"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp"

	"github.com/hyperledger/fabric-sdk-go/pkg/logging/utils"
)

const (
	mockIdentifier   = "mock-test"
	signedIdentifier = "-signed"
	signingKey       = "signing-key"
	hashMessage      = "-msg-bytes"
	sampleKey        = "sample-key"
	getKey           = "-getkey"
)

func TestCryptoSuite(t *testing.T) {

	//Get BCCSP implementation
	samplebccsp := getMockBCCSP(mockIdentifier)

	//Get cryptosuite
	samplecryptoSuite := GetSuite(samplebccsp)

	//Test cryptosuite.Sign
	signedBytes, err := samplecryptoSuite.Sign(GetKey(getMockKey(signingKey)), nil, nil)
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey")
	utils.VerifyTrue(t, string(signedBytes) == mockIdentifier+signedIdentifier, "Got unexpected result from samplecryptoSuite.Sign")

	//Test cryptosuite.Hash
	hashBytes, err := samplecryptoSuite.Hash([]byte(hashMessage), &bccsp.SHAOpts{})
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey")
	utils.VerifyTrue(t, string(hashBytes) == mockIdentifier+hashMessage, "Got unexpected result from samplecryptoSuite.Hash")

	//Test cryptosuite.GetKey
	key, err := samplecryptoSuite.GetKey([]byte(sampleKey))
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey")
	utils.VerifyNotEmpty(t, key, "Not supposed to get empty key for samplecryptoSuite.GetKey")

	keyBytes, err := key.Bytes()
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey().GetBytes()")
	utils.VerifyTrue(t, string(keyBytes) == sampleKey+getKey, "Not supposed to get empty bytes for samplecryptoSuite.GetKey().GetBytes()")

	skiBytes := key.SKI()
	utils.VerifyTrue(t, string(skiBytes) == sampleKey+getKey, "Not supposed to get empty bytes for samplecryptoSuite.GetKey().GetSKI()")

	utils.VerifyTrue(t, key.Private(), "Not supposed to get false for samplecryptoSuite.GetKey().Private()")
	utils.VerifyTrue(t, key.Symmetric(), "Not supposed to get false for samplecryptoSuite.GetKey().Symmetric()")

	publikey, err := key.PublicKey()
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey().PublicKey()")
	utils.VerifyNotEmpty(t, publikey, "Not supposed to get empty key for samplecryptoSuite.GetKey().PublicKey()")
}

/*
	Mock implementation of bccsp.BCCSP and bccsp.Key
*/

func getMockBCCSP(identifier string) bccsp.BCCSP {
	return &mockBCCSP{identifier}
}

func getMockKey(identifier string) bccsp.Key {
	return &mockKey{identifier}
}

type mockBCCSP struct {
	identifier string
}

func (mock *mockBCCSP) KeyGen(opts bccsp.KeyGenOpts) (k bccsp.Key, err error) {
	return &mockKey{"keygen"}, nil
}

func (mock *mockBCCSP) KeyDeriv(k bccsp.Key, opts bccsp.KeyDerivOpts) (dk bccsp.Key, err error) {
	return &mockKey{"keyderiv"}, nil
}

func (mock *mockBCCSP) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (k bccsp.Key, err error) {
	return &mockKey{"keyimport"}, nil
}

func (mock *mockBCCSP) GetKey(ski []byte) (k bccsp.Key, err error) {
	return &mockKey{string(ski) + getKey}, nil
}

func (mock *mockBCCSP) Hash(msg []byte, opts bccsp.HashOpts) (hash []byte, err error) {
	return []byte(mock.identifier + string(msg)), nil
}

func (mock *mockBCCSP) GetHash(opts bccsp.HashOpts) (h hash.Hash, err error) {
	return nil, errors.New("Not able to Get Hash")
}

func (mock *mockBCCSP) Sign(k bccsp.Key, digest []byte, opts bccsp.SignerOpts) (signature []byte, err error) {
	return []byte(mock.identifier + signedIdentifier), nil
}

func (mock *mockBCCSP) Verify(k bccsp.Key, signature, digest []byte, opts bccsp.SignerOpts) (valid bool, err error) {
	return false, nil
}

func (mock *mockBCCSP) Encrypt(k bccsp.Key, plaintext []byte, opts bccsp.EncrypterOpts) (ciphertext []byte, err error) {
	return []byte(mock.identifier + "-encrypted"), nil
}

func (mock *mockBCCSP) Decrypt(k bccsp.Key, ciphertext []byte, opts bccsp.DecrypterOpts) (plaintext []byte, err error) {
	return []byte(mock.identifier + "-decrypted"), nil
}

type mockKey struct {
	identifier string
}

func (k *mockKey) Bytes() ([]byte, error) {
	return []byte(k.identifier), nil
}

func (k *mockKey) SKI() []byte {
	return []byte(k.identifier)
}

func (k *mockKey) Symmetric() bool {
	return true
}

func (k *mockKey) Private() bool {
	return true
}

func (k *mockKey) PublicKey() (bccsp.Key, error) {
	return &mockKey{k.identifier + "-public"}, nil
}
