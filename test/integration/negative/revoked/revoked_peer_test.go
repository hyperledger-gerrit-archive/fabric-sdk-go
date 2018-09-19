/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package revoked

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"

	"io/ioutil"

	"bytes"

	"io"

	"time"

	"encoding/pem"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/utils"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	msp2 "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	org1AdminUser      = "Admin"
	org2AdminUser      = "Admin"
	org1User           = "User1"
	org1               = "Org1"
	org2               = "Org2"
	channelID          = "orgchannel"
	configFilename     = "config_test.yaml"
	configFileOverride = "overrides/revoke_test_override.yaml"
	pathRevokeCaRoot   = "peerOrganizations/org1.example.com/ca/"
	pathParentCert     = "peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem"
	pathCertToBeRevokd = "peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp/signcerts/peer0.org1.example.com-cert.pem"
)

var CRLTestRetryOpts = retry.Opts{
	Attempts:       20,
	InitialBackoff: 1 * time.Second,
	MaxBackoff:     10 * time.Second,
	BackoffFactor:  2.0,
	RetryableCodes: retry.TestRetryableCodes,
}

// Peers used for testing
var orgTestPeer0 fab.Peer
var orgTestPeer1 fab.Peer

var msps = []string{"Org1MSP", "Org2MSP"}

//TestPeerRevoke performs peer revoke test
// step 1: generate CRL
// step 2: update MSP revocation_list in channel config
// step 3: perform revoke peer test
func TestPeerRevoke(t *testing.T) {

	//generate CRL
	crlBytes, err := generateCRL()
	require.NoError(t, err, "failed to generate CRL")
	require.NotEmpty(t, crlBytes, "CRL is empty")

	//update revocation list in channel config
	updateRevocationList(t, crlBytes, true)

	//wait for config update
	waitForConfigUpdate(t)

	//test if peer has been revoked
	testRevokedPeer(t)

	//reset revocation list in channel config for other tests
	updateRevocationList(t, nil, false)
}

//updateRevocationList update MSP revocation_list in channel config
func updateRevocationList(t *testing.T, crlBytes []byte, joinCh bool) {

	sdk, err := fabsdk.New(config.FromFile(integration.GetConfigPath(configFilename)))
	require.NoError(t, err)
	defer sdk.Close()

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)
	defer integration.CleanupUserData(t, sdk)

	//prepare contexts
	org1AdminClientContext := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))
	org2AdminClientContext := sdk.Context(fabsdk.WithUser(org2AdminUser), fabsdk.WithOrg(org2))
	org1AdminChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))

	org1ResMgmt, err := resmgmt.New(org1AdminClientContext)
	require.NoError(t, err)

	if joinCh {
		//join channel
		org2ResMgmt, err := resmgmt.New(org2AdminClientContext)
		require.NoError(t, err)

		//join channel
		joinChannel(t, org1ResMgmt, org2ResMgmt)
	}

	ledgerClient1, err := ledger.New(org1AdminChannelClientContext)
	require.NoError(t, err)

	org1MspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org1))
	require.NoError(t, err)

	org2MspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org2))
	require.NoError(t, err)

	//create read write set for channel config update
	readSet, writeSet := prepareReadWriteSets(t, crlBytes, ledgerClient1)
	//update channel config MSP revocation lists to generated CRL
	updateChannelConfig(t, readSet, writeSet, org1ResMgmt, org1MspClient, org2MspClient)
}

//waitForConfigUpdate waits for all peer till they are updated with latest channel config
func waitForConfigUpdate(t *testing.T) {

	sdk, err := fabsdk.New(config.FromFile(integration.GetConfigPath(configFilename)))
	require.NoError(t, err)
	defer sdk.Close()

	org1AdminChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))

	ledgerClient1, err := ledger.New(org1AdminChannelClientContext)
	require.NoError(t, err)

	ctx, err := org1AdminChannelClientContext()
	require.NoError(t, err)

	ready := queryRevocationListUpdates(t, ledgerClient1, ctx.EndpointConfig(), channelID)
	require.True(t, ready, "all peers are not updated with latest channel config")
}

//testRevokedPeer performs revoke peer test
func testRevokedPeer(t *testing.T) {

	sdk1, err := fabsdk.New(getCustomConfigBackend())
	require.NoError(t, err)
	defer sdk1.Close()

	//prepare contexts
	org1AdminClientContext := sdk1.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))
	org2AdminClientContext := sdk1.Context(fabsdk.WithUser(org2AdminUser), fabsdk.WithOrg(org2))
	org1ChannelClientContext := sdk1.ChannelContext(channelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1))

	org1ResMgmt, err := resmgmt.New(org1AdminClientContext)
	require.NoError(t, err)

	org2ResMgmt, err := resmgmt.New(org2AdminClientContext)
	require.NoError(t, err)

	// Create chaincode package for example cc
	createCC(t, org1ResMgmt, org2ResMgmt)

	// Load specific targets for move funds test - one of the
	//targets has its certificate revoked
	loadOrgPeers(t, org1AdminClientContext)

	queryCC(t, org1ChannelClientContext)
}

//prepareReadWriteSets prepares read write sets for channel config update
func prepareReadWriteSets(t *testing.T, crlBytes []byte, ledgerClient *ledger.Client) (*common.ConfigGroup, *common.ConfigGroup) {

	var readSet, writeSet *common.ConfigGroup

	chCfg, err := ledgerClient.QueryConfig(ledger.WithTargetEndpoints("peer1.org2.example.com"))
	require.NoError(t, err)

	block, err := ledgerClient.QueryBlock(chCfg.BlockNumber(), ledger.WithTargetEndpoints("peer1.org2.example.com"))
	require.NoError(t, err)

	configEnv, err := resource.CreateConfigUpdateEnvelope(block.Data.Data[0])
	require.NoError(t, err)

	configUpdate := &common.ConfigUpdate{}
	proto.Unmarshal(configEnv.ConfigUpdate, configUpdate)
	readSet = configUpdate.ReadSet

	//prepare write set
	configEnv, err = resource.CreateConfigUpdateEnvelope(block.Data.Data[0])
	require.NoError(t, err)

	configUpdate = &common.ConfigUpdate{}
	proto.Unmarshal(configEnv.ConfigUpdate, configUpdate)
	writeSet = configUpdate.ReadSet

	//change write set for MSP revocation list update
	for _, org := range msps {
		val := writeSet.Groups["Application"].Groups[org].Values["MSP"].Value

		mspCfg := &msp.MSPConfig{}
		err = proto.Unmarshal(val, mspCfg)
		require.NoError(t, err)

		fabMspCfg := &msp.FabricMSPConfig{}
		err = proto.Unmarshal(mspCfg.Config, fabMspCfg)
		require.NoError(t, err)

		if len(crlBytes) > 0 {
			//append valid crl bytes to existing revocation list
			fabMspCfg.RevocationList = append(fabMspCfg.RevocationList, crlBytes)
		} else {
			//reset
			fabMspCfg.RevocationList = nil
		}

		fabMspBytes, err := proto.Marshal(fabMspCfg)
		require.NoError(t, err)

		mspCfg.Config = fabMspBytes

		mspBytes, err := proto.Marshal(mspCfg)
		require.NoError(t, err)

		writeSet.Groups["Application"].Groups[org].Values["MSP"].Version++
		writeSet.Groups["Application"].Groups[org].Values["MSP"].Value = mspBytes
	}

	return readSet, writeSet
}

func updateChannelConfig(t *testing.T, readSet *common.ConfigGroup, writeSet *common.ConfigGroup, resmgmtClient *resmgmt.Client, org1MspClient, org2MspClient *mspclient.Client) {

	//read block template and update read/write sets
	txBytes, err := ioutil.ReadFile(integration.GetChannelConfigPath("twoorgs.genesis.block"))
	require.NoError(t, err)

	block := &common.Block{}
	err = proto.Unmarshal(txBytes, block)
	require.NoError(t, err)

	configUpdateEnv, err := resource.CreateConfigUpdateEnvelope(block.Data.Data[0])
	require.NoError(t, err)

	configUpdate := &common.ConfigUpdate{}
	proto.Unmarshal(configUpdateEnv.ConfigUpdate, configUpdate)
	configUpdate.ChannelId = channelID
	configUpdate.ReadSet = readSet
	configUpdate.WriteSet = writeSet

	rawBytes, err := proto.Marshal(configUpdate)
	require.NoError(t, err)

	configUpdateEnv.ConfigUpdate = rawBytes
	configUpdateBytes, err := proto.Marshal(configUpdateEnv)
	require.NoError(t, err)

	//create config envelope
	reader := createConfigEnvelopeReader(t, block.Data.Data[0], configUpdateBytes)

	org1AdminIdentity, err := org1MspClient.GetSigningIdentity(org1AdminUser)
	require.NoError(t, err, "failed to get org1AdminIdentity")

	org2AdminIdenity, err := org2MspClient.GetSigningIdentity(org2AdminUser)
	require.NoError(t, err, "failed to get org2AdminIdentity")

	require.NoError(t, err, "failed to get a new channel management client for org1Admin")

	//perform save channel for channel config update
	req := resmgmt.SaveChannelRequest{ChannelID: channelID,
		ChannelConfig:     reader,
		SigningIdentities: []msp2.SigningIdentity{org1AdminIdentity, org2AdminIdenity}}
	txID, err := resmgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))

	require.Nil(t, err, "error should be nil for SaveChannel ")
	require.NotEmpty(t, txID, "transaction ID should be populated ")
}

func createConfigEnvelopeReader(t *testing.T, blockData []byte, configUpdateBytes []byte) io.Reader {
	envelope := &common.Envelope{}
	err := proto.Unmarshal(blockData, envelope)
	require.NoError(t, err)

	payload := &common.Payload{}
	err = proto.Unmarshal(envelope.Payload, payload)
	require.NoError(t, err)

	payload.Data = configUpdateBytes
	payloadBytes, err := proto.Marshal(payload)
	require.NoError(t, err)

	envelope.Payload = payloadBytes
	envelopeBytes, err := proto.Marshal(envelope)
	require.NoError(t, err)

	reader := bytes.NewReader(envelopeBytes)
	return reader
}

func joinChannel(t *testing.T, org1ResMgmt, org2ResMgmt *resmgmt.Client) {

	// Org1 peers join channel
	if err := org1ResMgmt.JoinChannel("orgchannel", resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com")); err != nil {
		t.Fatalf("Org1 peers failed to JoinChannel: %s", err)
	}

	// Org2 peers join channel
	if err := org2ResMgmt.JoinChannel("orgchannel", resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com")); err != nil {
		t.Fatalf("Org2 peers failed to JoinChannel: %s", err)
	}

}

func queryCC(t *testing.T, org1ChannelClientContext contextAPI.ChannelProvider) {
	// Org1 user connects to 'orgchannel'
	chClientOrg1User, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org1 user: %s", err)
	}
	// Org1 user queries initial value on both peers
	// Since one of the peers on channel has certificate revoked, eror is expected here
	// Error in container is :
	// .... identity 0 does not satisfy principal:
	// Could not validate identity against certification chain, err The certificate has been revoked
	_, err = chClientOrg1User.Query(channel.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCDefaultQueryArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err == nil {
		t.Fatal("Expected error: '....Description: could not find chaincode with name 'exampleCC',,, ")
	}
}

func createCC(t *testing.T, org1ResMgmt *resmgmt.Client, org2ResMgmt *resmgmt.Client) {
	ccPkg, err := packager.NewCCPackage("github.com/example_cc", integration.GetDeployPath())
	if err != nil {
		t.Fatal(err)
	}
	installCCReq := resmgmt.InstallCCRequest{Name: "exampleCC", Path: "github.com/example_cc", Version: "0", Package: ccPkg}
	// Install example cc to Org1 peers
	_, err = org1ResMgmt.InstallCC(installCCReq)
	if err != nil {
		t.Fatal(err)
	}
	// Install example cc to Org2 peers
	_, err = org2ResMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	// Set up chaincode policy to 'two-of-two msps'
	ccPolicy, err := cauthdsl.FromString("AND ('Org1MSP.member','Org2MSP.member')")
	require.NoErrorf(t, err, "Error creating cc policy with both orgs to approve")
	// Org1 resource manager will instantiate 'example_cc' on 'orgchannel'
	_, err = org1ResMgmt.InstantiateCC(
		"orgchannel",
		resmgmt.InstantiateCCRequest{
			Name:    "exampleCC",
			Path:    "github.com/example_cc",
			Version: "0",
			Args:    integration.ExampleCCInitArgs(),
			Policy:  ccPolicy,
		},
	)
	require.Errorf(t, err, "Expecting error instantiating CC on peer with revoked certificate")
	stat, ok := status.FromError(err)
	require.Truef(t, ok, "Expecting error to be a status error, but got ", err)
	require.Equalf(t, stat.Code, int32(status.SignatureVerificationFailed), "Expecting signature verification error due to revoked cert, but got", err)
	require.Truef(t, strings.Contains(err.Error(), "the creator certificate is not valid"), "Expecting error message to contain 'the creator certificate is not valid' but got", err)
}

func loadOrgPeers(t *testing.T, ctxProvider contextAPI.ClientProvider) {

	ctx, err := ctxProvider()
	if err != nil {
		t.Fatalf("context creation failed: %s", err)
	}

	org1Peers, ok := ctx.EndpointConfig().PeersConfig(org1)
	assert.True(t, ok)

	org2Peers, ok := ctx.EndpointConfig().PeersConfig(org2)
	assert.True(t, ok)

	orgTestPeer0, err = ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: org1Peers[0]})
	if err != nil {
		t.Fatal(err)
	}

	orgTestPeer1, err = ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: org2Peers[0]})
	if err != nil {
		t.Fatal(err)
	}

}

func queryRevocationListUpdates(t *testing.T, client *ledger.Client, config fab.EndpointConfig, chID string) bool {
	installed, err := retry.NewInvoker(retry.New(CRLTestRetryOpts)).Invoke(
		func() (interface{}, error) {
			ok := isChannelConfigUpdated(t, client, config, chID)
			if !ok {
				return &ok, status.New(status.TestStatus, status.GenericTransient.ToInt32(), "Revocation list is not updated in all peers", nil)
			}
			return &ok, nil
		},
	)

	require.NoErrorf(t, err, "Got error checking if chaincode was installed")
	return *(installed).(*bool)
}

func isChannelConfigUpdated(t *testing.T, client *ledger.Client, config fab.EndpointConfig, chID string) bool {
	chPeers, ok := config.ChannelPeers(chID)
	if !ok {
		return false
	}
	t.Logf("Performing config update check on %d channel peers in channel '%s'", len(chPeers), chID)
	updated := len(chPeers) > 0
	for _, chPeer := range chPeers {
		t.Logf("waiting for [%s] msp update", chPeer.URL)
		chCfg, err := client.QueryConfig(ledger.WithTargetEndpoints(chPeer.URL))
		if err != nil || len(chCfg.MSPs()) == 0 {
			return false
		}
		for _, mspCfg := range chCfg.MSPs() {
			fabMspCfg := &msp.FabricMSPConfig{}
			err = proto.Unmarshal(mspCfg.Config, fabMspCfg)
			if err != nil {
				return false
			}
			if fabMspCfg.Name == "OrdererMSP" {
				continue
			}
			t.Logf("length of revocation list found in peer[%s] is %d", chPeer.URL, len(fabMspCfg.RevocationList))
			updated = updated && len(fabMspCfg.RevocationList) > 0
		}
	}
	t.Logf("check result :%v \n\n", updated)
	return updated
}

func generateCRL() ([]byte, error) {

	root := integration.GetCryptoConfigPath(pathRevokeCaRoot)
	var parentKey string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "_sk") {
			parentKey = path
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	key, err := loadPrivateKey(parentKey)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to load private key")
	}

	cert, err := loadCert(integration.GetCryptoConfigPath(pathParentCert))
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to load cert")
	}

	certToBeRevoked, err := loadCert(integration.GetCryptoConfigPath(pathCertToBeRevokd))
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to load cert")
	}

	crlBytes, err := revokeCert(certToBeRevoked, cert, key)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to revoke cert")
	}

	return crlBytes, nil
}

func loadPrivateKey(path string) (interface{}, error) {

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	key, err := utils.PEMtoPrivateKey(raw, []byte(""))
	if err != nil {
		return nil, err
	}

	return key, nil
}

func loadCert(path string) (*x509.Certificate, error) {

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}

	return x509.ParseCertificate(block.Bytes)
}

func revokeCert(certToBeRevoked *x509.Certificate, parentCert *x509.Certificate, parentKey interface{}) ([]byte, error) {

	//Create a revocation record for the user
	clientRevocation := pkix.RevokedCertificate{
		SerialNumber:   certToBeRevoked.SerialNumber,
		RevocationTime: time.Now().UTC(),
	}

	curRevokedCertificates := []pkix.RevokedCertificate{clientRevocation}
	//Generate new CRL that includes the user's revocation
	newCrlList, err := parentCert.CreateCRL(rand.Reader, parentKey, curRevokedCertificates, time.Now().UTC(), time.Now().UTC().AddDate(20, 0, 0))
	if err != nil {
		return nil, err
	}

	//CRL pem Block
	crlPemBlock := &pem.Block{
		Type:  "X509 CRL",
		Bytes: newCrlList,
	}
	var crlBuffer bytes.Buffer
	//Encode it to X509 CRL pem format print it out
	err = pem.Encode(&crlBuffer, crlPemBlock)
	if err != nil {
		return nil, err
	}

	return crlBuffer.Bytes(), nil
}

//getCustomConfigBackend overrides test config backend with a backend which skips 'peer0.org1' so that
// all traffic goes to peer1.org1
func getCustomConfigBackend() core.ConfigProvider {

	return func() ([]core.ConfigBackend, error) {
		configBackends, err := config.FromFile(integration.GetConfigPath(configFilename))()
		if err != nil {
			return nil, errors.Wrap(err, "failed to read config backend from file, %v")
		}

		configBackendsOverrides, err := config.FromFile(integration.GetConfigPath(configFileOverride))()
		if err != nil {
			return nil, errors.Wrap(err, "failed to read config backend from file, %v")
		}

		return append(configBackendsOverrides, configBackends...), nil
	}
}
