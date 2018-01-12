package mock_apiconfig

import (
	tls "crypto/tls"
	x509 "crypto/x509"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
)

var GoodCert = &x509.Certificate{Raw: []byte{0, 1, 2}}
var BadCert = &x509.Certificate{Raw: []byte{1, 2}}
var TLSCert = tls.Certificate{Certificate: [][]byte{{3}, {4}}}
var CertPool = x509.NewCertPool()

const ErrorMessage = "default error message"

func DefaultMockConfig(t *testing.T) *MockConfig {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := NewMockConfig(mockCtrl)
	config.EXPECT().TLSCACertPool(GoodCert).Return(CertPool, nil).AnyTimes()
	config.EXPECT().TLSCACertPool(BadCert).Return(CertPool, errors.New(ErrorMessage)).AnyTimes()
	config.EXPECT().TLSCACertPool().Return(CertPool, nil).AnyTimes()
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5).AnyTimes()
	config.EXPECT().TLSClientCerts().Return([]tls.Certificate{TLSCert}, nil).AnyTimes()

	return config
}

func BadTLSClientMockConfig(t *testing.T) *MockConfig {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	config := NewMockConfig(mockCtrl)
	config.EXPECT().TLSCACertPool(GoodCert).Return(CertPool, nil).AnyTimes()
	config.EXPECT().TLSCACertPool(BadCert).Return(CertPool, errors.New(ErrorMessage)).AnyTimes()
	config.EXPECT().TLSCACertPool().Return(CertPool, nil).AnyTimes()
	config.EXPECT().TimeoutOrDefault(apiconfig.Endorser).Return(time.Second * 5).AnyTimes()
	config.EXPECT().TLSClientCerts().Return(nil, errors.Errorf(ErrorMessage)).AnyTimes()

	return config
}
