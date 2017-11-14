/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	"crypto"
)

//CryptoSuite adaptor for all bccsp functionalities used by SDK
type CryptoSuite interface {
	GetKey(ski []byte) (k Key, err error)

	Hash(msg []byte, opts HashOpts) (hash []byte, err error)

	Sign(k Key, digest []byte, opts SignerOpts) (signature []byte, err error)
}

// Key represents a cryptographic key
type Key interface {

	// Bytes converts this key to its byte representation,
	// if this operation is allowed.
	Bytes() ([]byte, error)

	// SKI returns the subject key identifier of this key.
	SKI() []byte

	// Symmetric returns true if this key is a symmetric key,
	// false is this key is asymmetric
	Symmetric() bool

	// Private returns true if this key is a private key,
	// false otherwise.
	Private() bool

	// PublicKey returns the corresponding public key part of an asymmetric public/private key pair.
	// This method returns an error in symmetric key schemes.
	PublicKey() (Key, error)
}

// HashOpts contains options for hashing with a CSP.
type HashOpts interface {

	// Algorithm returns the hash algorithm identifier (to be used).
	Algorithm() string
}

// SignerOpts contains options for signing with a CSP.
type SignerOpts interface {
	crypto.SignerOpts
}
