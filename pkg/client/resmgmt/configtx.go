/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	"encoding/json"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/cmd/configtxgen"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkinternal/configtxgen/localconfig"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt/genesisconfig"
)

// CreateGenesisBlock creates a genesis block for a channel
func (rc *Client) CreateGenesisBlock(config *genesisconfig.Profile, channelID string) ([]byte, error) {
	localConfig, err := genesisToLocalConfig(config)
	if err != nil {
		return nil, err
	}
	return configtxgen.OutputBlock(localConfig, channelID)
}

func genesisToLocalConfig(config *genesisconfig.Profile) (*localconfig.Profile, error) {
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

// InspectGenesisBlock inspects a block
func (rc *Client) InspectGenesisBlock(block []byte) (string, error) {
	return configtxgen.InspectBlock(block)
}
