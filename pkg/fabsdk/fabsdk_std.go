// +build !pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabsdk enables client usage of a Hyperledger Fabric network.
package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/operations"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/metrics"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
)

func (sdk *FabricSDK) extraInit(cfg *configs, userStore msp.UserStore, signingManager core.SigningManager, identityManagerProvider msp.IdentityManagerProvider,
	infraProvider fab.InfraProvider, localDiscoveryProvider fab.LocalDiscoveryProvider, channelProvider fab.ChannelProvider) error {

	//disabled metrics for standard build
	sdk.system = operations.NewSystem(operations.Options{
		Metrics: operations.MetricsOptions{
			Provider: "disabled",
		},
		Version: "latest",
	},
	)

	sdk.clientMetrics = &metrics.ClientMetrics{} // empty channel ClientMetrics for standard build.

	return nil
}
