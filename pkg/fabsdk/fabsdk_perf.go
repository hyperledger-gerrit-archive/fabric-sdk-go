// +build pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabsdk enables client usage of a Hyperledger Fabric network.
package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/operations"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/metrics"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/pkg/errors"
)

// FabricSDK provides access (and context) to clients being managed by the SDK.
type FabricSDK struct {
	opts          options
	provider      *context.Provider
	cryptoSuite   core.CryptoSuite
	cfgLookup     *lookup.ConfigLookup
	system        *operations.System
	clientMetrics *metrics.ClientMetrics
}

func (sdk *FabricSDK) extraInit(cfg *configs, userStore msp.UserStore, signingManager core.SigningManager, identityManagerProvider msp.IdentityManagerProvider,
	infraProvider fab.InfraProvider, localDiscoveryProvider fab.LocalDiscoveryProvider, channelProvider fab.ChannelProvider) (*context.Provider, error) {
	if sdk.opts.ConfigBackend == nil {
		return nil, errors.New("unable to find config backend")
	}
	if sdk.cfgLookup == nil {
		sdk.cfgLookup = lookup.New(sdk.opts.ConfigBackend...)
	}
	sdk.initMetrics(sdk.cfgLookup)

	//update sdk providers list since all required providers are initialized
	sdk.provider = context.NewProvider(context.WithCryptoSuiteConfig(cfg.cryptoSuiteConfig),
		context.WithEndpointConfig(cfg.endpointConfig),
		context.WithIdentityConfig(cfg.identityConfig),
		context.WithCryptoSuite(sdk.cryptoSuite),
		context.WithSigningManager(signingManager),
		context.WithUserStore(userStore),
		context.WithLocalDiscoveryProvider(localDiscoveryProvider),
		context.WithIdentityManagerProvider(identityManagerProvider),
		context.WithInfraProvider(infraProvider),
		context.WithChannelProvider(channelProvider),
		context.WithClientMetrics(sdk.clientMetrics),
	)

	return sdk.provider, nil
}

//Config returns config backend used by all SDK config types
func (sdk *FabricSDK) Config() (core.ConfigBackend, error) {
	return sdk.cfgLookup, nil
}
