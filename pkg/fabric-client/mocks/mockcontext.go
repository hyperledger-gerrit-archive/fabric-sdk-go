/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/crypto"
	config "github.com/hyperledger/fabric-sdk-go/pkg/context/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/apicryptosuite"

	"github.com/hyperledger/fabric-sdk-go/pkg/context"
)

// MockProviderContext holds core providers to enable mocking.
type MockProviderContext struct {
	config         config.Config
	cryptoSuite    apicryptosuite.CryptoSuite
	signingManager context.SigningManager
}

// NewMockProviderContext creates a MockProviderContext consisting of defaults
func NewMockProviderContext() *MockProviderContext {
	context := MockProviderContext{
		config:         NewMockConfig(),
		signingManager: NewMockSigningManager(),
		cryptoSuite:    &MockCryptoSuite{},
	}
	return &context
}

// NewMockProviderContextCustom creates a MockProviderContext consisting of the arguments
func NewMockProviderContextCustom(config config.Config, cryptoSuite apicryptosuite.CryptoSuite, signer context.SigningManager) *MockProviderContext {
	context := MockProviderContext{
		config:         config,
		signingManager: signer,
		cryptoSuite:    cryptoSuite,
	}
	return &context
}

// Config returns the mock configuration.
func (pc *MockProviderContext) Config() config.Config {
	return pc.config
}

// SetConfig sets the mock configuration.
func (pc *MockProviderContext) SetConfig(config config.Config) {
	pc.config = config
}

// CryptoSuite returns the mock crypto suite.
func (pc *MockProviderContext) CryptoSuite() apicryptosuite.CryptoSuite {
	return pc.cryptoSuite
}

// SigningManager returns the mock signing manager.
func (pc *MockProviderContext) SigningManager() context.SigningManager {
	return pc.signingManager
}

// MockContext holds core providers and identity to enable mocking.
type MockContext struct {
	*MockProviderContext
	context.IdentityContext
}

// NewMockContext creates a MockContext consisting of defaults and an identity
func NewMockContext(ic context.IdentityContext) *MockContext {
	ctx := MockContext{
		MockProviderContext: NewMockProviderContext(),
		IdentityContext:     ic,
	}
	return &ctx
}

// NewMockTxnID creates mock TxnID based on mock user.
func NewMockTxnID() (context.TransactionID, error) {
	user := NewMockUser("test")

	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return context.TransactionID{}, err
	}

	creator, err := user.Identity()
	if err != nil {
		return context.TransactionID{}, err
	}

	h := sha256.New()
	id, err := computeTxnID(nonce, creator, h)
	if err != nil {
		return context.TransactionID{}, err
	}

	txnID := context.TransactionID{
		ID:    id,
		Nonce: nonce,
	}

	return txnID, nil
}

func computeTxnID(nonce, creator []byte, h hash.Hash) (string, error) {
	b := append(nonce, creator...)

	_, err := h.Write(b)
	if err != nil {
		return "", err
	}
	digest := h.Sum(nil)
	id := hex.EncodeToString(digest)

	return id, nil
}
