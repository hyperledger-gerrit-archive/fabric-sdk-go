/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/pkg/errors"
)

// IdentityContext supplies the serialized identity and key reference.
//
// TODO - refactor SigningIdentity and this interface.
type IdentityContext interface {
	MspID() string
	Identity() ([]byte, error)
	PrivateKey() core.Key
}

// SessionContext primarily represents the session and identity context
type SessionContext interface {
	IdentityContext
}

// Context supplies the configuration and signing identity to client objects.
type Context interface {
	ProviderContext
	IdentityContext
}

// ProviderContext supplies the configuration to client objects.
type ProviderContext interface {
	SigningManager() api.SigningManager
	Config() core.Config
	CryptoSuite() core.CryptoSuite
}

//Option provides parameters for creating context
type Option func(s *Options, orgName string) error

//Options supplies identity options to caller
type Options struct {
	orgID    string
	identity IdentityContext
	ok       bool
}

// New returns identity context
func New(orgName string, options ...Option) (IdentityContext, error) {
	opts := Options{}

	for _, option := range options {
		err := option(&opts, orgName)
		if err != nil {
			return nil, errors.WithMessage(err, "Error in option passed to client")
		}
	}
	if !opts.ok {
		return nil, errors.New("Missing identity")
	}
	return opts.identity, nil
}

// WithIdentity returns option with identity
func WithIdentity(identity IdentityContext) Option {
	return func(o *Options, orgName string) error {
		if o.ok {
			return errors.New("Identity already determined")
		}
		o.identity = identity
		o.ok = true
		return nil
	}
}

// WithOrg returns option with orgID
func WithOrg(orgName string) Option {
	return func(o *Options, orgName string) error {
		o.orgID = orgName
		return nil
	}
}
