/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defprovider

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	discovery "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/discovery/staticdiscovery"
	selection "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/selection/staticselection"
)

// DefaultProviderFactory represents the default SDK provider factory.
type DefaultProviderFactory struct{}

// NewDefaultProviderFactory returns the default SDK provider factory.
func NewDefaultProviderFactory() *DefaultProviderFactory {
	f := DefaultProviderFactory{}
	return &f
}

// NewDiscoveryProvider returns a new default implementation of discovery provider
func (f *DefaultProviderFactory) NewDiscoveryProvider(config apiconfig.Config) (fab.DiscoveryProvider, error) {
	return discovery.NewDiscoveryProvider(config)
}

// NewSelectionProvider returns a new default implementation of selection service
func (f *DefaultProviderFactory) NewSelectionProvider(config apiconfig.Config) (fab.SelectionProvider, error) {
	return selection.NewSelectionProvider(config)
}
