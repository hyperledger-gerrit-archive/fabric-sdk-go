/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package defpkgsuite provides a default implementation of the fabric API for fabsdk
package defpkgsuite

import (
	"github.com/hyperledger/fabric-sdk-go/def/factory/defclient"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defsvc"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	sdkapi "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
)

// SDKOpt provides the default implementation for the SDK
// TODO: rename to FromPkgSuite?
func SDKOpt() fabsdk.Option {
	return fabsdk.FromPkgSuite(newPkgSuite())
}

// WithConfigFile sets the SDK to use the named file for loading configuration.
func WithConfigFile(name string) fabsdk.Option {
	return func(opts *fabsdk.Options) error {
		config, err := configImpl.InitConfig(name)

		if err != nil {
			return errors.WithMessage(err, "Unable to initialize configuration")
		}

		return fabsdk.WithConfig(config)(opts)
	}
}

// WithConfigRaw sets the SDK to load configuration from the passed bytes.
func WithConfigRaw(raw []byte, format string) fabsdk.Option {
	return func(opts *fabsdk.Options) error {
		config, err := configImpl.InitConfigFromBytes(raw, format)

		if err != nil {
			return errors.WithMessage(err, "Unable to initialize configuration")
		}

		return fabsdk.WithConfig(config)(opts)
	}
}

func newPkgSuite() sdkapi.PkgSuite {
	pkgSuite := sdkapi.PkgSuite{
		Core:    defcore.NewProviderFactory(),
		Service: defsvc.NewProviderFactory(),
		Context: defclient.NewOrgClientFactory(),
		Session: defclient.NewSessionClientFactory(),
		Logger:  deflogger.LoggerProvider(),
	}
	return pkgSuite
}
