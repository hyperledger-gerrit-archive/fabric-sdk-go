/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// SigningIdentity is the identity object that encapsulates the user's private key for signing
// and the user's enrollment certificate (identity)
type SigningIdentity struct {
	MspID          string
	EnrollmentCert []byte
	PrivateKey     core.Key
}

// EnrollmentService provides user enrollment services
type EnrollmentService interface {
	CAName() string
	Enroll(enrollmentID string, enrollmentSecret string) error
	Reenroll(enrollmentID string) error
}

// RegistrarService handle operations on identities that require
// the CA's registrar identity
type RegistrarService interface {
	CAName() string
	Register(request *RegistrationRequest) (string, error)
	Revoke(request *RevocationRequest) (*RevocationResponse, error)
}

// Manager provides management of identities that is local to SDK client
type Manager interface {
	GetSigningIdentity(name string) (*SigningIdentity, error)
	GetUser(name string) (core.User, error)
}
