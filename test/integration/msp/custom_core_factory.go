/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
)

// ========== MSP Provider Factory with custom crypto provider ============= //

type customCoreFactory struct {
	defaultFactory    *defcore.ProviderFactory
	customCryptoSuite core.CryptoSuite
}

func NewCustomCoreFactory(customCryptoSuite core.CryptoSuite) *customCoreFactory {
	return &customCoreFactory{defaultFactory: defcore.NewProviderFactory(), customCryptoSuite: customCryptoSuite}
}

func (f *customCoreFactory) CreateCryptoSuiteProvider(config core.Config) (core.CryptoSuite, error) {
	return f.customCryptoSuite, nil
}

func (f *customCoreFactory) CreateSigningManager(cryptoProvider core.CryptoSuite, config core.Config) (core.SigningManager, error) {
	return f.defaultFactory.CreateSigningManager(cryptoProvider, config)
}

func (f *customCoreFactory) CreateInfraProvider(config core.Config) (fab.InfraProvider, error) {
	return f.defaultFactory.CreateInfraProvider(config)
}
