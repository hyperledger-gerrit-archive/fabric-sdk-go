/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"fmt"
	"testing"
	"time"

	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// MockInfraProvider represents the default implementation of Fabric objects.
type MockInfraProvider struct {
	mocks.MockInfraProvider
	mockPeer *mocks.MockPeer
	lck      sync.RWMutex
}

// CreatePeerFromConfig returns a new default implementation of Peer based configuration
func (f *MockInfraProvider) CreatePeerFromConfig(peerCfg *fabApi.NetworkPeer) (fab.Peer, error) {
	if peerCfg.MSPID != mspID {
		panic(fmt.Sprintf("CreatePeerFromConfig create peer not same user msp(%v)", mspID))
	}
	return f.GetMockPeer(), nil
}

// GetMockPeer will return the mock infra provider mock peer in a thread safe way
func (f *MockInfraProvider) GetMockPeer() *mocks.MockPeer {
	f.lck.RLock()
	defer f.lck.RUnlock()
	return f.mockPeer
}

// SetMockPeer will write the mock infra provider mock peer in a thread safe way
func (f *MockInfraProvider) SetMockPeer(mPeer *mocks.MockPeer) {
	f.lck.Lock()
	defer f.lck.Unlock()
	f.mockPeer = mPeer
}

func TestDiscoveryService(t *testing.T) {
	mockInfraProv := &MockInfraProvider{
		MockInfraProvider: mocks.MockInfraProvider{
			Endpoints: []*mocks.MockDiscoverPeerEndpoint{
				&mocks.MockDiscoverPeerEndpoint{
					MSPID:        mspID,
					Endpoint:     "localhost:7051",
					LedgerHeight: 5,
				},
			},
		},
	}

	fabCtx := setupCustomTestContext(t, nil, nil, mockInfraProv)
	ctx := createChannelContext(fabCtx, ch)
	channel, err := ctx()
	if err != nil {
		t.Fatalf("get channel context failed %s", err)
	}

	membershipService := newService(
		options{
			refreshInterval: 5 * time.Second,
			responseTimeout: 2 * time.Second,
		},
	)
	defer membershipService.Close()

	err = membershipService.Initialize(channel)
	assert.NoError(t, err)

	peers, err := membershipService.GetPeers()
	if err != nil {
		t.Fatalf("Testing membership service's private getChannelTransactorClient function failed: %s", err)
	}

	if len(peers) != 1 {
		t.Fatalf("got list of peers(%v) from membership service not equal to 1", len(peers))
	}

	mockInfraProv.MockInfraProvider.SetPeerEndpoints(
		&mocks.MockDiscoverPeerEndpoint{
			MSPID:        mspID,
			Endpoint:     "localhost:7051",
			LedgerHeight: 5,
		},
		&mocks.MockDiscoverPeerEndpoint{
			MSPID:        mspID,
			Endpoint:     "localhost:8051",
			LedgerHeight: 5,
		},
	)

	peers, err = membershipService.GetPeers()
	if err != nil {
		t.Fatalf("Testing membership service's private getChannelTransactorClient function failed: %s", err)
	}
	//previous GetPeers() should get the list of peers from the cache, therefore len == 1 as set in the first SetMockPeer call above
	if len(peers) != 1 {
		t.Fatalf("got list of peers(%v) from membership service not equal to 1 ", len(peers))
	}

	// sleep to force cache to fetch new list of peers (which is now 5 as per the second SetMockPeer call above)
	time.Sleep(6 * time.Second)
	peers, err = membershipService.GetPeers()
	if err != nil {
		t.Fatalf("Testing membership service's private getChannelTransactorClient function failed: %s", err)
	}

	if len(peers) != 2 {
		t.Fatalf("expecting 2 peerEndponts, got %d", len(peers))
	}
}

func setupCustomTestContext(t *testing.T, selectionService fab.SelectionService, discoveryService fab.DiscoveryService, customInfraProvider fab.InfraProvider) context.ClientProvider {
	user := mspmocks.NewMockSigningIdentity("test", mspID)

	ctx := mocks.NewMockContext(user)

	ctx.SetCustomInfraProvider(customInfraProvider)

	testChannelSvc, err := setupTestChannelService(ctx)
	if err != nil {
		panic(err.Error())
	}
	//Modify for custom mocks to test scenarios
	selectionProvider := ctx.MockProviderContext.SelectionProvider()
	selectionProvider.(*mocks.MockSelectionProvider).SetCustomSelectionService(selectionService)

	channelProvider := ctx.MockProviderContext.ChannelProvider()
	channelProvider.(*mocks.MockChannelProvider).SetCustomChannelService(testChannelSvc)

	discoveryProvider := ctx.MockProviderContext.DiscoveryProvider()
	discoveryProvider.(*mocks.MockStaticDiscoveryProvider).SetCustomDiscoveryService(discoveryService)

	return createClientContext(ctx)
}

func createClientContext(client context.Client) context.ClientProvider {
	return func() (context.Client, error) {
		return client, nil
	}
}

func createChannelContext(clientContext context.ClientProvider, channelID string) context.ChannelProvider {
	channelProvider := func() (context.Channel, error) {
		return contextImpl.NewChannel(clientContext, channelID)
	}

	return channelProvider
}

func setupTestChannelService(ctx context.Client) (fab.ChannelService, error) {
	chProvider, err := mocks.NewMockChannelProvider(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "mock channel provider creation failed")
	}

	chService, err := chProvider.ChannelService(ctx, ch)
	if err != nil {
		return nil, errors.WithMessage(err, "mock channel service creation failed")
	}

	return chService, nil
}
