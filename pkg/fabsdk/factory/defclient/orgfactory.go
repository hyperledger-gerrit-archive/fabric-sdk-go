/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defclient

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/core/identity"
	credentialMgr "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/credentialmgr"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/idmgmtclient"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
)

// OrgClientFactory represents the default org provider factory.
type OrgClientFactory struct{}

// NewOrgClientFactory returns the default org provider factory.
func NewOrgClientFactory() *OrgClientFactory {
	f := OrgClientFactory{}
	return &f
}

// NewIdentityManager returns a client that manages identities on the network
func (f *OrgClientFactory) NewIdentityManager(providers apisdk.Providers, org string) (identity.Manager, error) {

	ctx := idmgmtclient.Context{
		ProviderContext: providers,
		MspID:           org,
	}
	return idmgmtclient.New(ctx)
}

/*
// NewMSPClient returns a new default implementation of the MSP client
// TODO: duplicate of core factory method (remove one) or at least call the core one like in sessfactory
func (f *OrgClientFactory) NewMSPClient(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (fabca.FabricCAClient, error) {
	return fabricCAClient.NewFabricCAClient(orgName, config, cryptoProvider)
}
*/

// NewCredentialManager returns a new default implementation of the credential manager
func (f *OrgClientFactory) NewCredentialManager(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (fab.CredentialManager, error) {
	return credentialMgr.NewCredentialManager(orgName, config, cryptoProvider)
}
