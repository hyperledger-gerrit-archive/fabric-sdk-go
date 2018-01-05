/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apicore

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabca"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// FabricProvider allows overriding of fabric objects such as peer and user
type FabricProvider interface {
	NewClient(user apifabclient.User) (apifabclient.FabricClient, error)
	NewPeer(url string, certificate string, serverHostOverride string) (apifabclient.Peer, error)
	NewPeerFromConfig(peerCfg *apiconfig.NetworkPeer) (apifabclient.Peer, error)
	NewUser(msp apifabca.FabricCAClient, name string, pwd string, mspID string) (apifabca.User, error)
	NewPreEnrolledUser(name string, signingIdentity *apifabclient.SigningIdentity) (apifabca.User, error)
}
