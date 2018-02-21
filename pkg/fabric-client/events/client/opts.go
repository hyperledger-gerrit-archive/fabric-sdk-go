/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/options"
)

type params struct {
	eventConsumerBufferSize uint
	reconn                  bool
	maxConnAttempts         uint
	maxReconnAttempts       uint
	reconnInitialDelay      time.Duration
	timeBetweenConnAttempts time.Duration
	connEventCh             chan *apifabclient.ConnectionEvent
	respTimeout             time.Duration
}

func defaultParams() *params {
	return &params{
		eventConsumerBufferSize: 100,
		reconn:                  true,
		maxConnAttempts:         1,
		maxReconnAttempts:       0, // Try forever
		reconnInitialDelay:      0,
		timeBetweenConnAttempts: 5 * time.Second,
		respTimeout:             5 * time.Second,
	}
}

// WithReconnect indicates whether the client should automatically attempt to reconnect
// to the server after a connection has been lost
func WithReconnect(value bool) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(reconnectSetter); ok {
			setter.SetReconnect(value)
		}
	}
}

// WithMaxConnectAttempts sets the maximum number of times that the client will attempt
// to connect to the server. If set to 0 then the client will try until it is stopped.
func WithMaxConnectAttempts(value uint) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(maxConnectAttemptsSetter); ok {
			setter.SetMaxConnectAttempts(value)
		}
	}
}

// WithMaxReconnectAttempts sets the maximum number of times that the client will attempt
// to reconnect to the server after a connection has been lost. If set to 0 then the
// client will try until it is stopped.
func WithMaxReconnectAttempts(value uint) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(maxReconnectAttemptsSetter); ok {
			setter.SetMaxReconnectAttempts(value)
		}
	}
}

// WithReconnectInitialDelay sets the initial delay before attempting to reconnect.
func WithReconnectInitialDelay(value time.Duration) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(reconnectInitialDelaySetter); ok {
			setter.SetReconnectInitialDelay(value)
		}
	}
}

// WithConnectEventCh sets the channel that is to receive connection events, i.e. when the client connects and/or
// disconnects from the channel event service.
func WithConnectEventCh(value chan *apifabclient.ConnectionEvent) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(connectEventChSetter); ok {
			setter.SetConnectEventCh(value)
		}
	}
}

// WithTimeBetweenConnectAttempts sets the time between connection attempts.
func WithTimeBetweenConnectAttempts(value time.Duration) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(timeBetweenConnectAttemptsSetter); ok {
			setter.SetTimeBetweenConnectAttempts(value)
		}
	}
}

// WithResponseTimeout sets the timeout when waiting for a response from the event server
func WithResponseTimeout(value time.Duration) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(responseTimeoutSetter); ok {
			setter.SetResponseTimeout(value)
		}
	}
}

func (p *params) SetEventConsumerBufferSize(value uint) {
	p.eventConsumerBufferSize = value
}

func (p *params) SetReconnect(value bool) {
	logger.Debugf("Reconnect: %t", value)
	p.reconn = value
}

func (p *params) SetMaxConnectAttempts(value uint) {
	logger.Debugf("MaxConnectAttempts: %d", value)
	p.maxConnAttempts = value
}

func (p *params) SetMaxReconnectAttempts(value uint) {
	logger.Debugf("MaxReconnectAttempts: %d", value)
	p.maxReconnAttempts = value
}

func (p *params) SetReconnectInitialDelay(value time.Duration) {
	logger.Debugf("ReconnectInitialDelay: %s", value)
	p.reconnInitialDelay = value
}

func (p *params) SetTimeBetweenConnectAttempts(value time.Duration) {
	logger.Debugf("TimeBetweenConnectAttempts: %d", value)
	p.timeBetweenConnAttempts = value
}

func (p *params) SetConnectEventCh(value chan *apifabclient.ConnectionEvent) {
	logger.Debugf("ConnectEventCh: %#v", value)
	p.connEventCh = value
}

func (p *params) SetResponseTimeout(value time.Duration) {
	logger.Debugf("ResponseTimeout: %s", value)
	p.respTimeout = value
}

type reconnectSetter interface {
	SetReconnect(value bool)
}

type maxConnectAttemptsSetter interface {
	SetMaxConnectAttempts(value uint)
}

type maxReconnectAttemptsSetter interface {
	SetMaxReconnectAttempts(value uint)
}

type reconnectInitialDelaySetter interface {
	SetReconnectInitialDelay(value time.Duration)
}

type connectEventChSetter interface {
	SetConnectEventCh(value chan *apifabclient.ConnectionEvent)
}

type timeBetweenConnectAttemptsSetter interface {
	SetTimeBetweenConnectAttempts(value time.Duration)
}

type responseTimeoutSetter interface {
	SetResponseTimeout(value time.Duration)
}
