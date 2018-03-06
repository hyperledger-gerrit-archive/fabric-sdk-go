/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ca

// Provider enables access CA services
type Provider interface {
	CreateCAService(orgName string) (Client, error)
}

// Providers represents the SDK configured service providers context.
type Providers interface {
	CAProvider() Provider
}
