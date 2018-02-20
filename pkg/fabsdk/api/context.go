/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api/core"
)

// Providers represents the SDK configured providers context.
type Providers interface {
	CoreProviders
	SvcProviders
}

// CoreProviders represents the SDK configured core providers context.
type CoreProviders interface {
	CryptoSuite() core.CryptoSuite
	StateStore() core.KVStore
	Config() apiconfig.Config
	SigningManager() fab.SigningManager
	FabricProvider() FabricProvider
}

// SvcProviders represents the SDK configured service providers context.
type SvcProviders interface {
	DiscoveryProvider() fab.DiscoveryProvider
	SelectionProvider() fab.SelectionProvider
	ChannelProvider() fab.ChannelProvider
}
