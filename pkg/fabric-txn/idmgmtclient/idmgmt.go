/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package idmgmtclient enables identity management client
package idmgmtclient

import (
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	idmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/idmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/api/core/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// IdentityMgmtClient enables managing organization idenitites in Fabric network.
type IdentityMgmtClient struct {
	provider fab.ProviderContext
	identity identity.Context
}

// Context holds the providers and services needed to create an IdentityMgmtClient.
type Context struct {
	fab.ProviderContext
	identity.Context
	Resource fab.Resource
}

// New returns an identity management client instance
func New(c Context) (*IdentityMgmtClient, error) {
	cc := &IdentityMgmtClient{
		provider: c.ProviderContext,
		identity: c.Context,
	}
	return cc, nil
}

// Enroll enrolls a registered user with the org's Fabric CA
func (cc *IdentityMgmtClient) Enroll(req idmgmt.EnrollmentRequest) (*idmgmt.EnrollmentResponse, error) {

	if req.Name == "" || req.Secret == "" {
		return nil, errors.New("must provide user name and enrollment secret")
	}

	logger.Debugf("***** Enrilling user: %s *****\n", req.Name)

	// // Signing user has to belong to one of configured channel organisations
	// // In case that order org is one of channel orgs we can use context user
	// signer := cc.identity
	// if req.SigningIdentity != nil {
	// 	// Retrieve custom signing identity here
	// 	signer = req.SigningIdentity
	// }

	// if signer == nil {
	// 	return errors.New("must provide signing user")
	// }

	// configTx, err := ioutil.ReadFile(req.ChannelConfig)
	// if err != nil {
	// 	return errors.WithMessage(err, "reading channel config file failed")
	// }

	// chConfig, err := cc.resource.ExtractChannelConfig(configTx)
	// if err != nil {
	// 	return errors.WithMessage(err, "extracting channel config failed")
	// }

	// configSignature, err := cc.resource.SignChannelConfig(chConfig, signer)
	// if err != nil {
	// 	return errors.WithMessage(err, "signing configuration failed")
	// }

	// var configSignatures []*common.ConfigSignature
	// configSignatures = append(configSignatures, configSignature)

	// // Figure out orderer configuration
	// var ordererCfg *config.OrdererConfig
	// if opts.OrdererID != "" {
	// 	ordererCfg, err = cc.provider.Config().OrdererConfig(opts.OrdererID)
	// } else {
	// 	// Default is random orderer from configuration
	// 	ordererCfg, err = cc.provider.Config().RandomOrdererConfig()
	// }

	// // Check if retrieving orderer configuration went ok
	// if err != nil || ordererCfg == nil {
	// 	return errors.Errorf("failed to retrieve orderer config: %s", err)
	// }

	// orderer, err := orderer.New(cc.provider.Config(), orderer.FromOrdererConfig(ordererCfg))
	// if err != nil {
	// 	return errors.WithMessage(err, "failed to create new orderer from config")
	// }

	// request := fab.CreateChannelRequest{
	// 	Name:       req.ChannelID,
	// 	Orderer:    orderer,
	// 	Config:     chConfig,
	// 	Signatures: configSignatures,
	// }

	// _, err = cc.resource.CreateChannel(request)
	// if err != nil {
	// 	return errors.WithMessage(err, "create channel failed")
	// }

	return nil, nil
}
