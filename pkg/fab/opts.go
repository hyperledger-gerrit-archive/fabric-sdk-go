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

type setter func()
type predicate func() bool

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
	PeerConfig(nameOrURL string) (*fab.PeerConfig, error)
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
	TLSCACertPool(certConfig ...*x509.Certificate) *x509.CertPool
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
// provided in fabsdk's WithEndpointConfig(opts...) call
func BuildConfigEndpointFromOptions(opts ...interface{}) (fab.EndpointConfig, error) {
	// build a new EndpointConfig with overridden function implementations
	c := &EndpointConfigOptions{}
	for i, option := range opts {
		logger.Debugf("option %d: %#v", i, option)
		err := setEndpointConfigWithOptionInterface(c, option)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

// will override EndpointConfig interface with functions provided by o (option)
func setEndpointConfigWithOptionInterface(c *EndpointConfigOptions, o interface{}) error {
	var isSetArr []bool
	isSetArr = append(isSetArr, set(c.timeout, func() bool { _, ok := o.(timeout); return ok }, func() { c.timeout = o.(timeout) }))
	isSetArr = append(isSetArr, set(c.mspID, func() bool { _, ok := o.(mspID); return ok }, func() { c.mspID = o.(mspID) }))
	isSetArr = append(isSetArr, set(c.peerMSPID, func() bool { _, ok := o.(peerMSPID); return ok }, func() { c.peerMSPID = o.(peerMSPID) }))
	isSetArr = append(isSetArr, set(c.orderersConfig, func() bool { _, ok := o.(orderersConfig); return ok }, func() { c.orderersConfig = o.(orderersConfig) }))
	isSetArr = append(isSetArr, set(c.ordererConfig, func() bool { _, ok := o.(ordererConfig); return ok }, func() { c.ordererConfig = o.(ordererConfig) }))
	isSetArr = append(isSetArr, set(c.peersConfig, func() bool { _, ok := o.(peersConfig); return ok }, func() { c.peersConfig = o.(peersConfig) }))
	isSetArr = append(isSetArr, set(c.peerConfig, func() bool { _, ok := o.(peerConfig); return ok }, func() { c.peerConfig = o.(peerConfig) }))
	isSetArr = append(isSetArr, set(c.networkConfig, func() bool { _, ok := o.(networkConfig); return ok }, func() { c.networkConfig = o.(networkConfig) }))
	isSetArr = append(isSetArr, set(c.networkPeers, func() bool { _, ok := o.(networkPeers); return ok }, func() { c.networkPeers = o.(networkPeers) }))
	isSetArr = append(isSetArr, set(c.channelConfig, func() bool { _, ok := o.(channelConfig); return ok }, func() { c.channelConfig = o.(channelConfig) }))
	isSetArr = append(isSetArr, set(c.channelPeers, func() bool { _, ok := o.(channelPeers); return ok }, func() { c.channelPeers = o.(channelPeers) }))
	isSetArr = append(isSetArr, set(c.channelOrderers, func() bool { _, ok := o.(channelOrderers); return ok }, func() { c.channelOrderers = o.(channelOrderers) }))
	isSetArr = append(isSetArr, set(c.tlsCACertPool, func() bool { _, ok := o.(tlsCACertPool); return ok }, func() { c.tlsCACertPool = o.(tlsCACertPool) }))
	isSetArr = append(isSetArr, set(c.eventServiceType, func() bool { _, ok := o.(eventServiceType); return ok }, func() { c.eventServiceType = o.(eventServiceType) }))
	isSetArr = append(isSetArr, set(c.tlsClientCerts, func() bool { _, ok := o.(tlsClientCerts); return ok }, func() { c.tlsClientCerts = o.(tlsClientCerts) }))
	isSetArr = append(isSetArr, set(c.cryptoConfigPath, func() bool { _, ok := o.(cryptoConfigPath); return ok }, func() { c.cryptoConfigPath = o.(cryptoConfigPath) }))

	// TODO for now, isSetArr is used to loop through the results of set() to avoid meta-linter error, find a better way
	isAnySet := false
	for _, isSet := range isSetArr {
		isAnySet = isSet || isAnySet
		if isAnySet {
			break
		}
	}

	if !isAnySet {
		return errors.Errorf("option %#v is not a sub interface of EndpointConfig, at least one of its functions must be implemented.", o)
	}
	return nil
}

// needed to avoid meta-linter errors (too many if conditions)
func set(current interface{}, check predicate, apply setter) bool {
	if current == nil && check() {
		apply()
		return true
	}

	return false
}

// UpdateMissingOptsWithDefaultConfig will verify if any functions of the EndpointConfig were not updated with fabsdk's
// WithConfigEndpoint(opts...) call, then use default EndpointConfig interface for these functions instead
func UpdateMissingOptsWithDefaultConfig(c *EndpointConfigOptions, d fab.EndpointConfig) fab.EndpointConfig {
	trueCheckFunc := func() bool { return true }
	set(c.timeout, trueCheckFunc, func() { c.timeout = d })
	set(c.mspID, trueCheckFunc, func() { c.mspID = d })
	set(c.peerMSPID, trueCheckFunc, func() { c.peerMSPID = d })
	set(c.orderersConfig, trueCheckFunc, func() { c.orderersConfig = d })
	set(c.ordererConfig, trueCheckFunc, func() { c.ordererConfig = d })
	set(c.peersConfig, trueCheckFunc, func() { c.peersConfig = d })
	set(c.peerConfig, trueCheckFunc, func() { c.peerConfig = d })
	set(c.networkConfig, trueCheckFunc, func() { c.networkConfig = d })
	set(c.networkPeers, trueCheckFunc, func() { c.networkPeers = d })
	set(c.channelConfig, trueCheckFunc, func() { c.channelConfig = d })
	set(c.channelPeers, trueCheckFunc, func() { c.channelPeers = d })
	set(c.channelOrderers, trueCheckFunc, func() { c.channelOrderers = d })
	set(c.tlsCACertPool, trueCheckFunc, func() { c.tlsCACertPool = d })
	set(c.eventServiceType, trueCheckFunc, func() { c.eventServiceType = d })
	set(c.tlsClientCerts, trueCheckFunc, func() { c.tlsClientCerts = d })
	set(c.cryptoConfigPath, trueCheckFunc, func() { c.cryptoConfigPath = d })

	return c
}

// IsEndpointConfigFullyOverridden will return true if all of the argument's sub interfaces is not nil
// (ie EndpointConfig interface not fully overridden)
func IsEndpointConfigFullyOverridden(c *EndpointConfigOptions) bool {
	var hasNilArr []bool
	hasNilArr = append(hasNilArr, c.timeout == nil)
	hasNilArr = append(hasNilArr, c.mspID == nil)
	hasNilArr = append(hasNilArr, c.peerMSPID == nil)
	hasNilArr = append(hasNilArr, c.orderersConfig == nil)
	hasNilArr = append(hasNilArr, c.ordererConfig == nil)
	hasNilArr = append(hasNilArr, c.peersConfig == nil)
	hasNilArr = append(hasNilArr, c.peerConfig == nil)
	hasNilArr = append(hasNilArr, c.networkConfig == nil)
	hasNilArr = append(hasNilArr, c.networkPeers == nil)
	hasNilArr = append(hasNilArr, c.channelPeers == nil)
	hasNilArr = append(hasNilArr, c.channelOrderers == nil)
	hasNilArr = append(hasNilArr, c.tlsCACertPool == nil)
	hasNilArr = append(hasNilArr, c.eventServiceType == nil)
	hasNilArr = append(hasNilArr, c.tlsClientCerts == nil)
	hasNilArr = append(hasNilArr, c.cryptoConfigPath == nil)

	hasNil := false
	for _, isNil := range hasNilArr {
		if isNil {
			hasNil = true
			break
		}
	}
	return !hasNil
}
