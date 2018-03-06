/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package core

// SigningIdentity is the identity object that encapsulates the user's private key for signing
// and the user's enrollment certificate (identity)
type SigningIdentity struct {
	MspID          string
	EnrollmentCert []byte
	PrivateKey     Key
}

// IdentityManager provides management of identities that is local to SDK client
type IdentityManager interface {
	GetSigningIdentity(orgName string, iserName string) (*SigningIdentity, error)
	GetUser(orgName string, iserName string) (User, error)
}
