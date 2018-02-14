/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabca

import (
	"github.com/hyperledger/fabric-sdk-go/api/core/identity"
)

// FabricCAClient is the client interface for fabric-ca
// Deprecated.
type FabricCAClient interface {
	CAName() string
	identity.Manager
	identity.EnrollmentService
}
