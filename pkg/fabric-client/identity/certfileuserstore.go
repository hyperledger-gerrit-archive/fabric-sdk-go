/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/config/cryptoutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"

	"github.com/pkg/errors"
)

// CertFileUserStore stores each user in a separate file.
// Only user's enrollment cert is stored, in pem format.
// File naming is <user>@<org>-cert.pem
type CertFileUserStore struct {
	store       *keyvaluestore.FileKeyValueStore
	cryptoSuite apicryptosuite.CryptoSuite
}

func userKeyFromUser(user context.User) context.UserKey {
	return context.UserKey{
		MspID: user.MspID(),
		Name:  user.Name(),
	}
}

func storeKeyFromUserKey(key context.UserKey) string {
	return key.Name + "@" + key.MspID + "-cert.pem"
}

// NewCertFileUserStore creates a new instance of CertFileUserStore
func NewCertFileUserStore(path string, cryptoSuite apicryptosuite.CryptoSuite) (*CertFileUserStore, error) {
	if path == "" {
		return nil, errors.New("path is empty")
	}
	if cryptoSuite == nil {
		return nil, errors.New("cryptoSuite is nil")
	}
	store, err := keyvaluestore.NewFileKeyValueStore(&keyvaluestore.FileKeyValueStoreOptions{
		Path: path,
	})
	if err != nil {
		return nil, errors.Wrap(err, "user store creation failed")
	}
	return &CertFileUserStore{
		store:       store,
		cryptoSuite: cryptoSuite,
	}, nil
}

// Load returns the User stored in the store for a key.
func (s *CertFileUserStore) Load(key context.UserKey) (context.User, error) {
	cert, err := s.store.Load(storeKeyFromUserKey(key))
	if err != nil {
		if err == context.ErrNotFound {
			return nil, context.ErrUserNotFound
		}
		return nil, err
	}
	certBytes, ok := cert.([]byte)
	if !ok {
		return nil, errors.New("user is not of proper type")
	}
	pubKey, err := cryptoutil.GetPublicKeyFromCert(certBytes, s.cryptoSuite)
	if err != nil {
		return nil, errors.WithMessage(err, "fetching public key from cert failed")
	}
	pk, err := s.cryptoSuite.GetKey(pubKey.SKI())
	if err != nil {
		return nil, errors.Wrap(err, "cryptoSuite GetKey failed")
	}
	u := &User{
		mspID: key.MspID,
		name:  key.Name,
		enrollmentCertificate: certBytes,
		privateKey:            pk,
	}
	return u, nil
}

// Store stores a User into store
func (s *CertFileUserStore) Store(user context.User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	key := storeKeyFromUserKey(userKeyFromUser(user))
	return s.store.Store(key, user.EnrollmentCertificate())
}

// Delete deletes a User from store
func (s *CertFileUserStore) Delete(user context.User) error {
	return s.store.Delete(storeKeyFromUserKey(userKeyFromUser(user)))
}
