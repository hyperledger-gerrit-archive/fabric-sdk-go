/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/identity"
)

// NewIdentityManager creates an Identity Manager for the given organization
func (sdk *FabricSDK) NewIdentityManager(orgName string) (identity.IdentityManager, error) {
	im, err := sdk.identityProvider.CreateIdentityManager(orgName)
	if err != nil {
		return nil, err
	}
	return im, nil
}
