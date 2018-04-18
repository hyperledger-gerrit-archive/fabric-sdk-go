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
	"github.com/pkg/errors"
)

// EndpointConfigOptions represents EndpointConfig interface with overridable interface functions
// if a function is not overridden, the default EndpointConfig implementation will be used.
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

// BuildConfigEndpointFromOptions will return an EndpointConfig instance pre-built with Optional interfaces
func BuildConfigEndpointFromOptions(opts ...interface{}) (fab.EndpointConfig, error) {
	// if first interface is fab.EndpointConfig, then return it immediately (if the user prefers to override the whole interface at once)
	if i, ok := opts[0].([]interface{}); ok {
		logger.Debugf(" %#v", i)
		if j, ok := i[0].(fab.EndpointConfig); ok {
			return j, nil
		}
	}

	// if the user prefers to override only some sub interfaces of EndpointConfig and use the default ones for other functions,
	// then build a new EndpointConfig with overridden functions
	c := &EndpointConfigOptions{}
	for _, option := range opts {
		logger.Infof("type of option is %T, options is: %s. percentage v is %#v", option, option, option)

		err := setEndpointConfigWithOptionInterface(c, option)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func setEndpointConfigWithOptionInterface(c *EndpointConfigOptions, o interface{}) error {
	wasSet := firstSetofEndpointConfig(c, o)

	if !wasSet {
		wasSet = secondSetofEndpointConfig(c, o)
	}

	if !wasSet {
		return errors.New("option received is not a sub interface of EndpointConfig. It must declare one of its functions")
	}
	return nil
}

func firstSetofEndpointConfig(c *EndpointConfigOptions, o interface{}) bool {
	if v, ok := o.(timeout); ok {
		c.timeout = v
		return true
	}
	if v, ok := o.(mspID); ok {
		c.mspID = v
		return true
	}

	if v, ok := o.(peerMSPID); ok {
		c.peerMSPID = v
		return true
	}

	if v, ok := o.(orderersConfig); ok {
		c.orderersConfig = v
		return true
	}

	if v, ok := o.(ordererConfig); ok {
		c.ordererConfig = v
		return true
	}

	if v, ok := o.(peersConfig); ok {
		c.peersConfig = v
		return true
	}

	if v, ok := o.(peerConfig); ok {
		c.peerConfig = v
		return true
	}

	if v, ok := o.(peerConfigByURL); ok {
		c.peerConfigByURL = v
		return true
	}

	if v, ok := o.(networkConfig); ok {
		c.networkConfig = v
		return true
	}

	return false
}

func secondSetofEndpointConfig(c *EndpointConfigOptions, o interface{}) bool {
	if v, ok := o.(networkPeers); ok {
		c.networkPeers = v
		return true
	}

	if v, ok := o.(channelConfig); ok {
		c.channelConfig = v
		return true
	}

	if v, ok := o.(channelPeers); ok {
		c.channelPeers = v
		return true
	}

	if v, ok := o.(channelOrderers); ok {
		c.channelOrderers = v
		return true
	}

	if v, ok := o.(tlsCACertPool); ok {
		c.tlsCACertPool = v
		return true
	}

	if v, ok := o.(eventServiceType); ok {
		c.eventServiceType = v
		return true
	}

	if v, ok := o.(tlsClientCerts); ok {
		c.tlsClientCerts = v
		return true
	}

	if v, ok := o.(cryptoConfigPath); ok {
		c.cryptoConfigPath = v
		return true
	}

	return false
}

// UpdateMissingOptsWithDefaultConfig will verify if any functions of the EndpointConfig were not updated WithConfigEndpoint,
// then use default EndpointConfig interface instead
func UpdateMissingOptsWithDefaultConfig(c *EndpointConfigOptions, d fab.EndpointConfig) fab.EndpointConfig {
	firstUpdateMissingOptsWithDefault(c, d)
	secondUpdateMissingOptsWithDefault(c, d)
	return c
}
func firstUpdateMissingOptsWithDefault(c *EndpointConfigOptions, d fab.EndpointConfig) {
	if c.timeout == nil {
		c.timeout = d
	}
	if c.mspID == nil {
		c.mspID = d
	}
	if c.peerMSPID == nil {
		c.peerMSPID = d
	}
	if c.orderersConfig == nil {
		c.orderersConfig = d
	}
	if c.ordererConfig == nil {
		c.ordererConfig = d
	}
	if c.peersConfig == nil {
		c.peersConfig = d
	}
	if c.peerConfig == nil {
		c.peerConfig = d
	}
	if c.peerConfigByURL == nil {
		c.peerConfigByURL = d
	}
}

func secondUpdateMissingOptsWithDefault(c *EndpointConfigOptions, d fab.EndpointConfig) {
	if c.networkConfig == nil {
		c.networkConfig = d
	}
	if c.networkPeers == nil {
		c.networkPeers = d
	}
	if c.channelConfig == nil {
		c.channelConfig = d
	}
	if c.channelPeers == nil {
		c.channelPeers = d
	}
	if c.channelOrderers == nil {
		c.channelOrderers = d
	}
	if c.tlsCACertPool == nil {
		c.tlsCACertPool = d
	}
	if c.eventServiceType == nil {
		c.eventServiceType = d
	}
	if c.tlsClientCerts == nil {
		c.tlsClientCerts = d
	}
	if c.cryptoConfigPath == nil {
		c.cryptoConfigPath = d
	}
}
