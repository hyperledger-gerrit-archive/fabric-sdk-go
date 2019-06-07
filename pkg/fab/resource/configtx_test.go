/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/genesisconfig"
	"github.com/stretchr/testify/require"
)

// See https://github.com/hyperledger/fabric/blob/v2.0.0-alpha/cmd/configtxgen/main_test.go

var yamlPath = "testdata"

func TestInspectMissing(t *testing.T) {
	_, err := InspectBlock(nil)
	require.Error(t, err, "Missing block")
}

func TestMissingOrdererSection(t *testing.T) {
	config, err := ProfileFromYaml("SampleInsecureSolo", yamlPath)
	require.NoError(t, err, "Failed to create profile configuration")

	config.Orderer = nil

	_, err = CreateGenesisBlock(config, "mychannel")
	require.Error(t, err, "Missing orderer section")
}

func TestMissingConsortiumSection(t *testing.T) {
	config, err := ProfileFromYaml("SampleInsecureSolo", yamlPath)
	require.NoError(t, err, "Failed to create profile configuration")

	config.Consortiums = nil

	_, err = CreateGenesisBlock(config, "mychannel")
	require.NoError(t, err, "Missing consortiums section")
}

func TestCreateAndInspectGenesiBlock(t *testing.T) {

	config, err := ProfileFromYaml("SampleSingleMSPChannel", yamlPath)
	require.NoError(t, err, "Failed to create profile configuration")

	b, err := CreateGenesisBlock(config, "mychannel")
	require.NoError(t, err, "Failed to create genesis block")
	require.NotNil(t, b, "Failed to create genesis block")

	s, err := InspectBlock(b)
	require.NoError(t, err, "Failed to inspect genesis block")
	require.False(t, s == "", "Failed to inspect genesis block")
}

func TestCreateAndInspectGenesiBlockForOrderer(t *testing.T) {

	config, err := ProfileFromYaml("SampleSingleMSPSolo", yamlPath)
	require.NoError(t, err, "Failed to create profile configuration")

	b, err := CreateGenesisBlockForOrderer(config, "mychannel")
	require.NoError(t, err, "Failed to create genesis block")
	require.NotNil(t, b, "Failed to create genesis block")

	s, err := InspectBlock(b)
	require.NoError(t, err, "Failed to inspect genesis block")
	require.False(t, s == "", "Failed to inspect genesis block")
}

func TestMissingConsortiumValue(t *testing.T) {
	config, err := ProfileFromYaml("SampleSingleMSPChannel", yamlPath)
	require.NoError(t, err, "Failed to create profile configuration")

	config.Consortium = ""

	_, err = CreateChannelCreateTx(config, nil, "configtx")
	require.Error(t, err, "Missing Consortium value in Application Profile definition")
}

func TestMissingApplicationValue(t *testing.T) {
	config, err := ProfileFromYaml("SampleSingleMSPChannel", yamlPath)
	require.NoError(t, err, "Failed to create profile configuration")

	config.Application = nil

	_, err = CreateChannelCreateTx(config, nil, "configtx")
	require.Error(t, err, "Missing Application value in Application Profile definition")
}

func TestCreateAndInspectConfigTx(t *testing.T) {

	config, err := ProfileFromYaml("SampleSingleMSPChannel", yamlPath)
	require.NoError(t, err, "Failed to create profile configuration")

	e, err := CreateChannelCreateTx(config, nil, "foo")
	require.NoError(t, err, "Failed to create channel create tx")
	require.NotNil(t, e, "Failed to create channel create tx")

	s, err := InspectChannelCreateTx(e)
	require.NoError(t, err, "Failed to inspect channel create tx")
	require.False(t, s == "", "Failed to inspect channel create tx")
}

func TestGenerateAnchorPeersUpdate(t *testing.T) {

	config, err := ProfileFromYaml("SampleSingleMSPChannel", yamlPath)
	require.NoError(t, err, "Failed to create profile configuration")

	e, err := CreateAnchorPeersUpdate(config, "foo", "SampleOrg")
	require.NoError(t, err, "Failed to create anchor peers update")
	require.NotNil(t, e, "Failed to create anchor peers update")
}

func TestBadAnchorPeersUpdates(t *testing.T) {

	config, err := ProfileFromYaml("SampleSingleMSPChannel", yamlPath)
	require.NoError(t, err, "Failed to create profile configuration")

	_, err = CreateAnchorPeersUpdate(config, "foo", "")
	require.Error(t, err, "Bad anchorPeerUpdate request - asOrg empty")

	backupApplication := config.Application
	config.Application = nil
	_, err = CreateAnchorPeersUpdate(config, "foo", "SampleOrg")
	require.Error(t, err, "Bad anchorPeerUpdate request")

	config.Application = backupApplication

	config.Application.Organizations[0] = &genesisconfig.Organization{Name: "FakeOrg", ID: "FakeOrg"}
	_, err = CreateAnchorPeersUpdate(config, "foo", "SampleOrg")
	require.Error(t, err, "Bad anchorPeerUpdate request - fake org")
}
