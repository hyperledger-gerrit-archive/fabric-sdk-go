/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabapi provides a default implementation of the fabric API for fabsdk
package fabapi

import (
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/defprovider"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
)

// NewSDKOptions returns SDK options populated with the default implementation referenced by the fabapi package
func NewSDKOptions() fabsdk.Options {
	opts := fabsdk.Options{}

	PopulateSDKOptions(&opts)

	return opts
}

// PopulateSDKOptions populates an SDK options with the default implementation referenced by the fabapi package
func PopulateSDKOptions(opts *fabsdk.Options) {
	if opts.LoggerFactory == nil {
		opts.LoggerFactory = deflogger.LoggerProvider()
	}
	if opts.ProviderFactory == nil {
		opts.ProviderFactory = defprovider.NewDefaultProviderFactory()
	}
	if opts.ContextFactory == nil {
		opts.ContextFactory = defprovider.NewOrgClientFactory()
	}
	if opts.SessionFactory == nil {
		opts.SessionFactory = defprovider.NewSessionClientFactory()
	}
	if opts.FabricSystemFactory == nil {
		opts.FabricSystemFactory = defprovider.NewFabricSystemFactory()
	}
}
