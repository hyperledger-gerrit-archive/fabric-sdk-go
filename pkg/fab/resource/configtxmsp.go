/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gogo/protobuf/proto"

	mspcfg "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
)

const (
	cacerts              = "cacerts"
	admincerts           = "admincerts"
	signcerts            = "signcerts"
	keystore             = "keystore"
	intermediatecerts    = "intermediatecerts"
	crlsfolder           = "crls"
	configfilename       = "config.yaml"
	tlscacerts           = "tlscacerts"
	tlsintermediatecerts = "tlsintermediatecerts"
)

// GenerateMspDir generates a MSP directory, using values from the provided MSP config.
// The intended usage is within the scope of creating a genesis block.
func GenerateMspDir(mspDir string, config *msp.MSPConfig) error {

	if mspcfg.ProviderTypeToString(mspcfg.ProviderType(config.Type)) != "bccsp" {
		return fmt.Errorf("Unsupported MSP config type")
	}

	cfg := &msp.FabricMSPConfig{}
	err := proto.Unmarshal(config.Config, cfg)
	if err == nil {
		err = generateCertDir(filepath.Join(mspDir, cacerts), cfg.RootCerts)
	}
	if err == nil {
		err = generateCertDir(filepath.Join(mspDir, admincerts), cfg.Admins)
	}
	if err == nil {
		err = generateCertDir(filepath.Join(mspDir, intermediatecerts), cfg.IntermediateCerts)
	}
	if err == nil {
		err = generateCertDir(filepath.Join(mspDir, tlscacerts), cfg.TlsRootCerts)
	}
	if err == nil {
		err = generateCertDir(filepath.Join(mspDir, tlsintermediatecerts), cfg.TlsIntermediateCerts)
	}
	if err == nil {
		err = generateCertDir(filepath.Join(mspDir, crlsfolder), cfg.RevocationList)
	}

	return err
}

func generateCertDir(certDir string, certs [][]byte) error {
	err := os.MkdirAll(certDir, 0755)
	if err != nil {
		return err
	}
	if len(certs) == 0 {
		return nil
	}
	for counter, certBytes := range certs {
		fileName := filepath.Join(certDir, "cert"+fmt.Sprintf("%d", counter)+".pem")
		err = ioutil.WriteFile(fileName, certBytes, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
