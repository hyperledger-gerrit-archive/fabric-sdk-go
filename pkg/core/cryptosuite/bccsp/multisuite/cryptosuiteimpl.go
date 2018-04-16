/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package multisuite

import (
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/pkcs11"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	"github.com/pkg/errors"
)

//GetSuiteByConfig returns cryptosuite adaptor for bccsp loaded according to given config
func GetSuiteByConfig(config core.CryptoSuiteConfig) (core.CryptoSuite, error) {
	securityProvider := config.SecurityProvider()

	switch {
	case strings.EqualFold(securityProvider, "SW"):
		return sw.GetSuiteByConfig(config)
	case strings.EqualFold(securityProvider, "PKCS11"):
		return pkcs11.GetSuiteByConfig(config)
	}

	return nil, errors.Errorf("Unsupported security provider requested: %s", config.SecurityProvider())
}
