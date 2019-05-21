/*
Copyright IBM Corp. 2017 All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package configtxgen

import (
	"bytes"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/tools/protolator"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/public/configtxgen/encoder"
	genesisconfig "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/public/configtxgen/localconfig"
	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	"github.com/pkg/errors"
)

var exitCode = 0

var logger = flogging.MustGetLogger("common.tools.configtxgen")

// OutputBlock generates a genesis block for the provided profile
func OutputBlock(config *genesisconfig.Profile, channelID string) ([]byte, error) {
	pgen, err := encoder.NewBootstrapper(config)
	if err != nil {
		return nil, errors.WithMessage(err, "could not create bootstrapper")
	}
	logger.Info("Generating genesis block")
	if config.Orderer == nil {
		return nil, errors.Errorf("refusing to generate block which is missing orderer section")
	}
	if config.Consortiums == nil {
		logger.Warning("Genesis block does not contain a consortiums group definition.  This block cannot be used for orderer bootstrap.")
	}
	genesisBlock := pgen.GenesisBlockForChannel(channelID)
	logger.Info("Writing genesis block")
	return protoutil.Marshal(genesisBlock)
}

// InspectBlock generates a string representation of a genesis block
func InspectBlock(data []byte) (string, error) {
	logger.Info("Parsing genesis block")
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
