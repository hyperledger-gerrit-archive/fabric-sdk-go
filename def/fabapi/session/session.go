/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package session

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric/bccsp"
)

// Session represents an identity being used with clients.
// TODO: Better description.
type Session struct {
	factory ProviderFactory
	user    fab.User
}

// NewSession creates a session from a context and a user (TODO)
func NewSession(user fab.User, factory ProviderFactory) *Session {
	s := Session{
		factory: factory,
		user:    user,
	}

	return &s
}

type sdkcxt interface {
	CryptoSuiteProvider() bccsp.BCCSP
	StateStoreProvider() fab.KeyValueStore
	ConfigProvider() apiconfig.Config
}

// ProviderFactory allows overriding default clients and providers of an SDK session
// TODO: Change to a context & session interface
type ProviderFactory interface {
	NewSystemClient(context sdkcxt, session Session, config apiconfig.Config) fab.FabricClient
	//NewChannelClient(session Session) fab.Channel
}

type DefaultSessionFactory struct{}

func NewDefaultSessionFactory() *DefaultSessionFactory {
	f := DefaultSessionFactory{}
	return &f
}

func (f *DefaultSessionFactory) NewSystemClient(sdk sdkcxt, session Session, config apiconfig.Config) fab.FabricClient {
	client := clientImpl.NewClient(config)

	client.SetCryptoSuite(sdk.CryptoSuiteProvider())
	client.SetStateStore(sdk.StateStoreProvider())
	client.SetUserContext(session.user)

	return client
}
