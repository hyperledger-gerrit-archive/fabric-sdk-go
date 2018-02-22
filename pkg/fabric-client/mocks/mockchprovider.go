/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/pkg/errors"
)

// MockChannelProvider holds a mock channel provider.
type MockChannelProvider struct {
	ctx        context.ProviderContext
	channels   map[string]context.Channel
	transactor context.Transactor
}

// MockChannelService holds a mock channel service.
type MockChannelService struct {
	provider   *MockChannelProvider
	channelID  string
	transactor context.Transactor
}

// NewMockChannelProvider returns a mock ChannelProvider
func NewMockChannelProvider(ctx context.Context) (*MockChannelProvider, error) {
	channels := make(map[string]context.Channel)

	// Create a mock client with the mock channel
	cp := MockChannelProvider{
		ctx:      ctx,
		channels: channels,
	}
	return &cp, nil
}

// SetChannel convenience method to set channel
func (cp *MockChannelProvider) SetChannel(id string, channel context.Channel) {
	cp.channels[id] = channel
}

// SetTransactor sets the default transactor for all mock channel services
func (cp *MockChannelProvider) SetTransactor(transactor context.Transactor) {
	cp.transactor = transactor
}

// NewChannelService returns a mock ChannelService
func (cp *MockChannelProvider) NewChannelService(ic context.IdentityContext, channelID string) (context.ChannelService, error) {
	cs := MockChannelService{
		provider:   cp,
		channelID:  channelID,
		transactor: cp.transactor,
	}
	return &cs, nil
}

// EventHub ...
func (cs *MockChannelService) EventHub() (context.EventHub, error) {
	return NewMockEventHub(), nil
}

// Channel ...
func (cs *MockChannelService) Channel() (context.Channel, error) {
	ch, ok := cs.provider.channels[cs.channelID]
	if !ok {
		return nil, errors.New("No channel")
	}

	return ch, nil
}

// Transactor ...
func (cs *MockChannelService) Transactor() (context.Transactor, error) {
	return cs.transactor, nil
}

// SetTransactor changes the return value of Transactor
func (cs *MockChannelService) SetTransactor(t context.Transactor) {
	cs.transactor = t
}

// Config ...
func (cs *MockChannelService) Config() (context.ChannelConfig, error) {
	return nil, nil
}

// Ledger ...
func (cs *MockChannelService) Ledger() (context.ChannelLedger, error) {
	return nil, nil
}
