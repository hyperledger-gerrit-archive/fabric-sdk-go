/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/
package sw

import (
	"errors"

	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/utils"
)

// MarshalKey marshals a BCCSP key to PEM format
func MarshalKey(k bccsp.Key, pwd []byte) (raw []byte, err error) {
	if k == nil {
		return nil, errors.New("Invalid key. It must be different from nil.")
	}
	switch k.(type) {
	case *ecdsaPrivateKey:
		kk := k.(*ecdsaPrivateKey)

		raw, err = utils.PrivateKeyToPEM(kk.privKey, pwd)
		if err != nil {
			logger.Errorf("Failed converting private key to PEM: [%s]", err)
			return nil, err
		}

	case *ecdsaPublicKey:
		kk := k.(*ecdsaPublicKey)

		raw, err = utils.PublicKeyToPEM(kk.pubKey, pwd)
		if err != nil {
			logger.Errorf("Failed converting public key to PEM: [%s]", err)
			return nil, err
		}

	case *rsaPrivateKey:
		kk := k.(*rsaPrivateKey)

		raw, err = utils.PrivateKeyToPEM(kk.privKey, pwd)
		if err != nil {
			logger.Errorf("Failed converting private key to PEM: [%s]", err)
			return nil, err
		}

	case *rsaPublicKey:
		kk := k.(*rsaPublicKey)

		raw, err = utils.PublicKeyToPEM(kk.pubKey, pwd)
		if err != nil {
			logger.Errorf("Failed converting public key to PEM: [%s]", err)
			return nil, err
		}

	case *aesPrivateKey:
		kk := k.(*aesPrivateKey)

		raw, err = utils.AEStoEncryptedPEM(kk.privKey, pwd)
		if err != nil {
			logger.Errorf("Failed converting key to PEM: [%s]", err)
			return nil, err
		}

	default:
		return nil, fmt.Errorf("Key type not reconigned [%s]", k)
	}

	return
}

// UnmarshalKey unmarshals a symetric AES BCCSP key from PEM format
func UnmarshalKey(pem []byte, pwd []byte) (bccsp.Key, error) {
	aes, err := utils.PEMtoAES(pem, pwd)
	if err != nil {
		logger.Errorf("Failed parsing key: [%s]", err)

		return nil, err
	}

	return &aesPrivateKey{aes, false}, nil
}

// UnmarshalKey unmarshals a private BCCSP key from PEM format
func UnmarshalPrivateKey(raw []byte, pwd []byte) (bccsp.Key, error) {
	key, err := utils.PEMtoPrivateKey(raw, pwd)
	if err != nil {
		logger.Errorf("Failed parsing private key: [%s].", err.Error())
		return nil, err
	}
	switch key.(type) {
	case *ecdsa.PrivateKey:
		return &ecdsaPrivateKey{key.(*ecdsa.PrivateKey)}, nil
	case *rsa.PrivateKey:
		return &rsaPrivateKey{key.(*rsa.PrivateKey)}, nil
	default:
		return nil, errors.New("Secret key type not recognized")
	}
}

// UnmarshalKey unmarshals a public BCCSP key from PEM format
func UnmarshalPublicKey(raw []byte, pwd []byte) (bccsp.Key, error) {
	key, err := utils.PEMtoPublicKey(raw, pwd)
	if err != nil {
		logger.Errorf("Failed parsing private key: [%s].", err.Error())
		return nil, err
	}
	switch key.(type) {
	case *ecdsa.PublicKey:
		return &ecdsaPublicKey{key.(*ecdsa.PublicKey)}, nil
	case *rsa.PublicKey:
		return &rsaPublicKey{key.(*rsa.PublicKey)}, nil
	default:
		return nil, errors.New("Public key type not recognized")
	}
}
