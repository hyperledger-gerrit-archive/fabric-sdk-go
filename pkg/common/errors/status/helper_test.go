/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientCodeCheckers(t *testing.T) {

	s := WithStack(ClientStatus, OK.ToInt32(), "test", nil)
	assert.True(t, IsOK(s))

	s = WithStack(ClientStatus, Unknown.ToInt32(), "test", nil)
	assert.True(t, IsUnknown(s))

	s = WithStack(ClientStatus, ConnectionFailed.ToInt32(), "test", nil)
	assert.True(t, IsConnectionFailed(s))

	s = WithStack(ClientStatus, EndorsementMismatch.ToInt32(), "test", nil)
	assert.True(t, IsEndorsementMismatch(s))

	s = WithStack(ClientStatus, EmptyCert.ToInt32(), "test", nil)
	assert.True(t, IsEmptyCert(s))

	s = WithStack(ClientStatus, Timeout.ToInt32(), "test", nil)
	assert.True(t, IsTimeout(s))

	s = WithStack(ClientStatus, NoPeersFound.ToInt32(), "test", nil)
	assert.True(t, IsNoPeersFound(s))

	s = WithStack(ClientStatus, MultipleErrors.ToInt32(), "test", nil)
	assert.True(t, IsMultipleErrors(s))

	s = WithStack(ClientStatus, SignatureVerificationFailed.ToInt32(), "test", nil)
	assert.True(t, IsSignatureVerificationFailed(s))

	s = WithStack(ClientStatus, MissingEndorsement.ToInt32(), "test", nil)
	assert.True(t, IsMissingEndorsement(s))

	s = WithStack(ClientStatus, ChaincodeError.ToInt32(), "test", nil)
	assert.True(t, IsChaincodeError(s))

	s = WithStack(ClientStatus, NoMatchingCertificateAuthorityEntity.ToInt32(), "test", nil)
	assert.True(t, IsNoMatchingCertificateAuthorityEntity(s))

	s = WithStack(ClientStatus, NoMatchingPeerEntity.ToInt32(), "test", nil)
	assert.True(t, IsNoMatchingPeerEntity(s))

	s = WithStack(ClientStatus, NoMatchingOrdererEntity.ToInt32(), "test", nil)
	assert.True(t, IsNoMatchingOrdererEntity(s))

}
