/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabca

import (
	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/api/core/identity"
)

// Context supplies the configuration client objects.
type Context interface {
	ProviderContext
}

// ProviderContext supplies the configuration to client objects.
type ProviderContext interface {
	Config() config.Config
	CryptoSuite() apicryptosuite.CryptoSuite
}

// IdentityProvider supplies Channel related-objects for the named channel.
type IdentityProvider interface {
	NewIdentityService(ic identity.Context) (IdentityService, error)
}

// IdentityService supplies services related to a channel.
type IdentityService interface {
	CAConfig(org string) (*config.CAConfig, error)
}
