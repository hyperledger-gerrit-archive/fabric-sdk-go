/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabnet"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// MockClient ...
type MockClient struct {
	channels    map[string]fab.Channel
	cryptoSuite bccsp.BCCSP
	stateStore  fab.KeyValueStore
	userContext fab.User
	config      config.Config
}

// NewMockClient ...
/*
 * Returns a FabricClient instance
 */
func NewMockClient() fab.FabricClient {
	channels := make(map[string]fab.Channel)
	c := &MockClient{channels: channels, cryptoSuite: nil, stateStore: nil, userContext: nil, config: NewMockConfig()}
	return c
}

// NewChannel ...
func (c *MockClient) NewChannel(name string) (fab.Channel, error) {
	return nil, nil
}

// GetChannel ...
func (c *MockClient) GetChannel(name string) fab.Channel {
	return c.channels[name]
}

// GetConfig ...
func (c *MockClient) GetConfig() config.Config {
	return c.config
}

// QueryChannelInfo ...
func (c *MockClient) QueryChannelInfo(name string, peers []fab.Peer) (fab.Channel, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

// SetStateStore ...
func (c *MockClient) SetStateStore(stateStore fab.KeyValueStore) {
	c.stateStore = stateStore
}

// GetStateStore ...
func (c *MockClient) GetStateStore() fab.KeyValueStore {
	return c.stateStore
}

// SetCryptoSuite ...
func (c *MockClient) SetCryptoSuite(cryptoSuite bccsp.BCCSP) {
	c.cryptoSuite = cryptoSuite
}

// GetCryptoSuite ...
func (c *MockClient) GetCryptoSuite() bccsp.BCCSP {
	return c.cryptoSuite
}

// SaveUserToStateStore ...
func (c *MockClient) SaveUserToStateStore(user fab.User, skipPersistence bool) error {
	return fmt.Errorf("Not implemented yet")

}

// LoadUserFromStateStore ...
func (c *MockClient) LoadUserFromStateStore(name string) (fab.User, error) {
	return NewMockUser("test"), nil
}

// ExtractChannelConfig ...
func (c *MockClient) ExtractChannelConfig(configEnvelope []byte) ([]byte, error) {
	return nil, fmt.Errorf("Not implemented yet")

}

// SignChannelConfig ...
func (c *MockClient) SignChannelConfig(config []byte) (*common.ConfigSignature, error) {
	return nil, fmt.Errorf("Not implemented yet")

}

// CreateChannel ...
func (c *MockClient) CreateChannel(request *fab.CreateChannelRequest) error {
	return fmt.Errorf("Not implemented yet")

}

// CreateOrUpdateChannel ...
func (c *MockClient) CreateOrUpdateChannel(request *fab.CreateChannelRequest, haveEnvelope bool) error {
	return fmt.Errorf("Not implemented yet")

}

//QueryChannels ...
func (c *MockClient) QueryChannels(peer fab.Peer) (*pb.ChannelQueryResponse, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

//QueryInstalledChaincodes ...
func (c *MockClient) QueryInstalledChaincodes(peer fab.Peer) (*pb.ChaincodeQueryResponse, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

// InstallChaincode ...
func (c *MockClient) InstallChaincode(chaincodeName string, chaincodePath string, chaincodeVersion string,
	chaincodePackage []byte, targets []fab.Peer) ([]*apitxn.TransactionProposalResponse, string, error) {
	return nil, "", fmt.Errorf("Not implemented yet")

}

// GetIdentity returns MockClient's serialized identity
func (c *MockClient) GetIdentity() ([]byte, error) {
	return []byte("test"), nil

}

// GetUserContext ...
func (c *MockClient) GetUserContext() fab.User {
	return c.userContext
}

// SetUserContext ...
func (c *MockClient) SetUserContext(user fab.User) {
	c.userContext = user
}
