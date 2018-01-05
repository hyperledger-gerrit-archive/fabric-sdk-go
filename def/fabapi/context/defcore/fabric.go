/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defcore

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/api/apifabca"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	identityImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

// FabricProvider represents the default implementation of Fabric objects.
type FabricProvider struct {
	config      apiconfig.Config
	stateStore  apifabclient.KeyValueStore
	cryptoSuite apicryptosuite.CryptoSuite
	signer      apifabclient.SigningManager
}

// NewFabricProvider creates a FabricProvider enabling access to core Fabric objects and functionality.
func NewFabricProvider(config apiconfig.Config, stateStore apifabclient.KeyValueStore, cryptoSuite apicryptosuite.CryptoSuite, signer apifabclient.SigningManager) *FabricProvider {
	f := FabricProvider{
		config,
		stateStore,
		cryptoSuite,
		signer,
	}
	return &f
}

// NewClient returns a new FabricClient.
func (f *FabricProvider) NewClient(user apifabclient.User) (apifabclient.FabricClient, error) {
	client := clientImpl.NewClient(f.config)

	client.SetCryptoSuite(f.cryptoSuite)
	client.SetStateStore(f.stateStore)
	client.SetUserContext(user)
	client.SetSigningManager(f.signer)

	return client, nil
}

// NewUser returns a new default implementation of a User.
func (f *FabricProvider) NewUser(msp apifabca.FabricCAClient, name string, pwd string,
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
func (f *FabricProvider) NewPreEnrolledUser(name string, signingIdentity *apifabclient.SigningIdentity) (apifabca.User, error) {

	user := identityImpl.NewUser(name, signingIdentity.MspID)

	user.SetPrivateKey(signingIdentity.PrivateKey)
	user.SetEnrollmentCertificate(signingIdentity.EnrollmentCert)

	return user, nil
}

// NewPeer returns a new default implementation of Peer
func (f *FabricProvider) NewPeer(url string, certificate string, serverHostOverride string) (apifabclient.Peer, error) {
	return peerImpl.NewPeerTLSFromCert(url, certificate, serverHostOverride, f.config)
}

// NewPeerFromConfig returns a new default implementation of Peer based configuration
func (f *FabricProvider) NewPeerFromConfig(peerCfg *apiconfig.NetworkPeer) (apifabclient.Peer, error) {
	return peerImpl.NewPeerFromConfig(peerCfg, f.config)
}
