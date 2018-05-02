/*
Copyright Unchain BV Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package vault

import (
	"hash"

	"encoding/hex"

	"encoding/base64"

	vault "github.com/hashicorp/vault/api"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/pkg/errors"
)

const (
	ECDSAP256 = "ecdsa-p256"
	RSA2048   = "rsa-2048"
	RSA4096   = "rsa-4096"
)

var logger = logging.NewLogger("fabsdk/core")

type CryptoSuite struct {
	hash   func() hash.Hash
	client *vault.Client
}

type Options struct {
	client *vault.Client
	config core.CryptoSuiteConfig
}

type OptionFunc func(*Options) error

func NewCryptoSuite(optFuncs ...OptionFunc) (*CryptoSuite, error) {
	var err error
	opts := &Options{}

	for _, optFunc := range optFuncs {
		err = optFunc(opts)

		if err != nil {
			return nil, err
		}
	}

	if opts.config != nil {
		opts.client, err = getVaultClient(opts.config)

		if err != nil {
			return nil, err
		}
	}

	if opts.client == nil {
		return nil, errors.New("No suitable vault client configuration has been provided")
	}

	return &CryptoSuite{
		client: opts.client,
	}, nil
}

func WithClient(client *vault.Client) OptionFunc {
	return func(o *Options) error {
		o.client = client
		return nil
	}
}

func FromConfig(config core.CryptoSuiteConfig) OptionFunc {
	return func(o *Options) error {
		o.config = config
		return nil
	}
}

func getVaultClient(config core.CryptoSuiteConfig) (*vault.Client, error) {
	vaultConfig := &vault.Config{
		Address: config.SecurityProviderAddress(),
	}

	client, err := vault.NewClient(vaultConfig)

	client.SetToken(config.SecurityProviderToken())

	if err != nil {
		return nil, errors.Wrapf(err, "Could not initialize Vault BCCSP for address: %s", vaultConfig.Address)
	}

	return client, nil
}

func (csp *CryptoSuite) KeyGen(opts core.KeyGenOpts) (core.Key, error) {
	var err error

	// Validate arguments
	if opts == nil {
		return nil, errors.New("Invalid Opts parameter. It must not be nil.")
	}

	if opts.Ephemeral() {
		return nil, errors.New("Vault does not support ephemeral keys")
	}

	keyType, err := parseKeyType(opts.Algorithm())

	if err != nil {
		return nil, err
	}

	keyID, err := csp.keyGen(keyType)

	sw, err := csp.getKey(keyID)

	if err != nil {
		return nil, err
	}

	key, err := sw.parseKey()

	if err != nil {
		return nil, err
	}

	err = csp.storeKeyID(key.SKI(), sw.keyID())

	if err != nil {
		return nil, err
	}

	return key, nil
}

func (csp *CryptoSuite) storeKeyID(ski []byte, keyID string) error {
	_, err := csp.client.Logical().Write(
		"kv/"+hex.EncodeToString(ski),

		map[string]interface{}{
			"value": keyID,
		},
	)

	return err
}

func (csp *CryptoSuite) loadKeyID(ski []byte) (string, error) {
	secret, err := csp.client.Logical().Read("kv/" + hex.EncodeToString(ski))

	if err != nil {
		return "", err
	}

	return newSecretWrapper(secret).parseValue(), nil
}

func parseKeyType(algorithm string) (string, error) {
	switch algorithm {
	case bccsp.ECDSAP256:
		return ECDSAP256, nil
	case bccsp.RSA2048:
		return RSA2048, nil
	case bccsp.RSA4096:
		return RSA4096, nil

	default:
		return "", errors.Errorf("the algorithm %s is not supported.", algorithm)
	}
}

func (csp *CryptoSuite) keyGen(keyType string) (string, error) {
	var err error

	keyID := randomString(24)

	_, err = csp.client.Logical().Write(
		"transit/keys/"+keyID,

		map[string]interface{}{
			"type": keyType,
		},
	)

	if err != nil {
		return "", errors.Wrapf(err, "failed to generate a key of type %s", keyType)
	}

	return keyID, nil
}

func (csp *CryptoSuite) GetKey(ski []byte) (core.Key, error) {
	keyID, err := csp.loadKeyID(ski)

	if err != nil {
		return nil, err
	}

	sw, err := csp.getKey(keyID)

	if err != nil {
		return nil, err
	}

	return sw.parseKey()
}

func (csp *CryptoSuite) getKey(keyID string) (*secretWrapper, error) {
	secret, err := csp.client.Logical().Read("transit/keys/" + keyID)

	if err != nil {
		return nil, errors.Errorf("failed to find key with id `%s`", keyID)
	}

	return newSecretWrapper(secret), nil
}

func (csp *CryptoSuite) KeyImport(raw interface{}, opts core.KeyImportOpts) (k core.Key, err error) {
	// TODO implement importing
	panic("implement me")
}

func (csp *CryptoSuite) Hash(msg []byte, opts core.HashOpts) (hash []byte, err error) {
	h := csp.hash()
	h.Write(msg)
	return h.Sum(nil), nil
}

func (csp *CryptoSuite) GetHash(opts core.HashOpts) (h hash.Hash, err error) {
	return csp.hash(), nil
}

func (csp *CryptoSuite) Sign(k core.Key, digest []byte, opts core.SignerOpts) (signature []byte, err error) {
	keyID, err := csp.loadKeyID(k.SKI())

	if err != nil {
		return nil, err
	}

	secret, err := csp.client.Logical().Write(
		"transit/sign/"+keyID,

		map[string]interface{}{
			"input": base64.StdEncoding.EncodeToString(digest),
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to sign the digest")
	}

	return newSecretWrapper(secret).parseSignature(), nil
}

func (csp *CryptoSuite) Verify(k core.Key, signature, digest []byte, opts core.SignerOpts) (valid bool, err error) {
	keyID, err := csp.loadKeyID(k.SKI())

	if err != nil {
		return false, err
	}

	secret, err := csp.client.Logical().Write(
		"transit/verify/"+keyID,

		map[string]interface{}{
			"input":     base64.StdEncoding.EncodeToString(digest),
			"signature": string(signature),
		},
	)

	if err != nil {
		return false, errors.Wrapf(err, "failed to verify the signature")
	}

	return newSecretWrapper(secret).parseVerification(), nil
}
