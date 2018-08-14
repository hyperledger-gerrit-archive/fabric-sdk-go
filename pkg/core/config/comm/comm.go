/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"crypto/tls"

	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/pkg/errors"
)

// TLSConfig returns the appropriate config for TLS including the root CAs,
// certs for mutual TLS, and server host override. Works with certs loaded either from a path or embedded pem.
func TLSConfig(cert *x509.Certificate, serverName string, config fab.EndpointConfig) (*tls.Config, error) {

	if cert != nil {
		config.TLSCACertPool().Add(cert)
	}

	certPool, err := config.TLSCACertPool().Get()
	if err != nil {
		return nil, err
	}
	return &tls.Config{RootCAs: certPool, Certificates: config.TLSClientCerts(), ServerName: serverName}, nil
}

// TLSCertHash is a utility method to calculate the SHA256 hash of the configured certificate (for usage in channel headers)
func TLSCertHash(config fab.EndpointConfig) ([]byte, bool, error) {
	certs := config.TLSClientCerts()
	if len(certs) == 0 {
		return nil, true, nil
	}

	cert := certs[0]
	if len(cert.Certificate) == 0 {
		return nil, true, nil
	}

	h, err := cryptosuite.GetDefault().Hash(cert.Certificate[0], cryptosuite.GetSHA256Opts())
	if err != nil {
		return nil, false, errors.WithMessage(err, "failed to compute tls cert hash")
	}
	return h, true, nil
}
