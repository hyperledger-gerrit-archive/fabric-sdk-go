/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeCheckers(t *testing.T) {

	s := New(ClientStatus, OK.ToInt32(), "test", nil)
	assert.True(t, IsOK(s))

	s := New(ClientStatus, Unknown.ToInt32(), "test", nil)
	assert.True(t, IsUnknown(s))

	s := New(ClientStatus, ConnectionFailed.ToInt32(), "test", nil)
	assert.True(t, IsConnectionFailed(s))

	s := New(ClientStatus, EndorsementMismatch.ToInt32(), "test", nil)
	assert.True(t, IsEndorsementMismatch(s))

	s := New(ClientStatus, EmptyCert.ToInt32(), "test", nil)
	assert.True(t, IsEmptyCert(s))

	s := New(ClientStatus, Timeout.ToInt32(), "test", nil)
	assert.True(t, IsTimeout(s))

	s := New(ClientStatus, NoPeersFound.ToInt32(), "test", nil)
	assert.True(t, IsNoPeersFound(s))

	s := New(ClientStatus, MultipleErrors.ToInt32(), "test", nil)
	assert.True(t, IsMultipleErrors(s))

}
