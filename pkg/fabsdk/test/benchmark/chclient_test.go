/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package benchmark

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	mspctx "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	channelID = "testChannel"
)

func BenchmarkExecuteTx(b *testing.B) {
	// setup mocked CA
	f := testFixture{}
	sdk := f.setup()
	defer f.close()

	ctxProvider := sdk.Context()

	// Get the msp Client.
	// Without WithOrg option, it uses default client organization.
	mspClient, err := msp.New(ctxProvider)
	if err != nil {
		b.Fatalf("failed to create CA client: %s", err)
	}

	// get a new enrolled user
	err = mspClient.Enroll("someuser", msp.WithSecret("enrollmentSecret"))
	if err != nil {
		b.Fatalf("Enroll return error %s", err)
	}

	enrolledUser, err := mspClient.GetSigningIdentity("someuser")
	if err != nil {
		b.Fatal("Expected to find user")
	}

	// using enrolled user, let's start the benchmark
	for n := 0; n < b.N; n++ {
		chClient := setupChannelClientForBenchmarck(nil, b, enrolledUser)

		_, err := chClient.Execute(channel.Request{})
		if err == nil {
			b.Fatal("Should have failed for empty invoke request")
		}

		_, err = chClient.Execute(channel.Request{Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}})
		if err == nil {
			b.Fatal("Should have failed for empty chaincode ID")
		}

		_, err = chClient.Execute(channel.Request{ChaincodeID: "testCC", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}})
		if err == nil {
			b.Fatal("Should have failed for empty function")
		}

		// Test return different payload
		testPeer1 := fcmocks.NewMockPeer("Peer1", "http://peer1.com")
		testPeer1.Payload = []byte("test1")
		testPeer2 := fcmocks.NewMockPeer("Peer2", "http://peer2.com")
		testPeer2.Payload = []byte("test2")
		chClient = setupChannelClientForBenchmarck([]fab.Peer{testPeer1, testPeer2}, b, enrolledUser)
		_, err = chClient.Execute(channel.Request{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("b")}})
		if err == nil {
			b.Fatal("Should have failed")
		}
		s, ok := status.FromError(err)
		assert.True(b, ok, "expected status error")
		assert.EqualValues(b, status.EndorsementMismatch.ToInt32(), s.Code, "expected mismatch error")

	}
}

func setupCustomTestContextForBenchmark(b *testing.B, selectionService fab.SelectionService, discoveryService fab.DiscoveryService, orderers []fab.Orderer, signingIdentity mspctx.SigningIdentity) context.ClientProvider {
	user := signingIdentity
	ctx := fcmocks.NewMockContext(user)

	if orderers == nil {
		orderer := fcmocks.NewMockOrderer("", nil)
		orderers = []fab.Orderer{orderer}
	}

	transactor := txnmocks.MockTransactor{
		Ctx:       ctx,
		ChannelID: channelID,
		Orderers:  orderers,
	}

	testChannelSvc, err := setupTestChannelService(ctx, orderers)
	assert.Nil(b, err, "Got error %s", err)

	mockChService := testChannelSvc.(*fcmocks.MockChannelService)
	mockChService.SetTransactor(&transactor)
	mockChService.SetDiscovery(discoveryService)
	mockChService.SetSelection(selectionService)

	channelProvider := ctx.MockProviderContext.ChannelProvider()
	channelProvider.(*fcmocks.MockChannelProvider).SetCustomChannelService(testChannelSvc)

	return createClientContext(ctx)
}

func setupChannelClientForBenchmarck(peers []fab.Peer, b *testing.B, signingIdentity mspctx.SigningIdentity) *channel.Client {

	return setupChannelClientWithErrorForBenchmarck(nil, nil, peers, b, signingIdentity)
}

func setupChannelClientWithErrorForBenchmarck(discErr error, selectionErr error, peers []fab.Peer, b *testing.B, signingIdentity mspctx.SigningIdentity) *channel.Client {

	fabCtx := setupCustomTestContextForBenchmark(b, txnmocks.NewMockSelectionService(selectionErr, peers...), txnmocks.NewMockDiscoveryService(discErr), nil, signingIdentity)

	ctx := createChannelContext(fabCtx, channelID)

	ch, err := channel.New(ctx)
	if err != nil {
		b.Fatalf("Failed to create new channel client: %s", err)
	}

	return ch
}

func setupTestChannelService(ctx context.Client, orderers []fab.Orderer) (fab.ChannelService, error) {
	chProvider, err := fcmocks.NewMockChannelProvider(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "mock channel provider creation failed")
	}

	chService, err := chProvider.ChannelService(ctx, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "mock channel service creation failed")
	}

	return chService, nil
}

func createChannelContext(clientContext context.ClientProvider, channelID string) context.ChannelProvider {

	channelProvider := func() (context.Channel, error) {
		return contextImpl.NewChannel(clientContext, channelID)
	}

	return channelProvider
}

func createClientContext(client context.Client) context.ClientProvider {
	return func() (context.Client, error) {
		return client, nil
	}
}
