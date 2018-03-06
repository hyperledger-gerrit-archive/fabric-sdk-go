/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/identity"
)

// NewIdentityManager creates an Identity Manager for the given organization
func (sdk *FabricSDK) NewIdentityManager(orgName string) (identity.Manager, error) {
	im, err := sdk.provider.IdentityProvider().CreateIdentityManager(orgName)
	if err != nil {
		return nil, err
	}
	return im, nil
}

// NewEnrollmentService creates an Enrollment Service for the given organization
func (sdk *FabricSDK) NewEnrollmentService(orgName string) (identity.EnrollmentService, error) {
	im, err := sdk.provider.IdentityProvider().CreateEnrollmentService(orgName)
	if err != nil {
		return nil, err
	}
	return im, nil
}
