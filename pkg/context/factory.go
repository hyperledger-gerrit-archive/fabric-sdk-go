/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/apicryptosuite"
)

// CoreProviderFactory allows overriding of primitives and the fabric core object provider
type CoreProviderFactory interface {
	NewStateStoreProvider(config apiconfig.Config) (KVStore, error)
	NewCryptoSuiteProvider(config apiconfig.Config) (apicryptosuite.CryptoSuite, error)
	NewSigningManager(cryptoProvider apicryptosuite.CryptoSuite, config apiconfig.Config) (SigningManager, error)
	NewFabricProvider(context ProviderContext) (FabricProvider, error)
}

// ServiceProviderFactory allows overriding default service providers (such as peer discovery)
type ServiceProviderFactory interface {
	NewDiscoveryProvider(config apiconfig.Config) (DiscoveryProvider, error)
	NewSelectionProvider(config apiconfig.Config) (SelectionProvider, error)
	//	NewChannelProvider(ctx Context, channelID string) (ChannelProvider, error)
}

// OrgClientFactory allows overriding default clients and providers of an organization
// Currently, a context is created for each organization that the client app needs.
type OrgClientFactory interface {
	//NewMSPClient(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (fabca.FabricCAClient, error)
	NewCredentialManager(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (CredentialManager, error)
}
