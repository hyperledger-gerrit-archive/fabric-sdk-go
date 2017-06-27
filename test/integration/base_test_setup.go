/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"

	api "github.com/hyperledger/fabric-sdk-go/api"
	fabricTxn "github.com/hyperledger/fabric-sdk-go/fabric-txn"
	admin "github.com/hyperledger/fabric-sdk-go/fabric-txn/admin"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// BaseSetupImpl implementation of BaseTestSetup
type BaseSetupImpl struct {
	Client          api.FabricClient
	Channel         api.Channel
	EventHub        api.EventHub
	ConnectEventHub bool
	ConfigFile      string
	ChannelID       string
	ChainCodeID     string
	Initialized     bool
	ChannelConfig   string
	AdminUser       api.User
	NormalUser      api.User
}

// Initialize reads configuration from file and sets up client, channel and event hub
func (setup *BaseSetupImpl) Initialize() error {
	configImpl, err := setup.InitConfig()
	if err != nil {
		return fmt.Errorf("Init from config failed: %v", err)
	}

	// Initialize bccsp factories before calling get client
	err = bccspFactory.InitFactories(configImpl.GetCSPConfig())
	if err != nil {
		return fmt.Errorf("Failed getting ephemeral software-based BCCSP [%s]", err)
	}

	mspClient, err := fabricTxn.GetDefaultImplMspClient(configImpl)
	if err != nil {
		return fmt.Errorf("Failed to get default msp client: %v", err)
	}

	client, err := fabricTxn.GetDefaultImplClientWithUser("admin", "adminpw", "/tmp/enroll_user", configImpl, mspClient)
	if err != nil {
		return fmt.Errorf("Create client failed: %v", err)
	}

	setup.Client = client

	org1Admin, err := GetAdmin(client, "org1")
	org1User, err := GetUser(client, "org1")
	if err != nil {
		return fmt.Errorf("Error getting org admin user: %v", err)
	}

	setup.AdminUser = org1Admin
	setup.NormalUser = org1User

	channel, err := setup.GetChannel(setup.Client, setup.ChannelID)
	if err != nil {
		return fmt.Errorf("Create channel (%s) failed: %v", setup.ChannelID, err)
	}
	setup.Channel = channel

	ordererAdmin, err := GetOrdererAdmin(client)
	if err != nil {
		return fmt.Errorf("Error getting orderer admin user: %v", err)
	}

	// Create and join channel
	if err = admin.CreateChannel(client, ordererAdmin, org1Admin, channel, setup.ChannelConfig); err != nil {
		return fmt.Errorf("CreateChannel returned error: %v", err)
	}
	time.Sleep(time.Second * 3)
	if err = admin.JoinChannel(client, org1Admin, channel); err != nil {
		return fmt.Errorf("JoinChannel returned error: %v", err)
	}

	//by default client's user context should use regular user, for admin actions, UserContext must be set to AdminUser
	client.SetUserContext(org1User)

	if err := setup.setupEventHub(client); err != nil {
		return err
	}

	setup.Initialized = true

	return nil
}

func (setup *BaseSetupImpl) setupEventHub(client api.FabricClient) error {
	eventHub, err := getEventHub(client)
	if err != nil {
		return err
	}

	if setup.ConnectEventHub {
		if err := eventHub.Connect(); err != nil {
			return fmt.Errorf("Failed eventHub.Connect() [%s]", err)
		}
	}
	setup.EventHub = eventHub

	return nil
}

// InitConfig ...
func (setup *BaseSetupImpl) InitConfig() (api.Config, error) {
	configImpl, err := config.InitConfig(setup.ConfigFile)
	if err != nil {
		return nil, err
	}
	return configImpl, nil
}

// InstantiateCC ...
func (setup *BaseSetupImpl) InstantiateCC(chainCodeID string, channelID string, chainCodePath string, chainCodeVersion string, args []string) error {
	// InstantiateCC requires AdminUser privileges so setting user context with Admin User
	setup.Client.SetUserContext(setup.AdminUser)

	// must reset client user context to normal user once done with Admin privilieges
	defer setup.Client.SetUserContext(setup.NormalUser)

	if err := admin.SendInstantiateCC(setup.Channel, chainCodeID, channelID, args, chainCodePath, chainCodeVersion, []api.Peer{setup.Channel.GetPrimaryPeer()}, setup.EventHub); err != nil {
		return err
	}
	return nil
}

// InstallCC ...
func (setup *BaseSetupImpl) InstallCC(chainCodeID string, chainCodePath string, chainCodeVersion string, chaincodePackage []byte) error {
	// installCC requires AdminUser privileges so setting user context with Admin User
	setup.Client.SetUserContext(setup.AdminUser)

	// must reset client user context to normal user once done with Admin privilieges
	defer setup.Client.SetUserContext(setup.NormalUser)

	if err := admin.SendInstallCC(setup.Client, setup.Channel, chainCodeID, chainCodePath, chainCodeVersion, chaincodePackage, setup.Channel.GetPeers(), setup.GetDeployPath()); err != nil {
		return fmt.Errorf("SendInstallProposal return error: %v", err)
	}

	return nil
}

// GetDeployPath ..
func (setup *BaseSetupImpl) GetDeployPath() string {
	pwd, _ := os.Getwd()
	return path.Join(pwd, "../fixtures")
}

// InstallAndInstantiateExampleCC ..
func (setup *BaseSetupImpl) InstallAndInstantiateExampleCC() error {

	chainCodePath := "github.com/example_cc"
	chainCodeVersion := "v0"

	if setup.ChainCodeID == "" {
		setup.ChainCodeID = GenerateRandomID()
	}

	if err := setup.InstallCC(setup.ChainCodeID, chainCodePath, chainCodeVersion, nil); err != nil {
		return err
	}

	var args []string
	args = append(args, "init")
	args = append(args, "a")
	args = append(args, "100")
	args = append(args, "b")
	args = append(args, "200")

	return setup.InstantiateCC(setup.ChainCodeID, setup.ChannelID, chainCodePath, chainCodeVersion, args)
}

// Query ...
func (setup *BaseSetupImpl) Query(channelID string, chainCodeID string, args []string) (string, error) {
	return fabricTxn.QueryChaincode(setup.Client, setup.Channel, chainCodeID, args)
}

// QueryAsset ...
func (setup *BaseSetupImpl) QueryAsset() (string, error) {

	var args []string
	args = append(args, "invoke")
	args = append(args, "query")
	args = append(args, "b")
	return setup.Query(setup.ChannelID, setup.ChainCodeID, args)
}

// MoveFunds ...
func (setup *BaseSetupImpl) MoveFunds() error {

	var args []string
	args = append(args, "invoke")
	args = append(args, "move")
	args = append(args, "a")
	args = append(args, "b")
	args = append(args, "1")

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	return fabricTxn.InvokeChaincode(setup.Client, setup.Channel, []api.Peer{setup.Channel.GetPrimaryPeer()}, setup.EventHub, setup.ChainCodeID, args, transientDataMap)
}

// MoveFundsAndGetTxID ...
func (setup *BaseSetupImpl) MoveFundsAndGetTxID() (string, error) {

	var args []string
	args = append(args, "invoke")
	args = append(args, "move")
	args = append(args, "a")
	args = append(args, "b")
	args = append(args, "1")

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	transactionProposalResponse, txID, err := setup.CreateAndSendTransactionProposal(setup.Channel, setup.ChainCodeID, setup.ChannelID, args, []api.Peer{setup.Channel.GetPrimaryPeer()}, transientDataMap)
	if err != nil {
		return "", fmt.Errorf("CreateAndSendTransactionProposal return error: %v", err)
	}
	// Register for commit event
	done, fail := setup.RegisterTxEvent(txID, setup.EventHub)

	txResponse, err := setup.CreateAndSendTransaction(setup.Channel, transactionProposalResponse)
	if err != nil {
		return "", fmt.Errorf("CreateAndSendTransaction return error: %v", err)
	}
	fmt.Println(txResponse)
	select {
	case <-done:
	case <-fail:
		return "", fmt.Errorf("invoke Error received from eventhub for txid(%s) error(%v)", txID, fail)
	case <-time.After(time.Second * 30):
		return "", fmt.Errorf("invoke Didn't receive block event for txid(%s)", txID)
	}
	return txID, nil
}

// GetChannel initializes and returns a channel based on config
func (setup *BaseSetupImpl) GetChannel(client api.FabricClient, channelID string) (api.Channel, error) {

	channel, err := client.NewChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("NewChannel return error: %v", err)
	}
	orderer, err := fabricTxn.GetDefaultImplOrderer(fmt.Sprintf("%s:%s", client.GetConfig().GetOrdererHost(), client.GetConfig().GetOrdererPort()),
		client.GetConfig().GetOrdererTLSCertificate(), client.GetConfig().GetOrdererTLSServerHostOverride(), client.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("NewOrderer return error: %v", err)
	}
	err = channel.AddOrderer(orderer)
	if err != nil {
		return nil, fmt.Errorf("Error adding orderer: %v", err)
	}

	peerConfig, err := client.GetConfig().GetPeersConfig()
	if err != nil {
		return nil, fmt.Errorf("Error reading peer config: %v", err)
	}
	for _, p := range peerConfig {
		endorser, err := fabricTxn.GetDefaultImplPeer(fmt.Sprintf("%s:%d", p.Host, p.Port),
			p.TLS.Certificate, p.TLS.ServerHostOverride, client.GetConfig())
		if err != nil {
			return nil, fmt.Errorf("NewPeer return error: %v", err)
		}
		err = channel.AddPeer(endorser)
		if err != nil {
			return nil, fmt.Errorf("Error adding peer: %v", err)
		}
		if p.Primary {
			channel.SetPrimaryPeer(endorser)
		}
	}

	return channel, nil
}

// CreateAndSendTransactionProposal ...
func (setup *BaseSetupImpl) CreateAndSendTransactionProposal(channel api.Channel, chainCodeID string, channelID string,
	args []string, targets []api.Peer, transientData map[string][]byte) ([]*api.TransactionProposalResponse, string, error) {

	signedProposal, err := channel.CreateTransactionProposal(chainCodeID, channelID, args, true, transientData)
	if err != nil {
		return nil, "", fmt.Errorf("SendTransactionProposal returned error: %v", err)
	}

	transactionProposalResponses, err := channel.SendTransactionProposal(signedProposal, 0, targets)
	if err != nil {
		return nil, "", fmt.Errorf("SendTransactionProposal returned error: %v", err)
	}

	for _, v := range transactionProposalResponses {
		if v.Err != nil {
			return nil, signedProposal.TransactionID, fmt.Errorf("invoke Endorser %s returned error: %v", v.Endorser, v.Err)
		}
		fmt.Printf("invoke Endorser '%s' returned ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	return transactionProposalResponses, signedProposal.TransactionID, nil
}

// CreateAndSendTransaction ...
func (setup *BaseSetupImpl) CreateAndSendTransaction(channel api.Channel, resps []*api.TransactionProposalResponse) ([]*api.TransactionResponse, error) {

	tx, err := channel.CreateTransaction(resps)
	if err != nil {
		return nil, fmt.Errorf("CreateTransaction return error: %v", err)
	}

	transactionResponse, err := channel.SendTransaction(tx)
	if err != nil {
		return nil, fmt.Errorf("SendTransaction return error: %v", err)

	}
	for _, v := range transactionResponse {
		if v.Err != nil {
			return nil, fmt.Errorf("Orderer %s return error: %v", v.Orderer, v.Err)
		}
	}

	return transactionResponse, nil
}

// RegisterTxEvent registers on the given eventhub for the give transaction
// returns a boolean channel which receives true when the event is complete
// and an error channel for errors
func (setup *BaseSetupImpl) RegisterTxEvent(txID string, eventHub api.EventHub) (chan bool, chan error) {
	done := make(chan bool)
	fail := make(chan error)

	eventHub.RegisterTxEvent(txID, func(txId string, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			fmt.Printf("Received error event for txid(%s)\n", txId)
			fail <- err
		} else {
			fmt.Printf("Received success event for txid(%s)\n", txId)
			done <- true
		}
	})

	return done, fail
}

// getEventHub initilizes the event hub
func getEventHub(client api.FabricClient) (api.EventHub, error) {
	eventHub, err := events.NewEventHub(client)
	if err != nil {
		return nil, fmt.Errorf("Error creating new event hub: %v", err)
	}
	foundEventHub := false
	peerConfig, err := client.GetConfig().GetPeersConfig()
	if err != nil {
		return nil, fmt.Errorf("Error reading peer config: %v", err)
	}
	for _, p := range peerConfig {
		if p.EventHost != "" && p.EventPort != 0 {
			fmt.Printf("******* EventHub connect to peer (%s:%d) *******\n", p.EventHost, p.EventPort)
			eventHub.SetPeerAddr(fmt.Sprintf("%s:%d", p.EventHost, p.EventPort),
				p.TLS.Certificate, p.TLS.ServerHostOverride)
			foundEventHub = true
			break
		}
	}

	if !foundEventHub {
		return nil, fmt.Errorf("No EventHub configuration found")
	}

	return eventHub, nil
}
