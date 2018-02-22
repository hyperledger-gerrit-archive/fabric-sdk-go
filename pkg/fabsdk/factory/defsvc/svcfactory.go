/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defsvc

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/apiconfig"

	discovery "github.com/hyperledger/fabric-sdk-go/pkg/client/discovery/staticdiscovery"
	selection "github.com/hyperledger/fabric-sdk-go/pkg/client/selection/staticselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
)

// ProviderFactory represents the default SDK provider factory for services.
type ProviderFactory struct{}

// NewProviderFactory returns the default SDK provider factory for services.
func NewProviderFactory() *ProviderFactory {
	f := ProviderFactory{}
	return &f
}

// NewDiscoveryProvider returns a new default implementation of discovery provider
func (f *ProviderFactory) NewDiscoveryProvider(config apiconfig.Config) (context.DiscoveryProvider, error) {
	return discovery.NewDiscoveryProvider(config)
}

// NewSelectionProvider returns a new default implementation of selection service
func (f *ProviderFactory) NewSelectionProvider(config apiconfig.Config) (context.SelectionProvider, error) {
	return selection.NewSelectionProvider(config)
}
