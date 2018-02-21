/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/pkg/options"
	"google.golang.org/grpc/keepalive"
)

type params struct {
	hostOverride    string
	certificate     *x509.Certificate
	keepAliveParams keepalive.ClientParameters
	failFast        bool
}

func defaultParams() *params {
	return &params{
		failFast: true,
	}
}

// WithHostOverride sets the host name that will be used to resolve the TLS certificate
func WithHostOverride(value string) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(hostOverrideSetter); ok {
			logger.Debugf("Applying option HostOverride: %s", value)
			setter.SetHostOverride(value)
		}
	}
}

// WithCertificate sets the X509 certificate used for the TLS connection
func WithCertificate(value *x509.Certificate) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(certificateSetter); ok {
			logger.Debugf("Applying option Certificate: %s", value)
			setter.SetCertificate(value)
		}
	}
}

// WithKeepAliveParams sets the GRPC keep-alive parameters
func WithKeepAliveParams(value keepalive.ClientParameters) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(keepAliveParamsSetter); ok {
			logger.Debugf("Applying option KeepAliveParams: %#v", value)
			setter.SetKeepAliveParams(value)
		}
	}
}

// WithFailFast sets the GRPC fail-fast parameter
func WithFailFast(value bool) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(failFastSetter); ok {
			logger.Debugf("Applying option FailFast: %t", value)
			setter.SetFailFast(value)
		}
	}
}

func (p *params) SetHostOverride(value string) {
	p.hostOverride = value
}

func (p *params) SetCertificate(value *x509.Certificate) {
	p.certificate = value
}

func (p *params) SetKeepAliveParams(value keepalive.ClientParameters) {
	p.keepAliveParams = value
}

func (p *params) SetFailFast(value bool) {
	p.failFast = value
}

type hostOverrideSetter interface {
	SetHostOverride(value string)
}

type certificateSetter interface {
	SetCertificate(value *x509.Certificate)
}

type keepAliveParamsSetter interface {
	SetKeepAliveParams(value keepalive.ClientParameters)
}

type failFastSetter interface {
	SetFailFast(value bool)
}
