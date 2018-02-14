/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabca"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/core/identity"
)

// FabricProvider enables access to fabric objects such as peer and user based on config or context.
type FabricProvider interface {
	CreateIdentityClient(orgID string) (apifabca.FabricCAClient, error)
	CreateChannelClient(user identity.Context, cfg apifabclient.ChannelCfg) (apifabclient.Channel, error)
	CreateChannelLedger(ic identity.Context, name string) (apifabclient.ChannelLedger, error)
	CreateChannelConfig(user identity.Context, name string) (apifabclient.ChannelConfig, error)
	CreateResourceClient(user identity.Context) (apifabclient.Resource, error)
	CreateEventHub(ic identity.Context, name string) (apifabclient.EventHub, error)

	CreatePeerFromConfig(peerCfg *apiconfig.NetworkPeer) (apifabclient.Peer, error)
	CreateOrdererFromConfig(cfg *apiconfig.OrdererConfig) (apifabclient.Orderer, error)
	CreateUser(name string, signingIdentity *apifabclient.SigningIdentity) (identity.User, error)
}
