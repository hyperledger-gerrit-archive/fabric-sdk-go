/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	logApi "github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
	"github.com/stretchr/testify/require"
)

var (
	m0 = &IdentityConfig{}
	m1 = &mockClient{}
	m2 = &mockCaConfig{}
	m3 = &mockCaServerCerts{}
	m4 = &mockCaClientKey{}
	m5 = &mockCaClientCert{}
	m6 = &mockCaKeyStorePath{}
	m7 = &mockCredentialStorePath{}
)

func TestCreateCustomFullIdentitytConfig(t *testing.T) {
	var opts []interface{}
	opts = append(opts, m0)
	// try to build with the overall interface (m0 is the overall interface's default implementation)
	identityConfigOption, err := BuildConfigIdentityFromOptions(opts...)
	if err != nil {
		t.Fatalf("BuildConfigIdentityFromOptions returned unexpected error %s", err)
	}
	if identityConfigOption == nil {
		t.Fatalf("BuildConfigIdentityFromOptions call returned nil")
	}
}

func TestCreateCustomIdentityConfig(t *testing.T) {
	// try to build with partial implementations
	identityConfigOption, err := BuildConfigIdentityFromOptions(m1, m2, m3, m4)
	if err != nil {
		t.Fatalf("BuildConfigIdentityFromOptions returned unexpected error %s", err)
	}
	var ico *IdentityConfigOptions
	var ok bool
	if ico, ok = identityConfigOption.(*IdentityConfigOptions); !ok {
		t.Fatalf("BuildConfigIdentityFromOptions did not return an Options instance %T", identityConfigOption)
	}
	require.NotNil(t, ico, "build ConfigIdentityOption returned is nil")

	// test m1 implementation
	clnt, err := ico.Client()
	require.NoError(t, err, "Client() function failed")
	require.NotEmpty(t, clnt, "client returned must not be empty")

	// test m2 implementation
	caCfg, err := ico.CAConfig("testORG")
	require.NoError(t, err, "CAConfig returned unexpected error")
	require.Equal(t, "test.url.com", caCfg.URL, "CAConfig did not return expected interface value")

	// test m3 implementation
	s, err := ico.CAServerCerts("testORG")
	require.NoError(t, err, "CAServerCerts returned unexpected error")
	require.Equal(t, []byte("testCAservercert1"), s[0], "CAServerCerts did not return the right cert")
	require.Equal(t, []byte("testCAservercert2"), s[1], "CAServerCerts did not return the right cert")

	// test m4 implementation
	c, err := ico.CAClientKey("testORG")
	require.NoError(t, err, "CAClientKey returned unexpected error")
	require.Equal(t, []byte("testCAclientkey"), c, "CAClientKey did not return the right cert")

	// verify if an interface was not passed as an option but was not nil, it should be nil (ie these implementations should not be populated in ico: m5, m6 and m7)
	require.Nil(t, ico.caClientCert, "caClientCert created with nil interface but got non nil one: %s. Expected nil interface", ico.caClientCert)
	require.Nil(t, ico.caKeyStorePath, "caKeyStorePath created with nil interface but got non nil one: %s. Expected nil interface", ico.caKeyStorePath)
	require.Nil(t, ico.credentialStorePath, "credentialStorePath created with nil interface but got non nil one: %s. Expected nil interface", ico.credentialStorePath)
}

func TestCreateCustomIdentityConfigRemainingFunctions(t *testing.T) {
	// try to build with the remaining implementations not tested above
	identityConfigOption, err := BuildConfigIdentityFromOptions(m5, m6, m7)
	if err != nil {
		t.Fatalf("BuildConfigIdentityFromOptions returned unexpected error %s", err)
	}
	var ico *IdentityConfigOptions
	var ok bool
	if ico, ok = identityConfigOption.(*IdentityConfigOptions); !ok {
		t.Fatalf("BuildConfigIdentityFromOptions did not return an Options instance %T", identityConfigOption)
	}
	require.NotNil(t, ico, "build ConfigIdentityOption returned is nil")

	// test m5 implementation
	c, err := ico.CAClientCert("")
	require.NoError(t, err, "CAClientCert returned unexpected error")
	require.Equal(t, []byte("testCAclientcert"), c, "CAClientCert did not return expected interface value")

	// test m6 implementation
	s := ico.CAKeyStorePath()
	require.Equal(t, "test/store/path", s, "CAKeyStorePath did not return expected interface value")

	// test m7 implementation
	s = ico.CredentialStorePath()
	require.Equal(t, "test/cred/store/path", s, "CredentialStorePath did not return expected interface value")

	// verify if an interface was not passed as an option but was not nil, it should be nil (ie these implementations should not be populated in ico: m1, m2, m3 and m4)
	require.Nil(t, ico.client, "client created with nil interface but got non nil one: %s. Expected nil interface", ico.client)
	require.Nil(t, ico.caConfig, "caConfig created with nil interface but got non nil one: %s. Expected nil interface", ico.caConfig)
	require.Nil(t, ico.caServerCerts, "caServerCerts created with nil interface but got non nil one: %s. Expected nil interface", ico.caServerCerts)
	require.Nil(t, ico.caClientKey, "caClientKey created with nil interface but got non nil one: %s. Expected nil interface", ico.caClientKey)

	// now try with a non related interface to test if an error returns
	var badType interface{}
	_, err = BuildConfigIdentityFromOptions(m4, m5, badType)
	require.Error(t, err, "BuildConfigIdentityFromOptions did not return error with badType")

}

func TestCreateCustomIdentityConfigWithSomeDefaultFunctions(t *testing.T) {
	// try to build with partial interfaces
	identityConfigOption, err := BuildConfigIdentityFromOptions(m1, m2, m3, m4)
	if err != nil {
		t.Fatalf("BuildConfigIdentityFromOptions returned unexpected error %s", err)
	}
	var ico *IdentityConfigOptions
	var ok bool
	if ico, ok = identityConfigOption.(*IdentityConfigOptions); !ok {
		t.Fatalf("BuildConfigIdentityFromOptions did not return an Options instance %T", identityConfigOption)
	}
	require.NotNil(t, ico, "build ConfigIdentityOption returned is nil")

	// now check if implementations that were not injected when building the config (ref first line in this function) are nil at this point
	// ie, verify these implementations should be nil: m5, m6 and m7
	require.Nil(t, ico.caClientCert, "caClientCert should be nil but got a non-nil one: %s. Expected nil interface", ico.caClientCert)
	require.Nil(t, ico.caKeyStorePath, "caKeyStorePath should be nil but got non-nil one: %s. Expected nil interface", ico.caKeyStorePath)
	require.Nil(t, ico.credentialStorePath, "credentialStorePath should be nil but got non-nil one: %s. Expected nil interface", ico.credentialStorePath)

	// do the same test using IsIdentityConfigFullyOverridden() call
	require.False(t, IsIdentityConfigFullyOverridden(ico), "IsIdentityConfigFullyOverridden is supposed to return false with an Options instance not implementing all the interface functions")

	// now inject default interfaces (using m0 as default full implementation for the sake of this test) for the ones that were not overridden by options above
	identityConfigOptionWithSomeDefaults := UpdateMissingOptsWithDefaultConfig(ico, m0)

	// test implementations m1-m4 are still working

	// test m1 implementation
	clnt, err := identityConfigOptionWithSomeDefaults.Client()
	require.NoError(t, err, "Client() function failed")
	require.NotEmpty(t, clnt, "client returned must not be empty")

	// test m2 implementation
	caCfg, err := identityConfigOptionWithSomeDefaults.CAConfig("testORG")
	require.NoError(t, err, "CAConfig returned unexpected error")
	require.Equal(t, "test.url.com", caCfg.URL, "CAConfig did not return expected interface value")

	// test m3 implementation
	s, err := identityConfigOptionWithSomeDefaults.CAServerCerts("testORG")
	require.NoError(t, err, "CAServerCerts returned unexpected error")
	require.Equal(t, []byte("testCAservercert1"), s[0], "CAServerCerts did not return the right cert")
	require.Equal(t, []byte("testCAservercert2"), s[1], "CAServerCerts did not return the right cert")

	// test m4 implementation
	c, err := identityConfigOptionWithSomeDefaults.CAClientKey("testORG")
	require.NoError(t, err, "CAClientKey returned unexpected error")
	require.Equal(t, []byte("testCAclientkey"), c, "CAClientKey did not return the right cert")

	if ico, ok = identityConfigOptionWithSomeDefaults.(*IdentityConfigOptions); !ok {
		t.Fatalf("UpdateMissingOptsWithDefaultConfig() call did not return an implementation of IdentityConfigOptions")
	}

	// now check if implementations that were not injected when building the config (ref first line in this function) are defaulted with m0 this time
	// ie, verify these implementations should now be populated in ico: m5, m6, m7
	require.NotNil(t, ico.caClientCert, "caClientCert should be populated with default interface but got nil one: %s. Expected default interface", ico.caClientCert)
	require.NotNil(t, ico.caKeyStorePath, "caKeyStorePath should be populated with default interface but got nil one: %s. Expected default interface", ico.caKeyStorePath)
	require.NotNil(t, ico.credentialStorePath, "credentialStorePath should be populated with default interface but got nil one: %s. Expected default interface", ico.credentialStorePath)

	// do the same test using IsIdentityConfigFullyOverridden() call
	require.True(t, IsIdentityConfigFullyOverridden(ico), "IsIdentityConfigFullyOverridden is supposed to return true since all the interface functions should be implemented")
}

type mockClient struct {
}

func (m *mockClient) Client() (*msp.ClientConfig, error) {
	return &msp.ClientConfig{
		CryptoConfig:    msp.CCType{Path: ""},
		CredentialStore: msp.CredentialStoreType{Path: "", CryptoStore: msp.CCType{Path: ""}},
		Logging:         logApi.LoggingType{Level: "INFO"},
		Organization:    "org1",
		TLSCerts:        endpoint.MutualTLSConfig{Path: "", Client: endpoint.TLSKeyPair{Cert: endpoint.TLSConfig{Path: ""}, Key: endpoint.TLSConfig{Path: ""}}},
	}, nil
}

type mockCaConfig struct{}

func (m *mockCaConfig) CAConfig(org string) (*msp.CAConfig, error) {
	return &msp.CAConfig{
		URL:        "test.url.com",
		Registrar:  msp.EnrollCredentials{EnrollSecret: "secret", EnrollID: ""},
		TLSCACerts: endpoint.MutualTLSConfig{Path: "", Client: endpoint.TLSKeyPair{Cert: endpoint.TLSConfig{Path: ""}, Key: endpoint.TLSConfig{Path: ""}}},
	}, nil
}

type mockCaServerCerts struct{}

func (m *mockCaServerCerts) CAServerCerts(org string) ([][]byte, error) {
	return [][]byte{[]byte("testCAservercert1"), []byte("testCAservercert2")}, nil
}

type mockCaClientKey struct{}

func (m *mockCaClientKey) CAClientKey(org string) ([]byte, error) {
	return []byte("testCAclientkey"), nil
}

type mockCaClientCert struct{}

func (m *mockCaClientCert) CAClientCert(org string) ([]byte, error) {
	return []byte("testCAclientcert"), nil
}

type mockCaKeyStorePath struct{}

func (m *mockCaKeyStorePath) CAKeyStorePath() string {
	return "test/store/path"
}

type mockCredentialStorePath struct{}

func (m *mockCredentialStorePath) CredentialStorePath() string {
	return "test/cred/store/path"
}
