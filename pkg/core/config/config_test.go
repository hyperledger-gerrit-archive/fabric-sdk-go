/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"reflect"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	api "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var configImpl *Config

const (
	org0                            = "org0"
	org1                            = "org1"
	configTestFilePath              = "testdata/config_test.yaml"
	configEmptyTestFilePath         = "testdata/empty.yaml"
	configPemTestFilePath           = "testdata/config_test_pem.yaml"
	configEmbeddedUsersTestFilePath = "testdata/config_test_embedded_pems.yaml"
	configType                      = "yaml"
	defaultConfigPath               = "testdata/template"
)

func TestCAConfig(t *testing.T) {
	//Test config
	vConfig := viper.New()
	vConfig.SetConfigFile(configTestFilePath)
	vConfig.ReadInConfig()
	vc := vConfig.ConfigFileUsed()

	if vc == "" {
		t.Fatalf("Failed to load config file")
	}

	//Test network name
	if vConfig.GetString("name") != "global-trade-network" {
		t.Fatalf("Incorrect Network name")
	}

	//Test network description
	if vConfig.GetString("description") != "The network to be in if you want to stay in the global trade business" {
		t.Fatalf("Incorrect Network description")
	}

	//Test network config version
	if vConfig.GetString("version") != "1.0.0" {
		t.Fatalf("Incorrect network version")
	}

	//Test client organization
	if vConfig.GetString("client.organization") != org1 {
		t.Fatalf("Incorrect Client organization")
	}

	//Test Crypto config path
	crossCheckWithViperConfig(configImpl.configViper.GetString("client.cryptoconfig.path"), configImpl.CryptoConfigPath(), "Incorrect crypto config path", t)

	//Testing CA Client File Location
	certfile, err := configImpl.CAClientCertPath(org1)

	if certfile == "" || err != nil {
		t.Fatalf("CA Cert file location read failed %s", err)
	}

	//Testing CA Key File Location
	keyFile, err := configImpl.CAClientKeyPath(org1)

	if keyFile == "" || err != nil {
		t.Fatal("CA Key file location read failed")
	}

	//Testing CA Server Cert Files
	sCertFiles, err := configImpl.CAServerCertPaths(org1)

	if sCertFiles == nil || len(sCertFiles) == 0 || err != nil {
		t.Fatal("Getting CA server cert files failed")
	}

	//Testing MSPID
	mspID, err := configImpl.MSPID(org1)
	if mspID != "Org1MSP" || err != nil {
		t.Fatal("Get MSP ID failed")
	}

	//Testing CAConfig
	caConfig, err := configImpl.CAConfig(org1)
	if caConfig == nil || err != nil {
		t.Fatal("Get CA Config failed")
	}

	// Test User Store Path
	if vConfig.GetString("client.credentialStore.path") != configImpl.CredentialStorePath() {
		t.Fatalf("Incorrect User Store path")
	}

	// Test CA KeyStore Path
	if vConfig.GetString("client.credentialStore.cryptoStore.path") != configImpl.CAKeyStorePath() {
		t.Fatalf("Incorrect CA keystore path")
	}

	// Test KeyStore Path
	if path.Join(vConfig.GetString("client.credentialStore.cryptoStore.path"), "keystore") != configImpl.KeyStorePath() {
		t.Fatalf("Incorrect keystore path ")
	}

	// Test BCCSP security is enabled
	if vConfig.GetBool("client.BCCSP.security.enabled") != configImpl.IsSecurityEnabled() {
		t.Fatalf("Incorrect BCCSP Security enabled flag")
	}

	// Test SecurityAlgorithm
	if vConfig.GetString("client.BCCSP.security.hashAlgorithm") != configImpl.SecurityAlgorithm() {
		t.Fatalf("Incorrect BCCSP Security Hash algorithm")
	}

	// Test Security Level
	if vConfig.GetInt("client.BCCSP.security.level") != configImpl.SecurityLevel() {
		t.Fatalf("Incorrect BCCSP Security Level")
	}

	// Test SecurityProvider provider
	if vConfig.GetString("client.BCCSP.security.default.provider") != configImpl.SecurityProvider() {
		t.Fatalf("Incorrect BCCSP SecurityProvider provider")
	}

	// Test Ephemeral flag
	if vConfig.GetBool("client.BCCSP.security.ephemeral") != configImpl.Ephemeral() {
		t.Fatalf("Incorrect BCCSP Ephemeral flag")
	}

	// Test SoftVerify flag
	if vConfig.GetBool("client.BCCSP.security.softVerify") != configImpl.SoftVerify() {
		t.Fatalf("Incorrect BCCSP Ephemeral flag")
	}

	// Test SecurityProviderPin
	if vConfig.GetString("client.BCCSP.security.pin") != configImpl.SecurityProviderPin() {
		t.Fatalf("Incorrect BCCSP SecurityProviderPin flag")
	}

	// Test SecurityProviderPin
	if vConfig.GetString("client.BCCSP.security.label") != configImpl.SecurityProviderLabel() {
		t.Fatalf("Incorrect BCCSP SecurityProviderPin flag")
	}

	// test Client
	c, err := configImpl.Client()
	if err != nil {
		t.Fatalf("Received error when fetching Client info, error is %s", err)
	}
	if c == nil {
		t.Fatal("Received empty client when fetching Client info")
	}

	// testing empty OrgMSP
	mspID, err = configImpl.MSPID("dummyorg1")
	if err == nil {
		t.Fatal("Get MSP ID did not fail for dummyorg1")
	}
}

func TestCAConfigFailsByNetworkConfig(t *testing.T) {

	//Tamper 'client.network' value and use a new config to avoid conflicting with other tests
	configProvider, err := FromFile(configTestFilePath)()
	if err != nil {
		t.Fatalf("Unexpected error reading config: %v", err)
	}
	sampleConfig := configProvider.(*Config)

	sampleConfig.networkConfigCached = false
	sampleConfig.configViper.Set("client", "INVALID")
	sampleConfig.configViper.Set("peers", "INVALID")
	sampleConfig.configViper.Set("organizations", "INVALID")
	sampleConfig.configViper.Set("orderers", "INVALID")
	sampleConfig.configViper.Set("channels", "INVALID")

	_, err = sampleConfig.NetworkConfig()
	if err == nil {
		t.Fatal("Network config load supposed to fail")
	}

	//Test CA client cert file failure scenario
	certfile, err := sampleConfig.CAClientCertPath("peerorg1")
	if certfile != "" || err == nil {
		t.Fatal("CA Cert file location read supposed to fail")
	}

	//Test CA client cert file failure scenario
	keyFile, err := sampleConfig.CAClientKeyPath("peerorg1")
	if keyFile != "" || err == nil {
		t.Fatal("CA Key file location read supposed to fail")
	}

	//Testing CA Server Cert Files failure scenario
	sCertFiles, err := sampleConfig.CAServerCertPaths("peerorg1")
	if len(sCertFiles) > 0 || err == nil {
		t.Fatal("Getting CA server cert files supposed to fail")
	}

	//Testing MSPID failure scenario
	mspID, err := sampleConfig.MSPID("peerorg1")
	if mspID != "" || err == nil {
		t.Fatal("Get MSP ID supposed to fail")
	}

	//Testing CAConfig failure scenario
	caConfig, err := sampleConfig.CAConfig("peerorg1")
	if caConfig != nil || err == nil {
		t.Fatal("Get CA Config supposed to fail")
	}

	//Testing RandomOrdererConfig failure scenario
	oConfig, err := sampleConfig.RandomOrdererConfig()
	if oConfig != nil || err == nil {
		t.Fatal("Testing get RandomOrdererConfig supposed to fail")
	}

	//Testing RandomOrdererConfig failure scenario
	oConfig, err = sampleConfig.OrdererConfig("peerorg1")
	if oConfig != nil || err == nil {
		t.Fatal("Testing get OrdererConfig supposed to fail")
	}

	//Testing PeersConfig failure scenario
	pConfigs, err := sampleConfig.PeersConfig("peerorg1")
	if pConfigs != nil || err == nil {
		t.Fatal("Testing PeersConfig supposed to fail")
	}

	//Testing PeersConfig failure scenario
	pConfig, err := sampleConfig.PeerConfig("peerorg1", "peer1")
	if pConfig != nil || err == nil {
		t.Fatal("Testing PeerConfig supposed to fail")
	}

	//Testing ChannelConfig failure scenario
	chConfig, err := sampleConfig.ChannelConfig("invalid")
	if chConfig != nil || err == nil {
		t.Fatal("Testing ChannelConfig supposed to fail")
	}

	//Testing ChannelPeers failure scenario
	cpConfigs, err := sampleConfig.ChannelPeers("invalid")
	if cpConfigs != nil || err == nil {
		t.Fatal("Testing ChannelPeeers supposed to fail")
	}

	//Testing ChannelOrderers failure scenario
	coConfigs, err := sampleConfig.ChannelOrderers("invalid")
	if coConfigs != nil || err == nil {
		t.Fatal("Testing ChannelOrderers supposed to fail")
	}

	// test empty network objects
	sampleConfig.configViper.Set("organizations", nil)
	_, err = sampleConfig.NetworkConfig()
	if err == nil {
		t.Fatalf("Organizations were empty, it should return an error")
	}
}

func TestTLSCAConfig(t *testing.T) {
	//Test TLSCA Cert Pool (Positive test case)

	certFile, _ := configImpl.CAClientCertPath(org1)
	certConfig := endpoint.TLSConfig{Path: certFile}

	cert, err := certConfig.TLSCert()
	if err != nil {
		t.Fatalf("Failed to get TLS CA Cert, reason: %v", err)
	}

	_, err = configImpl.TLSCACertPool(cert)
	if err != nil {
		t.Fatalf("TLS CA cert pool fetch failed, reason: %v", err)
	}

	//Try again with same cert
	_, err = configImpl.TLSCACertPool(cert)
	if err != nil {
		t.Fatalf("TLS CA cert pool fetch failed, reason: %v", err)
	}

	assert.False(t, len(configImpl.tlsCerts) > 1, "number of certs in cert list shouldn't accept duplicates")

	//Test TLSCA Cert Pool (Negative test case)

	badCertConfig := endpoint.TLSConfig{Path: "some random invalid path"}

	badCert, err := badCertConfig.TLSCert()

	if err == nil {
		t.Fatalf("TLS CA cert pool was supposed to fail")
	}

	_, err = configImpl.TLSCACertPool(badCert)

	keyFile, _ := configImpl.CAClientKeyPath(org1)

	keyConfig := endpoint.TLSConfig{Path: keyFile}

	key, err := keyConfig.TLSCert()

	if err == nil {
		t.Fatalf("TLS CA cert pool was supposed to fail when provided with wrong cert file")
	}

	_, err = configImpl.TLSCACertPool(key)
}

func TestTLSCAConfigFromPems(t *testing.T) {
	c, err := FromFile(configEmbeddedUsersTestFilePath)()
	if err != nil {
		t.Fatal(err)
	}

	//Test TLSCA Cert Pool (Positive test case)

	certPem, _ := c.CAClientCertPem(org1)
	certConfig := endpoint.TLSConfig{Pem: certPem}

	cert, err := certConfig.TLSCert()

	if err != nil {
		t.Fatalf("TLS CA cert parse failed, reason: %v", err)
	}

	_, err = configImpl.TLSCACertPool(cert)

	if err != nil {
		t.Fatalf("TLS CA cert pool fetch failed, reason: %v", err)
	}
	//Test TLSCA Cert Pool (Negative test case)

	badCertConfig := endpoint.TLSConfig{Pem: "some random invalid pem"}

	badCert, err := badCertConfig.TLSCert()

	if err == nil {
		t.Fatalf("TLS CA cert parse was supposed to fail")
	}

	_, err = configImpl.TLSCACertPool(badCert)

	keyPem, _ := configImpl.CAClientKeyPem(org1)

	keyConfig := endpoint.TLSConfig{Pem: keyPem}

	key, err := keyConfig.TLSCert()

	if err == nil {
		t.Fatalf("TLS CA cert pool was supposed to fail when provided with wrong cert file")
	}

	_, err = configImpl.TLSCACertPool(key)
}

func TestTimeouts(t *testing.T) {
	configImpl.configViper.Set("client.peer.timeout.connection", "2s")
	configImpl.configViper.Set("client.peer.timeout.response", "6s")
	configImpl.configViper.Set("client.eventService.timeout.connection", "2m")
	configImpl.configViper.Set("client.eventService.timeout.registrationResponse", "2h")
	configImpl.configViper.Set("client.orderer.timeout.connection", "2ms")
	configImpl.configViper.Set("client.global.timeout.query", "7h")
	configImpl.configViper.Set("client.global.timeout.execute", "8h")
	configImpl.configViper.Set("client.global.timeout.resmgmt", "118s")
	configImpl.configViper.Set("client.global.cache.connectionIdle", "1m")
	configImpl.configViper.Set("client.global.cache.eventServiceIdle", "2m")
	configImpl.configViper.Set("client.orderer.timeout.response", "6s")

	t1 := configImpl.TimeoutOrDefault(api.EndorserConnection)
	if t1 != time.Second*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.EventHubConnection)
	if t1 != time.Minute*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.EventReg)
	if t1 != time.Hour*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.Query)
	if t1 != time.Hour*7 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.Execute)
	if t1 != time.Hour*8 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.OrdererConnection)
	if t1 != time.Millisecond*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.OrdererResponse)
	if t1 != time.Second*6 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.ConnectionIdle)
	if t1 != time.Minute*1 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.EventServiceIdle)
	if t1 != time.Minute*2 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.PeerResponse)
	if t1 != time.Second*6 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}
	t1 = configImpl.TimeoutOrDefault(api.ResMgmt)
	if t1 != time.Second*118 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}

	// Test default
	configImpl.configViper.Set("client.orderer.timeout.connection", "")
	t1 = configImpl.TimeoutOrDefault(api.OrdererConnection)
	if t1 != time.Second*5 {
		t.Fatalf("Timeout not read correctly. Got: %s", t1)
	}

}

func TestOrdererConfig(t *testing.T) {
	oConfig, err := configImpl.RandomOrdererConfig()

	if oConfig == nil || err != nil {
		t.Fatal("Testing get RandomOrdererConfig failed")
	}

	oConfig, err = configImpl.OrdererConfig("invalid")

	if oConfig != nil || err == nil {
		t.Fatal("Testing non-existing OrdererConfig failed")
	}

	orderers, err := configImpl.OrderersConfig()
	if err != nil {
		t.Fatal(err)
	}

	if orderers[0].TLSCACerts.Path != "" {
		if !filepath.IsAbs(orderers[0].TLSCACerts.Path) {
			t.Fatal("Expected GOPATH relative path to be replaced")
		}
	} else if len(orderers[0].TLSCACerts.Pem) == 0 {
		t.Fatalf("Orderer %v must have at least a TlsCACerts.Path or TlsCACerts.Pem set", orderers[0])
	}
}

func TestChannelOrderers(t *testing.T) {
	orderers, err := configImpl.ChannelOrderers("mychannel")
	if orderers == nil || err != nil {
		t.Fatal("Testing ChannelOrderers failed")
	}

	if len(orderers) != 1 {
		t.Fatalf("Expecting one channel orderer got %d", len(orderers))
	}

	if orderers[0].TLSCACerts.Path != "" {
		if !filepath.IsAbs(orderers[0].TLSCACerts.Path) {
			t.Fatal("Expected GOPATH relative path to be replaced")
		}
	} else if len(orderers[0].TLSCACerts.Pem) == 0 {
		t.Fatalf("Orderer %v must have at least a TlsCACerts.Path or TlsCACerts.Pem set", orderers[0])
	}
}

func testCommonConfigPeerByURL(t *testing.T, expectedConfigURL string, fetchedConfigURL string) {
	expectedConfig, err := configImpl.peerConfig(expectedConfigURL)
	if err != nil {
		t.Fatalf(err.Error())
	}

	fetchedConfig, err := configImpl.PeerConfigByURL(fetchedConfigURL)

	if fetchedConfig.URL == "" {
		t.Fatalf("Url value for the host is empty")
	}

	if len(fetchedConfig.GRPCOptions) != len(expectedConfig.GRPCOptions) || fetchedConfig.TLSCACerts.Pem != expectedConfig.TLSCACerts.Pem {
		t.Fatalf("Expected Config and fetched config differ")
	}

	if fetchedConfig.URL != expectedConfig.URL || fetchedConfig.EventURL != expectedConfig.EventURL || fetchedConfig.GRPCOptions["ssl-target-name-override"] != expectedConfig.GRPCOptions["ssl-target-name-override"] {
		t.Fatalf("Expected Config and fetched config differ")
	}
}

func TestPeerConfigByUrl_directMatching(t *testing.T) {
	testCommonConfigPeerByURL(t, "peer0.org1.example.com", "peer0.org1.example.com:7051")
}

func TestPeerConfigByUrl_entityMatchers(t *testing.T) {
	testCommonConfigPeerByURL(t, "peer0.org1.example.com", "peer1.org1.example.com:7051")
}

func testCommonConfigOrderer(t *testing.T, expectedConfigHost string, fetchedConfigHost string) (expectedConfig *api.OrdererConfig, fetchedConfig *api.OrdererConfig) {

	expectedConfig, err := configImpl.OrdererConfig(expectedConfigHost)
	if err != nil {
		t.Fatalf(err.Error())
	}

	fetchedConfig, err = configImpl.OrdererConfig(fetchedConfigHost)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if expectedConfig.URL == "" {
		t.Fatalf("Url value for the host is empty")
	}
	if fetchedConfig.URL == "" {
		t.Fatalf("Url value for the host is empty")
	}

	if len(fetchedConfig.GRPCOptions) != len(expectedConfig.GRPCOptions) || fetchedConfig.TLSCACerts.Pem != expectedConfig.TLSCACerts.Pem {
		t.Fatalf("Expected Config and fetched config differ")
	}

	return expectedConfig, fetchedConfig
}

func TestOrdererWithSubstitutedConfig_WithADifferentSubstituteUrl(t *testing.T) {
	expectedConfig, fetchedConfig := testCommonConfigOrderer(t, "orderer.example.com", "orderer.example2.com")

	if fetchedConfig.URL == "orderer.example2.com:7050" || fetchedConfig.URL == expectedConfig.URL {
		t.Fatalf("Expected Config should have url that is given in urlSubstitutionExp of match pattern")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "localhost" {
		t.Fatalf("Config should have got localhost as its ssl-target-name-override url as per the matched config")
	}
}

func TestOrdererWithSubstitutedConfig_WithEmptySubstituteUrl(t *testing.T) {
	_, fetchedConfig := testCommonConfigOrderer(t, "orderer.example.com", "orderer.example3.com")

	if fetchedConfig.URL != "orderer.example3.com:7050" {
		t.Fatalf("Fetched Config should have the same url")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "orderer.example3.com" {
		t.Fatalf("Fetched config should have the same ssl-target-name-override as its hostname")
	}
}

func TestOrdererWithSubstitutedConfig_WithSubstituteUrlExpression(t *testing.T) {
	expectedConfig, fetchedConfig := testCommonConfigOrderer(t, "orderer.example.com", "orderer.example4.com:7050")

	if fetchedConfig.URL != expectedConfig.URL {
		t.Fatalf("fetched Config url should be same as expected config url as given in the substituteexp in yaml file")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "orderer.example.com" {
		t.Fatalf("Fetched config should have the ssl-target-name-override as per sslTargetOverrideUrlSubstitutionExp in yaml file")
	}
}

func TestPeersConfig(t *testing.T) {
	pc, err := configImpl.PeersConfig(org0)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for _, value := range pc {
		if value.URL == "" {
			t.Fatalf("Url value for the host is empty")
		}
		if value.EventURL == "" {
			t.Fatalf("EventUrl value is empty")
		}
	}

	pc, err = configImpl.PeersConfig(org1)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for _, value := range pc {
		if value.URL == "" {
			t.Fatalf("Url value for the host is empty")
		}
		if value.EventURL == "" {
			t.Fatalf("EventUrl value is empty")
		}
	}
}

func TestPeerConfig(t *testing.T) {
	pc, err := configImpl.PeerConfig(org1, "peer0.org1.example.com")
	if err != nil {
		t.Fatalf(err.Error())
	}

	if pc.URL == "" {
		t.Fatalf("Url value for the host is empty")
	}

	if pc.TLSCACerts.Path != "" {
		if !filepath.IsAbs(pc.TLSCACerts.Path) {
			t.Fatalf("Expected cert path to be absolute")
		}
	} else if len(pc.TLSCACerts.Pem) == 0 {
		t.Fatalf("Peer %s must have at least a TlsCACerts.Path or TlsCACerts.Pem set", "peer0")
	}
	if len(pc.GRPCOptions) == 0 || pc.GRPCOptions["ssl-target-name-override"] != "peer0.org1.example.com" {
		t.Fatalf("Peer %s must have grpcOptions set in config_test.yaml", "peer0")
	}
}

func testCommonConfigPeer(t *testing.T, expectedConfigHost string, fetchedConfigHost string) (expectedConfig *api.PeerConfig, fetchedConfig *api.PeerConfig) {

	expectedConfig, err := configImpl.peerConfig(expectedConfigHost)
	if err != nil {
		t.Fatalf(err.Error())
	}

	fetchedConfig, err = configImpl.peerConfig(fetchedConfigHost)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if expectedConfig.URL == "" {
		t.Fatalf("Url value for the host is empty")
	}
	if fetchedConfig.URL == "" {
		t.Fatalf("Url value for the host is empty")
	}

	if fetchedConfig.TLSCACerts.Path != expectedConfig.TLSCACerts.Path || len(fetchedConfig.GRPCOptions) != len(expectedConfig.GRPCOptions) {
		t.Fatalf("Expected Config and fetched config differ")
	}

	return expectedConfig, fetchedConfig
}

func TestPeerWithSubstitutedConfig_WithADifferentSubstituteUrl(t *testing.T) {
	expectedConfig, fetchedConfig := testCommonConfigPeer(t, "peer0.org1.example.com", "peer3.org1.example5.com")

	if fetchedConfig.URL == "peer3.org1.example5.com:7051" || fetchedConfig.URL == expectedConfig.URL {
		t.Fatalf("Expected Config should have url that is given in urlSubstitutionExp of match pattern")
	}

	if fetchedConfig.EventURL == "peer3.org1.example5.com:7053" || fetchedConfig.EventURL == expectedConfig.EventURL {
		t.Fatalf("Expected Config should have event url that is given in eventUrlSubstitutionExp of match pattern")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "localhost" {
		t.Fatalf("Config should have got localhost as its ssl-target-name-override url as per the matched config")
	}
}

func TestPeerWithSubstitutedConfig_WithEmptySubstituteUrl(t *testing.T) {
	_, fetchedConfig := testCommonConfigPeer(t, "peer0.org1.example.com", "peer4.org1.example3.com")

	if fetchedConfig.URL != "peer4.org1.example3.com:7051" {
		t.Fatalf("Fetched Config should have the same url")
	}

	if fetchedConfig.EventURL != "peer4.org1.example3.com:7053" {
		t.Fatalf("Fetched Config should have the same event url")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "peer4.org1.example3.com" {
		t.Fatalf("Fetched config should have the same ssl-target-name-override as its hostname")
	}
}

func TestPeerWithSubstitutedConfig_WithSubstituteUrlExpression(t *testing.T) {
	_, fetchedConfig := testCommonConfigPeer(t, "peer0.org1.example.com", "peer5.example4.com:1234")

	if fetchedConfig.URL != "peer5.org1.example.com:1234" {
		t.Fatalf("fetched Config url should change to include org1 as given in the substituteexp in yaml file")
	}

	if fetchedConfig.EventURL != "peer5.org1.example.com:7053" {
		t.Fatalf("fetched Config event url should change to include org1 as given in the eventsubstituteexp in yaml file")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "peer5.org1.example.com" {
		t.Fatalf("Fetched config should have the ssl-target-name-override as per sslTargetOverrideUrlSubstitutionExp in yaml file")
	}
}

func TestPeerWithSubstitutedConfig_WithMultipleMatchings(t *testing.T) {
	_, fetchedConfig := testCommonConfigPeer(t, "peer0.org2.example.com", "peer2.example2.com:1234")

	//Both 2nd and 5th entityMatchers match, however we are only taking 2nd one as its the first one to match
	if fetchedConfig.URL == "peer0.org2.example.com:7051" {
		t.Fatalf("fetched Config url should be matched with the first suitable matcher")
	}

	if fetchedConfig.EventURL != "localhost:7053" {
		t.Fatalf("fetched Config event url should have the config from first suitable matcher")
	}

	if fetchedConfig.GRPCOptions["ssl-target-name-override"] != "localhost" {
		t.Fatalf("Fetched config should have the ssl-target-name-override as per first suitable matcher in yaml file")
	}
}

func TestPeerNotInOrgConfig(t *testing.T) {
	_, err := configImpl.PeerConfig(org1, "peer1.org0.example.com")
	if err == nil {
		t.Fatalf("Fetching peer config not for an unassigned org should fail")
	}
}

func TestFromRawSuccess(t *testing.T) {
	// get a config byte for testing
	cBytes, err := loadConfigBytesFromFile(t, configTestFilePath)

	// test init config from bytes
	_, err = FromRaw(cBytes, configType)()
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}
}

func TestFromReaderSuccess(t *testing.T) {
	// get a config byte for testing
	cBytes, err := loadConfigBytesFromFile(t, configTestFilePath)
	buf := bytes.NewBuffer(cBytes)

	// test init config from bytes
	_, err = FromReader(buf, configType)()
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}
}

func TestFromFileEmptyFilename(t *testing.T) {
	_, err := FromFile("")()
	if err == nil {
		t.Fatalf("Expected error when passing empty string to FromFile")
	}
}

func loadConfigBytesFromFile(t *testing.T, filePath string) ([]byte, error) {
	// read test config file into bytes array
	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to read config file. Error: %s", err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		t.Fatalf("Failed to read config file stat. Error: %s", err)
	}
	s := fi.Size()
	cBytes := make([]byte, s, s)
	n, err := f.Read(cBytes)
	if err != nil {
		t.Fatalf("Failed to read test config for bytes array testing. Error: %s", err)
	}
	if n == 0 {
		t.Fatalf("Failed to read test config for bytes array testing. Mock bytes array is empty")
	}
	return cBytes, err
}

func TestInitConfigSuccess(t *testing.T) {
	//Test init config
	//...Positive case
	_, err := FromFile(configTestFilePath)()
	if err != nil {
		t.Fatalf("Failed to initialize config. Error: %s", err)
	}
}

func TestInitConfigWithCmdRoot(t *testing.T) {
	TestInitConfigSuccess(t)
	fileLoc := configTestFilePath
	cmdRoot := "fabric_sdk"
	var logger = logging.NewLogger("fabsdk/core")
	logger.Infof("fileLoc is %s", fileLoc)

	logger.Infof("fileLoc right before calling InitConfigWithCmdRoot is %s", fileLoc)
	configProvider, err := FromFile(fileLoc, WithEnvPrefix(cmdRoot))()
	if err != nil {
		t.Fatalf("Failed to initialize config with cmd root. Error: %s", err)
	}

	config := configProvider.(*Config)

	//Test if Viper is initialized after calling init config
	if config.configViper.GetString("client.BCCSP.security.hashAlgorithm") != configImpl.SecurityAlgorithm() {
		t.Fatal("Config initialized with incorrect viper configuration")
	}

}

func TestInitConfigPanic(t *testing.T) {

	os.Setenv("FABRIC_SDK_CLIENT_LOGGING_LEVEL", "INVALID")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Init config with cmdroot was supposed to panic")
		} else {
			//Setting it back during panic so as not to fail other tests
			os.Unsetenv("FABRIC_SDK_CLIENT_LOGGING_LEVEL")
		}
	}()

	FromFile(configTestFilePath)()
}

func TestInitConfigInvalidLocation(t *testing.T) {
	//...Negative case
	_, err := FromFile("invalid file location")()
	if err == nil {
		t.Fatalf("Config file initialization is supposed to fail. Error: %s", err)
	}
}

// Test case to create a new viper instance to prevent conflict with existing
// viper instances in applications that use the SDK
func TestMultipleVipers(t *testing.T) {
	viper.SetConfigFile("./test.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		t.Log(err.Error())
	}
	testValue1 := viper.GetString("test.testkey")
	// Read initial value from test.yaml
	if testValue1 != "testvalue" {
		t.Fatalf("Expected testValue before config initialization got: %s", testValue1)
	}
	// initialize go sdk
	configProvider, err := FromFile(configTestFilePath)()
	if err != nil {
		t.Log(err.Error())
	}

	config := configProvider.(*Config)

	// Make sure initial value is unaffected
	testValue2 := viper.GetString("test.testkey")
	if testValue2 != "testvalue" {
		t.Fatalf("Expected testvalue after config initialization")
	}
	// Make sure Go SDK config is unaffected
	testValue3 := config.configViper.GetBool("client.BCCSP.security.softVerify")
	if testValue3 != true {
		t.Fatalf("Expected existing config value to remain unchanged")
	}
}

func TestEnvironmentVariablesDefaultCmdRoot(t *testing.T) {
	testValue := configImpl.configViper.GetString("env.test")
	if testValue != "" {
		t.Fatalf("Expected environment variable value to be empty but got: %s", testValue)
	}

	err := os.Setenv("FABRIC_SDK_ENV_TEST", "123")
	defer os.Unsetenv("FABRIC_SDK_ENV_TEST")

	if err != nil {
		t.Log(err.Error())
	}

	testValue = configImpl.configViper.GetString("env.test")
	if testValue != "123" {
		t.Fatalf("Expected environment variable value but got: %s", testValue)
	}
}

func TestEnvironmentVariablesSpecificCmdRoot(t *testing.T) {
	testValue := configImpl.configViper.GetString("env.test")
	if testValue != "" {
		t.Fatalf("Expected environment variable value to be empty but got: %s", testValue)
	}

	err := os.Setenv("TEST_ROOT_ENV_TEST", "456")
	defer os.Unsetenv("TEST_ROOT_ENV_TEST")

	if err != nil {
		t.Log(err.Error())
	}

	configProvider, err := FromFile(configTestFilePath, WithEnvPrefix("test_root"))()
	if err != nil {
		t.Log(err.Error())
	}

	config := configProvider.(*Config)
	testValue = config.configViper.GetString("env.test")
	if testValue != "456" {
		t.Fatalf("Expected environment variable value but got: %s", testValue)
	}
}

func TestNetworkConfig(t *testing.T) {
	conf, err := configImpl.NetworkConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(conf.Orderers) == 0 {
		t.Fatal("Expected orderers to be set")
	}
	if len(conf.Organizations) == 0 {
		t.Fatal("Expected atleast one organisation to be set")
	}
	// viper map keys are lowercase
	if len(conf.Organizations[strings.ToLower(org1)].Peers) == 0 {
		t.Fatalf("Expected org %s to be present in network configuration and peers to be set", org1)
	}
}

func TestMain(m *testing.M) {
	setUp(m)
	r := m.Run()
	teardown()
	os.Exit(r)
}

func setUp(m *testing.M) {
	// do any test setup here...
	var err error
	configProvider, err := FromFile(configTestFilePath)()
	if err != nil {
		fmt.Println(err.Error())
	}
	configImpl = configProvider.(*Config)
}

func teardown() {
	// do any teadown activities here ..
	configImpl = nil
}

func crossCheckWithViperConfig(expected string, actual string, message string, t *testing.T) {
	expected = SubstPathVars(expected)
	if actual != expected {
		t.Fatalf(message)
	}
}

func TestInterfaces(t *testing.T) {
	var apiConfig api.Config
	var config Config

	apiConfig = &config
	if apiConfig == nil {
		t.Fatalf("this shouldn't happen. Config should not be nil.")
	}
}

func TestSystemCertPoolDisabled(t *testing.T) {

	// get a config file with pool disabled
	configProvider, err := FromFile(configTestFilePath)()
	if err != nil {
		t.Fatal(err)
	}

	certPool, err := configProvider.TLSCACertPool()
	if err != nil {
		t.Fatal("not supposed to get error")
	}
	// cert pool should be empty
	if len(certPool.Subjects()) > 0 {
		t.Fatal("Expecting empty tls cert pool due to disabled system cert pool")
	}
}

func TestSystemCertPoolEnabled(t *testing.T) {

	// get a config file with pool enabled
	configProvider, err := FromFile(configPemTestFilePath)()
	if err != nil {
		t.Fatal(err)
	}

	certPool, err := configProvider.TLSCACertPool()
	if err != nil {
		t.Fatal("not supposed to get error")
	}

	if len(certPool.Subjects()) == 0 {
		t.Fatal("System Cert Pool not loaded even though it is enabled")
	}

	// Org2 'mychannel' peer is missing cert + pem (it should not fail when systemCertPool enabled)
	_, err = configProvider.ChannelPeers("mychannel")
	if err != nil {
		t.Fatalf("Should have skipped verifying ca cert + pem: %s", err)
	}

}

func TestInitConfigFromRawWithPem(t *testing.T) {
	// get a config byte for testing
	cBytes, err := loadConfigBytesFromFile(t, configPemTestFilePath)
	if err != nil {
		t.Fatalf("Failed to load sample bytes from File. Error: %s", err)
	}

	// test init config from bytes
	c, err := FromRaw(cBytes, configType)()
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}

	o, err := c.OrderersConfig()
	if err != nil {
		t.Fatalf("Failed to load orderers from config. Error: %s", err)
	}

	if o == nil || len(o) == 0 {
		t.Fatalf("orderer cannot be nil or empty")
	}

	oPem := `-----BEGIN CERTIFICATE-----
MIICNjCCAdygAwIBAgIRAILSPmMB3BzoLIQGsFxwZr8wCgYIKoZIzj0EAwIwbDEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xFDASBgNVBAoTC2V4YW1wbGUuY29tMRowGAYDVQQDExF0bHNjYS5l
eGFtcGxlLmNvbTAeFw0xNzA3MjgxNDI3MjBaFw0yNzA3MjYxNDI3MjBaMGwxCzAJ
BgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1TYW4gRnJh
bmNpc2NvMRQwEgYDVQQKEwtleGFtcGxlLmNvbTEaMBgGA1UEAxMRdGxzY2EuZXhh
bXBsZS5jb20wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQfgKb4db53odNzdMXn
P5FZTZTFztOO1yLvCHDofSNfTPq/guw+YYk7ZNmhlhj8JHFG6dTybc9Qb/HOh9hh
gYpXo18wXTAOBgNVHQ8BAf8EBAMCAaYwDwYDVR0lBAgwBgYEVR0lADAPBgNVHRMB
Af8EBTADAQH/MCkGA1UdDgQiBCBxaEP3nVHQx4r7tC+WO//vrPRM1t86SKN0s6XB
8LWbHTAKBggqhkjOPQQDAgNIADBFAiEA96HXwCsuMr7tti8lpcv1oVnXg0FlTxR/
SQtE5YgdxkUCIHReNWh/pluHTxeGu2jNCH1eh6o2ajSGeeizoapvdJbN
-----END CERTIFICATE-----`
	loadedOPem := strings.TrimSpace(o[0].TLSCACerts.Pem) // viper's unmarshall adds a \n to the end of a string, hence the TrimeSpace
	if loadedOPem != oPem {
		t.Fatalf("Orderer Pem doesn't match. Expected \n'%s'\n, but got \n'%s'\n", oPem, loadedOPem)
	}

	pc, err := configImpl.PeersConfig(org1)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if pc == nil || len(pc) == 0 {
		t.Fatalf("peers list of %s cannot be nil or empty", org1)
	}
	peer0 := "peer0.org1.example.com"
	p0, err := c.PeerConfig(org1, peer0)
	if err != nil {
		t.Fatalf("Failed to load %s of %s from the config. Error: %s", peer0, org1, err)
	}
	if p0 == nil {
		t.Fatalf("%s of %s cannot be nil", peer0, org1)
	}
	pPem := `-----BEGIN CERTIFICATE-----
MIICSTCCAfCgAwIBAgIRAPQIzfkrCZjcpGwVhMSKd0AwCgYIKoZIzj0EAwIwdjEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHzAdBgNVBAMTFnRs
c2NhLm9yZzEuZXhhbXBsZS5jb20wHhcNMTcwNzI4MTQyNzIwWhcNMjcwNzI2MTQy
NzIwWjB2MQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UE
BxMNU2FuIEZyYW5jaXNjbzEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTEfMB0G
A1UEAxMWdGxzY2Eub3JnMS5leGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqGSM49
AwEHA0IABMOiG8UplWTs898zZ99+PhDHPbKjZIDHVG+zQXopw8SqNdX3NAmZUKUU
sJ8JZ3M49Jq4Ms8EHSEwQf0Ifx3ICHujXzBdMA4GA1UdDwEB/wQEAwIBpjAPBgNV
HSUECDAGBgRVHSUAMA8GA1UdEwEB/wQFMAMBAf8wKQYDVR0OBCIEID9qJz7xhZko
V842OVjxCYYQwCjPIY+5e9ORR+8pxVzcMAoGCCqGSM49BAMCA0cAMEQCIGZ+KTfS
eezqv0ml1VeQEmnAEt5sJ2RJA58+LegUYMd6AiAfEe6BKqdY03qFUgEYmtKG+3Dr
O94CDp7l2k7hMQI0zQ==
-----END CERTIFICATE-----`

	loadedPPem := strings.TrimSpace(p0.TLSCACerts.Pem) // viper's unmarshall adds a \n to the end of a string, hence the TrimeSpace
	if loadedPPem != pPem {
		t.Fatalf("%s Pem doesn't match. Expected \n'%s'\n, but got \n'%s'\n", peer0, pPem, loadedPPem)
	}

	// get CA Server cert pems (embedded) for org1
	certs, err := c.CAServerCertPems("org1")
	if err != nil {
		t.Fatalf("Failed to load CAServerCertPems from config. Error: %s", err)
	}
	if len(certs) == 0 {
		t.Fatalf("Got empty PEM certs for CAServerCertPems")
	}

	// get the client cert pem (embedded) for org1
	c.CAClientCertPem("org1")
	if err != nil {
		t.Fatalf("Failed to load CAClientCertPem from config. Error: %s", err)
	}

	// get CA Server certs paths for org1
	certs, err = c.CAServerCertPaths("org1")
	if err != nil {
		t.Fatalf("Failed to load CAServerCertPaths from config. Error: %s", err)
	}
	if len(certs) == 0 {
		t.Fatalf("Got empty cert file paths for CAServerCertPaths")
	}

	// get the client cert path for org1
	c.CAClientCertPath("org1")
	if err != nil {
		t.Fatalf("Failed to load CAClientCertPath from config. Error: %s", err)
	}

	// get the client key pem (embedded) for org1
	c.CAClientKeyPem("org1")
	if err != nil {
		t.Fatalf("Failed to load CAClientKeyPem from config. Error: %s", err)
	}

	// get the client key file path for org1
	c.CAClientKeyPath("org1")
	if err != nil {
		t.Fatalf("Failed to load CAClientKeyPath from config. Error: %s", err)
	}
}

func TestLoadConfigWithEmbeddedUsersWithPems(t *testing.T) {
	// get a config file with embedded users
	c, err := FromFile(configEmbeddedUsersTestFilePath)()
	if err != nil {
		t.Fatal(err)
	}

	conf, err := c.NetworkConfig()

	if err != nil {
		t.Fatal(err)
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUser")].Cert.Pem == "" {
		t.Fatal("Failed to parse the embedded cert for user EmbeddedUser")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUser")].Key.Pem == "" {
		t.Fatal("Failed to parse the embedded key for user EmbeddedUser")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("NonExistentEmbeddedUser")].Key.Pem != "" {
		t.Fatal("Mistakenly found an embedded key for user NonExistentEmbeddedUser")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("NonExistentEmbeddedUser")].Cert.Pem != "" {
		t.Fatal("Mistakenly found an embedded cert for user NonExistentEmbeddedUser")
	}
}

func TestLoadConfigWithEmbeddedUsersWithPaths(t *testing.T) {
	// get a config file with embedded users
	c, err := FromFile(configEmbeddedUsersTestFilePath)()
	if err != nil {
		t.Fatal(err)
	}

	conf, err := c.NetworkConfig()

	if err != nil {
		t.Fatal(err)
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUserWithPaths")].Cert.Path == "" {
		t.Fatal("Failed to parse the embedded cert for user EmbeddedUserWithPaths")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("EmbeddedUserWithPaths")].Key.Path == "" {
		t.Fatal("Failed to parse the embedded key for user EmbeddedUserWithPaths")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("NonExistentEmbeddedUser")].Key.Path != "" {
		t.Fatal("Mistakenly found an embedded key for user NonExistentEmbeddedUser")
	}

	if conf.Organizations[strings.ToLower(org1)].Users[strings.ToLower("NonExistentEmbeddedUser")].Cert.Path != "" {
		t.Fatal("Mistakenly found an embedded cert for user NonExistentEmbeddedUser")
	}
}

func TestInitConfigFromRawWrongType(t *testing.T) {
	// get a config byte for testing
	cBytes, err := loadConfigBytesFromFile(t, configPemTestFilePath)
	if err != nil {
		t.Fatalf("Failed to load sample bytes from File. Error: %s", err)
	}

	// test init config with empty type
	c, err := FromRaw(cBytes, "")()
	if err == nil {
		t.Fatalf("Expected error when initializing config with wrong config type but got no error.")
	}

	// test init config with wrong type
	c, err = FromRaw(cBytes, "json")()
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}

	o, err := c.OrderersConfig()
	if len(o) > 0 {
		t.Fatalf("Expected to get an empty list of orderers for wrong config type")
	}

	np, err := c.NetworkPeers()
	if len(np) > 0 {
		t.Fatalf("Expected to get an empty list of peers for wrong config type")
	}
}

func TestTLSClientCertsFromFiles(t *testing.T) {
	configImpl.networkConfig.Client.TLSCerts.Client.Cert.Path = "../../../test/fixtures/config/mutual_tls/client_sdk_go.pem"
	configImpl.networkConfig.Client.TLSCerts.Client.Key.Path = "../../../test/fixtures/config/mutual_tls/client_sdk_go-key.pem"
	configImpl.networkConfig.Client.TLSCerts.Client.Cert.Pem = ""
	configImpl.networkConfig.Client.TLSCerts.Client.Key.Pem = ""

	certs, err := configImpl.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertsFromFilesIncorrectPaths(t *testing.T) {
	// incorrect paths to files
	configImpl.networkConfig.Client.TLSCerts.Client.Cert.Path = "/test/fixtures/config/mutual_tls/client_sdk_go.pem"
	configImpl.networkConfig.Client.TLSCerts.Client.Key.Path = "/test/fixtures/config/mutual_tls/client_sdk_go-key.pem"
	configImpl.networkConfig.Client.TLSCerts.Client.Cert.Pem = ""
	configImpl.networkConfig.Client.TLSCerts.Client.Key.Pem = ""

	_, err := configImpl.TLSClientCerts()
	if err == nil {
		t.Fatalf("Expected error but got no errors instead")
	}

	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Fatalf("Expected no such file or directory error")
	}
}

func TestTLSClientCertsFromPem(t *testing.T) {
	configImpl.networkConfig.Client.TLSCerts.Client.Cert.Path = ""
	configImpl.networkConfig.Client.TLSCerts.Client.Key.Path = ""

	configImpl.networkConfig.Client.TLSCerts.Client.Cert.Pem = `-----BEGIN CERTIFICATE-----
MIIC5TCCAkagAwIBAgIUMYhiY5MS3jEmQ7Fz4X/e1Dx33J0wCgYIKoZIzj0EAwQw
gYwxCzAJBgNVBAYTAkNBMRAwDgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3Jv
bnRvMREwDwYDVQQKEwhsaW51eGN0bDEMMAoGA1UECxMDTGFiMTgwNgYDVQQDEy9s
aW51eGN0bCBFQ0MgUm9vdCBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eSAoTGFiKTAe
Fw0xNzEyMDEyMTEzMDBaFw0xODEyMDEyMTEzMDBaMGMxCzAJBgNVBAYTAkNBMRAw
DgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3JvbnRvMREwDwYDVQQKEwhsaW51
eGN0bDEMMAoGA1UECxMDTGFiMQ8wDQYDVQQDDAZzZGtfZ28wdjAQBgcqhkjOPQIB
BgUrgQQAIgNiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv2PotEEGuO6rMyaOu
pulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMhYZjcDi7YEOZ3Fs1h
xKmIxR+TTR2vf9KjgZAwgY0wDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsG
AQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFDwS3xhpAWs81OVWvZt+iUNL
z26DMB8GA1UdIwQYMBaAFLRasbknomawJKuQGiyKs/RzTCujMBgGA1UdEQQRMA+C
DWZhYnJpY19zZGtfZ28wCgYIKoZIzj0EAwQDgYwAMIGIAkIAk1MxMogtMtNO0rM8
gw2rrxqbW67ulwmMQzp6EJbm/28T2pIoYWWyIwpzrquypI7BOuf8is5b7Jcgn9oz
7sdMTggCQgF7/8ZFl+wikAAPbciIL1I+LyCXKwXosdFL6KMT6/myYjsGNeeDeMbg
3YkZ9DhdH1tN4U/h+YulG/CkKOtUATtQxg==
-----END CERTIFICATE-----`

	configImpl.networkConfig.Client.TLSCerts.Client.Key.Pem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDByldj7VTpqTQESGgJpR9PFW9b6YTTde2WN6/IiBo2nW+CIDmwQgmAl
c/EOc9wmgu+gBwYFK4EEACKhZANiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv
2PotEEGuO6rMyaOupulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMh
YZjcDi7YEOZ3Fs1hxKmIxR+TTR2vf9I=
-----END EC PRIVATE KEY-----`

	certs, err := configImpl.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertFromPemAndKeyFromFile(t *testing.T) {
	configImpl.networkConfig.Client.TLSCerts.Client.Cert.Path = ""
	configImpl.networkConfig.Client.TLSCerts.Client.Key.Path = "../../../test/fixtures/config/mutual_tls/client_sdk_go-key.pem"

	configImpl.networkConfig.Client.TLSCerts.Client.Cert.Pem = `-----BEGIN CERTIFICATE-----
MIIC5TCCAkagAwIBAgIUMYhiY5MS3jEmQ7Fz4X/e1Dx33J0wCgYIKoZIzj0EAwQw
gYwxCzAJBgNVBAYTAkNBMRAwDgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3Jv
bnRvMREwDwYDVQQKEwhsaW51eGN0bDEMMAoGA1UECxMDTGFiMTgwNgYDVQQDEy9s
aW51eGN0bCBFQ0MgUm9vdCBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eSAoTGFiKTAe
Fw0xNzEyMDEyMTEzMDBaFw0xODEyMDEyMTEzMDBaMGMxCzAJBgNVBAYTAkNBMRAw
DgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3JvbnRvMREwDwYDVQQKEwhsaW51
eGN0bDEMMAoGA1UECxMDTGFiMQ8wDQYDVQQDDAZzZGtfZ28wdjAQBgcqhkjOPQIB
BgUrgQQAIgNiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv2PotEEGuO6rMyaOu
pulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMhYZjcDi7YEOZ3Fs1h
xKmIxR+TTR2vf9KjgZAwgY0wDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsG
AQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFDwS3xhpAWs81OVWvZt+iUNL
z26DMB8GA1UdIwQYMBaAFLRasbknomawJKuQGiyKs/RzTCujMBgGA1UdEQQRMA+C
DWZhYnJpY19zZGtfZ28wCgYIKoZIzj0EAwQDgYwAMIGIAkIAk1MxMogtMtNO0rM8
gw2rrxqbW67ulwmMQzp6EJbm/28T2pIoYWWyIwpzrquypI7BOuf8is5b7Jcgn9oz
7sdMTggCQgF7/8ZFl+wikAAPbciIL1I+LyCXKwXosdFL6KMT6/myYjsGNeeDeMbg
3YkZ9DhdH1tN4U/h+YulG/CkKOtUATtQxg==
-----END CERTIFICATE-----`

	configImpl.networkConfig.Client.TLSCerts.Client.Key.Pem = ""

	certs, err := configImpl.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertFromFileAndKeyFromPem(t *testing.T) {
	configImpl.networkConfig.Client.TLSCerts.Client.Cert.Path = "../../../test/fixtures/config/mutual_tls/client_sdk_go.pem"
	configImpl.networkConfig.Client.TLSCerts.Client.Key.Path = ""

	configImpl.networkConfig.Client.TLSCerts.Client.Cert.Pem = ""

	configImpl.networkConfig.Client.TLSCerts.Client.Key.Pem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDByldj7VTpqTQESGgJpR9PFW9b6YTTde2WN6/IiBo2nW+CIDmwQgmAl
c/EOc9wmgu+gBwYFK4EEACKhZANiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv
2PotEEGuO6rMyaOupulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMh
YZjcDi7YEOZ3Fs1hxKmIxR+TTR2vf9I=
-----END EC PRIVATE KEY-----`

	certs, err := configImpl.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertsPemBeforeFiles(t *testing.T) {
	// files have incorrect paths, but pems are loaded first
	configImpl.networkConfig.Client.TLSCerts.Client.Cert.Path = "/test/fixtures/config/mutual_tls/client_sdk_go.pem"
	configImpl.networkConfig.Client.TLSCerts.Client.Key.Path = "/test/fixtures/config/mutual_tls/client_sdk_go-key.pem"

	configImpl.networkConfig.Client.TLSCerts.Client.Cert.Pem = `-----BEGIN CERTIFICATE-----
MIIC5TCCAkagAwIBAgIUMYhiY5MS3jEmQ7Fz4X/e1Dx33J0wCgYIKoZIzj0EAwQw
gYwxCzAJBgNVBAYTAkNBMRAwDgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3Jv
bnRvMREwDwYDVQQKEwhsaW51eGN0bDEMMAoGA1UECxMDTGFiMTgwNgYDVQQDEy9s
aW51eGN0bCBFQ0MgUm9vdCBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eSAoTGFiKTAe
Fw0xNzEyMDEyMTEzMDBaFw0xODEyMDEyMTEzMDBaMGMxCzAJBgNVBAYTAkNBMRAw
DgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3JvbnRvMREwDwYDVQQKEwhsaW51
eGN0bDEMMAoGA1UECxMDTGFiMQ8wDQYDVQQDDAZzZGtfZ28wdjAQBgcqhkjOPQIB
BgUrgQQAIgNiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv2PotEEGuO6rMyaOu
pulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMhYZjcDi7YEOZ3Fs1h
xKmIxR+TTR2vf9KjgZAwgY0wDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsG
AQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFDwS3xhpAWs81OVWvZt+iUNL
z26DMB8GA1UdIwQYMBaAFLRasbknomawJKuQGiyKs/RzTCujMBgGA1UdEQQRMA+C
DWZhYnJpY19zZGtfZ28wCgYIKoZIzj0EAwQDgYwAMIGIAkIAk1MxMogtMtNO0rM8
gw2rrxqbW67ulwmMQzp6EJbm/28T2pIoYWWyIwpzrquypI7BOuf8is5b7Jcgn9oz
7sdMTggCQgF7/8ZFl+wikAAPbciIL1I+LyCXKwXosdFL6KMT6/myYjsGNeeDeMbg
3YkZ9DhdH1tN4U/h+YulG/CkKOtUATtQxg==
-----END CERTIFICATE-----`

	configImpl.networkConfig.Client.TLSCerts.Client.Key.Pem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDByldj7VTpqTQESGgJpR9PFW9b6YTTde2WN6/IiBo2nW+CIDmwQgmAl
c/EOc9wmgu+gBwYFK4EEACKhZANiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv
2PotEEGuO6rMyaOupulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMh
YZjcDi7YEOZ3Fs1hxKmIxR+TTR2vf9I=
-----END EC PRIVATE KEY-----`

	certs, err := configImpl.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only one tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is empty")
	}
}

func TestTLSClientCertsNoCerts(t *testing.T) {
	configImpl.networkConfig.Client.TLSCerts.Client.Cert.Path = ""
	configImpl.networkConfig.Client.TLSCerts.Client.Key.Path = ""
	configImpl.networkConfig.Client.TLSCerts.Client.Cert.Pem = ""
	configImpl.networkConfig.Client.TLSCerts.Client.Key.Pem = ""

	certs, err := configImpl.TLSClientCerts()
	if err != nil {
		t.Fatalf("Expected no errors but got error instead: %s", err)
	}

	if len(certs) != 1 {
		t.Fatalf("Expected only empty tls cert struct")
	}

	emptyCert := tls.Certificate{}

	if !reflect.DeepEqual(certs[0], emptyCert) {
		t.Fatalf("Actual cert is not equal to empty cert")
	}
}

func TestNetworkPeerConfigFromURL(t *testing.T) {
	configProvider, err := FromFile(configTestFilePath)()
	if err != nil {
		t.Fatalf("Unexpected error reading config: %v", err)
	}
	sampleConfig := configProvider.(*Config)

	_, err = NetworkPeerConfigFromURL(sampleConfig, "invalid")
	assert.NotNil(t, err, "invalid url should return err")

	np, err := NetworkPeerConfigFromURL(sampleConfig, "peer0.org2.example.com:8051")
	assert.Nil(t, err, "valid url should not return err")
	assert.Equal(t, "peer0.org2.example.com:8051", np.URL, "wrong URL")
	assert.Equal(t, "Org2MSP", np.MSPID, "wrong MSP")

	np, err = NetworkPeerConfigFromURL(sampleConfig, "peer0.org1.example.com:7051")
	assert.Nil(t, err, "valid url should not return err")
	assert.Equal(t, "peer0.org1.example.com:7051", np.URL, "wrong URL")
	assert.Equal(t, "Org1MSP", np.MSPID, "wrong MSP")
}

func TestNewGoodOpt(t *testing.T) {
	_, err := FromFile("../../../test/fixtures/config/config_test.yaml", goodOpt())()
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	cBytes, err := loadConfigBytesFromFile(t, configTestFilePath)
	if err != nil || len(cBytes) == 0 {
		t.Fatalf("Unexpected error from loadConfigBytesFromFile")
	}

	buf := bytes.NewBuffer(cBytes)

	_, err = FromReader(buf, configType, goodOpt())()
	if err != nil {
		t.Fatalf("Unexpected error from FromReader: %v", err)
	}

	_, err = FromRaw(cBytes, configType, goodOpt())()
	if err != nil {
		t.Fatalf("Unexpected error from FromRaw %v", err)
	}

	err = os.Setenv("FABRIC_SDK_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	/*
		_, err = FromDefaultPath(goodOpt())
		if err != nil {
			t.Fatalf("Unexpected error from FromRaw: %v", err)
		}
	*/
}

func goodOpt() Option {
	return func(opts *options) error {
		return nil
	}
}

func TestNewBadOpt(t *testing.T) {
	_, err := FromFile("../../../test/fixtures/config/config_test.yaml", badOpt())()
	if err == nil {
		t.Fatalf("Expected error from FromFile")
	}

	cBytes, err := loadConfigBytesFromFile(t, configTestFilePath)
	if err != nil || len(cBytes) == 0 {
		t.Fatalf("Unexpected error from loadConfigBytesFromFile")
	}

	buf := bytes.NewBuffer(cBytes)

	_, err = FromReader(buf, configType, badOpt())()
	if err == nil {
		t.Fatalf("Expected error from FromReader")
	}

	_, err = FromRaw(cBytes, configType, badOpt())()
	if err == nil {
		t.Fatalf("Expected error from FromRaw")
	}

	err = os.Setenv("FABRIC_SDK_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	/*
		_, err = FromDefaultPath(badOpt())
		if err == nil {
			t.Fatalf("Expected error from FromRaw")
		}
	*/
}

func badOpt() Option {
	return func(opts *options) error {
		return errors.New("Bad Opt")
	}
}

/*
func TestDefaultConfigFromFile(t *testing.T) {
	c, err := FromFile(configEmptyTestFilePath, WithTemplatePath(defaultConfigPath))

	if err != nil {
		t.Fatalf("Unexpected error from FromFile: %s", err)
	}

	n, err := c.NetworkConfig()
	if err != nil {
		t.Fatalf("Failed to load default network config: %v", err)
	}

	if n.Name != "default-network" {
		t.Fatalf("Default network was not loaded. Network name loaded is: %s", n.Name)
	}

	if n.Description != "hello" {
		t.Fatalf("Incorrect Network name from default config. Got %s", n.Description)
	}
}

func TestDefaultConfigFromRaw(t *testing.T) {
	cBytes, err := loadConfigBytesFromFile(t, configEmptyTestFilePath)
	c, err := FromRaw(cBytes, configType, WithTemplatePath(defaultConfigPath))

	if err != nil {
		t.Fatalf("Unexpected error from FromFile: %s", err)
	}

	n, err := c.NetworkConfig()
	if err != nil {
		t.Fatalf("Failed to load default network config: %v", err)
	}

	if n.Name != "default-network" {
		t.Fatalf("Default network was not loaded. Network name loaded is: %s", n.Name)
	}

	if n.Description != "hello" {
		t.Fatalf("Incorrect Network name from default config. Got %s", n.Description)
	}
}
*/

/*
func TestFromDefaultPathSuccess(t *testing.T) {
	err := os.Setenv("FABRIC_SDK_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	// test init config from bytes
	_, err = FromDefaultPath()
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}
}

func TestFromDefaultPathCustomPrefixSuccess(t *testing.T) {
	err := os.Setenv("FABRIC_SDK2_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK2_CONFIG_PATH")

	// test init config from bytes
	_, err = FromDefaultPath(WithEnvPrefix("FABRIC_SDK2"))
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}
}

func TestFromDefaultPathCustomPathSuccess(t *testing.T) {
	err := os.Setenv("FABRIC_SDK2_CONFIG_PATH", defaultConfigPath)
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK2_CONFIG_PATH")

	// test init config from bytes
	_, err = FromDefaultPath(WithTemplatePath(defaultConfigPath))
	if err != nil {
		t.Fatalf("Failed to initialize config from bytes array. Error: %s", err)
	}
}

func TestFromDefaultPathEmptyFailure(t *testing.T) {
	os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	// test init config from bytes
	_, err := FromDefaultPath()
	if err == nil {
		t.Fatalf("Expected failure from unset FABRIC_SDK_CONFIG_PATH")
	}
}

func TestFromDefaultPathFailure(t *testing.T) {
	err := os.Setenv("FABRIC_SDK_CONFIG_PATH", defaultConfigPath+"/bad")
	if err != nil {
		t.Fatalf("Unexpected problem setting environment. Error: %s", err)
	}
	defer os.Unsetenv("FABRIC_SDK_CONFIG_PATH")

	// test init config from bytes
	_, err = FromDefaultPath()
	if err == nil {
		t.Fatalf("Expected failure from bad default path")
	}
}
*/
