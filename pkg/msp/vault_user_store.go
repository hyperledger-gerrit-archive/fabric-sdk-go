/*
Copyright Hyperledger and its contributors.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
)

// VaultUserStore is a vault implementation of UserStore
type VaultUserStore struct {
	store map[string][]byte
}

// NewVaultUserStore creates a new VaultUserStore instance
func NewVaultUserStore() *VaultUserStore {
	store := make(map[string][]byte)
	return &VaultUserStore{store: store}
}

// Store stores a user into store
func (s *VaultUserStore) Store(user *msp.UserData) error {
	s.store[user.ID+"@"+user.MSPID] = user.EnrollmentCertificate
	return nil
}

// Load loads a user from store
func (s *VaultUserStore) Load(id msp.IdentityIdentifier) (*msp.UserData, error) {
	cert, ok := s.store[id.ID+"@"+id.MSPID]
	if !ok {
		return nil, msp.ErrUserNotFound
	}
	userData := msp.UserData{
		ID:    id.ID,
		MSPID: id.MSPID,
		EnrollmentCertificate: cert,
	}
	return &userData, nil
}
