/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"encoding/hex"
	"path"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/api/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/keyvaluestore"
	"github.com/pkg/errors"
)

// NewFileKeyStore ...
func NewFileKeyStore(cryptoConfigMSPPath string) (core.KVStore, error) {
	opts := &keyvaluestore.FileKeyValueStoreOptions{
		Path: cryptoConfigMSPPath,
		KeySerializer: func(key interface{}) (string, error) {
			pkk, ok := key.(*msp.PrivKeyKey)
			if !ok {
				return "", errors.New("converting key to PrivKeyKey failed")
			}
			if pkk == nil || pkk.MSPID == "" || pkk.Username == "" || pkk.SKI == nil {
				return "", errors.New("invalid key")
			}

			// TODO: refactor to case insensitive or remove eventually.
			r := strings.NewReplacer("{userName}", pkk.Username, "{username}", pkk.Username)
			keyDir := path.Join(r.Replace(cryptoConfigMSPPath), "keystore")

			return path.Join(keyDir, hex.EncodeToString(pkk.SKI)+"_sk"), nil
		},
	}
	return keyvaluestore.New(opts)
}
