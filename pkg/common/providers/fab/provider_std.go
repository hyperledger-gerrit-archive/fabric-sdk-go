// +build !pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

// Providers represents the SDK configured service providers context.
type Providers interface {
	LocalDiscoveryProvider() LocalDiscoveryProvider
	ChannelProvider() ChannelProvider
	InfraProvider() InfraProvider
	EndpointConfig() EndpointConfig
}
