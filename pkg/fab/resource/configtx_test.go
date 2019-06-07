/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	mspcfg "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"

	"github.com/stretchr/testify/assert"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/genesisconfig"
	"github.com/stretchr/testify/require"
)

// See https://github.com/hyperledger/fabric/blob/v2.0.0-alpha/cmd/configtxgen/main_test.go

var yamlPath = "testdata"

const mspDir = "./testdata/msp"

func TestMain(m *testing.M) {
	err := genMspDir(mspDir)
	if err != nil {
		fmt.Printf("Error generating msp dir: %v\n", err)
	}
	defer func() {
		_ = os.RemoveAll(mspDir)
	}()

	m.Run()
}

func genMspDir(dir string) error {
	ordererMspDir := filepath.Join(metadata.GetProjectPath(), "test/fixtures/fabric/v1/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp")
	cfg, err := mspcfg.GetVerifyingMspConfig(ordererMspDir, "mymspid", "bccsp")
	if err != nil {
		return fmt.Errorf("Error generating msp config from dir: %v\n", err)
	}
	mspConfig := &msp.FabricMSPConfig{}
	err = proto.Unmarshal(cfg.Config, mspConfig)
	if err != nil {
		return fmt.Errorf("Error unmarshaling msp config: %v\n", err)
	}

	return GenerateMspDir(dir, cfg)
}

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

func TestCreateAndInspectGenesisBlock(t *testing.T) {
	config, err := ProfileFromYaml("SampleSingleMSPSolo", yamlPath)
	require.NoError(t, err, "Failed to create profile configuration")

	b, err := CreateGenesisBlock(config, "mychannel")
	require.NoError(t, err, "Failed to create genesis block")
	require.NotNil(t, b, "Failed to create genesis block")

	s, err := InspectBlock(b)
	require.NoError(t, err, "Failed to inspect genesis block")
	require.False(t, s == "", "Failed to inspect genesis block")
}

func TestCreateAndInspectGenesisBlockForOrderer(t *testing.T) {
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

func TestOrgAsJSON(t *testing.T) {
	config, err := TopLevelFromYaml(yamlPath)
	require.NoError(t, err, "Failed to create profile configuration")

	// TODO: check json
	_, err = OrgAsJSON(config, "SampleOrg")
	assert.NoError(t, err, "Good org to print")

	// TODO: check json
	_, err = OrgAsJSON(config, "SampleOrg.wrong")
	assert.Error(t, err, "Bad org name")
	assert.Regexp(t, "organization [^ ]* not found", err.Error())

	config.Organizations[0] = &genesisconfig.Organization{Name: "FakeOrg", ID: "FakeOrg"}
	// TODO: check json
	_, err = OrgAsJSON(config, "FakeOrg")
	assert.Error(t, err, "Fake org")
	assert.Regexp(t, "bad org definition", err.Error())
}
