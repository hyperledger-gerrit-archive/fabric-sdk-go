/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package metrics provides the ability to gather metrics information of a client.
// for now, since metrics are only used in channel client, creating the system operations instance (the same way as in Fabric)
// is only configured for this specific client.
// For additional metrics elsewhere in the SDK, add new metrics here and add metrics calls as in:
// fabric-sdk-go/pkg/client/channel/chclientrun_perf.go
// this package assumes the same peer configs are available in the SDK (code copied from fabric/peer/node/start.go)
package metrics

import (
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/operations"
	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	"github.com/spf13/viper"
)

var system *operations.System

func init() {
	system = newOperationsSystem()

	SdkMetrics = NewClientMetrics(system.Provider)

	err := system.Start()
	if err != nil {
		panic("metrics failed to started: " + err.Error())
	}
}

func newOperationsSystem() *operations.System {
	return operations.NewSystem(operations.Options{
		Logger:        flogging.MustGetLogger("peer.sdk.operations"),
		ListenAddress: viper.GetString("operations.listenAddress"),
		Metrics: operations.MetricsOptions{
			Provider: viper.GetString("metrics.provider"),
			Statsd: &operations.Statsd{
				Network:       viper.GetString("metrics.statsd.network"),
				Address:       viper.GetString("metrics.statsd.address"),
				WriteInterval: viper.GetDuration("metrics.statsd.writeInterval"),
				Prefix:        viper.GetString("metrics.statsd.prefix"),
			},
		},
		TLS: operations.TLS{
			Enabled:            viper.GetBool("operations.tls.enabled"),
			CertFile:           viper.GetString("operations.tls.cert.file"),
			KeyFile:            viper.GetString("operations.tls.key.file"),
			ClientCertRequired: viper.GetBool("operations.tls.clientAuthRequired"),
			ClientCACertFiles:  viper.GetStringSlice("operations.tls.clientRootCAs.files"),
		},
		Version: "latest",
	})
}
