/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package consclient enables consortium client
package consclient

import (
	"io/ioutil"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	cons "github.com/hyperledger/fabric-sdk-go/api/apitxn/consclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

var logger = logging.NewLogger("fabric_sdk_go")

// ConsortiumClient enables managing consortium resources (channels) in Fabric network.
type ConsortiumClient struct {
	client fab.FabricClient
	config config.Config
}

// NewConsortiumClient returns a consortium client instance
func NewConsortiumClient(client fab.FabricClient, config config.Config) (*ConsortiumClient, error) {
	cc := &ConsortiumClient{client: client, config: config}
	return cc, nil
}

// SaveChannel creates or updates channel
func (cc *ConsortiumClient) SaveChannel(req cons.SaveChannelRequest) error {
	return cc.SaveChannelWithOpts(req, cons.SaveChannelOpts{})
}

// SaveChannelWithOpts creates or updates channel with custom options
func (cc *ConsortiumClient) SaveChannelWithOpts(req cons.SaveChannelRequest, opts cons.SaveChannelOpts) error {

	if req.ChannelID == "" || req.ChannelConfig == "" {
		return errors.New("must provide channel ID and channel config")
	}

	logger.Debugf("***** Saving channel: %s *****\n", req.ChannelID)

	// Signing user has to belong to one of configured channel organisations
	// In case that order org is one of channel orgs we can use context user
	signer := cc.client.UserContext()
	if req.SigningUser != nil {
		// Retrieve custom signing identity here
		signer = req.SigningUser
	}

	if signer == nil {
		return errors.New("must provide signing user")
	}

	configTx, err := ioutil.ReadFile(req.ChannelConfig)
	if err != nil {
		return errors.WithMessage(err, "reading channel config file failed")
	}

	chConfig, err := cc.client.ExtractChannelConfig(configTx)
	if err != nil {
		return errors.WithMessage(err, "extracting channel config failed")
	}

	configSignature, err := cc.client.SignChannelConfig(chConfig, signer)
	if err != nil {
		return errors.WithMessage(err, "signing configuration failed")
	}

	var configSignatures []*common.ConfigSignature
	configSignatures = append(configSignatures, configSignature)

	// Figure out orderer configuration
	var ordererCfg *config.OrdererConfig
	if opts.OrdererID != "" {
		ordererCfg, err = cc.config.OrdererConfig(opts.OrdererID)
	} else {
		// Default is random orderer from configuration
		ordererCfg, err = cc.config.RandomOrdererConfig()
	}

	// Check if retrieving orderer configuration went ok
	if err != nil || ordererCfg == nil {
		return errors.Errorf("failed to retrieve orderer config: %s", err)
	}

	orderer, err := orderer.NewOrdererFromConfig(ordererCfg, cc.config)
	if err != nil {
		return errors.WithMessage(err, "failed to create new orderer from config")
	}

	request := fab.CreateChannelRequest{
		Name:       req.ChannelID,
		Orderer:    orderer,
		Config:     chConfig,
		Signatures: configSignatures,
	}

	_, err = cc.client.CreateChannel(request)
	if err != nil {
		return errors.WithMessage(err, "create channel failed")
	}

	return nil
}
