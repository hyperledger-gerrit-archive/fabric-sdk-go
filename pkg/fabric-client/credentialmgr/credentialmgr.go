/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialmgr

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	fabricCaUtil "github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric/bccsp"
)

// CredentialManager is used for retriving user's signing identity (ecert + private key)
type CredentialManager struct {
	orgName        string
	keyDir         string
	certDir        string
	config         apiconfig.Config
	cryptoProvider bccsp.BCCSP
}

// NewCredentialManager Constructor for a credential manager.
// @param {string} orgName - organisation id
// @returns {CredentialManager} new credential manager
func NewCredentialManager(orgName string, config apiconfig.Config, cryptoProvider bccsp.BCCSP) (apifabclient.CredentialManager, error) {

	// TODO: Remove hard-coded domains
	orgToDomain := map[string]string{
		"peerorg1": "org1.example.com",
		"peerorg2": "org2.example.com",
	}

	domain := orgToDomain[orgName]

	mspDir := strings.Replace("peerOrganizations/{domain}/users/{userName}@{domain}/msp/", "{domain}", domain, -1)
	return &CredentialManager{orgName: orgName, config: config, keyDir: mspDir + "keystore", certDir: mspDir + "signcerts", cryptoProvider: cryptoProvider}, nil
}

// GetSigningIdentity will sign the given object with provided key,
func (mgr *CredentialManager) GetSigningIdentity(userName string) (*apifabclient.SigningIdentity, error) {

	privateKeyDir := filepath.Join(mgr.config.CryptoConfigPath(), strings.Replace(mgr.keyDir, "{userName}", userName, -1))
	privateKeyPath, err := getFirstPathFromDir(privateKeyDir)
	if err != nil {
		return nil, fmt.Errorf("Error finding the private key path: %v", err)
	}

	enrollmentCertDir := filepath.Join(mgr.config.CryptoConfigPath(), strings.Replace(mgr.certDir, "{userName}", userName, -1))
	enrollmentCertPath, err := getFirstPathFromDir(enrollmentCertDir)
	if err != nil {
		return nil, fmt.Errorf("Error finding the enrollment cert path: %v", err)
	}

	return mgr.GetSigningIdentityFromPath(privateKeyPath, enrollmentCertPath)
}

// GetSigningIdentityFromPath will create signing identity the given private key path and enrollment certificate path
func (mgr *CredentialManager) GetSigningIdentityFromPath(privateKeyPath string, enrollmentCertPath string) (*apifabclient.SigningIdentity, error) {

	mspID, err := mgr.config.MspID(mgr.orgName)
	if err != nil {
		return nil, fmt.Errorf("Error reading MSP ID config: %s", err)
	}

	privateKey, err := fabricCaUtil.ImportBCCSPKeyFromPEM(privateKeyPath, mgr.cryptoProvider, true)
	if err != nil {
		return nil, fmt.Errorf("Error importing private key: %v", err)
	}
	enrollmentCert, err := ioutil.ReadFile(enrollmentCertPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading from the enrollment cert path: %v", err)
	}

	signingIdentity := &apifabclient.SigningIdentity{MspID: mspID, PrivateKey: privateKey, EnrollmentCert: enrollmentCert}

	return signingIdentity, nil

}

// Gets the first path from the dir directory
func getFirstPathFromDir(dir string) (string, error) {

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("Could not read directory %s, err %s", err, dir)
	}

	for _, p := range files {
		if p.IsDir() {
			continue
		}

		fullName := filepath.Join(dir, string(filepath.Separator), p.Name())
		fmt.Printf("Reading file %s\n", fullName)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		fullName := filepath.Join(dir, string(filepath.Separator), f.Name())
		return fullName, nil
	}

	return "", fmt.Errorf("No paths found in directory: %s", dir)
}
