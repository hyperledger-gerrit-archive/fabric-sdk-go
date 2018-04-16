/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package multisuite

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/pkcs11"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/wrapper"
)

func TestBadConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := mockcore.NewMockCryptoSuiteConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("UNKNOWN")
	mockConfig.EXPECT().SecurityProvider().Return("UNKNOWN")

	//Get cryptosuite using config
	_, err := GetSuiteByConfig(mockConfig)
	if err == nil {
		t.Fatalf("Unknown security provider should return error")
	}
}

func TestCryptoSuiteByConfigSW(t *testing.T) {

	// test multiple string cases
	securityProviderTypes := []string{
		// upper case
		"SW",
		// lower case
		"sw",
		// mixed case
		"Sw",
	}

	for _, securityProviderType := range securityProviderTypes {

		mockCtrl := gomock.NewController(t)

		mockConfig := mockcore.NewMockCryptoSuiteConfig(mockCtrl)
		mockConfig.EXPECT().SecurityProvider().Return(securityProviderType)
		mockConfig.EXPECT().SecurityProvider().Return(securityProviderType)
		mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
		mockConfig.EXPECT().SecurityLevel().Return(256)
		mockConfig.EXPECT().KeyStorePath().Return("")
		mockConfig.EXPECT().Ephemeral().Return(true)

		//Get cryptosuite using config
		c, err := GetSuiteByConfig(mockConfig)
		if err != nil {
			t.Fatalf("Not supposed to get error, but got: %v", err)
		}

		verifySuiteType(t, c, "*sw.impl")
		mockCtrl.Finish()
	}

}

func TestCryptoSuiteByConfigPKCS11(t *testing.T) {

	// test multiple string cases
	securityProviderTypes := []string{
		// upper case
		"PKCS11",
		// lower case
		"pkcs11",
		// mixed case
		"Pkcs11",
	}

	for _, securityProviderType := range securityProviderTypes {

		mockCtrl := gomock.NewController(t)

		//Prepare Config
		providerLib, softHSMPin, softHSMTokenLabel := pkcs11.FindPKCS11Lib()

		mockConfig := mockcore.NewMockCryptoSuiteConfig(mockCtrl)
		mockConfig.EXPECT().SecurityProvider().Return(securityProviderType)
		mockConfig.EXPECT().SecurityProvider().Return(securityProviderType)
		mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
		mockConfig.EXPECT().SecurityLevel().Return(256)
		mockConfig.EXPECT().KeyStorePath().Return("")
		mockConfig.EXPECT().Ephemeral().Return(true)
		mockConfig.EXPECT().SecurityProviderLibPath().Return(providerLib)
		mockConfig.EXPECT().SecurityProviderLabel().Return(softHSMTokenLabel)
		mockConfig.EXPECT().SecurityProviderPin().Return(softHSMPin)
		mockConfig.EXPECT().SoftVerify().Return(true)

		//Get cryptosuite using config
		c, err := GetSuiteByConfig(mockConfig)
		if err != nil {
			t.Fatalf("Not supposed to get error, but got: %v", err)
		}

		verifySuiteType(t, c, "*pkcs11.impl")
		mockCtrl.Finish()
	}

}

func verifySuiteType(t *testing.T, c core.CryptoSuite, expectedType string) {
	w, ok := c.(*wrapper.CryptoSuite)
	if !ok {
		t.Fatal("Unexpected cryptosuite type")
	}

	suiteType := reflect.TypeOf(w.BCCSP)
	if suiteType.String() != expectedType {
		t.Fatalf("Unexpected cryptosuite type: %s", suiteType)
	}
}
