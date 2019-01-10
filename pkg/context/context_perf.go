// +build pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/metrics"
)

//Channel supplies the configuration for channel context client
type Channel struct {
	context.Client
	channelService fab.ChannelService
	channelID      string
	metrics        *metrics.ClientMetrics
}

//Provider implementation of Providers interface
type Provider struct {
	cryptoSuiteConfig      core.CryptoSuiteConfig
	endpointConfig         fab.EndpointConfig
	identityConfig         msp.IdentityConfig
	userStore              msp.UserStore
	cryptoSuite            core.CryptoSuite
	localDiscoveryProvider fab.LocalDiscoveryProvider
	signingManager         core.SigningManager
	idMgmtProvider         msp.IdentityManagerProvider
	infraProvider          fab.InfraProvider
	channelProvider        fab.ChannelProvider
	clientMetrics          *metrics.ClientMetrics
}

// GetMetrics will return the SDK's metrics instance
func (c *Provider) GetMetrics() *metrics.ClientMetrics {
	return c.clientMetrics
}

//WithClientMetrics sets clientMetrics to Context Provider
func WithClientMetrics(cm *metrics.ClientMetrics) SDKContextParams {
	return func(ctx *Provider) {
		ctx.clientMetrics = cm
	}
}

func createChannel(client context.Client, channelService fab.ChannelService, channelID string) *Channel {
	return &Channel{
		Client:         client,
		channelService: channelService,
		channelID:      channelID,
		metrics:        client.GetMetrics(),
	}
}
