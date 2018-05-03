/*
Copyright Hyperledger and its contributors.

SPDX-License-Identifier: Apache-2.0
*/

package internal

import (
	vault "github.com/hashicorp/vault/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
)

// SecretWrapper is a wrapper around Vault secrets, used for parsing various items of interest
type SecretWrapper struct {
	secret *vault.Secret
}

// NewSecretWrapper constructs a new SecretWrapper out of a vault secret
func NewSecretWrapper(secret *vault.Secret) *SecretWrapper {
	return &SecretWrapper{
		secret: secret,
	}
}

// ParseKey parses a core.Key out of a vault secret's Data field
func (sw *SecretWrapper) ParseKey() (core.Key, error) {
	pubBytes := sw.PublicKey()
	return ParseKey(pubBytes)
}

// PublicKey parses a public key out of a vault secret's Data field
func (sw *SecretWrapper) PublicKey() []byte {
	keys, _ := sw.secret.Data["keys"].(map[string]interface{})
	onlyKey, _ := keys["1"].(map[string]interface{})
	pub, _ := onlyKey["public_key"].(string)

	return []byte(pub)
}

// KeyID parses a key's ID out of a vault secret's Data field
func (sw *SecretWrapper) KeyID() string {
	keyID, _ := sw.secret.Data["name"].(string)
	return keyID
}

// ParseValue parses the value out of a vault secret's Data field
func (sw *SecretWrapper) ParseValue() string {
	value, _ := sw.secret.Data["value"].(string)
	return value
}

// ParseSignature parses the signature out of a vault secret's Data field
func (sw *SecretWrapper) ParseSignature() []byte {
	signature, _ := sw.secret.Data["signature"].(string)
	return []byte(signature)
}

// ParseVerification parses the verification out of a vault secret's Data field
func (sw *SecretWrapper) ParseVerification() bool {
	valid, _ := sw.secret.Data["valid"].(bool)
	return valid
}
