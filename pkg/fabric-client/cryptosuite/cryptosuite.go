/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp"
)

//GetSuite returns cryptosuite adaptor
func GetSuite(bccsp bccsp.BCCSP) apifabclient.CryptoSuite {
	return &cryptoSuite{bccsp}
}

//GetKey returns implementation of of cryptosuite.Key
func GetKey(newkey bccsp.Key) apifabclient.Key {
	return &key{newkey}
}

type cryptoSuite struct {
	bccsp bccsp.BCCSP
}

func (c *cryptoSuite) GetKey(ski []byte) (k apifabclient.Key, err error) {
	key, err := c.bccsp.GetKey(ski)
	return GetKey(key), err
}

func (c *cryptoSuite) Hash(msg []byte, opts apifabclient.HashOpts) (hash []byte, err error) {
	return c.bccsp.Hash(msg, opts)
}

func (c *cryptoSuite) Sign(k apifabclient.Key, digest []byte, opts apifabclient.SignerOpts) (signature []byte, err error) {
	return c.bccsp.Sign(k.(*key).key, digest, opts)
}

type key struct {
	key bccsp.Key
}

func (k *key) Bytes() ([]byte, error) {
	return k.key.Bytes()
}

func (k *key) SKI() []byte {
	return k.key.SKI()
}

func (k *key) Symmetric() bool {
	return k.key.Symmetric()
}

func (k *key) Private() bool {
	return k.key.Private()
}

func (k *key) PublicKey() (apifabclient.Key, error) {
	key, err := k.key.PublicKey()
	return GetKey(key), err
}
