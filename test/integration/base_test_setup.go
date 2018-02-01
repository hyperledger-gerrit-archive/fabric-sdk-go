/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// BaseSetupImpl implementation of BaseTestSetup
type BaseSetupImpl struct {
	SDK             *fabsdk.FabricSDK
	Identity        fab.IdentityContext
	Client          fab.Resource
	Channel         fab.Channel
	EventHub        fab.EventHub
	ConnectEventHub bool
	ConfigFile      string
	OrgID           string
	ChannelID       string
	ChainCodeID     string
	Initialized     bool
	ChannelConfig   string
}

// Initial B values for ExampleCC
const (
	ExampleCCInitB    = "200"
	ExampleCCUpgradeB = "400"
)

// ExampleCC query and transaction arguments
var queryArgs = [][]byte{[]byte("query"), []byte("b")}
var txArgs = [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}

// ExampleCC init and upgrade args
var initArgs = [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b"), []byte(ExampleCCInitB)}
var upgradeArgs = [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b"), []byte(ExampleCCUpgradeB)}

var resMgmtClient resmgmt.ResourceMgmtClient

// ExampleCCQueryArgs returns example cc query args
func ExampleCCQueryArgs() [][]byte {
	return queryArgs
}

// ExampleCCTxArgs returns example cc move funds args
func ExampleCCTxArgs() [][]byte {
	return txArgs
}

//ExampleCCInitArgs returns example cc initialization args
func ExampleCCInitArgs() [][]byte {
	return initArgs
}

//ExampleCCUpgradeArgs returns example cc upgrade args
func ExampleCCUpgradeArgs() [][]byte {
	return upgradeArgs
}

// Initialize reads configuration from file and sets up client, channel and event hub
func (setup *BaseSetupImpl) Initialize(t *testing.T) error {
	// Create SDK setup for the integration tests
	sdk, err := fabsdk.New(config.FromFile(setup.ConfigFile))
	if err != nil {
		return errors.WithMessage(err, "SDK init failed")
	}
	setup.SDK = sdk

	session, err := sdk.NewClient(fabsdk.WithUser("Admin"), fabsdk.WithOrg(setup.OrgID)).Session()
	if err != nil {
		return errors.WithMessage(err, "failed getting admin user session for org")
	}

	setup.Identity = session

	sc, err := sdk.FabricProvider().NewResourceClient(setup.Identity)
	if err != nil {
		return errors.WithMessage(err, "NewResourceClient failed")
	}

	setup.Client = sc

	channel, err := setup.GetChannel(sdk, setup.Identity, sdk.Config(), setup.ChannelID, []string{setup.OrgID})
	if err != nil {
		return errors.Wrapf(err, "create channel (%s) failed: %v", setup.ChannelID)
	}
	setup.Channel = channel

	// Channel management client is responsible for managing channels (create/update)
	chMgmtClient, err := sdk.NewClient(fabsdk.WithUser("Admin"), fabsdk.WithOrg("ordererorg")).ChannelMgmt()
	if err != nil {
		t.Fatalf("Failed to create new channel management client: %s", err)
	}

	// Resource management client is responsible for managing resources (joining channels, install/instantiate/upgrade chaincodes)
	resMgmtClient, err = sdk.NewClient(fabsdk.WithUser("Admin")).ResourceMgmt()
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Check if primary peer has joined channel
	alreadyJoined, err := HasPrimaryPeerJoinedChannel(sc, channel)
	if err != nil {
		return errors.WithMessage(err, "failed while checking if primary peer has already joined channel")
	}

	if !alreadyJoined {

		// Create channel (or update if it already exists)
		req := chmgmt.SaveChannelRequest{ChannelID: setup.ChannelID, ChannelConfig: setup.ChannelConfig, SigningIdentity: session}

		if err = chMgmtClient.SaveChannel(req); err != nil {
			return errors.WithMessage(err, "SaveChannel failed")
		}

		time.Sleep(time.Second * 5)

		if err = channel.Initialize(nil); err != nil {
			return errors.WithMessage(err, "channel init failed")
		}

		if err = resMgmtClient.JoinChannel(setup.ChannelID); err != nil {
			return errors.WithMessage(err, "JoinChannel failed")
		}
	}

	if err := setup.setupEventHub(t, sdk, setup.Identity); err != nil {
		return err
	}

	setup.Initialized = true

	return nil
}

// GetChannel initializes and returns a channel based on config
func (setup *BaseSetupImpl) GetChannel(sdk *fabsdk.FabricSDK, ic fab.IdentityContext, config apiconfig.Config, channelID string, orgs []string) (fab.Channel, error) {

	channel, err := sdk.FabricProvider().NewChannelClient(ic, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "NewChannel failed")
	}

	for _, org := range orgs {
		peerConfig, err := config.PeersConfig(org)
		if err != nil {
			return nil, errors.WithMessage(err, "reading peer config failed")
		}
		for _, p := range peerConfig {
			endorser, err := sdk.FabricProvider().NewPeerFromConfig(&apiconfig.NetworkPeer{PeerConfig: p})
			if err != nil {
				return nil, errors.WithMessage(err, "NewPeer failed")
			}
			err = channel.AddPeer(endorser)
			if err != nil {
				return nil, errors.WithMessage(err, "adding peer failed")
			}
		}
	}

	return channel, nil
}

func (setup *BaseSetupImpl) setupEventHub(t *testing.T, client *fabsdk.FabricSDK, identity fab.IdentityContext) error {
	eventHub, err := client.FabricProvider().NewEventHub(identity, setup.ChannelID)
	if err != nil {
		return err
	}

	if setup.ConnectEventHub {
		if err := eventHub.Connect(); err != nil {
			return errors.WithMessage(err, "eventHub connect failed")
		}
	}
	setup.EventHub = eventHub

	return nil
}

// InitConfig ...
func (setup *BaseSetupImpl) InitConfig() apiconfig.ConfigProvider {
	return config.FromFile(setup.ConfigFile)
}

// InstallCC use low level client to install chaincode
func (setup *BaseSetupImpl) InstallCC(name string, path string, version string, ccPackage *fab.CCPackage) error {

	icr := fab.InstallChaincodeRequest{Name: name, Path: path, Version: version, Package: ccPackage, Targets: peer.PeersToTxnProcessors(setup.Channel.Peers())}

	transactionProposalResponse, _, err := setup.Client.InstallChaincode(icr)

	if err != nil {
		return errors.WithMessage(err, "InstallChaincode failed")
	}
	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return errors.WithMessage(v.Err, "InstallChaincode endorser failed")
		}
	}

	return nil
}

// GetDeployPath ..
func (setup *BaseSetupImpl) GetDeployPath() string {
	pwd, _ := os.Getwd()
	return path.Join(pwd, "../../fixtures/testdata")
}

// InstallAndInstantiateExampleCC install and instantiate using resource management client
func (setup *BaseSetupImpl) InstallAndInstantiateExampleCC() error {

	if setup.ChainCodeID == "" {
		setup.ChainCodeID = GenerateRandomID()
	}

	return setup.InstallAndInstantiateCC(setup.ChainCodeID, "github.com/example_cc", "v0", setup.GetDeployPath(), initArgs)
}

// InstallAndInstantiateCC install and instantiate using resource management client
func (setup *BaseSetupImpl) InstallAndInstantiateCC(ccName, ccPath, ccVersion, goPath string, ccArgs [][]byte) error {

	ccPkg, err := packager.NewCCPackage(ccPath, goPath)
	if err != nil {
		return err
	}

	_, err = resMgmtClient.InstallCC(resmgmt.InstallCCRequest{Name: ccName, Path: ccPath, Version: ccVersion, Package: ccPkg})
	if err != nil {
		return err
	}

	ccPolicy := cauthdsl.SignedByMspMember(setup.Identity.MspID())
	return resMgmtClient.InstantiateCC("mychannel", resmgmt.InstantiateCCRequest{Name: ccName, Path: ccPath, Version: ccVersion, Args: ccArgs, Policy: ccPolicy})
}

// CreateAndSendTransactionProposal ... TODO duplicate
func (setup *BaseSetupImpl) CreateAndSendTransactionProposal(channel fab.Channel, chainCodeID string,
	fcn string, args [][]byte, targets []apitxn.ProposalProcessor, transientData map[string][]byte) ([]*apitxn.TransactionProposalResponse, apitxn.TransactionID, error) {

	request := apitxn.ChaincodeInvokeRequest{
		Targets:      targets,
		Fcn:          fcn,
		Args:         args,
		TransientMap: transientData,
		ChaincodeID:  chainCodeID,
	}
	transactionProposalResponses, txnID, err := channel.SendTransactionProposal(request)
	if err != nil {
		return nil, txnID, err
	}

	for _, v := range transactionProposalResponses {
		if v.Err != nil {
			return nil, txnID, errors.Wrapf(v.Err, "endorser %s failed", v.Endorser)
		}
	}

	return transactionProposalResponses, txnID, nil
}

// CreateAndSendTransaction ...
func (setup *BaseSetupImpl) CreateAndSendTransaction(channel fab.Channel, resps []*apitxn.TransactionProposalResponse) (*apitxn.TransactionResponse, error) {

	tx, err := channel.CreateTransaction(resps)
	if err != nil {
		return nil, errors.WithMessage(err, "CreateTransaction failed")
	}

	transactionResponse, err := channel.SendTransaction(tx)
	if err != nil {
		return nil, errors.WithMessage(err, "SendTransaction failed")

	}

	if transactionResponse.Err != nil {
		return nil, errors.Wrapf(transactionResponse.Err, "orderer %s failed", transactionResponse.Orderer)
	}

	return transactionResponse, nil
}

// RegisterTxEvent registers on the given eventhub for the give transaction
// returns a boolean channel which receives true when the event is complete
// and an error channel for errors
// TODO - Duplicate
func (setup *BaseSetupImpl) RegisterTxEvent(t *testing.T, txID apitxn.TransactionID, eventHub fab.EventHub) (chan bool, chan error) {
	done := make(chan bool)
	fail := make(chan error)

	eventHub.RegisterTxEvent(txID, func(txId string, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			t.Logf("Received error event for txid(%s)", txId)
			fail <- err
		} else {
			t.Logf("Received success event for txid(%s)", txId)
			done <- true
		}
	})

	return done, fail
}
