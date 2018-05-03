/*
Copyright Hyperledger and its contributors.

SPDX-License-Identifier: Apache-2.0
*/

package vault

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"

	"crypto/ecdsa"
	"crypto/rsa"

	"crypto/elliptic"
	"crypto/sha256"

	vault "github.com/hashicorp/vault/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/pkg/errors"
)

type secretWrapper struct {
	secret *vault.Secret
}

func newSecretWrapper(secret *vault.Secret) *secretWrapper {
	return &secretWrapper{
		secret: secret,
	}
}

func (sw *secretWrapper) ski(pub interface{}) ([]byte, error) {
	switch tpub := pub.(type) {
	case *rsa.PublicKey:
		// Marshall the public key
		raw, err := asn1.Marshal(*tpub)

		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal key")
		}

		// Hash it
		hash := sha256.New()
		_, err = hash.Write(raw)

		if err != nil {
			return nil, errors.Wrap(err, "failed to calculate hash")
		}

		return hash.Sum(nil), nil

	case *ecdsa.PublicKey:
		// Marshall the public key
		raw := elliptic.Marshal(tpub.Curve, tpub.X, tpub.Y)

		// Hash it
		hash := sha256.New()
		_, err := hash.Write(raw)

		if err != nil {
			return nil, err
		}

		return hash.Sum(nil), nil

	default:
		return nil, errors.New("unsupported key type")
	}
}

func (sw *secretWrapper) parseKey() (core.Key, error) {
	pubBytes := sw.publicKey()

	block, _ := pem.Decode(pubBytes)

	if block == nil {
		return nil, errors.Errorf("failed to decode the bytes of the vault secret")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return nil, errors.Errorf("failed to parse the core.Key from the decoded vault secret")
	}

	ski, err := sw.ski(pub)

	if err != nil {
		return nil, err
	}

	return &privateKey{
		ski: ski,
		pub: publicKey{
			ski: ski,
			pub: pub,
		},
	}, nil
}

func (sw *secretWrapper) publicKey() []byte {
	keys, _ := sw.secret.Data["keys"].(map[string]interface{})
	onlyKey, _ := keys["1"].(map[string]interface{})
	pub, _ := onlyKey["public_key"].(string)

	return []byte(pub)
}

func (sw *secretWrapper) keyID() string {
	keyID, _ := sw.secret.Data["name"].(string)
	return keyID
}

func (sw *secretWrapper) parseValue() string {
	value, _ := sw.secret.Data["value"].(string)
	return value
}

func (sw *secretWrapper) parseSignature() []byte {
	signature, _ := sw.secret.Data["signature"].(string)
	return []byte(signature)
}

func (sw *secretWrapper) parseVerification() bool {
	valid, _ := sw.secret.Data["valid"].(bool)
	return valid
}
