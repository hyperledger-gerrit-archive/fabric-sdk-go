/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ca

// Client provides access to CA services
type Client interface {
	CAName() string
	Enroll(enrollmentID string, enrollmentSecret string) error
	Reenroll(enrollmentID string) error
	Register(request *RegistrationRequest) (string, error)
	Revoke(request *RevocationRequest) (*RevocationResponse, error)
}
