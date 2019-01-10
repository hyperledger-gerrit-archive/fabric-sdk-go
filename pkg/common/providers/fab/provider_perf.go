// +build pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/metrics"

// Providers represents the SDK configured service providers context.
type Providers interface {
	LocalDiscoveryProvider() LocalDiscoveryProvider
	ChannelProvider() ChannelProvider
	InfraProvider() InfraProvider
	EndpointConfig() EndpointConfig
	MetricsProvider
}

// MetricsProvider represents a provider of metrics.
type MetricsProvider interface {
	GetMetrics() *metrics.ClientMetrics
}
