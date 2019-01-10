// +build pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"strings"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/operations"
	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/metrics"
)

// initMetrics will initialize the Go SDK's metric's system instance to allow capturing metrics data by the SDK clients.
func (sdk *FabricSDK) initMetrics(configLookup *lookup.ConfigLookup) {
	if configLookup == nil {
		return
	}
	if sdk.system == nil {
		sdk.system = newOperationsSystem(configLookup)

		err := sdk.system.Start()
		if err != nil {
			panic("metrics failed to start: " + err.Error())
		}

		// for now NewClientMetrics supports channel client. TODO: if other client types require metrics tracking, update this function
		sdk.clientMetrics = metrics.NewClientMetrics(sdk.system.Provider)
	}
}

func newOperationsSystem(configLookup *lookup.ConfigLookup) *operations.System {
	c := configLookup.GetString("operations.tls.clientRootCAs.files")
	caCertFiles := strings.Split(c, ",")
	return operations.NewSystem(operations.Options{
		Logger:        flogging.MustGetLogger("operations.runner"),
		ListenAddress: configLookup.GetString("operations.listenAddress"),
		Metrics: operations.MetricsOptions{
			Provider: configLookup.GetString("metrics.provider"),
			Statsd: &operations.Statsd{
				Network:       configLookup.GetString("metrics.statsd.network"),
				Address:       configLookup.GetString("metrics.statsd.address"),
				WriteInterval: configLookup.GetDuration("metrics.statsd.writeInterval"),
				Prefix:        configLookup.GetString("metrics.statsd.prefix"),
			},
		},
		TLS: operations.TLS{
			Enabled:            configLookup.GetBool("operations.tls.enabled"),
			CertFile:           configLookup.GetString("operations.tls.cert.file"),
			KeyFile:            configLookup.GetString("operations.tls.key.file"),
			ClientCertRequired: configLookup.GetBool("operations.tls.clientAuthRequired"),
			ClientCACertFiles:  caCertFiles, // TODO expose and use configLookup.GetStringSlice here instead
		},
		Version: "latest", // TODO expose version somewhere, Fabric uses 'metadata.Version'
	})
}
