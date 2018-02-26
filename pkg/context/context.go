/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	//this import causes cyclic-dep
	//providers "github.com/hyperledger/fabric-sdk-go/pkg/context/providers"
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

//Context supplies the configuration and signing identity to client objects.
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

//Providers represents the SDK configured providers context.
//QQQ Use an empty interface just to check the build without cyclic-deps
type Providers interface {
}

//CoreOption provides parameters for creating core
type CoreOption func(opts *coreOptions) error

//ClientOption provides parameters for creating core
type ClientOption func(opts *clientOptions) error

//Options supplies identity options to caller
type coreOptions struct {
	providers Providers
}

//Options supplies identity options to caller
type clientOptions struct {
	orgID    string
	identity IdentityContext
	ok       bool
}

//Core returns core context
type Core struct {
	Providers Providers
}

//Client ty
type Client struct {
	Core
	IdentityContext
}

// New returns identity context takes providers
func New(options ...CoreOption) (*Core, error) {
	opts := coreOptions{} //

	for _, option := range options {
		err := option(&opts)
		if err != nil {
			return nil, errors.WithMessage(err, "Error in option passed to client")
		}
	}
	cc := Core{}
	cc.Providers = opts.providers

	//
	return &cc, nil
}

//NewClient new client
func NewClient(core *Core, options ...ClientOption) (*Client, error) {
	opts := clientOptions{}

	for _, option := range options {
		err := option(&opts)
		if err != nil {
			return nil, errors.WithMessage(err, "Error in option passed to client")
		}
	}
	if !opts.ok {
		return nil, errors.New("Missing identity")
	}
	cc := Client{}
	cc.Core = *core
	cc.IdentityContext = opts.identity

	//
	return &cc, nil
}

// WithIdentity returns option with identity
func WithIdentity(identity IdentityContext) ClientOption {
	return func(o *clientOptions) error {
		if o.ok {
			return errors.New("Identity already determined")
		}
		o.identity = identity
		o.ok = true
		return nil
	}
}

// WithOrg returns option with orgID
func WithOrg(orgName string) ClientOption {
	return func(o *clientOptions) error {
		o.orgID = orgName
		return nil
	}
}

// WithProvider returns option with provider
func WithProvider(provider Providers) CoreOption {
	return func(o *coreOptions) error {
		o.providers = provider
		return nil
	}
}
