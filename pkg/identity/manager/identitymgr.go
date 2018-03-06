/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package manager

import (
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/logging"

	"fmt"

	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

var logger = logging.NewLogger("fabsdk/core")

// IdentityManager implements Identity Manager
type IdentityManager struct {
	config          core.Config
	cryptoSuite     core.CryptoSuite
	mspIDMap        map[string]string
	embeddedUsers   map[string]map[string]core.TLSKeyPair
	mspPrivKeyStore map[string]core.KVStore
	mspCertStore    map[string]core.KVStore
	userStore       UserStore
}

// New creates a new instance of Identity Manager
// @param {Config} client config for fabric-ca services
// @returns {Manager} Identity Manager instance
// @returns {error} error, if any
func New(stateStore core.KVStore, cryptoSuite core.CryptoSuite, config core.Config) (*IdentityManager, error) {

	netConfig, err := config.NetworkConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "network config retrieval failed")
	}

	numOrgs := len(netConfig.Organizations)

	mspIDMap := make(map[string]string, numOrgs)
	embeddedUsers := make(map[string]map[string]core.TLSKeyPair, numOrgs)
	mspPrivKeyStore := make(map[string]core.KVStore, numOrgs)
	mspCertStore := make(map[string]core.KVStore, numOrgs)

	for orgName, orgConfig := range netConfig.Organizations {

		mspID := orgConfig.MspID
		if mspID == "" {
			return nil, fmt.Errorf("org config doesn't have MspID: %s", orgName)
		}

		orgName = strings.ToLower(orgName)

		mspIDMap[orgName] = orgConfig.MspID
		embeddedUsers[mspID] = orgConfig.Users

		if orgConfig.CryptoPath == "" && len(orgConfig.Users) == 0 {
			return nil, errors.New("Either a cryptopath or an embedded list of users is required")
		}

		orgCryptoPathTemplate := orgConfig.CryptoPath
		if orgCryptoPathTemplate != "" {
			if !filepath.IsAbs(orgCryptoPathTemplate) {
				orgCryptoPathTemplate = filepath.Join(config.CryptoConfigPath(), orgCryptoPathTemplate)
			}
			mspPrivKeyStore[mspID], err = NewFileKeyStore(orgCryptoPathTemplate)
			if err != nil {
				return nil, errors.Wrapf(err, "creating a private key store failed")
			}
			mspCertStore[mspID], err = NewFileCertStore(orgCryptoPathTemplate)
			if err != nil {
				return nil, errors.Wrapf(err, "creating a cert store failed")
			}
		} else {
			logger.Warnf("Cryptopath not provided for organization [%s], MSP stores not created", orgName)
		}
	}

	userStore, err := NewCertFileUserStore1(stateStore)
	if err != nil {
		return nil, errors.Wrapf(err, "creating a user store failed")
	}

	mgr := &IdentityManager{
		config:          config,
		cryptoSuite:     cryptoSuite,
		mspPrivKeyStore: mspPrivKeyStore,
		mspCertStore:    mspCertStore,
		embeddedUsers:   embeddedUsers,
		userStore:       userStore,
	}
	return mgr, nil
}
