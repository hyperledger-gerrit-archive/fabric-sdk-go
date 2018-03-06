/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

const (
	IdentityTypeUser = "User"

	org1Name = "Org1"
	org2Name = "Org2"
)

func TestRegisterEnroll(t *testing.T) {

	configProvider := config.FromFile(sdkConfigFile)

	// Instantiate the SDK
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		t.Fatalf("SDK init failed: %v", err)
	}

	// Get the Enrollment Service
	// It is used by registered users who need to
	// enroll with the org's CA.
	es, err := sdk.NewEnrollmentService(org2Name)
	if err != nil {
		t.Fatalf("NewEnrollmentService failed: %v", err)
	}

	// Get the Registrar Service
	// It is used by the org's CA registrars, to
	// perform user management functions on the CA.
	rs, err := sdk.NewRegistrarService(org2Name)
	if err != nil {
		t.Fatalf("NewRegistrarService failed: %v", err)
	}

	// Get the Identity Manager
	// It is used by the application to retrieve
	// enrolled users' identities.
	im, err := sdk.NewIdentityManager(org2Name)
	if err != nil {
		t.Fatalf("NewIdentityManager failed: %v", err)
	}

	// As this integration test spawns a fresh CA instance,
	// we have to enroll the CA registrar first. Otherwise,
	// CA operations that require the registrar's identity
	// will be rejected by the CA.
	registrarEnrollID, registrarEnrollSecret := getRegistrarEnrollmentCredentials(t, configProvider)
	err = es.Enroll(registrarEnrollID, registrarEnrollSecret)
	if err != nil {
		t.Fatalf("Enroll failed: %v", err)
	}

	// The enrollment process generates a new private key and
	// enrollment certificate for the user. The private key
	// is stored in the SDK crypto provider's key store, while the
	// enrollment certificate is stored in the SKD's user store
	// (state store). The Registrar Service will lookup the
	// registrar's identity information in these stores.

	// Generate a random user name
	userName := integration.GenerateRandomID()

	// Register the new user
	enrollmentSecret, err := rs.Register(&identity.RegistrationRequest{
		Name: userName,
		Type: IdentityTypeUser,
		// Affiliation is mandatory. "org1" and "org2" are hardcoded as CA defaults
		// See https://github.com/hyperledger/fabric-ca/blob/release/cmd/fabric-ca-server/config.go
		Affiliation: "org2",
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Enroll the new user
	err = es.Enroll(userName, enrollmentSecret)
	if err != nil {
		t.Fatalf("Enroll failed: %v", err)
	}

	// Get the new user's signing identity
	_, err = im.GetSigningIdentity(userName)
	if err != nil {
		t.Fatalf("GetSigningIdentity failed: %v", err)
	}

	// Get the new user's full information
	_, err = im.GetUser(userName)
	if err != nil {
		t.Fatalf("GetSigningIdentity failed: %v", err)
	}

}

func getRegistrarEnrollmentCredentials(t *testing.T, configProvider core.ConfigProvider) (string, string) {
	config, err := configProvider()
	if err != nil {
		t.Fatalf("configProvider failed: %v", err)
	}
	caConfig, err := config.CAConfig(org2Name)
	if err != nil {
		t.Fatalf("CAConfig failed: %v", err)
	}
	return caConfig.Registrar.EnrollID, caConfig.Registrar.EnrollSecret
}
