/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
)

// MockChannelSvc holds a mock channel service.
type MockChannelSvc struct {
	ctx      fab.ProviderContext
	channels map[string]fab.Channel
}

// NewMockChannelProvider returns a mock ChannelService
func NewMockChannelProvider(ctx fab.Context) (*MockChannelSvc, error) {
	channels := make(map[string]fab.Channel)

	// Create a mock client with the mock channel
	cp := MockChannelSvc{
		ctx,
		channels,
	}
	return &cp, nil
}

// EventHub ...
func (cs *MockChannelSvc) EventHub(channelID string) (fab.EventHub, error) {
	return NewMockEventHub(), nil
}

// SetChannel convenience method to set channel
func (cs *MockChannelSvc) SetChannel(id string, channel fab.Channel) {
	cs.channels[id] = channel
}

// Channel ...
func (cs *MockChannelSvc) Channel(channelID string) (fab.Channel, error) {
	ch, ok := cs.channels[channelID]
	if !ok {
		return nil, errors.New("No channel")
	}

	return ch, nil
	/*
		if channel != nil {
			return channel, nil
		}
		// Creating channel requires orderer information
		var orderers []apiconfig.OrdererConfig
		chCfg, err := cs.client.Config().ChannelConfig(channelID)
		if err != nil {
			return nil, err
		}

		if chCfg == nil {
			orderers, err = cs.client.Config().OrderersConfig()
		} else {
			orderers, err = cs.client.Config().ChannelOrderers(channelID)
		}

		// Check if retrieving orderer configuration went ok
		if err != nil {
			return nil, errors.WithMessage(err, "Failed to retrieve orderer configuration")
		}

		if len(orderers) == 0 {
			return nil, errors.Errorf("Must configure at least one order for channel and/or one orderer in the network")
		}

		channel, err = cs.client.NewChannel(channelID)
		if err != nil {
			return nil, errors.WithMessage(err, "NewChannel failed")
		}

		for _, ordererCfg := range orderers {
			orderer, err := orderer.New(cs.client.Config(), orderer.FromOrdererConfig(&ordererCfg))
			if err != nil {
				return nil, errors.WithMessage(err, "NewOrdererFromConfig failed")
			}
			err = channel.AddOrderer(orderer)
			if err != nil {
				return nil, errors.WithMessage(err, "adding orderer failed")
			}
		}
		return channel, nil*/
}
