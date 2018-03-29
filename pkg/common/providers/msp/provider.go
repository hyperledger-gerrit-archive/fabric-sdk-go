/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
)

// Context is the context required by MSP services
type Context interface {
	core.Providers
	Providers
}

// IdentityManagerProvider provides identity management services
type IdentityManagerProvider interface {
	IdentityManager(orgName string) (IdentityManager, bool)
}

//IdentityConfig contains identity configurations
type IdentityConfig interface {
	Client() (*ClientConfig, error)                 //msp
	CAConfig(org string) (*CAConfig, error)         //msp
	CAServerCertPems(org string) ([]string, error)  //msp
	CAServerCertPaths(org string) ([]string, error) //msp
	CAClientKeyPem(org string) (string, error)      //msp
	CAClientKeyPath(org string) (string, error)     //msp
	CAClientCertPem(org string) (string, error)     //msp
	CAClientCertPath(org string) (string, error)    //msp
	CAKeyStorePath() string                         //msp
	CredentialStorePath() string                    //msp
}

// ClientConfig provides the definition of the client configuration
type ClientConfig struct {
	Organization    string
	Logging         LoggingType
	CryptoConfig    CCType
	TLSCerts        MutualTLSConfig
	CredentialStore CredentialStoreType
}

// MutualTLSConfig Mutual TLS configurations
type MutualTLSConfig struct {
	Pem []string
	// Certfiles root certificates for TLS validation (Comma separated path list)
	Path string

	//Client TLS information
	Client TLSKeyPair
}

// TLSKeyPair contains the private key and certificate for TLS encryption
type TLSKeyPair struct {
	Key  endpoint.TLSConfig
	Cert endpoint.TLSConfig
}

// LoggingType defines the level of logging
type LoggingType struct {
	Level string
}

// CCType defines the path to crypto keys and certs
type CCType struct {
	Path string
}

// CredentialStoreType defines pluggable KV store properties
type CredentialStoreType struct {
	Path        string
	CryptoStore struct {
		Path string
	}
}

// EnrollCredentials holds credentials used for enrollment
type EnrollCredentials struct {
	EnrollID     string
	EnrollSecret string
}

// CAConfig defines a CA configuration
type CAConfig struct {
	URL        string
	TLSCACerts MutualTLSConfig
	Registrar  EnrollCredentials
	CAName     string
}

// Providers represents a provider of MSP service.
type Providers interface {
	UserStore() UserStore
	IdentityManagerProvider
	IdentityConfig() IdentityConfig
}
