/*
Copyright Hyperledger and its contributors.

SPDX-License-Identifier: Apache-2.0
*/

package vault

import (
	"hash"

	"encoding/hex"

	"encoding/base64"

	"crypto/sha256"
	"crypto/sha512"

	vault "github.com/hashicorp/vault/api"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

// Constants describing encryption algorithms
const (
	ECDSAP256 = "ecdsa-p256"
	RSA2048   = "rsa-2048"
	RSA4096   = "rsa-4096"
)

var logger = logging.NewLogger("fabsdk/core")

// CryptoSuite is a vault implementation of the core.CryptoSuite interface
type CryptoSuite struct {
	hashers map[string]Hasher
	client  *vault.Client
}

// options configure a new CryptoSuite. options are set by the OptionFunc values passed to NewCryptoSuite.
type options struct {
	client  *vault.Client
	hashers map[string]Hasher
	config  core.CryptoSuiteConfig
}

// OptionFunc configures how the CryptoSuite is set up.
type OptionFunc func(*options) error

// NewCryptoSuite constructs a new CryptoSuite, configured via provided OptionFuncs
func NewCryptoSuite(optFuncs ...OptionFunc) (*CryptoSuite, error) {
	var err error
	opts := &options{}

	for _, optFunc := range optFuncs {
		err = optFunc(opts)

		if err != nil {
			return nil, err
		}
	}

	if opts.client == nil {
		opts.client, err = getVaultClient(opts.config)

		if err != nil {
			return nil, err
		}
	}

	hashers := getHashers(opts.config)

	for key, hasher := range opts.hashers {
		hashers[key] = hasher
	}

	logger.Debug("Initialized the vault CryptoSuite")

	return &CryptoSuite{
		client:  opts.client,
		hashers: hashers,
	}, nil
}

func getHashers(cfg core.CryptoSuiteConfig) map[string]Hasher {
	if cfg == nil {
		return nil
	}

	defaultHasher := parseHasher(cfg.SecurityAlgorithm())

	// Set the hashers
	hashers := make(map[string]Hasher)

	if defaultHasher != nil {
		hashers[bccsp.SHA] = defaultHasher
	}

	hashers[bccsp.SHA256] = &hasher{hash: sha256.New}
	hashers[bccsp.SHA384] = &hasher{hash: sha512.New384}
	hashers[bccsp.SHA3_256] = &hasher{hash: sha3.New256}
	hashers[bccsp.SHA3_384] = &hasher{hash: sha3.New384}

	return hashers
}

func parseHasher(algorithm string) *hasher {
	switch algorithm {
	case bccsp.SHA256:
		return &hasher{hash: sha256.New}
	case bccsp.SHA384:
		return &hasher{hash: sha512.New384}
	case bccsp.SHA3_256:
		return &hasher{hash: sha3.New256}
	case bccsp.SHA3_384:
		return &hasher{hash: sha3.New384}

	default:
		return nil
	}
}

// WithClient allows to set the vault client of the CryptoSuite
func WithClient(client *vault.Client) OptionFunc {
	return func(o *options) error {
		o.client = client
		return nil
	}
}

// WithHashers allows to provide additional hashers to the CryptoSuite
func WithHashers(hashers map[string]Hasher) OptionFunc {
	return func(o *options) error {
		o.hashers = hashers
		return nil
	}
}

// FromConfig uses a core.CryptoSuiteConfig to configure the vault client of the CryptoSuite
func FromConfig(config core.CryptoSuiteConfig) OptionFunc {
	return func(o *options) error {
		o.config = config
		return nil
	}
}

func getVaultClient(config core.CryptoSuiteConfig) (*vault.Client, error) {
	if config == nil {
		return nil, errors.New("Unable to obtain vault client configuration from nil CryptoSuiteConfig")
	}

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

// KeyGen generates a key using opts.
func (csp *CryptoSuite) KeyGen(opts core.KeyGenOpts) (core.Key, error) {
	var err error

	// Validate arguments
	if opts == nil {
		return nil, errors.New("opts must not be nil")
	}

	if opts.Ephemeral() {
		return nil, errors.New("vault does not support ephemeral keys")
	}

	keyType, err := parseKeyType(opts.Algorithm())

	if err != nil {
		return nil, err
	}

	keyID, err := csp.keyGen(keyType)

	if err != nil {
		return nil, err
	}

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

// GetKey returns the key this CSP associates to
// the Subject Key Identifier ski.
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

// KeyImport imports a key from its raw representation using opts.
// The opts argument should be appropriate for the primitive used.
func (csp *CryptoSuite) KeyImport(raw interface{}, opts core.KeyImportOpts) (k core.Key, err error) {
	// TODO implement importing
	panic("implement me")
}

// Hash hashes messages msg using options opts.
func (csp *CryptoSuite) Hash(msg []byte, opts core.HashOpts) (digest []byte, err error) {
	// Validate arguments
	if opts == nil {
		return nil, errors.New("opts must not be nil")
	}

	hasher, found := csp.hashers[opts.Algorithm()]
	if !found {
		return nil, errors.Errorf("unsupported 'HashOpt' provided [%v]", opts)
	}

	digest, err = hasher.Hash(msg, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "failed hashing with opts [%v]", opts)
	}

	return
}

// GetHash returns and instance of hash.Hash using options opts.
// If opts is nil then the default hash function is returned.
func (csp *CryptoSuite) GetHash(opts core.HashOpts) (h hash.Hash, err error) {
	// Validate arguments
	if opts == nil {
		return nil, errors.New("opts must not be nil")
	}

	hasher, found := csp.hashers[opts.Algorithm()]
	if !found {
		return nil, errors.Errorf("unsupported 'HashOpt' provided [%v]", opts)
	}

	h, err = hasher.GetHash(opts)
	if err != nil {
		return nil, errors.Wrapf(err, "failed getting hash function with opts [%v]", opts)
	}

	return
}

// Sign signs digest using key k.
// The opts argument should be appropriate for the algorithm used.
//
// Note that when a signature of a hash of a larger message is needed,
// the caller is responsible for hashing the larger message and passing
// the hash (as digest).
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

// Verify verifies signature against key k and digest
// The opts argument should be appropriate for the algorithm used.
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
