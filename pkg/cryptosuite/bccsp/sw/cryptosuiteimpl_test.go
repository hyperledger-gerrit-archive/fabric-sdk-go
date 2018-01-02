/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sw

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
)

func TestCryptoSuiteByConfigPKCS11Unsupported(t *testing.T) {
	//Prepare Config
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	//Prepare Config
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("PKCS11")
	mockConfig.EXPECT().SecurityProvider().Return("PKCS11")

	//Get cryptosuite using config
	_, err := GetSuiteByConfig(mockConfig)
	if err == nil {
		t.Fatalf("Getting cryptosuite with unsupported pkcs11 security provider supposed to error")
	}
}
