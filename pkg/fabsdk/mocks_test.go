/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	idapi "github.com/hyperledger/fabric-sdk-go/pkg/context/api/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/api"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	sdkApi "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defidentity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/modlog"
	"github.com/pkg/errors"
)

type mockCorePkg struct {
	stateStore      core.KVStore
	cryptoSuite     core.CryptoSuite
	signingManager  core.SigningManager
	identityManager map[string]idapi.IdentityManager
	fabricProvider  fab.InfraProvider
}

func newMockCorePkg(config core.Config) (*mockCorePkg, error) {
	pkgSuite := defPkgSuite{}
	sdkcore, err := pkgSuite.Core()
	if err != nil {
		return nil, err
	}
	stateStore, err := sdkcore.CreateStateStoreProvider(config)
	if err != nil {
		return nil, err
	}
	cs, err := sdkcore.CreateCryptoSuiteProvider(config)
	if err != nil {
		return nil, err
	}
	sm, err := sdkcore.CreateSigningManager(cs, config)
	if err != nil {
		return nil, err
	}

	ctx := mocks.NewMockProviderContextCustom(config, cs, sm, stateStore)
	fp, err := sdkcore.CreateFabricProvider(ctx)
	if err != nil {
		return nil, err
	}

	c := mockCorePkg{
		stateStore:     stateStore,
		cryptoSuite:    cs,
		signingManager: sm,
		fabricProvider: fp,
	}

	return &c, nil
}

func (mc *mockCorePkg) CreateStateStoreProvider(config core.Config) (core.KVStore, error) {
	return mc.stateStore, nil
}

func (mc *mockCorePkg) CreateCryptoSuiteProvider(config core.Config) (core.CryptoSuite, error) {
	return mc.cryptoSuite, nil
}

func (mc *mockCorePkg) CreateSigningManager(cryptoProvider core.CryptoSuite, config core.Config) (core.SigningManager, error) {
	return mc.signingManager, nil
}

func (mc *mockCorePkg) CreateFabricProvider(ctx core.Providers) (fab.InfraProvider, error) {
	return mc.fabricProvider, nil
}

type mockPkgSuite struct {
	errOnCore    bool
	errOnService bool
	errOnSession bool
	errOnLogger  bool
}

func (ps *mockPkgSuite) Core() (sdkApi.CoreProviderFactory, error) {
	if ps.errOnCore {
		return nil, errors.New("Error")
	}
	return defcore.NewProviderFactory(), nil
}

func (ps *mockPkgSuite) Identity() (sdkApi.IdentityProviderFactory, error) {
	if ps.errOnCore {
		return nil, errors.New("Error")
	}
	return defidentity.NewProviderFactory(), nil
}

func (ps *mockPkgSuite) Service() (sdkApi.ServiceProviderFactory, error) {
	if ps.errOnService {
		return nil, errors.New("Error")
	}
	return defsvc.NewProviderFactory(), nil
}

func (ps *mockPkgSuite) Session() (sdkApi.SessionClientFactory, error) {
	if ps.errOnSession {
		return nil, errors.New("Error")
	}
	return defclient.NewSessionClientFactory(), nil
}

func (ps *mockPkgSuite) Logger() (api.LoggerProvider, error) {
	if ps.errOnLogger {
		return nil, errors.New("Error")
	}
	return modlog.LoggerProvider(), nil
}
