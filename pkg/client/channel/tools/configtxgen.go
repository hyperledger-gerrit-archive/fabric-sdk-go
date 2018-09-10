/*
Copyright IBM Corp. 2017 All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package tools

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hyperledger/fabric/common/channelconfig"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/common/tools/configtxgen/encoder"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	com "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	genesisconfig "github.com/hyperledger/fabric/common/tools/configtxgen/localconfig"
	"github.com/hyperledger/fabric/common/tools/protolator"
	cb "github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
)

var logger = flogging.MustGetLogger("common/tools/configtxgen")

// DoOutputChannelCreateTx returns orderer system channel genesis block.
//  Parameters:
//  conf is mandatory configuration profile
//  channelID is mandatory channel ID
//
//  Returns:
//  orderer system channel genesis block
func DoOutputBlock(config *genesisconfig.Profile, channelID string) ([]byte, error) {
	pgen := encoder.New(config)
	logger.Info("Generating genesis block")
	if config.Consortiums == nil {
		logger.Warning("Genesis block does not contain a consortiums group definition.  This block cannot be used for orderer bootstrap.")
	}
	genesisBlock := pgen.GenesisBlockForChannel(channelID)
	if genesisBlock == nil {
		return nil, fmt.Errorf("Error generating orderer channel genesis block")
	}
	genesisBlockBytes := utils.MarshalOrPanic(genesisBlock)
	if genesisBlockBytes == nil {
		return nil, fmt.Errorf("Error marshaling genesis block")
	}
	return genesisBlockBytes, nil
}

// DoOutputChannelCreateTx returns application channel genesis block.
//  Parameters:
//  conf is mandatory configuration profile
//  channelID is mandatory channel ID
//
//  Returns:
//  application channel genesis block
func DoOutputChannelCreateTx(conf *genesisconfig.Profile, channelID string) ([]byte, error) {
	logger.Info("Generating new channel configtx")

	configtx, err := encoder.MakeChannelCreationTransaction(channelID, nil, nil, conf)
	if err != nil {
		return nil, err
	}

	logger.Info("Marshaling new channel tx")
	cBytes := utils.MarshalOrPanic(configtx)
	if cBytes == nil {
		return nil, fmt.Errorf("Error marshaling application org channel genesis block tx")
	}
	return cBytes, nil
}

// DoOutputAnchorPeersUpdate returns channel update configuration envelope which can be sent to resmgmt.SaveChannel function in SaveChannelRequest along with signing identities necessary to make such a change.
//  Parameters:
//  conf is mandatory configuration profile
//  channelID is mandatory channel ID
//  asOrg is org being updated
//
//  Returns:
//  channel update configuration envelope
func DoOutputAnchorPeersUpdate(conf *genesisconfig.Profile, channelID string, asOrg string) ([]byte, error) {
	logger.Info("Generating anchor peer update")
	if asOrg == "" {
		return nil, fmt.Errorf("Must specify an organization to update the anchor peer for")
	}

	if conf.Application == nil {
		return nil, fmt.Errorf("Cannot update anchor peers without an application section")
	}

	var org *genesisconfig.Organization
	for _, iorg := range conf.Application.Organizations {
		if iorg.Name == asOrg {
			org = iorg
		}
	}

	if org == nil {
		return nil, fmt.Errorf("No organization name matching: %s", asOrg)
	}

	anchorPeers := make([]*pb.AnchorPeer, len(org.AnchorPeers))
	for i, anchorPeer := range org.AnchorPeers {
		anchorPeers[i] = &pb.AnchorPeer{
			Host: anchorPeer.Host,
			Port: int32(anchorPeer.Port),
		}
	}

	configUpdate := &cb.ConfigUpdate{
		ChannelId: channelID,
		WriteSet:  cb.NewConfigGroup(),
		ReadSet:   cb.NewConfigGroup(),
	}

	// Add all the existing config to the readset
	configUpdate.ReadSet.Groups[channelconfig.ApplicationGroupKey] = cb.NewConfigGroup()
	configUpdate.ReadSet.Groups[channelconfig.ApplicationGroupKey].Version = 1
	configUpdate.ReadSet.Groups[channelconfig.ApplicationGroupKey].ModPolicy = channelconfig.AdminsPolicyKey
	configUpdate.ReadSet.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name] = cb.NewConfigGroup()
	configUpdate.ReadSet.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name].Values[channelconfig.MSPKey] = &cb.ConfigValue{}
	configUpdate.ReadSet.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name].Policies[channelconfig.ReadersPolicyKey] = &cb.ConfigPolicy{}
	configUpdate.ReadSet.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name].Policies[channelconfig.WritersPolicyKey] = &cb.ConfigPolicy{}
	configUpdate.ReadSet.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name].Policies[channelconfig.AdminsPolicyKey] = &cb.ConfigPolicy{}

	// Add all the existing at the same versions to the writeset
	configUpdate.WriteSet.Groups[channelconfig.ApplicationGroupKey] = cb.NewConfigGroup()
	configUpdate.WriteSet.Groups[channelconfig.ApplicationGroupKey].Version = 1
	configUpdate.WriteSet.Groups[channelconfig.ApplicationGroupKey].ModPolicy = channelconfig.AdminsPolicyKey
	configUpdate.WriteSet.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name] = cb.NewConfigGroup()
	configUpdate.WriteSet.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name].Version = 1
	configUpdate.WriteSet.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name].ModPolicy = channelconfig.AdminsPolicyKey
	configUpdate.WriteSet.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name].Values[channelconfig.MSPKey] = &cb.ConfigValue{}
	configUpdate.WriteSet.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name].Policies[channelconfig.ReadersPolicyKey] = &cb.ConfigPolicy{}
	configUpdate.WriteSet.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name].Policies[channelconfig.WritersPolicyKey] = &cb.ConfigPolicy{}
	configUpdate.WriteSet.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name].Policies[channelconfig.AdminsPolicyKey] = &cb.ConfigPolicy{}
	configUpdate.WriteSet.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name].Values[channelconfig.AnchorPeersKey] = &cb.ConfigValue{
		Value:     utils.MarshalOrPanic(channelconfig.AnchorPeersValue(anchorPeers).Value()),
		ModPolicy: channelconfig.AdminsPolicyKey,
	}
	configUpdateEnvelope := &cb.ConfigUpdateEnvelope{
		ConfigUpdate: utils.MarshalOrPanic(configUpdate),
	}

	update := &cb.Envelope{
		Payload: utils.MarshalOrPanic(&cb.Payload{
			Header: &cb.Header{
				ChannelHeader: utils.MarshalOrPanic(&cb.ChannelHeader{
					ChannelId: channelID,
					Type:      int32(cb.HeaderType_CONFIG_UPDATE),
				}),
			},
			Data: utils.MarshalOrPanic(configUpdateEnvelope),
		}),
	}
	if err := protolator.DeepMarshalJSON(os.Stdout, update); err != nil {
		return nil, errors.Wrapf(err, "Error when json marshalling newly generated org")
	}

	logger.Info("Writing anchor peer update")
	cBytes := utils.MarshalOrPanic(update)
	if cBytes == nil {
		return nil, fmt.Errorf("Error marshaling channel anchor peer update")
	}
	return cBytes, nil
}

// DoOutputAddOrgToChannelUpdate returns channel update configuration envelope which can be sent to resmgmt.SaveChannel function in SaveChannelRequest along with signing identities necessary to make such a change.
//  Parameters:
//  org is mandatory organization that is being added
//  channelID is mandatory channel ID
//  block is latest channel configuration block
//  channelCfg is latest channel configuration struct
//
//  Returns:
//  channel update configuration envelope
func DoOutputAddOrgToChannelUpdate(org *genesisconfig.Organization, channelID string, block *com.Block, channelCfg fab.ChannelCfg) ([]byte, error) {
	logger.Info("Generating add org to channel update")

	og, err := encoder.NewOrdererOrgGroup(org)
	if err := protolator.DeepMarshalJSON(os.Stdout, &cb.DynamicConsortiumOrgGroup{ConfigGroup: og}); err != nil {
		return nil, errors.Wrapf(err, "Error when json marshalling newly generated org")
	}

	if err != nil {
		return nil, fmt.Errorf("Error getting organizations config group: %s", err)
	}

	configEnvelope, err := resource.CreateConfigEnvelope(block.Data.Data[0])
	if err != nil {
		return nil, err
	}
	// configtxlator proto_decode --input config_block.pb --type common.Block | jq .data.data[0].payload.data.config > config.json
	group := configEnvelope.Config.ChannelGroup // config.json

	// Type conversion
	b, err := json.Marshal(group)
	if err != nil {
		return nil, errors.Wrapf(err, "Error when json marshalling latest config block")
	}

	readset := &cb.ConfigGroup{}
	err = json.Unmarshal(b, readset)
	if err != nil {
		return nil, errors.Wrapf(err, "Error when json unmarshalling latest config block to the readset")
	}

	writeset := &cb.ConfigGroup{}
	err = json.Unmarshal(b, writeset)
	if err != nil {
		return nil, errors.Wrapf(err, "Error when json unmarshalling latest config block to the writeset")
	}

	// must increment version of element that needs to be changed
	readsetApplicationVersion := readset.Groups[channelconfig.ApplicationGroupKey].GetVersion()
	writeset.Groups[channelconfig.ApplicationGroupKey].Version = readsetApplicationVersion + 1
	writeset.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name] = og

	configUpdate := &cb.ConfigUpdate{
		ChannelId: channelID,
		WriteSet:  writeset,
		ReadSet:   readset,
	}

	// configtxlator compute_update --channel_id $CHANNEL_NAME --original config.pb --updated modified_config.pb --output org3_update.pb
	configUpdateEnvelope := &cb.ConfigUpdateEnvelope{ // configUpdateEnvelope = configtxlator proto_decode --input org3_update.pb --type common.ConfigUpdate | jq . > org3_update.json
		ConfigUpdate: utils.MarshalOrPanic(configUpdate),
	}

	// configtxlator proto_encode --input org3_update_in_envelope.json --type common.Envelope --output org3_update_in_envelope.pb
	update := &cb.Envelope{
		// echo '{"payload":{"header":{"channel_header":{"channel_id":"mychannel", "type":2}},"data":{"config_update":'$(cat org3_update.json)'}}}' | jq . > org3_update_in_envelope.json
		Payload: utils.MarshalOrPanic(&cb.Payload{
			Header: &cb.Header{
				ChannelHeader: utils.MarshalOrPanic(&cb.ChannelHeader{
					ChannelId: channelID,
					Type:      int32(cb.HeaderType_CONFIG_UPDATE),
				}),
			},
			Data: utils.MarshalOrPanic(configUpdateEnvelope),
		}),
	}

	logger.Info("Writing add org to channel update")
	cBytes := utils.MarshalOrPanic(update) // org3_update_in_envelope.pb
	if cBytes == nil {
		return nil, fmt.Errorf("Error marshaling add org to channel update")
	}
	return cBytes, nil
}
