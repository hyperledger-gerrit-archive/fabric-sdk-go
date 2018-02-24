/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defclient

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	credentialMgr "github.com/hyperledger/fabric-sdk-go/pkg/fab/credentialmgr"
)

// OrgClientFactory represents the default org provider factory.
type OrgClientFactory struct{}

// NewOrgClientFactory returns the default org provider factory.
func NewOrgClientFactory() *OrgClientFactory {
	f := OrgClientFactory{}
	return &f
}

/*
// NewMSPClient returns a new default implementation of the MSP client
// TODO: duplicate of core factory method (remove one) or at least call the core one like in sessfactory
func (f *OrgClientFactory) NewMSPClient(orgName string, config apiconfig.Config, cryptoProvider apicryptosuite.CryptoSuite) (fabca.IdentityManager, error) {
	return fabricCAClient.NewIdentityManager(orgName, config, cryptoProvider)
}
*/

// NewCredentialManager returns a new default implementation of the credential manager
func (f *OrgClientFactory) NewCredentialManager(orgName string, config core.Config, cryptoProvider core.CryptoSuite) (api.CredentialManager, error) {
	return credentialMgr.NewCredentialManager(orgName, config, cryptoProvider)
}
