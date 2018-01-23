/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/api/apifabca"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	channelImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	identityImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

// FabricProvider represents the default implementation of Fabric objects.
type FabricProvider struct {
	config         apiconfig.Config
	cryptoSuite    apicryptosuite.CryptoSuite
	signingManager apifabclient.SigningManager
}

// NewFabricProvider creates a FabricProvider enabling access to core Fabric objects and functionality.
func NewFabricProvider(config apiconfig.Config, cryptoSuite apicryptosuite.CryptoSuite, signingManager apifabclient.SigningManager) *FabricProvider {
	f := FabricProvider{
		config,
		cryptoSuite,
		signingManager,
	}
	return &f
}

// NewResourceClient returns a new client initialized for the current instance of the SDK.
func (f *FabricProvider) NewResourceClient(ic apifabclient.IdentityContext) (apifabclient.Resource, error) {
	context := &clientContext{fabProvider: f, identity: ic}
	client := clientImpl.New(context)

	return client, nil
}

// NewChannelClient returns a new client initialized for the current instance of the SDK.
//
// TODO - add argument with channel config interface (to enable channel configuration obtained from the network)
func (f *FabricProvider) NewChannelClient(ic apifabclient.IdentityContext, channelID string) (apifabclient.Channel, error) {
	context := &clientContext{fabProvider: f, identity: ic}
	channel, err := channelImpl.New(context, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "NewChannel failed")
	}

	chOrderers, err := f.config.ChannelOrderers(channel.Name())
	if err != nil {
		return nil, errors.WithMessage(err, "reading channel orderers failed")
	}

	for _, ordererCfg := range chOrderers {

		orderer, err := orderer.New(f.config, orderer.FromOrdererConfig(&ordererCfg))
		if err != nil {
			return nil, errors.WithMessage(err, "creating orderer failed")
		}
		err = channel.AddOrderer(orderer)
		if err != nil {
			return nil, errors.WithMessage(err, "adding orderer failed")
		}
	}

	return channel, nil
}

// NewCAClient returns a new FabricCAClient initialized for the current instance of the SDK.
func (f *FabricProvider) NewCAClient(orgID string) (apifabca.FabricCAClient, error) {
	return fabricCAClient.NewFabricCAClient(orgID, f.config, f.cryptoSuite)
}

/////////////
// TODO - refactor the below (see if we really need to create these objects from the factory rather than directly)

// NewUser returns a new default implementation of a User.
func (f *FabricProvider) NewUser(name string, signingIdentity *apifabclient.SigningIdentity) (apifabclient.User, error) {

	user := identityImpl.NewUser(name, signingIdentity.MspID)

	user.SetPrivateKey(signingIdentity.PrivateKey)
	user.SetEnrollmentCertificate(signingIdentity.EnrollmentCert)

	return user, nil
}

// NewPeer returns a new default implementation of Peer
func (f *FabricProvider) NewPeer(url string, certificate *x509.Certificate, serverHostOverride string) (apifabclient.Peer, error) {
	return peerImpl.New(f.config, peerImpl.WithURL(url), peerImpl.WithTLSCert(certificate), peerImpl.WithServerName(serverHostOverride))
}

// NewPeerFromConfig returns a new default implementation of Peer based configuration
func (f *FabricProvider) NewPeerFromConfig(peerCfg *apiconfig.NetworkPeer) (apifabclient.Peer, error) {
	return peerImpl.New(f.config, peerImpl.FromPeerConfig(peerCfg))
}

//////////////

type clientContext struct {
	fabProvider *FabricProvider
	identity    apifabclient.IdentityContext
}

func (c *clientContext) Config() apiconfig.Config {
	return c.fabProvider.config
}

func (c *clientContext) CryptoSuite() apicryptosuite.CryptoSuite {
	return c.fabProvider.cryptoSuite
}

func (c *clientContext) SigningManager() apifabclient.SigningManager {
	return c.fabProvider.signingManager
}

func (c *clientContext) IdentityContext() apifabclient.IdentityContext {
	return c.identity
}

/*
// getChannel is helper method to initializes and returns a channel based on config
func getChannel(client fab.Resource, channelID string) (fab.Channel, error) {

	channel, err := client.NewChannel(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "NewChannel failed")
	}

	chCfg, err := client.Config().ChannelConfig(channel.Name())
	if err != nil || chCfg == nil {
		return nil, errors.Errorf("reading channel config failed: %s", err)
	}

	chOrderers, err := client.Config().ChannelOrderers(channel.Name())
	if err != nil {
		return nil, errors.WithMessage(err, "reading channel orderers failed")
	}

	for _, ordererCfg := range chOrderers {

		orderer, err := orderer.New(client.Config(), orderer.FromOrdererConfig(&ordererCfg))
		if err != nil {
			return nil, errors.WithMessage(err, "creating orderer failed")
		}
		err = channel.AddOrderer(orderer)
		if err != nil {
			return nil, errors.WithMessage(err, "adding orderer failed")
		}
	}

	return channel, nil
}

// getChannel is helper method for instantiating channel. If channel is not configured it will use random orderer from global orderer configuration
func (rc *ResourceMgmtClient) getChannel(channelID string) (fab.Channel, error) {

	channel, err := rc.channelProvider.Channel(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get channel")
	}

	// Creating channel requires orderer information
	var orderers []config.OrdererConfig
	chCfg, err := rc.config.ChannelConfig(channelID)
	if err != nil {
		return nil, err
	}

	if chCfg == nil {
		orderers, err = rc.config.OrderersConfig()
	} else {
		orderers, err = rc.config.ChannelOrderers(channelID)
	}

	// Check if retrieving orderer configuration went ok
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to retrieve orderer configuration")
	}

	if len(orderers) == 0 {
		return nil, errors.Errorf("Must configure at least one order for channel and/or one orderer in the network")
	}

	channel, err = rc.channelProvider.NewChannel(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "NewChannel failed")
	}

	for _, ordererCfg := range orderers {
		orderer, err := orderer.New(rc.config, orderer.FromOrdererConfig(&ordererCfg))
		if err != nil {
			return nil, errors.WithMessage(err, "NewOrdererFromConfig failed")
		}
		err = channel.AddOrderer(orderer)
		if err != nil {
			return nil, errors.WithMessage(err, "adding orderer failed")
		}
	}

	return channel, nil
}
*/
