/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
)

// Client represents the fabric transaction clients
type Client struct {
	opts          *clientOptions
	session       *Session
	providers     apisdk.SDK
	clientFactory apisdk.SessionClientFactory
}

// ClientOption configures the clients created by the SDK.
type ClientOption func(opts *clientOptions) error

type clientOptions struct {
	orgName        string
	configProvider apiconfig.Config
	targetFilter   resmgmt.TargetFilter
}

// WithOrg uses the configuration and users from the named organization.
func WithOrg(name string) ClientOption {
	return func(opts *clientOptions) error {
		opts.orgName = name
		return nil
	}
}

// WithTargetFilter allows for filtering target peers.
func WithTargetFilter(targetFilter resmgmt.TargetFilter) ClientOption {
	return func(opts *clientOptions) error {
		opts.targetFilter = targetFilter
		return nil
	}
}

// withConfig allows for overriding the configuration of the client.
// TODO: What's the use case for this? Should be using the SDK's configuration.
func withConfig(configProvider apiconfig.Config) ClientOption {
	return func(opts *clientOptions) error {
		opts.configProvider = configProvider
		return nil
	}
}

// ClientFromUsername ...
func (sdk *FabricSDK) ClientFromUsername(userName string, opts ...ClientOption) (*Client, error) {
	o, err := newClientOptions(sdk.ConfigProvider(), opts)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to retrieve configuration from SDK")
	}

	session, err := sdk.NewPreEnrolledUserSession(o.orgName, userName)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get pre-enrolled user session")
	}

	client := Client{
		opts:          o,
		session:       session,
		providers:     sdk,
		clientFactory: sdk.opts.Session,
	}
	return &client, nil
}

func newClientOptions(config apiconfig.Config, in []ClientOption) (*clientOptions, error) {
	// Read default org name from configuration
	client, err := config.Client()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to retrieve client from network config")
	}

	opts := clientOptions{
		orgName:        client.Organization,
		configProvider: config,
	}

	for _, option := range in {
		err := option(&opts)
		if err != nil {
			return nil, errors.WithMessage(err, "Error in option passed to client")
		}
	}

	if opts.orgName == "" {
		return nil, errors.New("must provide default organisation name in configuration")
	}

	return &opts, nil
}

// ChannelMgmt returns a client API for managing channels
func (c *Client) ChannelMgmt() (chmgmt.ChannelMgmtClient, error) {
	client, err := c.clientFactory.NewChannelMgmtClient(c.providers, c.session, c.opts.configProvider)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create new channel management client")
	}

	return client, nil
}

// ResourceMgmt returns a client API for managing system resources
func (c *Client) ResourceMgmt() (resmgmt.ResourceMgmtClient, error) {
	client, err := c.clientFactory.NewResourceMgmtClient(c.providers, c.session, c.opts.configProvider, c.opts.targetFilter)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to created new resource management client")
	}

	return client, nil
}

// Channel returns a client API for transacting on a channel
func (c *Client) Channel(id string) (apitxn.ChannelClient, error) {
	client, err := c.clientFactory.NewChannelClient(c.providers, c.session, c.opts.configProvider, id)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to created new resource management client")
	}

	return client, nil
}
