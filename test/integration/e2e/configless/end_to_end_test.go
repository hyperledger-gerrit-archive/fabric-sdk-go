/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package configless

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration/e2e"
)

// this test mimics the original e2e test with the difference of injecting interface functions implementations
// to programmatically supply configs instead of using a yaml file. With this change, application developers can fetch
// configs from any source as long as they provide their own implementations.

func TestE2E(t *testing.T) {

	//Using same Run call as e2e package but with programmatically overriding interfaces
	e2e.RunWithoutSetup(t, nil,
		fabsdk.WithConfigEndpoint(endpointConfigImpls...),
		fabsdk.WithConfigCryptoSuite(cryptoConfigImpls...),
		fabsdk.WithConfigIdentity(identityConfigImpls...),
	)
}
