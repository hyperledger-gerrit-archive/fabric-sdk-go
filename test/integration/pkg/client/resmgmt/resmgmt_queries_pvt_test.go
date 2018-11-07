/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	"reflect"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/stretchr/testify/require"
)

func TestQueryCollectionsConfig(t *testing.T) {
	sdk := mainSDK

	orgsContext := setupMultiOrgContext(t, sdk)
	err := integration.EnsureChannelCreatedAndPeersJoined(t, sdk, orgChannelID, "orgchannel.tx", orgsContext)
	require.NoError(t, err)

	coll1 := "collection1"
	ccID := integration.GenerateExamplePvtID(true)
	collConfig, err := newCollectionConfig(coll1, "OR('Org1MSP.member','Org2MSP.member')", 0, 2, 1000)
	require.NoError(t, err)

	err = integration.InstallExamplePvtChaincode(orgsContext, ccID)
	require.NoError(t, err)
	err = integration.InstantiateExamplePvtChaincode(orgsContext, orgChannelID, ccID, "OR('Org1MSP.member','Org2MSP.member')", collConfig)
	require.NoError(t, err)

	org1AdminClientContext := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))
	client, err := resmgmt.New(org1AdminClientContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	resp, err := client.QueryCollectionsConfig(orgChannelID, ccID)
	if err != nil {
		t.Fatalf("QueryCollectionsConfig return error: %s", err)
	}
	if len(resp.Config) != 1 {
		t.Fatalf("The number of collection config is incorrect, expected 1, got %d", len(resp.Config))
	}

	conf := resp.Config[0]
	switch cconf := conf.Payload.(type) {
	case *cb.CollectionConfig_StaticCollectionConfig:
		if cconf.StaticCollectionConfig.Name != "collection1" {
			t.Fatalf("CollectionConfig'name is incorrect, expected collection1, got %s", cconf.StaticCollectionConfig.Name)
		}
	default:
		t.Fatalf("The CollectionConfig.Payload's type is incorrect, expected `CollectionConfig_StaticCollectionConfig`, got %+v", reflect.TypeOf(conf.Payload))
	}
}

func newCollectionConfig(colName, policy string, reqPeerCount, maxPeerCount int32, blockToLive uint64) (*cb.CollectionConfig, error) {
	p, err := cauthdsl.FromString(policy)
	if err != nil {
		return nil, err
	}
	cpc := &cb.CollectionPolicyConfig{
		Payload: &cb.CollectionPolicyConfig_SignaturePolicy{
			SignaturePolicy: p,
		},
	}
	return &cb.CollectionConfig{
		Payload: &cb.CollectionConfig_StaticCollectionConfig{
			StaticCollectionConfig: &cb.StaticCollectionConfig{
				Name:              colName,
				MemberOrgsPolicy:  cpc,
				RequiredPeerCount: reqPeerCount,
				MaximumPeerCount:  maxPeerCount,
				BlockToLive:       blockToLive,
			},
		},
	}, nil
}
