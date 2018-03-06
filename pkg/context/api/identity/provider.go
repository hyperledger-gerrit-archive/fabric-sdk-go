/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

// Provider enables access to fabric objects such as peer and user based on config or
type Provider interface {
	CreateIdentityManager(orgName string) (Manager, error)
	CreateRegistrarService(orgName string) (RegistrarService, error)
	CreateEnrollmentService(orgName string) (EnrollmentService, error)
}

// Providers represents the SDK configured service providers context.
type Providers interface {
	IdentityProvider() Provider
}
