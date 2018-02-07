/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkcs11

import (
	"bytes"
	"crypto/sha256"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	api "github.com/hyperledger/fabric-sdk-go/api/config"
	"github.com/hyperledger/fabric-sdk-go/api/config/mocks"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	pkcsFactory "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory/pkcs11"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/pkcs11"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/utils"
)

var configImpl api.Config
var securityLevel = 256

const (
	providerTypePKCS11 = "PKCS11"
)

func TestBadConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("UNKNOWN")
	mockConfig.EXPECT().SecurityProvider().Return("UNKNOWN")

	//Get cryptosuite using config
	_, err := GetSuiteByConfig(mockConfig)
	if err == nil {
		t.Fatalf("Unknown security provider should return error")
	}
}
func TestCryptoSuiteByConfigPKCS11(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	//Prepare Config
	providerLib, softHSMPin, softHSMTokenLabel := pkcs11.FindPKCS11Lib()

	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("PKCS11")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return("/tmp/msp")
	mockConfig.EXPECT().Ephemeral().Return(false)
	mockConfig.EXPECT().SecurityProviderLibPath().Return(providerLib)
	mockConfig.EXPECT().SecurityProviderLabel().Return(softHSMTokenLabel)
	mockConfig.EXPECT().SecurityProviderPin().Return(softHSMPin)
	mockConfig.EXPECT().SoftVerify().Return(true)

	//Get cryptosuite using config
	c, err := GetSuiteByConfig(mockConfig)
	if err != nil {
		t.Fatalf("Not supposed to get error, but got: %v", err)
	}

	verifyHashFn(t, c)
}

func TestCryptoSuiteByConfigPKCS11Failure(t *testing.T) {

	//Prepare Config
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	//Prepare Config
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("PKCS11")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return("/tmp/msp")
	mockConfig.EXPECT().Ephemeral().Return(false)
	mockConfig.EXPECT().SecurityProviderLibPath().Return("")
	mockConfig.EXPECT().SecurityProviderLabel().Return("")
	mockConfig.EXPECT().SecurityProviderPin().Return("")
	mockConfig.EXPECT().SoftVerify().Return(true)

	//Get cryptosuite using config
	samplecryptoSuite, err := GetSuiteByConfig(mockConfig)
	utils.VerifyNotEmpty(t, err, "Supposed to get error on GetSuiteByConfig call : %s", err)
	utils.VerifyEmpty(t, samplecryptoSuite, "Not supposed to get valid cryptosuite")
}

func TestPKCS11CSPConfigWithValidOptions(t *testing.T) {
	opts := configurePKCS11Options("SHA2", securityLevel)
	f := &pkcsFactory.PKCS11Factory{}

	csp, err := f.Get(opts)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if csp == nil {
		t.Fatalf("BCCSP PKCS11 was not configured")
	}
	t.Logf("TestPKCS11CSPConfigWithValidOptions passed. BCCSP PKCS11 provider was configured (%v)", csp)

}

func TestPKCS11CSPConfigWithEmptyHashFamily(t *testing.T) {

	opts := configurePKCS11Options("", securityLevel)

	f := &pkcsFactory.PKCS11Factory{}
	t.Logf("PKCS11 factory name: %s", f.Name())
	_, err := f.Get(opts)
	if err == nil {
		t.Fatalf("Expected error 'Hash Family not supported'")
	}
	t.Log("TestPKCS11CSPConfigWithEmptyHashFamily passed. ")

}

func TestPKCS11CSPConfigWithIncorrectLevel(t *testing.T) {

	opts := configurePKCS11Options("SHA2", 100)

	f := &pkcsFactory.PKCS11Factory{}
	t.Logf("PKCS11 factory name: %s", f.Name())
	_, err := f.Get(opts)
	if err == nil {
		t.Fatalf("Expected error 'Failed initializing configuration'")
	}

}

func TestPKCS11CSPConfigWithEmptyProviderName(t *testing.T) {
	f := &pkcsFactory.PKCS11Factory{}
	if f.Name() != providerTypePKCS11 {
		t.Fatalf("Expected default name for PKCS11. Got %s", f.Name())
	}
}

func configurePKCS11Options(hashFamily string, securityLevel int) *pkcs11.PKCS11Opts {
	providerLib, softHSMPin, softHSMTokenLabel := pkcs11.FindPKCS11Lib()

	pkks := pkcs11.FileKeystoreOpts{KeyStorePath: os.TempDir()}
	//PKCS11 options
	pkcsOpt := pkcs11.PKCS11Opts{
		SecLevel:     securityLevel,
		HashFamily:   hashFamily,
		FileKeystore: &pkks,
		Library:      providerLib,
		Pin:          softHSMPin,
		Label:        softHSMTokenLabel,
		Ephemeral:    false,
	}

	return &pkcsOpt

}

func verifyHashFn(t *testing.T, c apicryptosuite.CryptoSuite) {
	msg := []byte("Hello")
	e := sha256.Sum256(msg)
	a, err := c.Hash(msg, &bccsp.SHA256Opts{})
	if err != nil {
		t.Fatalf("Not supposed to get error, but got: %v", err)
	}

	if bytes.Compare(a, e[:]) != 0 {
		t.Fatalf("Expected SHA 256 hash function")
	}
}
