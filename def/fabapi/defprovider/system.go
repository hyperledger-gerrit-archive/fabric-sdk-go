/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defprovider

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabca"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	identityImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/context"
)

// FabricSystemFactory represents the default implementation of a session client.
type FabricSystemFactory struct{}

// NewFabricSystemFactory creates a new default session client factory.
func NewFabricSystemFactory() *FabricSystemFactory {
	f := FabricSystemFactory{}
	return &f
}

// NewClient returns a new FabricClient.
func (f *FabricSystemFactory) NewClient(sdk context.SDK, session context.Session, config apiconfig.Config) (apifabclient.FabricClient, error) {
	client := clientImpl.NewClient(config)

	client.SetCryptoSuite(sdk.CryptoSuiteProvider())
	client.SetStateStore(sdk.StateStoreProvider())
	client.SetUserContext(session.Identity())
	client.SetSigningManager(sdk.SigningManager())

	return client, nil
}

// NewUser returns a new default implementation of a User.
func (f *FabricSystemFactory) NewUser(config apiconfig.Config, msp apifabca.FabricCAClient, name string, pwd string,
	mspID string) (apifabca.User, error) {

	key, cert, err := msp.Enroll(name, pwd)
	if err != nil {
		return nil, errors.WithMessage(err, "Enroll failed")
	}
	user := identityImpl.NewUser(name, mspID)
	user.SetPrivateKey(key)
	user.SetEnrollmentCertificate(cert)

	return user, nil
}

// NewPreEnrolledUser returns a new default implementation of a User.
func (f *FabricSystemFactory) NewPreEnrolledUser(config apiconfig.Config, name string, signingIdentity *apifabclient.SigningIdentity) (apifabca.User, error) {

	user := identityImpl.NewUser(name, signingIdentity.MspID)

	user.SetPrivateKey(signingIdentity.PrivateKey)
	user.SetEnrollmentCertificate(signingIdentity.EnrollmentCert)

	return user, nil
}

// NewPeer returns a new default implementation of Peer
func (f *FabricSystemFactory) NewPeer(url string, certificate string, serverHostOverride string, config apiconfig.Config) (apifabclient.Peer, error) {
	return peerImpl.NewPeerTLSFromCert(url, certificate, serverHostOverride, config)
}

// NewPeerFromConfig returns a new default implementation of Peer based configuration
func (f *FabricSystemFactory) NewPeerFromConfig(peerCfg *apiconfig.NetworkPeer, config apiconfig.Config) (apifabclient.Peer, error) {
	return peerImpl.NewPeerFromConfig(peerCfg, config)
}
