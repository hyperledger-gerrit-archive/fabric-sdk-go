/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
)

// MockClient ...
type MockClient struct {
	fab.ProviderContext
	fab.IdentityContext

	channels       map[string]fab.Channel
	stateStore     fab.KeyValueStore
	errorScenario  bool
	signingManager fab.SigningManager
}

// NewMockClient ...
/*
 * Returns a FabricClient instance
 */
func NewMockClient() *MockClient {
	channels := make(map[string]fab.Channel)
	pc := NewMockProviderContext()
	c := &MockClient{
		channels:        channels,
		ProviderContext: pc,
	}
	return c
}

//NewMockInvalidClient : Returns new Mock FabricClient with error flag on used to test invalid scenarios
func NewMockInvalidClient() *MockClient {
	channels := make(map[string]fab.Channel)
	pc := NewMockProviderContext()
	c := &MockClient{
		channels:        channels,
		ProviderContext: pc,
		errorScenario:   true,
	}
	return c
}

// NewChannel ...
func (c *MockClient) NewChannel(name string) (fab.Channel, error) {
	if name == "error" {
		return nil, errors.New("Genererate error in new channel")
	}
	return nil, nil
}

// SetChannel convenience method to set channel
func (c *MockClient) SetChannel(id string, channel fab.Channel) {
	c.channels[id] = channel
}

// Channel ...
func (c *MockClient) Channel(id string) fab.Channel {
	return c.channels[id]
}

// SetConfig changes the configuration of the mock client.
func (c *MockClient) SetConfig(config config.Config) {
	mockPc := c.ProviderContext.(*MockProviderContext)
	mockPc.config = config
}

// SetStateStore ...
func (c *MockClient) SetStateStore(stateStore fab.KeyValueStore) {
	c.stateStore = stateStore
}

// StateStore ...
func (c *MockClient) StateStore() fab.KeyValueStore {
	return c.stateStore
}

// SetSigningManager mocks setting signing manager
func (c *MockClient) SetSigningManager(signingMgr fab.SigningManager) {
	mockPc := c.ProviderContext.(*MockProviderContext)
	mockPc.signingManager = signingMgr
}

// SaveUserToStateStore ...
func (c *MockClient) SaveUserToStateStore(user fab.User) error {
	return errors.New("Not implemented yet")

}

// LoadUserFromStateStore ...
func (c *MockClient) LoadUserFromStateStore(name string) (fab.User, error) {
	if c.errorScenario {
		return nil, errors.New("just to test error scenario")
	}
	return NewMockUser("test"), nil
}
