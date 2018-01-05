/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabapi provides a default implementation of the fabric API for fabsdk
package fabapi

import (
	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context/defprovider"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/opt"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
)

// Options encapsulates configuration for the SDK
type Options struct {
	// Quick access options
	ConfigFile string
	ConfigByte []byte
	ConfigType string

	// Options for default providers
	ConfigOpts     opt.ConfigOpts
	StateStoreOpts opt.StateStoreOpts

	// Factories to create clients and providers
	ProviderFactory context.SDKProviderFactory
	ContextFactory  context.OrgClientFactory
	SessionFactory  context.SessionClientFactory
	LoggerFactory   apilogging.LoggerProvider
}

// NewOptions populates the SDK options with the default implementation referenced by the fabapi package
func NewOptions(opt Options) fabsdk.Options {
	sdkOpt := fabsdk.Options{
		ConfigFile:      opt.ConfigFile,
		ConfigByte:      opt.ConfigByte,
		ConfigType:      opt.ConfigType,
		ConfigOpts:      opt.ConfigOpts,
		StateStoreOpts:  opt.StateStoreOpts,
		ProviderFactory: opt.ProviderFactory,
		ContextFactory:  opt.ContextFactory,
		SessionFactory:  opt.SessionFactory,
		LoggerFactory:   opt.LoggerFactory,
	}

	if sdkOpt.LoggerFactory == nil {
		sdkOpt.LoggerFactory = deflogger.LoggerProvider()
	}
	if sdkOpt.ProviderFactory == nil {
		sdkOpt.ProviderFactory = defprovider.NewDefaultProviderFactory()
	}
	if sdkOpt.ContextFactory == nil {
		sdkOpt.ContextFactory = defprovider.NewOrgClientFactory()
	}
	if sdkOpt.SessionFactory == nil {
		sdkOpt.SessionFactory = defprovider.NewSessionClientFactory()
	}

	return sdkOpt
}
