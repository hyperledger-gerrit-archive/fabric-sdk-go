/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper

import (
	"errors"
	"hash"
	"testing"

	"strings"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/testutils"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core/mocks"
)

const (
	mockIdentifier   = "mock-test"
	signedIdentifier = "-signed"
	signingKey       = "signing-key"
	hashMessage      = "-msg-bytes"
	sampleKey        = "sample-key"
	getKey           = "-getkey"
	keyImport        = "-keyimport"
	keyGen           = "-keygent"
)

func TestCryptoSuite(t *testing.T) {

	//Get BCCSP implementation
	samplebccsp := getMockBCCSP(mockIdentifier)

	//Get cryptosuite
	samplecryptoSuite := NewCryptoSuite(samplebccsp)

	//Verify CryptSuite
	verifyCryptoSuite(t, samplecryptoSuite)

}

func TestCryptoSuiteByConfig(t *testing.T) {

	//Prepare Config
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("SW")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return("/tmp/msp")
	mockConfig.EXPECT().Ephemeral().Return(false)

	//Get cryptosuite using config
	samplecryptoSuite, err := getSuiteByConfig(mockConfig)
	testutils.VerifyEmpty(t, err, "Not supposed to get error on GetSuiteByConfig call : %s", err)
	testutils.VerifyNotEmpty(t, samplecryptoSuite, "Supposed to get valid cryptosuite")

	hashbytes, err := samplecryptoSuite.Hash([]byte(hashMessage), &bccsp.SHAOpts{})
	testutils.VerifyEmpty(t, err, "Not supposed to get error on GetSuiteByConfig call : %s", err)
	testutils.VerifyNotEmpty(t, hashbytes, "Supposed to get valid hash from sample cryptosuite")

}

func TestCryptoSuiteByConfigFailures(t *testing.T) {

	//Prepare Config
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("SW")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(100)
	mockConfig.EXPECT().KeyStorePath().Return("/tmp/msp")
	mockConfig.EXPECT().Ephemeral().Return(false)

	//Get cryptosuite using config
	samplecryptoSuite, err := getSuiteByConfig(mockConfig)
	testutils.VerifyNotEmpty(t, err, "Supposed to get error on GetSuiteByConfig call : %s", err)
	testutils.VerifyEmpty(t, samplecryptoSuite, "Not supposed to get valid cryptosuite")

	if !strings.HasPrefix(err.Error(), "Failed initializing configuration") {
		t.Fatalf("Didn't get expected failure, got %s instead", err)
	}

}

// TestCreateInvalidBCCSPSecurityLevel will test cryptsuite creation with invalid BCCSP options
func TestCreateInvalidBCCSPSecurityLevel(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().SecurityProvider().Return("SW")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(100)
	mockConfig.EXPECT().KeyStorePath().Return("/tmp/msp")
	mockConfig.EXPECT().Ephemeral().Return(false)

	_, err := getSuiteByConfig(mockConfig)
	if !strings.Contains(err.Error(), "Security level not supported [100]") {
		t.Fatalf("Expected invalid security level error, but got %v", err.Error())
	}

}

// TestCreateInvalidBCCSPHashFamily will test cryptsuite creation with bad HashFamily
func TestCreateInvalidBCCSPHashFamily(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().SecurityProvider().Return("SW")
	mockConfig.EXPECT().SecurityAlgorithm().Return("ABC")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return("/tmp/msp")
	mockConfig.EXPECT().Ephemeral().Return(false)

	_, err := getSuiteByConfig(mockConfig)
	if !strings.Contains(err.Error(), "Hash Family not supported [ABC]") {
		t.Fatalf("Expected invalid hash family error, but got %v", err.Error())
	}
}

// TestCreateInvalidSecurityProviderPanic will test cryptsuite creation with bad HashFamily
func TestCreateInvalidSecurityProviderPanic(t *testing.T) {

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("was supposed to panic")
		}
	}()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().SecurityProvider().Return("XYZ")
	mockConfig.EXPECT().SecurityProvider().Return("XYZ")

	getSuiteByConfig(mockConfig)
	t.Fatalf("Getting cryptosuite with invalid security provider supposed to panic")
}

func verifyCryptoSuite(t *testing.T, samplecryptoSuite core.CryptoSuite) {
	//Test cryptosuite.Sign
	signedBytes, err := samplecryptoSuite.Sign(GetKey(getMockKey(signingKey)), nil, nil)
	testutils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey : %s", err)
	testutils.VerifyTrue(t, string(signedBytes) == mockIdentifier+signedIdentifier, "Got unexpected result from samplecryptoSuite.Sign")

	//Test cryptosuite.Hash
	hashBytes, err := samplecryptoSuite.Hash([]byte(hashMessage), &bccsp.SHAOpts{})
	testutils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey")
	testutils.VerifyTrue(t, string(hashBytes) == mockIdentifier+hashMessage, "Got unexpected result from samplecryptoSuite.Hash")

	//Test cryptosuite.GetKey
	key, err := samplecryptoSuite.GetKey([]byte(sampleKey))
	testutils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey")
	testutils.VerifyNotEmpty(t, key, "Not supposed to get empty key for samplecryptoSuite.GetKey")

	keyBytes, err := key.Bytes()
	testutils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey().GetBytes()")
	testutils.VerifyTrue(t, string(keyBytes) == sampleKey+getKey, "Not supposed to get empty bytes for samplecryptoSuite.GetKey().GetBytes()")

	skiBytes := key.SKI()
	testutils.VerifyTrue(t, string(skiBytes) == sampleKey+getKey, "Not supposed to get empty bytes for samplecryptoSuite.GetKey().GetSKI()")

	testutils.VerifyTrue(t, key.Private(), "Not supposed to get false for samplecryptoSuite.GetKey().Private()")
	testutils.VerifyTrue(t, key.Symmetric(), "Not supposed to get false for samplecryptoSuite.GetKey().Symmetric()")

	publikey, err := key.PublicKey()
	testutils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey().PublicKey()")
	testutils.VerifyNotEmpty(t, publikey, "Not supposed to get empty key for samplecryptoSuite.GetKey().PublicKey()")

	//Test cryptosuite.KeyImport
	key, err = samplecryptoSuite.KeyImport(nil, &bccsp.X509PublicKeyImportOpts{Temporary: true})
	testutils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.KeyImport")
	testutils.VerifyNotEmpty(t, key, "Not supposed to get empty key for samplecryptoSuite.KeyImport")

	keyBytes, err = key.Bytes()
	testutils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.KeyImport().GetBytes()")
	testutils.VerifyTrue(t, string(keyBytes) == mockIdentifier+keyImport, "Unexpected bytes for samplecryptoSuite.KeyImport().GetBytes()")

	skiBytes = key.SKI()
	testutils.VerifyTrue(t, string(skiBytes) == mockIdentifier+keyImport, "Unexpected bytes for samplecryptoSuite.KeyImport().GetSKI()")

	testutils.VerifyTrue(t, key.Private(), "Not supposed to get false for samplecryptoSuite.KeyImport().Private()")
	testutils.VerifyTrue(t, key.Symmetric(), "Not supposed to get false for samplecryptoSuite.KeyImport().Symmetric()")

	publikey, err = key.PublicKey()
	testutils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.KeyImport().PublicKey()")
	testutils.VerifyNotEmpty(t, publikey, "Not supposed to get empty key for samplecryptoSuite.KeyImport().PublicKey()")

	//Test cryptosuite.KeyGen
	key, err = samplecryptoSuite.KeyGen(nil)
	testutils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.KeyGen")
	testutils.VerifyNotEmpty(t, key, "Not supposed to get empty key for samplecryptoSuite.KeyGen")

	keyBytes, err = key.Bytes()
	testutils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.KeyGen().GetBytes()")
	testutils.VerifyTrue(t, string(keyBytes) == mockIdentifier+keyGen, "Unexpected bytes for samplecryptoSuite.KeyGen().GetBytes()")

	skiBytes = key.SKI()
	testutils.VerifyTrue(t, string(skiBytes) == mockIdentifier+keyGen, "Unexpected bytes for samplecryptoSuite.KeyGen().GetSKI()")

	testutils.VerifyTrue(t, key.Private(), "Not supposed to get false for samplecryptoSuite.KeyGen().Private()")
	testutils.VerifyTrue(t, key.Symmetric(), "Not supposed to get false for samplecryptoSuite.KeyGen().Symmetric()")

	publikey, err = key.PublicKey()
	testutils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.KeyGen().PublicKey()")
	testutils.VerifyNotEmpty(t, publikey, "Not supposed to get empty key for samplecryptoSuite.KeyGen().PublicKey()")

	//Test cryptosuite.GetHash
	hash, err := samplecryptoSuite.GetHash(&bccsp.SHA256Opts{})
	testutils.VerifyNotEmpty(t, err, "Supposed to get error for samplecryptoSuite.GetHash")
	testutils.VerifyEmpty(t, hash, "Supposed to get empty hash for samplecryptoSuite.GetHash")

	//Test cryptosuite.GetHash
	valid, err := samplecryptoSuite.Verify(GetKey(getMockKey(signingKey)), nil, nil, nil)
	testutils.VerifyEmpty(t, err, "Not supposed to get error for samplecryptoSuite.Verify")
	testutils.VerifyTrue(t, valid, "Supposed to get true for samplecryptoSuite.Verify")
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
	return &mockKey{mock.identifier + keyGen}, nil
}

func (mock *mockBCCSP) KeyDeriv(k bccsp.Key, opts bccsp.KeyDerivOpts) (dk bccsp.Key, err error) {
	return &mockKey{"keyderiv"}, nil
}

func (mock *mockBCCSP) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (k bccsp.Key, err error) {
	return &mockKey{mock.identifier + keyImport}, nil
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
	return true, nil
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
