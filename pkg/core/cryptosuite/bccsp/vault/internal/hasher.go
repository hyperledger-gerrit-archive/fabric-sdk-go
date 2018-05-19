/*
Copyright Hyperledger and its contributors.

SPDX-License-Identifier: Apache-2.0
*/

package internal

import (
	"hash"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
)

// Hasher is an implementation of the Hasher interface
type Hasher struct {
	HashFunc func() hash.Hash
}

// Hash hashes messages msg using options opts.
// If opts is nil, the default HashFunc function will be used.
func (c *Hasher) Hash(msg []byte, opts bccsp.HashOpts) (hash []byte, err error) {
	h := c.HashFunc()
	_, err = h.Write(msg)

	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// GetHash returns and instance of HashFunc.Hash using options opts.
// If opts is nil, the default HashFunc function will be returned.
func (c *Hasher) GetHash(opts bccsp.HashOpts) (h hash.Hash, err error) {
	return c.HashFunc(), nil
}
