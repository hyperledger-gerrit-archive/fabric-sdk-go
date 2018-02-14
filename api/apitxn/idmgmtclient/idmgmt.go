/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package idmgmtclient

import (
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
)

// EnrollmentRequest contains enrollment request parameters
type EnrollmentRequest struct {
	// The identity name to enroll
	Name string
	// The secret returned via Register
	Secret string
	// AttrReqs are requests for attributes to add to the certificate.
	// Each attribute is added only if the requestor owns the attribute.
	AttrReqs []*AttributeRequest
}

// EnrollmentResponse contains enrollment response.
type EnrollmentResponse struct {
	EnrollmentCert []byte
	PrivateKey     apicryptosuite.Key
}

// AttributeRequest is a request for an attribute.
type AttributeRequest struct {
	Name     string
	Optional bool
}

// IdentityMgmtClient is responsible for managing dientities.
type IdentityMgmtClient interface {

	// Enroll enrolls a registered user with the org's Fabric CA
	Enroll(req EnrollmentRequest) (*EnrollmentResponse, error)
}
