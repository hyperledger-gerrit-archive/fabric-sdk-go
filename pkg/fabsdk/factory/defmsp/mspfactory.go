/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defmsp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/msppvdr"
	mspimpl "github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/pkg/errors"
)

// MSPProviderFactory represents the default MSP provider factory.
type MSPProviderFactory struct {
}

// NewMSPProviderFactory returns the default MSP provider factory.
func NewMSPProviderFactory() *MSPProviderFactory {
	f := MSPProviderFactory{}
	return &f
}

// CreateUserStore creates a UserStore using the SDK's default implementation
func (f *MSPProviderFactory) CreateUserStore(config core.Config) (msp.UserStore, error) {

	clientCofig, err := config.Client()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to retrieve client config")
	}
	stateStorePath := clientCofig.CredentialStore.Path

	stateStore, err := kvs.New(&kvs.FileKeyValueStoreOptions{Path: stateStorePath})
	if err != nil {
		return nil, errors.WithMessage(err, "CreateNewFileKeyValueStore failed")
	}

	userStore, err := mspimpl.NewCertFileUserStore1(stateStore)
	if err != nil {
		return nil, errors.Wrapf(err, "creating a user store failed")
	}

	return userStore, nil
}

// CreateIdentityManagerProvider returns a new default implementation of MSP provider
func (f *MSPProviderFactory) CreateIdentityManagerProvider(config core.Config, cryptoProvider core.CryptoSuite, userStore msp.UserStore) (msp.IdentityManagerProvider, error) {
	return msppvdr.New(config, cryptoProvider, userStore)
}
