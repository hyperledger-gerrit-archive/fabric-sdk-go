/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabapi enables client usage of a Hyperledger Fabric network
package fabapi

import (
	"fmt"

	"github.com/hyperledger/fabric/bccsp"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabca"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// Options encapsulates configuration for the SDK
type Options struct {
	// Quick access options
	ConfigFile string
	OrgID      string

	// Options for default providers
	ConfigOpts     ConfigOpts
	StateStoreOpts StateStoreOpts

	// Implementations of client functionality (defaults are used if not specified)
	Config       apiconfig.Config
	FabricClient apifabclient.FabricClient
	MSPClient    apifabca.FabricCAClient
	StateStore   apifabclient.KeyValueStore
	CryptoSuite  bccsp.BCCSP // TODO - maybe copy this interface into the API package

	// TODO extract hard-coded logger
}

// FabricSDK provides access to clients being managed by the SDK
type FabricSDK struct {
	Options
}

// NewSDK initializes default clients
func NewSDK(options Options) (*FabricSDK, error) {
	// Construct SDK opts from the quick access options in setup
	sdkOpts := SDKOpts{
		ConfigFile: options.ConfigFile,
	}

	sdk := FabricSDK{
		Options: options,
	}

	// Initialize default config provider
	if sdk.Config == nil {
		config, err := NewDefaultConfig(sdk.ConfigOpts, sdkOpts)
		if err != nil {
			return nil, fmt.Errorf("Failed to initialize default config [%s]", err)
		}
		sdk.Config = config
	}

	// Initialize default crypto provider
	if sdk.CryptoSuite == nil {
		cryptosuite, err := NewCryptoSuite(sdk.Config.CSPConfig())
		if err != nil {
			return nil, fmt.Errorf("Failed to initialize default crypto suite [%s]", err)
		}
		sdk.CryptoSuite = cryptosuite
	}

	// Initialize default state store
	if sdk.StateStore == nil {
		store, err := NewDefaultStateStore(sdk.StateStoreOpts, sdk.Config)
		if err != nil {
			return nil, fmt.Errorf("Failed to initialize default state store [%s]", err)
		}
		sdk.StateStore = store
	}

	// Initialize default CA client
	if sdk.MSPClient == nil {
		client, err := NewCAClient(sdk.OrgID, sdk.Config)
		if err != nil {
			return nil, fmt.Errorf("Failed to initialize default client [%s]", err)
		}
		sdk.MSPClient = client
	}

	// Initialize default fabric client
	if sdk.FabricClient == nil {
		client := newDefaultClient(sdk.Config)
		sdk.FabricClient = client
	}

	sdk.FabricClient.SetCryptoSuite(sdk.CryptoSuite)
	sdk.FabricClient.SetStateStore(sdk.StateStore)

	return &sdk, nil
}
