/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package idpvdr

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/identity/manager"
	"github.com/pkg/errors"
)

// IdentityProvider represents the default implementation of identity provider.
type IdentityProvider struct {
	ctx core.Providers
}

// New creates an IdentityProvider.
func New(ctx core.Providers) *IdentityProvider {

	provider := IdentityProvider{
		ctx: ctx,
	}
	return &provider
}

// CreateIdentityManager returns an identity manager for the given organization.
func (p *IdentityProvider) CreateIdentityManager(orgName string) (identity.IdentityManager, error) {

	im, err := manager.New(orgName, p.ctx.StateStore(), p.ctx.CryptoSuite(), p.ctx.Config())
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create identity manager for organization: %s", orgName)
	}

	return im, nil
}
