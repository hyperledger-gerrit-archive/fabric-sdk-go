/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabapi enables client usage of a Hyperledger Fabric network
package fabapi

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// Setup encapsulates configuration for the SDK
type Setup struct {

	// Implementations of client functionality (defaults are used if not specified)
	Config apiconfig.Config
	Client apifabclient.FabricClient
}

// FabricSDK provides access to clients being managed by the SDK
type FabricSDK struct {
	Setup
}

// ConfigOpts provides bootstrap setup for Config
type ConfigOpts struct {
	ConfigFile string
}

// NewSDK initializes default clients
func NewSDK(setup Setup) (*FabricSDK, error) {
	sdk := FabricSDK{
		Setup: setup,
	}

	// Initialize defaults
	if sdk.Client == nil {
		client, err := NewClient(nil, false, "", sdk.Config)
		if err != nil {
			return nil, err
		}
		sdk.Client = client
	}

	return &sdk, nil
}
