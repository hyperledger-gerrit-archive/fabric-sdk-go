/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

// MockMemberID mock member id
type MockMemberID struct {
	ValidateErr error
	VerifyErr   error
}

// NewMockMemberID new mock member id
func NewMockMemberID() *MockMemberID {
	return &MockMemberID{}
}

// Validate if the given ID was issued by the channel's members
func (m *MockMemberID) Validate(serializedID []byte) error {
	return m.ValidateErr
}

// Verify the given signature
func (m *MockMemberID) Verify(serializedID []byte, msg []byte, sig []byte) error {
	return m.VerifyErr
}
