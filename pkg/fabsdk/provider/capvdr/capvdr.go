/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package capvdr

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/ca"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/identity/caclient"
	"github.com/pkg/errors"
)

// CAProvider represents the default implementation of identity provider.
type CAProvider struct {
	ctx core.Providers
}

// New creates an CAProvider.
func New(ctx core.Providers) *CAProvider {

	provider := CAProvider{
		ctx: ctx,
	}
	return &provider
}

// CreateCAService returns an enrollment service for the given organization.
func (p *CAProvider) CreateCAService(orgName string) (ca.Client, error) {

	im, err := caclient.New(orgName, p.ctx.IdentityManager(), p.ctx.StateStore(), p.ctx.CryptoSuite(), p.ctx.Config())
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create enrollment service for organization: %s", orgName)
	}

	return im, nil
}
