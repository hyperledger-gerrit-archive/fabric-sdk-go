/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabapi provides a default implementation of the fabric API for fabsdk
package fabapi

import (
	"github.com/hyperledger/fabric-sdk-go/def/factory/defclient"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/def/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
)

// Impl provides the default implementation for the SDK
func Impl() fabsdk.SDKOption {
	impl := fabsdk.ImplFactory{
		Core:    defcore.NewProviderFactory(),
		Service: defsvc.NewProviderFactory(),
		Context: defclient.NewOrgClientFactory(),
		Session: defclient.NewSessionClientFactory(),
		Logger:  deflogger.LoggerProvider(),
	}
	return fabsdk.Impl(impl)
}
