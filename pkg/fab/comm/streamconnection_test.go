/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"context"
	"testing"
	"time"

	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/keepalive"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var testStream = func(grpcconn *grpc.ClientConn) (grpc.ClientStream, error) {
	return pb.NewDeliverClient(grpcconn).Deliver(context.Background())
}

var invalidStream = func(grpcconn *grpc.ClientConn) (grpc.ClientStream, error) {
	return nil, errors.New("simulated error creating stream")
}

func TestStreamConnectionEmptyURL(t *testing.T) {
	const channelID = "testchannel"

	context := newMockContext()
	chConfig := fabmocks.NewMockChannelCfg(channelID)

	_, err := NewStreamConnection(context, chConfig, testStream, "")
	assert.Error(t, err, "expected error creating new connection with empty URL")
}

func TestStreamConnectionInvalidURL(t *testing.T) {
	const channelID = "testchannel"

	context := newMockContext()
	chConfig := fabmocks.NewMockChannelCfg(channelID)

	_, err := NewStreamConnection(context, chConfig, testStream, "invalidhost:0000",
		WithFailFast(true),
		WithCertificate(nil),
		WithInsecure(),
		WithHostOverride(""),
		WithKeepAliveParams(keepalive.ClientParameters{}),
		WithConnectTimeout(3*time.Second),
	)
	assert.Error(t, err, "expected error creating new connection with invalid URL")

	_, err = NewStreamConnection(context, chConfig, invalidStream, peerURL)
	assert.Error(t, err, "expected error creating new connection with invalid stream but got none")
}

func TestStreamConnection(t *testing.T) {
	const channelID = "testchannel"

	context := newMockContext()
	chConfig := fabmocks.NewMockChannelCfg(channelID)

	conn, err := NewStreamConnection(context, chConfig, testStream, peerURL)
	assert.NoError(t, err, "error creating new connection")
	assert.False(t, conn.Closed(), "expected connection to be open")
	assert.NotNil(t, conn.Stream(), "got invalid stream")

	conn.Close()
	assert.True(t, conn.Closed(), "expected connection to be closed")

	// Calling close again should be ignored
	conn.Close()
}
