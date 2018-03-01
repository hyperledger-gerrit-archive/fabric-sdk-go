/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package persistence

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"

	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/pkg/errors"
)

// CertFileUserStore stores each user in a separate file.
// Only user's enrollment cert is stored, in pem format.
// File naming is <user>@<org>-cert.pem
type CertFileUserStore struct {
	store *keyvaluestore.FileKeyValueStore
}

func userIdentifierFromUser(user contextApi.UserData) contextApi.UserIdentifier {
	return contextApi.UserIdentifier{
		MspID: user.MspID,
		Name:  user.Name,
	}
}

func storeKeyFromUserIdentifier(key contextApi.UserIdentifier) string {
	return key.Name + "@" + key.MspID + "-cert.pem"
}

// NewCertFileUserStore creates a new instance of CertFileUserStore
func NewCertFileUserStore(path string) (*CertFileUserStore, error) {
	if path == "" {
		return nil, errors.New("path is empty")
	}
	store, err := keyvaluestore.New(&keyvaluestore.FileKeyValueStoreOptions{
		Path: path,
	})
	if err != nil {
		return nil, errors.Wrap(err, "user store creation failed")
	}
	return &CertFileUserStore{
		store: store,
	}, nil
}

// Load returns the User stored in the store for a key.
func (s *CertFileUserStore) Load(key contextApi.UserIdentifier) (contextApi.UserData, error) {
	var userData contextApi.UserData
	cert, err := s.store.Load(storeKeyFromUserIdentifier(key))
	if err != nil {
		if err == contextApi.ErrNotFound {
			return userData, contextApi.ErrUserNotFound
		}
		return userData, err
	}
	certBytes, ok := cert.([]byte)
	if !ok {
		return userData, errors.New("user is not of proper type")
	}
	userData = contextApi.UserData{
		MspID: key.MspID,
		Name:  key.Name,
		EnrollmentCertificate: certBytes,
	}
	return userData, nil
}

// Store stores a User into store
func (s *CertFileUserStore) Store(user contextApi.UserData) error {
	key := storeKeyFromUserIdentifier(contextApi.UserIdentifier{MspID: user.MspID, Name: user.Name})
	return s.store.Store(key, user.EnrollmentCertificate)
}

// Delete deletes a User from store
func (s *CertFileUserStore) Delete(key contextApi.UserIdentifier) error {
	return s.store.Delete(storeKeyFromUserIdentifier(key))
}
