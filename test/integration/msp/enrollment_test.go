/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"testing"

	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/attrmgr"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/stretchr/testify/assert"
)

const (
	IdentityTypeUser = "User"
)

func TestRegisterEnroll(t *testing.T) {

	// Instantiate the SDK
	sdk, err := fabsdk.New(integration.ConfigBackend)

	if err != nil {
		t.Fatalf("SDK init failed: %v", err)
	}

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)
	defer integration.CleanupUserData(t, sdk)

	ctxProvider := sdk.Context()

	// Get the Client.
	// Without WithOrg option, uses default client organization.
	mspClient, err := msp.New(ctxProvider)
	if err != nil {
		t.Fatalf("failed to create CA client: %v", err)
	}

	// As this integration test spawns a fresh CA instance,
	// we have to enroll the CA registrar first. Otherwise,
	// CA operations that require the registrar's identity
	// will be rejected by the CA.
	registrarEnrollID, registrarEnrollSecret := getRegistrarEnrollmentCredentials(t, ctxProvider)
	err = mspClient.Enroll(registrarEnrollID, msp.WithSecret(registrarEnrollSecret))
	if err != nil {
		t.Fatalf("Enroll failed: %v", err)
	}

	// The enrollment process generates a new private key and
	// enrollment certificate for the user. The private key
	// is stored in the SDK crypto provider's key store, while the
	// enrollment certificate is stored in the SKD's user store
	// (state store). The CAClient will lookup the
	// registrar's identity information in these stores.

	// Generate a random user name
	username := integration.GenerateRandomID()

	testAttributes := []msp.Attribute{
		msp.Attribute{
			Name:  integration.GenerateRandomID(),
			Value: fmt.Sprintf("%s:ecert", integration.GenerateRandomID()),
			ECert: true,
		},
		msp.Attribute{
			Name:  integration.GenerateRandomID(),
			Value: fmt.Sprintf("%s:ecert", integration.GenerateRandomID()),
			ECert: true,
		},
	}

	// Register the new user
	enrollmentSecret, err := mspClient.Register(&msp.RegistrationRequest{
		Name:       username,
		Type:       IdentityTypeUser,
		Attributes: testAttributes,
		// Affiliation is mandatory. "org1" and "org2" are hardcoded as CA defaults
		// See https://github.com/hyperledger/fabric-ca/blob/release/cmd/fabric-ca-server/config.go
		Affiliation: "org2",
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Enroll the new user
	err = mspClient.Enroll(username, msp.WithSecret(enrollmentSecret))
	if err != nil {
		t.Fatalf("Enroll failed: %v", err)
	}

	// Get the new user's signing identity
	si, err := mspClient.GetSigningIdentity(username)
	if err != nil {
		t.Fatalf("GetSigningIdentity failed: %v", err)
	}

	checkCertAttributes(t, si.EnrollmentCertificate(), testAttributes)

}

func checkCertAttributes(t *testing.T, certBytes []byte, expected []msp.Attribute) {
	decoded, _ := pem.Decode(certBytes)
	if decoded == nil {
		t.Fatalf("Failed cert decoding")
	}
	cert, err := x509.ParseCertificate(decoded.Bytes)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}
	if cert == nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}
	mgr := attrmgr.New()
	attrs, err := mgr.GetAttributesFromCert(cert)
	if err != nil {
		t.Fatalf("Failed to GetAttributesFromCert: %s", err)
	}
	for _, a := range expected {
		v, ok, err := attrs.Value(a.Name)
		assert.NoError(t, err)
		assert.True(t, attrs.Contains(a.Name), "does not contain attribute '%s'", a.Name)
		assert.True(t, ok, "attribute '%s' was not found", a.Name)
		assert.True(t, v == a.Value, "incorrect value for '%s'; expected '%s' but found '%s'", a.Name, a.Value, v)
	}
}

func getRegistrarEnrollmentCredentials(t *testing.T, ctxProvider context.ClientProvider) (string, string) {

	ctx, err := ctxProvider()
	if err != nil {
		t.Fatalf("failed to get context: %v", err)
	}

	clientConfig, err := ctx.IdentityConfig().Client()
	if err != nil {
		t.Fatalf("config.Client() failed: %v", err)
	}

	myOrg := clientConfig.Organization

	caConfig, err := ctx.IdentityConfig().CAConfig(myOrg)
	if err != nil {
		t.Fatalf("CAConfig failed: %v", err)
	}

	return caConfig.Registrar.EnrollID, caConfig.Registrar.EnrollSecret
}
