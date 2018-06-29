/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"testing"

	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/assert"
)

func TestConnectionEmptyURL(t *testing.T) {
	context := newMockContext()

	_, err := NewConnection(context, "")
	assert.Error(t, err, "expected error creating new connection with empty URL")
}

func TestConnection(t *testing.T) {
	context := newMockContext()

	conn, err := NewConnection(context, peerURL)
	assert.NoError(t, err, "error creating new connection")
	assert.False(t, conn.Closed(), "expected connection to be open")

	conn.Close()
	assert.True(t, conn.Closed(), "expected connection to be closed")

	// Calling close again should be ignored
	conn.Close()
}

func newMockContext() *fabmocks.MockContext {
	context := fabmocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", "test"))
	context.SetCustomInfraProvider(NewMockInfraProvider())
	return context
}
