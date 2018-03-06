/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/ca"
)

// NewCAService creates an Enrollment Service for the given organization
func (sdk *FabricSDK) NewCAService(orgName string) (ca.Client, error) {
	im, err := sdk.provider.CAProvider().CreateCAService(orgName)
	if err != nil {
		return nil, err
	}
	return im, nil
}
