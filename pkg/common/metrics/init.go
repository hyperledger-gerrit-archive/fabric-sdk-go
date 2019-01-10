/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package metrics provides the ability to gather metrics information of a client.
// for now, since metrics are only used in "channel client", creating the system operations instance (the same way as in Fabric)
// is only configured for this specific client.
// For additional metrics elsewhere in the SDK, add new metrics here and add metrics calls as in:
// fabric-sdk-go/pkg/client/channel/chclientrun_perf.go
// this package assumes the same peer configs are available in the SDK (code copied from fabric/peer/node/start.go)
package metrics

import (
	"strings"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/operations"
	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
)

var system *operations.System

// InitMetrics will initialize the Go SDK's metric's system instance to allow capturing metrics data by the SDK.
func InitMetrics(configLookup *lookup.ConfigLookup) {
	if configLookup == nil {
		return
	}
	if system == nil {
		system = newOperationsSystem(configLookup)

		SdkMetrics = NewClientMetrics(system.Provider)

		err := system.Start()
		if err != nil {
			panic("metrics failed to start: " + err.Error())
		}
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
