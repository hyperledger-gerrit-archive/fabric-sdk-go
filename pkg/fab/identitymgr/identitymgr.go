/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identitymgr

import (
	"path/filepath"
	"strings"

	fabric_ca "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/identitymgr/persistence"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// IdentityManager manages user identities within an organization
type IdentityManager struct {
	config          core.Config
	orgName         string
	mspID           string
	embeddedUsers   map[string]core.TLSKeyPair
	mspPrivKeyStore api.KVStore
	mspCertStore    api.KVStore
	cryptoProvider  core.CryptoSuite
	userStore       api.UserStore

	// Fabric CA client
	caClient  *fabric_ca.Client
	registrar core.EnrollCredentials
}

// New Constructor for an identity manager.
// @param {string} orgName - organisation id
// @returns {IdentityManager} new credential manager
func New(orgName string, config core.Config, cryptoProvider core.CryptoSuite) (fab.IdentityManager, error) {

	netConfig, err := config.NetworkConfig()
	if err != nil {
		return nil, errors.New("network config retrieval failed")
	}

	// viper keys are case insensitive
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		return nil, errors.New("org config retrieval failed")
	}

	if orgConfig.CryptoPath == "" && len(orgConfig.Users) == 0 {
		return nil, errors.New("Either a cryptopath or an embedded list of users is required")
	}

	var mspPrivKeyStore api.KVStore
	var mspCertStore api.KVStore

	orgCryptoPathTemplate := orgConfig.CryptoPath
	if orgCryptoPathTemplate != "" {
		if !filepath.IsAbs(orgCryptoPathTemplate) {
			orgCryptoPathTemplate = filepath.Join(config.CryptoConfigPath(), orgCryptoPathTemplate)
		}
		mspPrivKeyStore, err = persistence.NewFileKeyStore(orgCryptoPathTemplate)
		if err != nil {
			return nil, errors.Wrapf(err, "creating a private key store failed")
		}
		mspCertStore, err = persistence.NewFileCertStore(orgCryptoPathTemplate)
		if err != nil {
			return nil, errors.Wrapf(err, "creating a cert store failed")
		}
	} else {
		logger.Warnf("Cryptopath not provided for organization [%s], MSP store(s) not created", orgName)
	}

	// In the future, shared UserStore from the SDK context will be used
	var userStore api.UserStore
	clientCofig, err := config.Client()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to retrieve client config")
	}
	if clientCofig.CredentialStore.Path != "" {
		userStore, err = identity.NewCertFileUserStore(clientCofig.CredentialStore.Path, cryptoProvider)
	}

	return &IdentityManager{
		orgName:         orgName,
		mspID:           orgConfig.MspID,
		config:          config,
		embeddedUsers:   orgConfig.Users,
		mspPrivKeyStore: mspPrivKeyStore,
		mspCertStore:    mspCertStore,
		cryptoProvider:  cryptoProvider,
		userStore:       userStore,
	}, nil
}

// CAName returns the CA name.
func (mgr *IdentityManager) CAName() string {
	return mgr.caClient.Config.CAName
}
