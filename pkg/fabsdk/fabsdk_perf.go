// +build pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabsdk enables client usage of a Hyperledger Fabric network.
package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
)

func (sdk *FabricSDK) extraInit(cfg *configs, userStore msp.UserStore, signingManager core.SigningManager, identityManagerProvider msp.IdentityManagerProvider,
	infraProvider fab.InfraProvider, localDiscoveryProvider fab.LocalDiscoveryProvider, channelProvider fab.ChannelProvider) error {

	sdk.initMetrics(cfg)

	return nil
}
