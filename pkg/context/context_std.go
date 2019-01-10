// +build !pprof

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
)

//Channel supplies the configuration for channel context client
type Channel struct {
	context.Client
	channelService fab.ChannelService
	channelID      string
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
}

func createChannel(client context.Client, channelService fab.ChannelService, channelID string) *Channel {
	return &Channel{
		Client:         client,
		channelService: channelService,
		channelID:      channelID,
	}
}
