/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msppvdr

import (
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	mspimpl "github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk")

// MspProvider provides the default implementation of MSP
type MspProvider struct {
	providerContext core.Providers
	identityManager map[string]msp.IdentityManager
}

// New creates a MSP context provider
func New(config core.Config, cryptoSuite core.CryptoSuite, stateStore core.KVStore) (*MspProvider, error) {

	identityManager := make(map[string]msp.IdentityManager)
	netConfig, err := config.NetworkConfig()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to retrieve network config")
	}
	for orgName := range netConfig.Organizations {
		mgr, err := mspimpl.NewIdentityManager(orgName, stateStore, cryptoSuite, config)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to initialize identity manager for organization: %s", orgName)
		}
		identityManager[orgName] = mgr
	}

	mspProvider := MspProvider{
		identityManager: identityManager,
	}

	return &mspProvider, nil
}

// Initialize sets the provider context
func (p *MspProvider) Initialize(providers core.Providers) error {
	p.providerContext = providers
	return nil
}

// IdentityManager returns the organization's identity manager
func (p *MspProvider) IdentityManager(orgName string) (msp.IdentityManager, bool) {
	im, ok := p.identityManager[strings.ToLower(orgName)]
	if !ok {
		return nil, false
	}
	return im, true
}
