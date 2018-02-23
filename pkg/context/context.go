/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// IdentityContext supplies the serialized identity and key reference.
//
// TODO - refactor SigningIdentity and this interface.
type IdentityContext interface {
	MspID() string
	Identity() ([]byte, error)
	PrivateKey() core.Key
}

// SessionContext primarily represents the session and identity context
type SessionContext interface {
	IdentityContext
}

// Context supplies the configuration and signing identity to client objects.
type Context interface {
	ProviderContext
	IdentityContext
}

// ProviderContext supplies the configuration to client objects.
type ProviderContext interface {
	SigningManager() api.SigningManager
	Config() core.Config
	CryptoSuite() core.CryptoSuite
}

// User represents users that have been enrolled and represented by
// an enrollment certificate (ECert) and a signing key. The ECert must have
// been signed by one of the CAs the blockchain network has been configured to trust.
// An enrolled user (having a signing key and ECert) can conduct chaincode deployments,
// transactions and queries with the Chain.
//
// User ECerts can be obtained from a CA beforehand as part of deploying the application,
// or it can be obtained from the optional Fabric COP service via its enrollment process.
//
// Sometimes User identities are confused with Peer identities. User identities represent
// signing capability because it has access to the private key, while Peer identities in
// the context of the application/SDK only has the certificate for verifying signatures.
// An application cannot use the Peer identity to sign things because the application doesn’t
// have access to the Peer identity’s private key.
type User interface {
	IdentityContext

	Name() string
	EnrollmentCertificate() []byte
	Roles() []string
}
