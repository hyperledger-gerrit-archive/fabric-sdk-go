/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricselection

import (
	"time"

	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
)

type params struct {
	refreshInterval time.Duration
	responseTimeout time.Duration
	retryInterval   time.Duration
	maxRetries      int
}

// WithRefreshInterval sets the interval in which the
// peer cache is refreshed
func WithRefreshInterval(value time.Duration) coptions.Opt {
	return func(p coptions.Params) {
		logger.Debug("Checking refreshIntervalSetter")
		if setter, ok := p.(refreshIntervalSetter); ok {
			setter.SetRefreshInterval(value)
		}
	}
}

// WithResponseTimeout sets the Discover service response timeout
func WithResponseTimeout(value time.Duration) coptions.Opt {
	return func(p coptions.Params) {
		logger.Debug("Checking responseTimeoutSetter")
		if setter, ok := p.(responseTimeoutSetter); ok {
			setter.SetResponseTimeout(value)
		}
	}
}

// WithRetryInterval sets the interval in which requests to the discovery
// service will be retried in case of a transient error
func WithRetryInterval(value time.Duration) coptions.Opt {
	return func(p coptions.Params) {
		logger.Debug("Checking retryIntervalSetter")
		if setter, ok := p.(retryIntervalSetter); ok {
			setter.SetRetryInterval(value)
		}
	}
}

// WithMaxRetries sets the maximum number of times a request to the discovery
// service will be retried in case of a transient error
func WithMaxRetries(value int) coptions.Opt {
	return func(p coptions.Params) {
		logger.Debug("Checking maxRetriesSetter")
		if setter, ok := p.(maxRetriesSetter); ok {
			setter.SetMaxRetries(value)
		}
	}
}

type refreshIntervalSetter interface {
	SetRefreshInterval(value time.Duration)
}

type responseTimeoutSetter interface {
	SetResponseTimeout(value time.Duration)
}

type retryIntervalSetter interface {
	SetRetryInterval(value time.Duration)
}

type maxRetriesSetter interface {
	SetMaxRetries(value int)
}

func (o *params) SetRefreshInterval(value time.Duration) {
	logger.Debugf("RefreshInterval: %s", value)
	o.refreshInterval = value
}

func (o *params) SetResponseTimeout(value time.Duration) {
	logger.Debugf("ResponseTimeout: %s", value)
	o.responseTimeout = value
}

func (o *params) SetRetryInterval(value time.Duration) {
	logger.Debugf("RetryInterval: %s", value)
	o.retryInterval = value
}

func (o *params) SetMaxRetries(value int) {
	logger.Debugf("MaxRetries: %d", value)
	o.maxRetries = value
}
