/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discover

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const (
	peerAddress  = "localhost:9999"
	peerURL      = "grpc://" + peerAddress
	peer2Address = "localhost:9998"
)

func TestDiscoverClient(t *testing.T) {
	channelID := "mychannel"
	clientCtx := newMockContext()
	chConfig := mocks.NewMockChannelCfg(channelID)

	client, err := New(clientCtx, chConfig, peerURL, WithIdleTimeout(5*time.Second))
	defer client.Close()

	req := discclient.NewRequest().OfChannel(channelID).AddPeersQuery()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := client.Send(ctx, req)
	cancel()

	assert.NoError(t, err)

	chResp := resp.ForChannel(channelID)
	peers, err := chResp.Peers()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(peers))
}

var discoverServer *mocks.MockDiscoverServer

func TestMain(m *testing.M) {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	lis, err := net.Listen("tcp", peerAddress)
	if err != nil {
		panic(fmt.Sprintf("Error starting events listener %s", err))
	}

	discoverServer = mocks.NewMockDiscoverServer(mocks.WithDiscoverServerPeers(
		&mocks.MockDiscoverPeerEndpoint{
			MSPID:        "Org1MSP",
			Endpoint:     peerAddress,
			LedgerHeight: 26,
		},
		&mocks.MockDiscoverPeerEndpoint{
			MSPID:        "Org2MSP",
			Endpoint:     peer2Address,
			LedgerHeight: 25,
		},
	))

	discovery.RegisterDiscoveryServer(grpcServer, discoverServer)

	go grpcServer.Serve(lis)

	time.Sleep(2 * time.Second)
	os.Exit(m.Run())
}

func newMockContext() *mocks.MockContext {
	context := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("user1", "test"))
	context.SetCustomInfraProvider(comm.NewMockInfraProvider())
	return context
}
