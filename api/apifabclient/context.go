/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
)

// Context supplies the configuration and signing identity to client objects.
type Context interface {
	ProviderContext
	IdentityContext
}

// ProviderContext supplies the configuration to client objects.
type ProviderContext interface {
	SigningManager() SigningManager
	Config() config.Config
	CryptoSuite() apicryptosuite.CryptoSuite
}

// ChannelService supplies Channel objects for the named channel.
type ChannelService interface {
	Channel(name string) (Channel, error)
	EventHub(name string) (EventHub, error) // TODO support new event delivery
}
