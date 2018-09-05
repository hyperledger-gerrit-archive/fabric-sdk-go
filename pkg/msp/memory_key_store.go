/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"encoding/hex"

	"fmt"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/sw"
)

// MemoryKeyStore is in-memory implementation of BCCSP key store
type MemoryKeyStore struct {
	store    map[string][]byte
	password []byte
}

// NewMemoryKeyStore creates a new MemoryKeyStore instance
func NewMemoryKeyStore(password []byte) *MemoryKeyStore {
	store := make(map[string][]byte)
	return &MemoryKeyStore{store: store, password: password}
}

// ReadOnly returns always false
func (s *MemoryKeyStore) ReadOnly() bool {
	return false
}

// GetKey returns a key for the provided SKI
func (s *MemoryKeyStore) GetKey(ski []byte) (bccsp.Key, error) {
	b, ok := s.store[hex.EncodeToString(ski)]
	if !ok {
		return nil, fmt.Errorf("Key not found [%s]", ski)
	}
	key, err := sw.UnmarshalPrivateKey(b, s.password)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling key [%s]: [%s]", ski, err)
	}
	return key, nil
}

// StoreKey stores a key
func (s *MemoryKeyStore) StoreKey(key bccsp.Key) error {
	ski := hex.EncodeToString(key.SKI())
	b, err := sw.MarshalKey(key, s.password)
	if err != nil {
		return fmt.Errorf("Error marshalling key [%s]: [%s]", ski, err)
	}
	s.store[ski] = b
	return nil
}
