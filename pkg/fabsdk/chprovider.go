/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apicore"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// channelProvider keeps context across ChannelService instances.
//
// TODO: add cache for dynamic channel configuration. This cache is updated
// by channel services, as only channel service have an identity context.
// TODO: add listener for channel config changes. Upon channel config change,
// underlying channel services need to recreate their channel clients.
type channelProvider struct {
	sdk *sdkContext
}

func newChannelProvider(sdk *sdkContext) (*channelProvider, error) {
	cp := channelProvider{sdk}
	return &cp, nil
}

func (cp *channelProvider) newChannelService(ic apifabclient.IdentityContext) apifabclient.ChannelService {
	cs := channelService{
		fabricProvider:  cp.sdk.FabricProvider(),
		identityContext: ic,
	}
	return &cs
}

// channelService provides Channel clients and maintains contexts for them.
// the identity context is used
//
// TODO: add cache for channel rather than reconstructing each time.
type channelService struct {
	fabricProvider  apicore.FabricProvider
	identityContext apifabclient.IdentityContext
}

// Channel returns the named Channel client.
func (cp *channelService) Channel(channelID string) (apifabclient.Channel, error) {
	return cp.fabricProvider.NewChannelClient(cp.identityContext, channelID)
}
