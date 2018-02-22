/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/apiconfig"
)

// FabricProvider enables access to fabric objects such as peer and user based on config or
type FabricProvider interface {
	CreateChannelClient(user IdentityContext, cfg ChannelCfg) (Channel, error)
	CreateChannelLedger(ic IdentityContext, name string) (ChannelLedger, error)
	CreateChannelConfig(user IdentityContext, name string) (ChannelConfig, error)
	CreateResourceClient(user IdentityContext) (Resource, error)
	CreateChannelTransactor(ic IdentityContext, cfg ChannelCfg) (Transactor, error)
	CreateEventHub(ic IdentityContext, name string) (EventHub, error)
	CreateCAClient(orgID string) (FabricCAClient, error)

	CreatePeerFromConfig(peerCfg *apiconfig.NetworkPeer) (Peer, error)
	CreateOrdererFromConfig(cfg *apiconfig.OrdererConfig) (Orderer, error)
	CreateUser(name string, signingIdentity *SigningIdentity) (User, error)
}
