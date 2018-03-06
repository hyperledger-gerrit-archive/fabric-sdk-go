/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package manager

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/cryptoutil"

	"strings"

	fabricCaUtil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/pkg/errors"
)

func newUser(userData UserData, cryptoSuite core.CryptoSuite) (*User, error) {
	pubKey, err := cryptoutil.GetPublicKeyFromCert(userData.EnrollmentCertificate, cryptoSuite)
	if err != nil {
		return nil, errors.WithMessage(err, "fetching public key from cert failed")
	}
	pk, err := cryptoSuite.GetKey(pubKey.SKI())
	if err != nil {
		return nil, errors.WithMessage(err, "cryptoSuite GetKey failed")
	}
	u := &User{
		MspID_: userData.MspID,
		Name_:  userData.Name,
		EnrollmentCertificate_: userData.EnrollmentCertificate,
		PrivateKey_:            pk,
	}
	return u, nil
}

// NewUser creates a new user instance from user data
func (mgr *IdentityManager) NewUser(userData UserData) (*User, error) {
	return newUser(userData, mgr.cryptoSuite)
}

func (mgr *IdentityManager) loadUserFromStore(mspID string, userName string) (*User, error) {
	if mgr.userStore == nil {
		return nil, core.ErrUserNotFound
	}
	var user *User
	userData, err := mgr.userStore.Load(UserIdentifier{MspID: mspID, Name: userName})
	if err != nil {
		return nil, err
	}
	user, err = mgr.NewUser(userData)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetSigningIdentity returns a signing identity for the given user Name
func (mgr *IdentityManager) GetSigningIdentity(mspID string, userName string) (*core.SigningIdentity, error) {

	userName = strings.ToLower(userName)

	if mspID == "" {
		return nil, errors.New("MspID is empty")
	}
	if userName == "" {
		return nil, errors.New("userName is empty")
	}
	user, err := mgr.GetUser(mspID, userName)
	if err != nil {
		return nil, err
	}
	signingIdentity := &core.SigningIdentity{MspID: user.MspID(), PrivateKey: user.PrivateKey(), EnrollmentCert: user.EnrollmentCertificate()}
	return signingIdentity, nil
}

// GetUser returns a user for the given user Name
func (mgr *IdentityManager) GetUser(mspID string, userName string) (core.User, error) {

	if mspID == "" {
		return nil, errors.New("MspID is empty")
	}
	if userName == "" {
		return nil, errors.New("userName is empty")
	}

	userName = strings.ToLower(userName)

	u, err := mgr.loadUserFromStore(mspID, userName)
	if err != nil {
		if err != core.ErrUserNotFound {
			return nil, errors.WithMessage(err, "getting private key from cert failed")
		}
		// Not found, continue
	}

	if u == nil {
		certBytes, err := mgr.getEmbeddedCertBytes(mspID, userName)
		if err != nil && err != core.ErrUserNotFound {
			return nil, errors.WithMessage(err, "fetching embedded cert failed")
		}
		if certBytes == nil {
			certBytes, err = mgr.getCertBytesFromCertStore(mspID, userName)
			if err != nil && err != core.ErrUserNotFound {
				return nil, errors.WithMessage(err, "fetching cert from store failed")
			}
		}
		if certBytes == nil {
			return nil, core.ErrUserNotFound
		}
		privateKey, err := mgr.getEmbeddedPrivateKey(mspID, userName)
		if err != nil && err != core.ErrUserNotFound {
			return nil, errors.WithMessage(err, "fetching embedded private key failed")
		}
		if privateKey == nil {
			privateKey, err = mgr.getPrivateKeyFromCert(mspID, userName, certBytes)
			if err != nil {
				return nil, errors.WithMessage(err, "getting private key from cert failed")
			}
		}
		if privateKey == nil {
			return nil, fmt.Errorf("unable to find private key for user [%s]", userName)
		}
		u = &User{
			MspID_: mspID,
			Name_:  userName,
			EnrollmentCertificate_: certBytes,
			PrivateKey_:            privateKey,
		}
	}
	return u, nil
}

func (mgr *IdentityManager) getEmbeddedUserTLSKeyPair(mspID string, userName string) (*core.TLSKeyPair, error) {

	orgMap, ok := mgr.embeddedUsers[mspID]
	if !ok {
		return nil, core.ErrUserNotFound
	}
	tlsKeyPair, ok := orgMap[userName]
	if !ok {
		return nil, core.ErrUserNotFound
	}
	return &tlsKeyPair, nil
}

func (mgr *IdentityManager) getEmbeddedCertBytes(mspID string, userName string) ([]byte, error) {

	tlsKeyPair, err := mgr.getEmbeddedUserTLSKeyPair(mspID, userName)
	if err != nil {
		return nil, err
	}

	certPem := tlsKeyPair.Cert.Pem
	certPath := tlsKeyPair.Cert.Path

	if certPem == "" && certPath == "" {
		return nil, core.ErrUserNotFound
	}

	var pemBytes []byte

	if certPem != "" {
		pemBytes = []byte(certPem)
	} else if certPath != "" {
		pemBytes, err = ioutil.ReadFile(certPath)
		if err != nil {
			return nil, errors.WithMessage(err, "reading cert from embedded path failed")
		}
	}

	return pemBytes, nil
}

func (mgr *IdentityManager) getEmbeddedPrivateKey(mspID string, userName string) (core.Key, error) {

	key, err := mgr.getEmbeddedUserTLSKeyPair(mspID, userName)
	if err != nil {
		return nil, err
	}

	keyPem := key.Key.Pem
	keyPath := key.Key.Path

	var privateKey core.Key
	var pemBytes []byte

	if keyPem != "" {
		// Try importing from the Embedded Pem
		pemBytes = []byte(keyPem)
	} else if keyPath != "" {
		// Try importing from the Embedded Path
		_, err := os.Stat(keyPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, errors.WithMessage(err, "OS stat embedded path failed")
			}
			// file doesn't exist, continue
		} else {
			// file exists, try to read it
			pemBytes, err = ioutil.ReadFile(keyPath)
			if err != nil {
				return nil, errors.WithMessage(err, "reading private key from embedded path failed")
			}
		}
	}

	if pemBytes != nil {
		// Try the crypto provider as a SKI
		privateKey, err = mgr.cryptoSuite.GetKey(pemBytes)
		if err != nil || privateKey == nil {
			// Try as a pem
			privateKey, err = fabricCaUtil.ImportBCCSPKeyFromPEMBytes(pemBytes, mgr.cryptoSuite, true)
			if err != nil {
				return nil, errors.Wrapf(err, "import private key failed %v", keyPem)
			}
		}
	}

	return privateKey, nil
}

func (mgr *IdentityManager) getPrivateKeyPemFromKeyStore(mspID string, userName string, ski []byte) ([]byte, error) {
	if mgr.mspPrivKeyStore == nil {
		return nil, nil
	}
	keyStore, ok := mgr.mspPrivKeyStore[mspID]
	if !ok {
		return nil, core.ErrUserNotFound
	}
	key, err := keyStore.Load(
		&PrivKeyKey{
			MspID:    mspID,
			UserName: userName,
			SKI:      ski,
		})
	if err != nil {
		return nil, err
	}
	keyBytes, ok := key.([]byte)
	if !ok {
		return nil, errors.New("key from store is not []byte")
	}
	return keyBytes, nil
}

func (mgr *IdentityManager) getCertBytesFromCertStore(mspID string, userName string) ([]byte, error) {
	if mgr.mspCertStore == nil {
		return nil, core.ErrUserNotFound
	}
	certStore, ok := mgr.mspCertStore[mspID]
	if !ok {
		return nil, core.ErrUserNotFound
	}
	cert, err := certStore.Load(&CertKey{
		MspID:    mspID,
		UserName: userName,
	})
	if err != nil {
		if err == core.ErrKeyValueNotFound {
			return nil, core.ErrUserNotFound
		}
		return nil, err
	}
	certBytes, ok := cert.([]byte)
	if !ok {
		return nil, errors.New("cert from store is not []byte")
	}
	return certBytes, nil
}

func (mgr *IdentityManager) getPrivateKeyFromCert(orgName string, userName string, cert []byte) (core.Key, error) {
	if cert == nil {
		return nil, errors.New("cert is nil")
	}
	pubKey, err := cryptoutil.GetPublicKeyFromCert(cert, mgr.cryptoSuite)
	if err != nil {
		return nil, errors.WithMessage(err, "fetching public key from cert failed")
	}
	privKey, err := mgr.getPrivateKeyFromKeyStore(orgName, userName, pubKey.SKI())
	if err == nil {
		return privKey, nil
	}
	if err != core.ErrKeyValueNotFound {
		return nil, errors.WithMessage(err, "fetching private key from key store failed")
	}
	return mgr.cryptoSuite.GetKey(pubKey.SKI())
}

func (mgr *IdentityManager) getPrivateKeyFromKeyStore(orgName string, userName string, ski []byte) (core.Key, error) {
	pemBytes, err := mgr.getPrivateKeyPemFromKeyStore(orgName, userName, ski)
	if err != nil {
		return nil, err
	}
	if pemBytes != nil {
		return fabricCaUtil.ImportBCCSPKeyFromPEMBytes(pemBytes, mgr.cryptoSuite, true)
	}
	return nil, core.ErrKeyValueNotFound
}
