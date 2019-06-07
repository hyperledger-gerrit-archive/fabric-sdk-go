/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/tools/protolator/protoext/ordererext"

	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/channelconfig"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/tools/protolator"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkinternal/configtxgen/encoder"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkinternal/configtxlator/update"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkinternal/configtxgen/localconfig"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/genesisconfig"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

// See https://github.com/hyperledger/fabric/blob/be235fd3a236f792a525353d9f9586c8b0d4a61a/cmd/configtxgen/main.go

// CreateGenesisBlock creates a genesis block for a channel
func CreateGenesisBlock(config *genesisconfig.Profile, channelID string) ([]byte, error) {
	localConfig, err := genesisToLocalProfile(config)
	if err != nil {
		return nil, err
	}
	pgen, err := encoder.NewBootstrapper(localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "could not create bootstrapper")
	}
	logger.Debug("Generating genesis block")
	if config.Orderer == nil {
		return nil, errors.Errorf("refusing to generate block which is missing orderer section")
	}
	genesisBlock := pgen.GenesisBlockForChannel(channelID)
	logger.Debug("Writing genesis block")
	return protoutil.Marshal(genesisBlock)
}

// CreateGenesisBlockForOrderer creates a genesis block for an orderer
func CreateGenesisBlockForOrderer(config *genesisconfig.Profile, channelID string) ([]byte, error) {
	if config.Consortiums == nil {
		return nil, errors.Errorf("Genesis block does not contain a consortiums group definition. This block cannot be used for orderer bootstrap.")
	}
	return CreateGenesisBlock(config, channelID)
}

func genesisToLocalProfile(config *genesisconfig.Profile) (*localconfig.Profile, error) {
	b, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	c := &localconfig.Profile{}
	err = json.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func genesisToLocalTopLevel(config *genesisconfig.TopLevel) (*localconfig.TopLevel, error) {
	b, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	c := &localconfig.TopLevel{}
	err = json.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func localToGenesisProfile(config *localconfig.Profile) (*genesisconfig.Profile, error) {
	b, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	c := &genesisconfig.Profile{}
	err = json.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func localToGenesisTopLevel(config *localconfig.TopLevel) (*genesisconfig.TopLevel, error) {
	b, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	c := &genesisconfig.TopLevel{}
	err = json.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// InspectBlock inspects a block
func InspectBlock(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("missing block")
	}
	logger.Debug("Parsing genesis block")
	block, err := protoutil.UnmarshalBlock(data)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling to block: %s", err)
	}
	var buf bytes.Buffer
	err = protolator.DeepMarshalJSON(&buf, block)
	if err != nil {
		return "", fmt.Errorf("malformed block contents: %s", err)
	}
	return buf.String(), nil
}

// CreateChannelCreateTx creates a Fabric transaction for creating a channel
func CreateChannelCreateTx(conf, baseProfile *genesisconfig.Profile, channelID string) ([]byte, error) {
	logger.Debug("Generating new channel configtx")

	localConf, err := genesisToLocalProfile(conf)
	if err != nil {
		return nil, err
	}
	localBaseProfile, err := genesisToLocalProfile(baseProfile)
	if err != nil {
		return nil, err
	}

	var configtx *cb.Envelope
	if baseProfile == nil {
		configtx, err = encoder.MakeChannelCreationTransaction(channelID, nil, localConf)
	} else {
		configtx, err = encoder.MakeChannelCreationTransactionWithSystemChannelContext(channelID, nil, localConf, localBaseProfile)
	}
	if err != nil {
		return nil, err
	}

	logger.Debug("Writing new channel tx")
	return protoutil.Marshal(configtx)
}

// InspectChannelCreateTx inspects a Fabric transaction for creating a channel
func InspectChannelCreateTx(data []byte) (string, error) {
	logger.Debug("Parsing transaction")
	env, err := protoutil.UnmarshalEnvelope(data)
	if err != nil {
		return "", fmt.Errorf("Error unmarshaling envelope: %s", err)
	}
	var buf bytes.Buffer
	err = protolator.DeepMarshalJSON(&buf, env)
	if err != nil {
		return "", fmt.Errorf("malformed transaction contents: %s", err)
	}
	return buf.String(), nil
}

// CreateAnchorPeersUpdate creates an anchor peers update transaction
func CreateAnchorPeersUpdate(conf *genesisconfig.Profile, channelID string, asOrg string) (*common.Envelope, error) {
	logger.Debug("Generating anchor peer update")
	if asOrg == "" {
		return nil, fmt.Errorf("Must specify an organization to update the anchor peer for")
	}

	if conf.Application == nil {
		return nil, fmt.Errorf("Cannot update anchor peers without an application section")
	}

	localConf, err := genesisToLocalProfile(conf)
	if err != nil {
		return nil, err
	}

	original, err := encoder.NewChannelGroup(localConf)
	if err != nil {
		return nil, errors.WithMessage(err, "error parsing profile as channel group")
	}
	original.Groups[channelconfig.ApplicationGroupKey].Version = 1

	updated := proto.Clone(original).(*cb.ConfigGroup)

	originalOrg, ok := original.Groups[channelconfig.ApplicationGroupKey].Groups[asOrg]
	if !ok {
		return nil, errors.Errorf("org with name '%s' does not exist in config", asOrg)
	}

	if _, ok = originalOrg.Values[channelconfig.AnchorPeersKey]; !ok {
		return nil, errors.Errorf("org '%s' does not have any anchor peers defined", asOrg)
	}

	delete(originalOrg.Values, channelconfig.AnchorPeersKey)

	updt, err := update.Compute(&cb.Config{ChannelGroup: original}, &cb.Config{ChannelGroup: updated})
	if err != nil {
		return nil, errors.WithMessage(err, "could not compute update")
	}
	updt.ChannelId = channelID

	newConfigUpdateEnv := &cb.ConfigUpdateEnvelope{
		ConfigUpdate: protoutil.MarshalOrPanic(updt),
	}

	return protoutil.CreateSignedEnvelope(cb.HeaderType_CONFIG_UPDATE, channelID, nil, newConfigUpdateEnv, 0, 0)

}

// ProfileFromYaml constructs channel genesis profile from standard configtxgen yaml file
func ProfileFromYaml(profile, yamlPath string) (*genesisconfig.Profile, error) {
	config, err := localconfig.Load(strings.ToLower(profile), yamlPath)
	if err != nil {
		return nil, err
	}

	keyMap := map[string]interface{}{
		"readers":              "Readers",
		"writers":              "Writers",
		"admins":               "Admins",
		"blockvalidation":      "BlockValidation",
		"lifecycleendorsement": "LifecycleEndorsement",
		"endorsement":          "Endorsement",
	}

	err = applyKeyReplacements(config.Policies, keyMap)
	if err != nil {
		return nil, err
	}

	if config.Application != nil {
		err = applyKeyReplacements(config.Application.Policies, keyMap)
		if err != nil {
			return nil, err
		}
	}

	if config.Orderer != nil {
		err = applyKeyReplacements(config.Orderer.Policies, keyMap)
		if err != nil {
			return nil, err
		}
	}
	return localToGenesisProfile(config)
}

// TopLevelFromYaml constructs top level configuration from standard configtxgen yaml file
func TopLevelFromYaml(yamlPath string) (*genesisconfig.TopLevel, error) {
	config, err := localconfig.LoadTopLevel(yamlPath)
	if err != nil {
		return nil, err
	}
	return localToGenesisTopLevel(config)
}

// OrgAsJSON returns a JSON string of the specified organization's definition
func OrgAsJSON(conf *genesisconfig.TopLevel, orgName string) (string, error) {

	localConf, err := genesisToLocalTopLevel(conf)
	if err != nil {
		return "", err
	}

	for _, org := range localConf.Organizations {
		if org.Name == orgName {
			og, err := encoder.NewOrdererOrgGroup(org)
			if err != nil {
				return "", errors.Wrapf(err, "bad org definition for org %s", org.Name)
			}

			var buf bytes.Buffer
			if err := protolator.DeepMarshalJSON(&buf, &ordererext.DynamicOrdererOrgGroup{ConfigGroup: og}); err != nil {
				return "", errors.Wrapf(err, "malformed org definition for org: %s", org.Name)
			}
			return buf.String(), nil
		}
	}
	return "", errors.Errorf("organization %s not found", orgName)
}

func applyKeyReplacements(policies map[string]*localconfig.Policy, dict map[string]interface{}) error {
	if len(policies) == 0 {
		return errors.New("policy map cannot be empty")
	}
	if len(dict) == 0 {
		return errors.New("dictionary cannot be empty")
	}

	for k, v := range dict {
		if _, ok := policies[k]; ok {
			policies[v.(string)] = policies[k]
			delete(policies, k)
		}
	}
	return nil
}
