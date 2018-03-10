/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

// Provider enables access to MSP services
type Provider interface {
	IdentityManager(orgName string) (IdentityManager, bool)
}

// Providers represents the MSP service providers context.
type Providers interface {
	IdentityManager(orgName string) (IdentityManager, bool)
}
