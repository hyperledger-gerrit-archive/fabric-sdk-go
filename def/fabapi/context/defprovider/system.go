/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defprovider

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
)

// FabricSystemFactory represents the default implementation of a session client.
type FabricSystemFactory struct{}

// NewFabricSystemFactory creates a new default session client factory.
func NewFabricSystemFactory() *FabricSystemFactory {
	f := FabricSystemFactory{}
	return &f
}

// NewClient returns a new FabricClient.
func (f *FabricSystemFactory) NewClient(sdk context.SDK, session context.Session, config apiconfig.Config) (apifabclient.FabricClient, error) {
	client := clientImpl.NewClient(config)

	client.SetCryptoSuite(sdk.CryptoSuiteProvider())
	client.SetStateStore(sdk.StateStoreProvider())
	client.SetUserContext(session.Identity())
	client.SetSigningManager(sdk.SigningManager())

	return client, nil
}
