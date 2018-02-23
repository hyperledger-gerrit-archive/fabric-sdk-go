/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricca

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"

	config "github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	cryptosuiteimpl "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	bccspwrapper "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/wrapper"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ca/mocks"
)

var configImp config.Config
var cryptoSuiteProvider core.CryptoSuite
var org1 = "peerorg1"
var caServerURL = "http://localhost:8090"
var wrongCAServerURL = "http://localhost:8091"
var userStorePath = "/tmp/userstore"

// TestMain Load testing config
func TestMain(m *testing.M) {
	cleanup(userStorePath)
	configImp = mocks.NewMockConfig(caServerURL)
	cryptoSuiteProvider, _ = cryptosuiteimpl.GetSuiteByConfig(configImp)
	if cryptoSuiteProvider == nil {
		panic("Failed initialize cryptoSuiteProvider")
	}
	// Start Http Server
	go mocks.StartFabricCAMockServer(strings.TrimPrefix(caServerURL, "http://"), cryptoSuiteProvider)
	// Allow HTTP server to start
	time.Sleep(1 * time.Second)
	os.Exit(m.Run())
}

// TestEnroll will test multiple enrol scenarios
func TestEnroll(t *testing.T) {

	fabricCAClient, err := New(org1, configImp, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}
	_, _, err = fabricCAClient.Enroll("", "user1")
	if err == nil {
		t.Fatalf("Enroll didn't return error")
	}
	_, _, err = fabricCAClient.Enroll("test", "")
	if err == nil {
		t.Fatalf("Enroll didn't return error")
	}
	_, _, err = fabricCAClient.Enroll("enrollmentID", "enrollmentSecret")
	if err != nil {
		t.Fatalf("fabricCAClient Enroll return error %v", err)
	}

	wrongConfigImp := mocks.NewMockConfig(wrongCAServerURL)
	fabricCAClient, err = New(org1, wrongConfigImp, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}
	_, _, err = fabricCAClient.Enroll("enrollmentID", "enrollmentSecret")
	if err == nil {
		t.Fatalf("Enroll didn't return error")
	}

}

// TestRegister tests multiple scenarios of registering a test (mocked or nil user) and their certs
func TestRegister(t *testing.T) {

	fabricCAClient, err := New(org1, configImp, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}

	// Register with nil request
	_, err = fabricCAClient.Register(nil)
	if err == nil {
		t.Fatalf("Expected error with nil request")
	}

	// Register without registration name parameter
	_, err = fabricCAClient.Register(&fab.RegistrationRequest{})
	if err == nil {
		t.Fatalf("Expected error without registration name parameter")
	}

	// Register with valid request
	var attributes []fab.Attribute
	attributes = append(attributes, fab.Attribute{Key: "test1", Value: "test2"})
	attributes = append(attributes, fab.Attribute{Key: "test2", Value: "test3"})
	secret, err := fabricCAClient.Register(&fab.RegistrationRequest{Name: "test", Affiliation: "test", Attributes: attributes})
	if err != nil {
		t.Fatalf("fabricCAClient Register return error %v", err)
	}
	if secret != "mockSecretValue" {
		t.Fatalf("fabricCAClient Register return wrong value %s", secret)
	}
}

// TestRevoke will test multiple revoking a user with a nil request or a nil user
func TestRevoke(t *testing.T) {

	cryptoSuiteProvider, err := cryptosuiteimpl.GetSuiteByConfig(configImp)
	if err != nil {
		t.Fatalf("cryptosuite.GetSuiteByConfig returned error: %v", err)
	}

	fabricCAClient, err := New(org1, configImp, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}
	mockKey := bccspwrapper.GetKey(&mocks.MockKey{})
	user := mocks.NewMockUser("test")
	// Revoke with nil request
	_, err = fabricCAClient.Revoke(nil)
	if err == nil {
		t.Fatalf("Expected error with nil request")
	}
	user.SetEnrollmentCertificate(readCert(t))
	user.SetPrivateKey(mockKey)
	_, err = fabricCAClient.Revoke(&fab.RevocationRequest{})
	if err == nil {
		t.Fatalf("Expected decoding error with test cert")
	}
}

// TestReenroll will test multiple scenarios of re enrolling a user
func TestReenroll(t *testing.T) {

	fabricCAClient, err := New(org1, configImp, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}
	user := mocks.NewMockUser("")
	// Reenroll with nil user
	_, _, err = fabricCAClient.Reenroll(nil)
	if err == nil {
		t.Fatalf("Expected error with nil user")
	}
	if err.Error() != "user required" {
		t.Fatalf("Expected error user required. Got: %s", err.Error())
	}
	// Reenroll with user.Name is empty
	_, _, err = fabricCAClient.Reenroll(user)
	if err == nil {
		t.Fatalf("Expected error with user.Name is empty")
	}
	if err.Error() != "user name missing" {
		t.Fatalf("Expected error user name missing. Got: %s", err.Error())
	}
	// Reenroll with user.EnrollmentCertificate is empty
	user = mocks.NewMockUser("testUser")
	_, _, err = fabricCAClient.Reenroll(user)
	if err == nil {
		t.Fatalf("Expected error with user.EnrollmentCertificate is empty")
	}
	if !strings.Contains(err.Error(), "createSigningIdentity failed") {
		t.Fatalf("Expected error createSigningIdentity failed. Got: %s", err.Error())
	}
	// Reenroll with appropriate user
	user.SetEnrollmentCertificate(readCert(t))
	key, err := cryptosuite.GetDefault().KeyGen(cryptosuite.GetECDSAP256KeyGenOpts(true))
	if err != nil {
		t.Fatalf("KeyGen return error %v", err)
	}
	user.SetPrivateKey(key)
	_, _, err = fabricCAClient.Reenroll(user)
	if err != nil {
		t.Fatalf("Reenroll return error %v", err)
	}

	// Reenroll with wrong fabric-ca server url
	wrongConfigImp := mocks.NewMockConfig(wrongCAServerURL)
	fabricCAClient, err = New(org1, wrongConfigImp, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}
	_, _, err = fabricCAClient.Reenroll(user)
	if err == nil {
		t.Fatalf("Expected error with wrong fabric-ca server url")
	}
	if !strings.Contains(err.Error(), "reenroll failed") {
		t.Fatalf("Expected error with wrong fabric-ca server url. Got: %s", err.Error())
	}
}

// TestGetCAName will test the CAName is properly created once a new FabricCAClient is created
func TestGetCAName(t *testing.T) {
	fabricCAClient, err := New(org1, configImp, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}
	if fabricCAClient.CAName() != "test" {
		t.Fatalf("CAName returned wrong value: %s", fabricCAClient.CAName())
	}
}

// TestCreateNewFabricCAClientCAConfigMissingFailure will test newFabricCA Client creation with with CAConfig
func TestCreateNewFabricCAClientCAConfigMissingFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().CAConfig(org1).Return(nil, errors.New("CAConfig error"))
	mockConfig.EXPECT().CredentialStorePath().Return(userStorePath)

	_, err := New(org1, mockConfig, cryptoSuiteProvider)
	if err.Error() != "CAConfig error" {
		t.Fatalf("Expected error from CAConfig. Got: %s", err.Error())
	}
}

// TestCreateNewFabricCAClientCertFilesMissingFailure will test newFabricCA Client creation with missing CA Cert files
func TestCreateNewFabricCAClientCertFilesMissingFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().CAConfig(org1).Return(&config.CAConfig{}, nil).AnyTimes()
	mockConfig.EXPECT().CredentialStorePath().Return(userStorePath)
	mockConfig.EXPECT().CAServerCertPaths(org1).Return(nil, errors.New("CAServerCertPaths error"))
	_, err := New(org1, mockConfig, cryptoSuiteProvider)
	if err.Error() != "CAServerCertPaths error" {
		t.Fatalf("Expected error from CAServerCertPaths. Got: %s", err.Error())
	}
}

// TestCreateNewFabricCAClientCertFileErrorFailure will test newFabricCA Client creation with missing CA Cert files, additional scenario
func TestCreateNewFabricCAClientCertFileErrorFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().CAConfig(org1).Return(&config.CAConfig{}, nil).AnyTimes()
	mockConfig.EXPECT().CredentialStorePath().Return(userStorePath)
	mockConfig.EXPECT().CAServerCertPaths(org1).Return([]string{"test"}, nil)
	mockConfig.EXPECT().CAClientCertPath(org1).Return("", errors.New("CAClientCertPath error"))
	_, err := New(org1, mockConfig, cryptoSuiteProvider)
	if err.Error() != "CAClientCertPath error" {
		t.Fatalf("Expected error from CAClientCertPath. Got: %s", err.Error())
	}
}

// TestCreateNewFabricCAClientKeyFileErrorFailure will test newFabricCA Client creation with missing CA Cert files and missing key
func TestCreateNewFabricCAClientKeyFileErrorFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().CAConfig(org1).Return(&config.CAConfig{}, nil).AnyTimes()
	mockConfig.EXPECT().CredentialStorePath().Return(userStorePath)
	mockConfig.EXPECT().CAServerCertPaths(org1).Return([]string{"test"}, nil)
	mockConfig.EXPECT().CAClientCertPath(org1).Return("", nil)
	mockConfig.EXPECT().CAClientKeyPath(org1).Return("", errors.New("CAClientKeyPath error"))
	_, err := New(org1, mockConfig, cryptoSuiteProvider)
	if err.Error() != "CAClientKeyPath error" {
		t.Fatalf("Expected error from CAClientKeyPath. Got: %s", err.Error())
	}
}

// TestCreateValidBCCSPOptsForNewFabricClient test newFabricCA Client creation with valid inputs, successful scenario
func TestCreateValidBCCSPOptsForNewFabricClient(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)
	clientMockObject := &config.ClientConfig{Organization: "org1", Logging: config.LoggingType{Level: "info"}, CryptoConfig: config.CCType{Path: "test/path"}}

	mockConfig.EXPECT().CAConfig(org1).Return(&config.CAConfig{}, nil).AnyTimes()
	mockConfig.EXPECT().CredentialStorePath().Return(userStorePath)
	mockConfig.EXPECT().CAServerCertPaths(org1).Return([]string{"test"}, nil)
	mockConfig.EXPECT().CAClientCertPath(org1).Return("", nil)
	mockConfig.EXPECT().CAClientKeyPath(org1).Return("", nil)
	mockConfig.EXPECT().CAKeyStorePath().Return(os.TempDir())
	mockConfig.EXPECT().Client().Return(clientMockObject, nil)
	mockConfig.EXPECT().SecurityProvider().Return("SW")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return("/tmp/msp")
	mockConfig.EXPECT().Ephemeral().Return(false)

	newCryptosuiteProvider, err := cryptosuiteimpl.GetSuiteByConfig(mockConfig)

	if err != nil {
		t.Fatalf("Expected fabric client ryptosuite to be created with SW BCCS provider, but got %v", err.Error())
	}

	_, err = New(org1, mockConfig, newCryptosuiteProvider)
	if err != nil {
		t.Fatalf("Expected fabric client to be created with SW BCCS provider, but got %v", err.Error())
	}
}

// readCert Reads a random cert for testing
func readCert(t *testing.T) []byte {
	cert, err := ioutil.ReadFile("testdata/root.pem")
	if err != nil {
		t.Fatalf("Error reading cert: %s", err.Error())
	}
	return cert
}

// TestInterfaces will test if the interface instantiation happens properly, ie no nil returned
func TestInterfaces(t *testing.T) {
	var apiCA fab.FabricCAClient
	var ca FabricCA

	apiCA = &ca
	if apiCA == nil {
		t.Fatalf("this shouldn't happen.")
	}
}

func cleanup(storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to remove dir %s: %v\n", storePath, err))
	}
}
