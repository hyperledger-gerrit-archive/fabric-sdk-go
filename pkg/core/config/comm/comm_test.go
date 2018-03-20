/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"bytes"
	"encoding/hex"
	"testing"

	"strings"

	"crypto/tls"

	"reflect"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockcore"
)

func TestTLSConfigErrorAddingCertificate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockcore.DefaultMockConfig(mockCtrl)

	_, err := TLSConfig(mockcore.BadCert, "", config)
	if err == nil {
		t.Fatal("Expected failure adding invalid certificate")
	}

	if !strings.Contains(err.Error(), mockcore.ErrorMessage) {
		t.Fatalf("Expected error: %s", mockcore.ErrorMessage)
	}
}

func TestTLSConfigErrorFromClientCerts(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockcore.BadTLSClientMockConfig(mockCtrl)

	_, err := TLSConfig(mockcore.GoodCert, "", config)

	if err == nil {
		t.Fatal("Expected failure from loading client certs")
	}

	if !strings.Contains(err.Error(), mockcore.ErrorMessage) {
		t.Fatalf("Expected error: %s", mockcore.ErrorMessage)
	}
}

func TestTLSConfigHappyPath(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := mockcore.DefaultMockConfig(mockCtrl)

	serverHostOverride := "servernamebeingoverriden"

	tlsConfig, err := TLSConfig(mockcore.GoodCert, serverHostOverride, config)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if tlsConfig.ServerName != serverHostOverride {
		t.Fatal("Incorrect server name!")
	}

	if tlsConfig.RootCAs != mockcore.CertPool {
		t.Fatal("Incorrect cert pool")
	}

	if len(tlsConfig.Certificates) != 1 {
		t.Fatal("Incorrect number of certs")
	}

	if !reflect.DeepEqual(tlsConfig.Certificates[0], mockcore.TLSCert) {
		t.Fatal("Certs do not match")
	}
}

func TestNoTlsCertHash(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mockcore.NewMockConfig(mockCtrl)

	config.EXPECT().TLSClientCerts().Return([]tls.Certificate{}, nil)

	tlsCertHash := TLSCertHash(config)

	if len(tlsCertHash) != 0 {
		t.Fatal("Unexpected non-empty cert hash")
	}
}

func TestEmptyTlsCertHash(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mockcore.NewMockConfig(mockCtrl)

	emptyCert := tls.Certificate{}
	config.EXPECT().TLSClientCerts().Return([]tls.Certificate{emptyCert}, nil)

	tlsCertHash := TLSCertHash(config)

	if len(tlsCertHash) != 0 {
		t.Fatal("Unexpected non-empty cert hash")
	}
}

func TestTlsCertHash(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	config := mockcore.NewMockConfig(mockCtrl)

	cert, err := tls.LoadX509KeyPair("testdata/server.crt", "testdata/server.key")
	if err != nil {
		t.Fatalf("Unexpected error loading cert %v", err)
	}

	config.EXPECT().TLSClientCerts().Return([]tls.Certificate{cert}, nil)
	tlsCertHash := TLSCertHash(config)

	// openssl x509 -fingerprint -sha256 -in testdata/server.crt
	// SHA256 Fingerprint=0D:D5:90:B8:A5:0E:A6:04:3E:A8:75:16:BF:77:A8:FE:E7:C5:62:2D:4C:B3:CB:99:12:74:72:2A:D8:BA:B8:92
	expectedHash, err := hex.DecodeString("0DD590B8A50EA6043EA87516BF77A8FEE7C5622D4CB3CB991274722AD8BAB892")
	if err != nil {
		t.Fatalf("Unexpected error decoding cert fingerprint %v", err)
	}

	if bytes.Compare(tlsCertHash, expectedHash) != 0 {
		t.Fatal("Cert hash calculated incorrectly")
	}
}