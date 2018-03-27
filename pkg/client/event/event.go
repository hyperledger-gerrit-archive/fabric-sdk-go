/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package event enables access to a channel events on a Fabric network.
package event

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/client")

// Client enables access to a channel events on a Fabric network.
type Client struct {
	fab.EventService
	permitBlockEvents bool
}

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*Client) error

// WithBlockEvents option
func WithBlockEvents() ClientOption {
	return func(c *Client) error {
		c.permitBlockEvents = true
		return nil
	}
}

// New returns a Client instance.
func New(channelProvider context.ChannelProvider, opts ...ClientOption) (*Client, error) {

	channelContext, err := channelProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create channel context")
	}

	eventClient := Client{}

	for _, param := range opts {
		err := param(&eventClient)
		if err != nil {
			return nil, errors.WithMessage(err, "option failed")
		}
	}

	if channelContext.ChannelService() == nil {
		return nil, errors.New("channel service not initialized")
	}

	var es fab.EventService
	if eventClient.permitBlockEvents {
		es, err = channelContext.ChannelService().EventService(client.WithBlockEvents())
	} else {
		es, err = channelContext.ChannelService().EventService()
	}

	if err != nil {
		return nil, errors.WithMessage(err, "event service creation failed")
	}

	eventClient.EventService = es

	return &eventClient, nil
}
