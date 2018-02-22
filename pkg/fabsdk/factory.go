/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/chclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/api"
)

// PkgSuite provides the package factories that create clients and providers
type PkgSuite interface {
	Core() (context.CoreProviderFactory, error)
	Service() (context.ServiceProviderFactory, error)
	Context() (context.OrgClientFactory, error)
	Session() (SessionClientFactory, error)
	Logger() (api.LoggerProvider, error)
}

// SessionClientFactory allows overriding default clients and providers of a session
type SessionClientFactory interface {
	NewChannelClient(sdk context.Providers, session context.SessionContext, channelID string, targetFilter context.TargetFilter) (*chclient.ChannelClient, error)
}
