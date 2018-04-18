/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"crypto/tls"
	"crypto/x509"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// EndpointConfigOptions represents EndpointConfig interface with overridable interface functions
// if a function is not overridden, the default EndpointConfig implementation will be called.
type EndpointConfigOptions struct {
	timeout
	mspID
	peerMSPID
	orderersConfig
	ordererConfig
	peersConfig
	peerConfig
	peerConfigByURL
	networkConfig
	networkPeers
	channelConfig
	channelPeers
	channelOrderers
	tlsCACertPool
	eventServiceType
	tlsClientCerts
	cryptoConfigPath
	defaultConfig EndpointConfig
}

// timeout interface allows to uniquely override EndpointConfig interface's Timeout() function
type timeout interface {
	Timeout(fab.TimeoutType) time.Duration
}

// mspID interface allows to uniquely override EndpointConfig interface's MSPID() function
type mspID interface {
	MSPID(org string) (string, error)
}

// peerMSPID interface allows to uniquely override EndpointConfig interface's PeerMSPID() function
type peerMSPID interface {
	PeerMSPID(name string) (string, error)
}

// orderersConfig interface allows to uniquely override EndpointConfig interface's OrderersConfig() function
type orderersConfig interface {
	OrderersConfig() ([]fab.OrdererConfig, error)
}

// ordererConfig interface allows to uniquely override EndpointConfig interface's OrdererConfig() function
type ordererConfig interface {
	OrdererConfig(name string) (*fab.OrdererConfig, error)
}

// peersConfig interface allows to uniquely override EndpointConfig interface's PeersConfig() function
type peersConfig interface {
	PeersConfig(org string) ([]fab.PeerConfig, error)
}

// peerConfig interface allows to uniquely override EndpointConfig interface's PeerConfig() function
type peerConfig interface {
	PeerConfig(org string, name string) (*fab.PeerConfig, error)
}

// peerConfigByURL interface allows to uniquely override EndpointConfig interface's PeerConfigByURL() function
type peerConfigByURL interface {
	PeerConfigByURL(url string) (*fab.PeerConfig, error)
}

// networkConfig interface allows to uniquely override EndpointConfig interface's NetworkConfig() function
type networkConfig interface {
	NetworkConfig() (*fab.NetworkConfig, error)
}

// networkPeers interface allows to uniquely override EndpointConfig interface's NetworkPeers() function
type networkPeers interface {
	NetworkPeers() ([]fab.NetworkPeer, error)
}

// channelConfig interface allows to uniquely override EndpointConfig interface's ChannelConfig() function
type channelConfig interface {
	ChannelConfig(name string) (*fab.ChannelNetworkConfig, error)
}

// channelPeers interface allows to uniquely override EndpointConfig interface's ChannelPeers() function
type channelPeers interface {
	ChannelPeers(name string) ([]fab.ChannelPeer, error)
}

// channelOrderers interface allows to uniquely override EndpointConfig interface's ChannelOrderers() function
type channelOrderers interface {
	ChannelOrderers(name string) ([]fab.OrdererConfig, error)
}

// tlsCACertPool interface allows to uniquely override EndpointConfig interface's TLSCACertPool() function
type tlsCACertPool interface {
	TLSCACertPool(certConfig ...*x509.Certificate) (*x509.CertPool, error)
}

// eventServiceType interface allows to uniquely override EndpointConfig interface's EventServiceType() function
type eventServiceType interface {
	EventServiceType() fab.EventServiceType
}

// tlsClientCerts interface allows to uniquely override EndpointConfig interface's TLSClientCerts() function
type tlsClientCerts interface {
	TLSClientCerts() ([]tls.Certificate, error)
}

// cryptoConfigPath interface allows to uniquely override EndpointConfig interface's CryptoConfigPath() function
type cryptoConfigPath interface {
	CryptoConfigPath() string
}

func BuildConfigEndpointFromOptions(opts ...interface{}) fab.EndpointConfig {
	c := &EndpointConfigOptions{defaultConfig: EndpointConfig{}}

	for _, option := range opts {
		overrideDefaultConfigEndpointInterface(c, option)
	}

	return c
}

func overrideDefaultConfigEndpointInterface(c *EndpointConfigOptions, o interface{}) {
	switch o.(type) {
	case timeout:
		c.timeout = o.(timeout)
	case mspID:
		c.mspID = o.(mspID)
	case peerMSPID:
		c.peerMSPID = o.(peerMSPID)
	case orderersConfig:
		c.orderersConfig = o.(orderersConfig)
	case ordererConfig:
		c.ordererConfig = o.(ordererConfig)
	case peersConfig:
		c.peersConfig = o.(peersConfig)
	case peerConfig:
		c.peerConfig = o.(peerConfig)
	case peerConfigByURL:
		c.peerConfigByURL = o.(peerConfigByURL)
	case networkConfig:
		c.networkConfig = o.(networkConfig)
	case networkPeers:
		c.networkPeers = o.(networkPeers)
	case channelConfig:
		c.channelConfig = o.(channelConfig)
	case channelPeers:
		c.channelPeers = o.(channelPeers)
	case channelOrderers:
		c.channelOrderers = o.(channelOrderers)
	case tlsCACertPool:
		c.tlsCACertPool = o.(tlsCACertPool)
	case eventServiceType:
		c.eventServiceType = o.(eventServiceType)
	case tlsClientCerts:
		c.tlsClientCerts = o.(tlsClientCerts)
	case cryptoConfigPath:
		c.cryptoConfigPath = o.(cryptoConfigPath)
	}
}
