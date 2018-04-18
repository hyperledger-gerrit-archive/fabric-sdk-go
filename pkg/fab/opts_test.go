/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"crypto/tls"
	"crypto/x509"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
)

var (
	m0  = &EndpointConfig{}
	m1  = &MockTimeoutConfig{}
	m2  = &MockMspID{}
	m3  = &MockPeerMSPID{}
	m4  = &MockrderersConfig{}
	m5  = &MockOrdererConfig{}
	m6  = &MockPeersConfig{}
	m7  = &MockPeerConfig{}
	m8  = &MockNetworkConfig{}
	m9  = &MockNetworkPeers{}
	m10 = &MockChannelConfig{}
	m11 = &MockChannelPeers{}
	m12 = &MockChannelOrderers{}
	m13 = &MockTLSCACertPool{}
	m14 = &mockEventServiceType{}
	m15 = &MockTLSClientCerts{}
	m16 = &MockCryptoConfigPath{}
)

func TestCreateCustomFullEndpointConfig(t *testing.T) {
	var opts []interface{}
	opts = append(opts, m0)
	// try to build with the overall interface (m0 is the overall interface implementation)
	endpointConfigOption, err := BuildConfigEndpointFromOptions(opts...)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}
	if endpointConfigOption == nil {
		t.Fatalf("BuildConfigEndpointFromOptions call returned nil")
	}
}

func TestCreateCustomEndpointConfig(t *testing.T) {
	// try to build with partial interfaces
	endpointConfigOption, err := BuildConfigEndpointFromOptions(m1, m2, m3, m4, m5, m6, m7, m8, m9, m10)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}
	var eco *EndpointConfigOptions
	var ok bool
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}
	if eco == nil {
		t.Fatalf("build ConfigEndpointOption returned is nil")
	}
	tmout := eco.Timeout(fab.EndorserConnection)
	if tmout < 0 {
		t.Fatalf("EndpointConfig was supposed to have Timeout function overridden from Options but was not %+v. Timeout: %s", eco, tmout)
	}
	m, err := eco.MSPID("")
	if err != nil {
		t.Fatalf("MSPID returned unexpected error %s", err)
	}
	if m != "testMSP" {
		t.Fatalf("MSPID did not return expected interface value. Expected: %s, Received: %s", "testMSP", m)
	}
	m, err = eco.PeerMSPID("")
	if err != nil {
		t.Fatalf("PeerMSPID returned unexpected error %s", err)
	}
	if m != "testPeerMSP" {
		t.Fatalf("MSPID did not return expected interface value. Expected: %s, Received: %s", "testPeerMSP", m)
	}

	// verify if an interface was not passed as an option but was not nil, it should be nil
	if eco.channelPeers != nil {
		t.Fatalf("channelPeers created with nil interface but got non nil one. %s", eco.channelPeers)
	}
}

func TestCreateCustomEndpointConfigRemainingFunctions(t *testing.T) {
	// test other sub interface functions
	endpointConfigOption, err := BuildConfigEndpointFromOptions(m11, m12, m13, m14, m15, m16)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}
	var eco *EndpointConfigOptions
	var ok bool
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}
	if eco == nil {
		t.Fatalf("build ConfigEndpointOption returned is nil")
	}
	// verify that their functions are available
	p, err := eco.ChannelPeers("")
	if err != nil {
		t.Fatalf("ChannelPeers returned unexpected error %s", err)
	}
	if len(p) != 1 {
		t.Fatalf("ChannelPeers did not return expected interface value. Expected: 1 ChannelPeer, Received: %d", len(p))
	}

	c, err := eco.TLSClientCerts()
	if err != nil {
		t.Fatalf("TLSClientCerts returned unexpected error %s", err)
	}
	if len(c) != 2 {
		t.Fatalf("TLSClientCerts did not return expected interface value. Expected: 2 Certificates, Received: %d", len(c))
	}

	// verify if an interface that was not passed as an option but was not nil, it should be nil
	if eco.timeout != nil {
		t.Fatalf("timeout created with nil timeout interface but got non nil one. %s", eco.timeout)
	}

	// now try with non related interface to test if an error returns
	var badType interface{}
	_, err = BuildConfigEndpointFromOptions(m12, m13, badType)
	if err == nil {
		t.Fatalf("BuildConfigEndpointFromOptions did not return error with badType")
	}
}

func TestCreateCustomEndpoitConfigWithSomeDefaultFunctions(t *testing.T) {
	// create a config with the first 7 interfaces to be overridden
	endpointConfigOption, err := BuildConfigEndpointFromOptions(m1, m2, m3, m4, m5, m6, m7)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}

	var eco *EndpointConfigOptions
	var ok bool
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}
	if eco == nil {
		t.Fatalf("build ConfigEndpointOption returned is nil")
	}

	// now inject default interfaces (using m0 as default interface for the sake of this test) for the ones that were not overridden by options above
	endpointConfigOptionWithSomeDefaults := UpdateMissingOptsWithDefaultConfig(eco, m0)

	// test if options updated interfaces with options are still working
	tmout := endpointConfigOptionWithSomeDefaults.Timeout(fab.EndorserConnection)
	if tmout < 0 {
		t.Fatalf("EndpointConfig was supposed to have Timeout function overridden from Options but was not %+v. Timeout: %s", eco, tmout)
	}
	m, err := endpointConfigOptionWithSomeDefaults.MSPID("")
	if err != nil {
		t.Fatalf("MSPID returned unexpected error %s", err)
	}
	if m != "testMSP" {
		t.Fatalf("MSPID did not return expected interface value. Expected: %s, Received: %s", "testMSP", m)
	}

	// now check if interfaces that are not updated are defaulted with m0
	if eco, ok = endpointConfigOptionWithSomeDefaults.(*EndpointConfigOptions); !ok {
		t.Fatalf("UpdateMissingOptsWithDefaultConfig did not return a Options instance %T", endpointConfigOptionWithSomeDefaults)
	}
	// cryptoConfigPath (m17) is among the interfaces that were not updated by options
	if eco.cryptoConfigPath == nil {
		t.Fatalf("UpdateMissingOptsWithDefaultConfig did not set CryptoConfigPath() with default function implementation")
	}
	// tlsClientCerts (m16) is among the interfaces that were not updated by options
	if eco.tlsClientCerts == nil {
		t.Fatalf("UpdateMissingOptsWithDefaultConfig did not set TLSClientCerts() with default function implementation")
	}
}

func TestCreateCustomEndpoitConfigWithSomeDefaultFunctionsRemainingFunctions(t *testing.T) {
	// do the same test with the other interfaces in reverse
	endpointConfigOption, err := BuildConfigEndpointFromOptions(m8, m9, m10, m11, m12, m13, m14, m15, m16)
	if err != nil {
		t.Fatalf("BuildConfigEndpointFromOptions returned unexpected error %s", err)
	}

	var eco *EndpointConfigOptions
	var ok bool
	if eco, ok = endpointConfigOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %T", endpointConfigOption)
	}
	if eco == nil {
		t.Fatalf("build ConfigEndpointOption returned is nil")
	}

	// now inject default interfaces
	endpointConfigOptionWithSomeDefaults := UpdateMissingOptsWithDefaultConfig(eco, m0)

	//test that interfaces overrident by the options are still working
	m := endpointConfigOptionWithSomeDefaults.CryptoConfigPath()
	if m != "" {
		t.Fatalf("CryptoConfigPath did not return expected interface value. Expected: '%s', Received: %s", "", m)
	}
	e := endpointConfigOptionWithSomeDefaults.EventServiceType()

	if e != fab.DeliverEventServiceType {
		t.Fatalf("MSPID did not return expected interface value. Expected: %d, Received: %d", fab.DeliverEventServiceType, e)

	}
}

type MockTimeoutConfig struct{}

func (M *MockTimeoutConfig) Timeout(timeoutType fab.TimeoutType) time.Duration {
	return 10 * time.Second
}

type MockMspID struct{}

func (M *MockMspID) MSPID(org string) (string, error) {
	return "testMSP", nil
}

type MockPeerMSPID struct{}

func (M *MockPeerMSPID) PeerMSPID(name string) (string, error) {
	return "testPeerMSP", nil
}

type MockrderersConfig struct{}

func (M *MockrderersConfig) OrderersConfig() ([]fab.OrdererConfig, error) {
	return []fab.OrdererConfig{{URL: "orderer1.com", GRPCOptions: nil, TLSCACerts: endpoint.TLSConfig{Path: "", Pem: ""}}}, nil
}

type MockOrdererConfig struct{}

func (M *MockOrdererConfig) OrdererConfig(name string) (*fab.OrdererConfig, error) {
	return &fab.OrdererConfig{URL: "o.com", GRPCOptions: nil, TLSCACerts: endpoint.TLSConfig{Path: "", Pem: ""}}, nil
}

type MockPeersConfig struct{}

func (M *MockPeersConfig) PeersConfig(org string) ([]fab.PeerConfig, error) {
	return []fab.PeerConfig{{URL: "peer.com", EventURL: "event.peer.com", GRPCOptions: nil, TLSCACerts: endpoint.TLSConfig{Path: "", Pem: ""}}}, nil
}

type MockPeerConfig struct{}

func (M *MockPeerConfig) PeerConfig(nameOrURL string) (*fab.PeerConfig, error) {
	return &fab.PeerConfig{URL: "p.com", EventURL: "event.p.com", GRPCOptions: nil, TLSCACerts: endpoint.TLSConfig{Path: "", Pem: ""}}, nil
}

type MockNetworkConfig struct{}

func (M *MockNetworkConfig) NetworkConfig() (*fab.NetworkConfig, error) {
	return &fab.NetworkConfig{}, nil
}

type MockNetworkPeers struct{}

func (M *MockNetworkPeers) NetworkPeers() ([]fab.NetworkPeer, error) {
	return []fab.NetworkPeer{{PeerConfig: fab.PeerConfig{URL: "p.com", EventURL: "event.p.com", GRPCOptions: nil, TLSCACerts: endpoint.TLSConfig{Path: "", Pem: ""}}, MSPID: ""}}, nil
}

type MockChannelConfig struct{}

func (M *MockChannelConfig) ChannelConfig(name string) (*fab.ChannelNetworkConfig, error) {
	return &fab.ChannelNetworkConfig{}, nil
}

type MockChannelPeers struct{}

func (M *MockChannelPeers) ChannelPeers(name string) ([]fab.ChannelPeer, error) {
	return []fab.ChannelPeer{{}}, nil
}

type MockChannelOrderers struct{}

func (M *MockChannelOrderers) ChannelOrderers(name string) ([]fab.OrdererConfig, error) {
	return []fab.OrdererConfig{}, nil
}

type MockTLSCACertPool struct{}

func (M *MockTLSCACertPool) TLSCACertPool(certConfig ...*x509.Certificate) *x509.CertPool {
	return nil
}

type mockEventServiceType struct{}

func (M *mockEventServiceType) EventServiceType() fab.EventServiceType {
	return fab.DeliverEventServiceType
}

type MockTLSClientCerts struct{}

func (M *MockTLSClientCerts) TLSClientCerts() ([]tls.Certificate, error) {
	return []tls.Certificate{{}, {}}, nil
}

type MockCryptoConfigPath struct{}

func (M *MockCryptoConfigPath) CryptoConfigPath() string {
	return ""
}
