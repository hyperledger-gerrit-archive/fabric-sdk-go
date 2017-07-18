/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package org

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fabca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
)

// Context currently represents an organization that the app is dealing with.
// TODO: better decription (e.g., possibility of holding discovery resources for the org & peers).
type Context struct {
	mspClient fabca.FabricCAClient
}

// NewContext creates a context based on the providers in the SDK
func NewContext(factory ProviderFactory, orgID string, config apiconfig.Config) (*Context, error) {
	c := Context{}

	// Initialize MSP client
	client, err := factory.NewMSPClient(orgID, config)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize MSP client [%s]", err)
	}
	c.mspClient = client

	return &c, nil
}

// MSPClient provides the MSP client of the context.
func (c *Context) MSPClient() fabca.FabricCAClient {
	return c.mspClient
}

// ProviderFactory allows overriding default clients and providers of an SDK context
// Currently, a context is created for each organization that the client app needs.
type ProviderFactory interface {
	NewMSPClient(orgName string, config apiconfig.Config) (fabca.FabricCAClient, error)
}

// DefaultContextFactory represents the default org provider factory.
type DefaultContextFactory struct{}

// NewDefaultContextFactory returns the default org provider factory.
func NewDefaultContextFactory() *DefaultContextFactory {
	f := DefaultContextFactory{}
	return &f
}

// NewMSPClient returns a new default implmentation of the MSP client
func (f *DefaultContextFactory) NewMSPClient(orgName string, config apiconfig.Config) (fabca.FabricCAClient, error) {
	mspClient, err := fabricCAClient.NewFabricCAClient(config, orgName)
	if err != nil {
		return nil, fmt.Errorf("NewFabricCAClient returned error: %v", err)
	}

	return mspClient, nil
}
