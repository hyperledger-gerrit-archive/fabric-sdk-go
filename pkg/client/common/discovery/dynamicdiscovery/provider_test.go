/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/stretchr/testify/assert"
)

const (
	ch    = "orgchannel"
	mspID = "Org1MSP"
)

func TestDiscoveryProvider(t *testing.T) {
	mockInfraProv := &MockInfraProvider{
		MockInfraProvider: mocks.MockInfraProvider{
			Endpoints: []*mocks.MockDiscoverPeerEndpoint{
				&mocks.MockDiscoverPeerEndpoint{
					MSPID:        mspID,
					Endpoint:     "localhost:7051",
					LedgerHeight: 5,
				},
				&mocks.MockDiscoverPeerEndpoint{
					MSPID:        mspID,
					Endpoint:     "localhost:8051",
					LedgerHeight: 4,
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

	p := New(WithRefreshInterval(0))
	defer p.Close()

	ds, err := p.CreateDiscoveryService(ch)
	if err != nil {
		t.Fatalf("Membership discovery provider returned an error when creating a new service. Error: %s", err)
	}

	err = ds.(*Service).Initialize(channel)
	assert.NoError(t, err)
}
