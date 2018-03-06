/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ca"
	caapi "github.com/hyperledger/fabric-sdk-go/pkg/context/api/ca"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

const (
	IdentityTypeUser = "User"
)

func TestRegisterEnroll(t *testing.T) {

	configProvider := config.FromFile(sdkConfigFile)

	myOrg, myMspID := getMyOrgInfo(t, configProvider)

	// Instantiate the SDK
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		t.Fatalf("SDK init failed: %v", err)
	}

	ctx := sdk.AnonymousContext()

	// Get the CA Client
	ca, err := ca.New(ctx)
	if err != nil {
		t.Fatalf("failed to create CA client: %v", err)
	}

	// Get the Identity Manager
	// It is used by the application to retrieve
	// enrolled users' identities.
	im := ctx.IdentityManager()
	if err != nil {
		t.Fatalf("New IdentityManager failed: %v", err)
	}

	// As this integration test spawns a fresh CA instance,
	// we have to enroll the CA registrar first. Otherwise,
	// CA operations that require the registrar's identity
	// will be rejected by the CA.
	registrarEnrollID, registrarEnrollSecret := getRegistrarEnrollmentCredentials(t, myOrg, configProvider)
	err = ca.Enroll(registrarEnrollID, registrarEnrollSecret)
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
	enrollmentSecret, err := ca.Register(&caapi.RegistrationRequest{
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
	err = ca.Enroll(userName, enrollmentSecret)
	if err != nil {
		t.Fatalf("Enroll failed: %v", err)
	}

	// Get the new user's signing identity
	_, err = im.GetSigningIdentity(myMspID, userName)
	if err != nil {
		t.Fatalf("GetSigningIdentity failed: %v", err)
	}

	// Get the new user's full information
	_, err = im.GetUser(myMspID, userName)
	if err != nil {
		t.Fatalf("GetSigningIdentity failed: %v", err)
	}

}

func getMyOrgInfo(t *testing.T, configProvider core.ConfigProvider) (string, string) {

	config, err := configProvider()
	if err != nil {
		t.Fatalf("configProvider() failed: %v", err)
	}

	clientConfig, err := config.Client()
	if err != nil {
		t.Fatalf("config.Client() failed: %v", err)
	}

	myOrg := clientConfig.Organization

	myMspID, err := config.MspID(myOrg)
	if err != nil {
		t.Fatalf("config.MspID(myOrg) failed: %v", err)
	}

	return myOrg, myMspID
}

func getRegistrarEnrollmentCredentials(t *testing.T, orgName string, configProvider core.ConfigProvider) (string, string) {

	config, err := configProvider()
	if err != nil {
		t.Fatalf("configProvider failed: %v", err)
	}

	caConfig, err := config.CAConfig(orgName)
	if err != nil {
		t.Fatalf("CAConfig failed: %v", err)
	}

	return caConfig.Registrar.EnrollID, caConfig.Registrar.EnrollSecret
}
