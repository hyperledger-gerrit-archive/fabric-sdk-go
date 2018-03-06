/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
)

// Session primarily represents the session and identity context
type Session interface {
	Identity
}

// Identity supplies the serialized identity and key reference.
type Identity interface {
	MspID() string
	SerializedIdentity() ([]byte, error)
	PrivateKey() core.Key
}

// Client supplies the configuration and signing identity to client objects.
type Client interface {
	Providers
	Identity
}

// Providers represents the SDK configured providers context.
type Providers interface {
	core.Providers
	fab.Providers
}

//Channel supplies the configuration for channel context client
type Channel interface {
	Client
	DiscoveryService() fab.DiscoveryService
	SelectionService() fab.SelectionService
	ChannelService() fab.ChannelService
}

//ClientProvider returns client context
type ClientProvider func() (Client, error)

//ChannelProvider returns channel client context
type ChannelProvider func() (Channel, error)
