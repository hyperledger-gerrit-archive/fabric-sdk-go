/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fabca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	identityImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

// NewSystemClient returns a new default implementation of Client
func NewSystemClient(config apiconfig.Config) *clientImpl.Client {
	return clientImpl.NewClient(config)
}

// NewPreEnrolledUser returns a new default implementation of a User.
func NewPreEnrolledUser(config apiconfig.Config, name string, signingIdentity *fab.SigningIdentity) (fabca.User, error) {

	user := identityImpl.NewUser(name, signingIdentity.MspID)

	user.SetPrivateKey(signingIdentity.PrivateKey)
	user.SetEnrollmentCertificate(signingIdentity.EnrollmentCert)

	return user, nil
}

// NewPeerFromConfig returns a new default implementation of Peer based configuration
func NewPeerFromConfig(peerCfg *apiconfig.NetworkPeer, config apiconfig.Config) (fab.Peer, error) {
	return peerImpl.NewPeerFromConfig(peerCfg, config)
}

// ChannelClientOpts provides options for creating channel client
type ChannelClientOpts struct {
	OrgName        string
	ConfigProvider apiconfig.Config
}

// ChannelMgmtClientOpts provides options for creating channel management client
type ChannelMgmtClientOpts struct {
	OrgName        string
	ConfigProvider apiconfig.Config
}

// ResourceMgmtClientOpts provides options for creating resource management client
type ResourceMgmtClientOpts struct {
	OrgName        string
	TargetFilter   resmgmt.TargetFilter
	ConfigProvider apiconfig.Config
}
