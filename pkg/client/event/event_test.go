/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package event

import (
	"testing"

	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	channelID = "testChannel"
)

func TestNewEventClient(t *testing.T) {

	fabCtx := setupCustomTestContext(t, nil)
	ctx := createChannelContext(fabCtx, channelID)

	_, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create new event client: %s", err)
	}

	_, err = New(ctx, WithBlockEvents())
	if err != nil {
		t.Fatalf("Failed to create new event client: %s", err)
	}

	ctxErr := createChannelContextWithError(fabCtx, channelID)
	_, err = New(ctxErr)
	if err == nil {
		t.Fatalf("Should have failed with 'Test Error'")
	}
}

func setupCustomTestContext(t *testing.T, orderers []fab.Orderer) context.ClientProvider {
	user := mspmocks.NewMockSigningIdentity("test", "test")
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

	ctx.InfraProvider().(*fcmocks.MockInfraProvider).SetCustomTransactor(&transactor)

	testChannelSvc, err := setupTestChannelService(ctx, orderers)
	assert.Nil(t, err, "Got error %s", err)

	channelProvider := ctx.MockProviderContext.ChannelProvider()
	channelProvider.(*fcmocks.MockChannelProvider).SetCustomChannelService(testChannelSvc)

	return createClientContext(ctx)
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

func createChannelContextWithError(clientContext context.ClientProvider, channelID string) context.ChannelProvider {

	channelProvider := func() (context.Channel, error) {
		return nil, errors.New("Test Error")
	}

	return channelProvider
}

func createClientContext(client context.Client) context.ClientProvider {
	return func() (context.Client, error) {
		return client, nil
	}
}
