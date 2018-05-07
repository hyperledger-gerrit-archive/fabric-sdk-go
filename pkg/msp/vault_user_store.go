/*
Copyright Hyperledger and its contributors.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"encoding/hex"

	vault "github.com/hashicorp/vault/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/pkg/errors"
)

// VaultUserStore is a vault implementation of UserStore
type VaultUserStore struct {
	client *vault.Client
}

// NewVaultUserStore creates a new VaultUserStore instance
func NewVaultUserStore(address, token string) (*VaultUserStore, error) {
	vaultConfig := &vault.Config{
		Address: address,
	}

	client, err := vault.NewClient(vaultConfig)

	if err != nil {
		return nil, errors.Wrapf(err, "could not initialize Vault BCCSP for address: %s", vaultConfig.Address)
	}

	client.SetToken(token)

	return &VaultUserStore{client: client}, nil
}

// Store stores a user into store
func (s *VaultUserStore) Store(user *msp.UserData) error {
	_, err := s.client.Logical().Write(
		"kv/"+user.ID+"@"+user.MSPID,

		map[string]interface{}{
			"value": hex.EncodeToString(user.EnrollmentCertificate),
		})

	if err != nil {
		return err
	}

	return nil
}

// Load loads a user from store
func (s *VaultUserStore) Load(id msp.IdentityIdentifier) (*msp.UserData, error) {
	secret, err := s.client.Logical().Read("kv/" + id.ID + "@" + id.MSPID)

	if err != nil {
		return nil, err
	}

	certString, ok := secret.Data["value"].(string)

	if !ok {
		return nil, msp.ErrUserNotFound
	}

	certBytes, err := hex.DecodeString(certString)

	if err != nil {
		return nil, err
	}

	userData := msp.UserData{
		ID:    id.ID,
		MSPID: id.MSPID,
		EnrollmentCertificate: certBytes,
	}

	return &userData, nil
}
