/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"

	api "github.com/hyperledger/fabric-sdk-go/api"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// MockClient ...
type MockClient struct {
	chains      map[string]api.Chain
	cryptoSuite bccsp.BCCSP
	stateStore  api.KeyValueStore
	userContext api.User
}

// NewMockClient ...
/*
 * Returns a Client instance
 */
func NewMockClient() api.Client {
	chains := make(map[string]api.Chain)
	c := &MockClient{chains: chains, cryptoSuite: nil, stateStore: nil, userContext: nil}
	return c
}

// NewChain ...
func (c *MockClient) NewChain(name string) (api.Chain, error) {
	//	if _, ok := c.chains[name]; ok {
	//		return nil, fmt.Errorf("Chain %s already exists", name)
	//	}
	//	var err error
	//	c.chains[name], err = chain.NewChain(name, c)
	//	if err != nil {
	//		return nil, err
	//	}
	//	return c.chains[name], nil
	return nil, nil
}

// GetChain ...
func (c *MockClient) GetChain(name string) api.Chain {
	return c.chains[name]
}

// QueryChainInfo ...
func (c *MockClient) QueryChainInfo(name string, peers []api.Peer) (api.Chain, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

// SetStateStore ...
func (c *MockClient) SetStateStore(stateStore api.KeyValueStore) {
	c.stateStore = stateStore
}

// GetStateStore ...
func (c *MockClient) GetStateStore() api.KeyValueStore {
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
func (c *MockClient) SaveUserToStateStore(user api.User, skipPersistence bool) error {
	return fmt.Errorf("Not implemented yet")

}

// LoadUserFromStateStore ...
func (c *MockClient) LoadUserFromStateStore(name string) (api.User, error) {
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
func (c *MockClient) CreateChannel(request *api.CreateChannelRequest) error {
	return fmt.Errorf("Not implemented yet")

}

// CreateOrUpdateChannel ...
func (c *MockClient) CreateOrUpdateChannel(request *api.CreateChannelRequest, haveEnvelope bool) error {
	return fmt.Errorf("Not implemented yet")

}

//QueryChannels ...
func (c *MockClient) QueryChannels(peer api.Peer) (*pb.ChannelQueryResponse, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

//QueryInstalledChaincodes ...
func (c *MockClient) QueryInstalledChaincodes(peer api.Peer) (*pb.ChaincodeQueryResponse, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

// InstallChaincode ...
func (c *MockClient) InstallChaincode(chaincodeName string, chaincodePath string, chaincodeVersion string,
	chaincodePackage []byte, targets []api.Peer) ([]*api.TransactionProposalResponse, string, error) {
	return nil, "", fmt.Errorf("Not implemented yet")

}

// GetIdentity returns MockClient's serialized identity
func (c *MockClient) GetIdentity() ([]byte, error) {
	return []byte("test"), nil

}

// GetUserContext ...
func (c *MockClient) GetUserContext() api.User {
	return c.userContext
}

// SetUserContext ...
func (c *MockClient) SetUserContext(user api.User) {
	c.userContext = user
}
