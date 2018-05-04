// +build !prev

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"testing"

	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

func TestIdentity(t *testing.T) {

	mspClient := setupClient(t)

	// Generate a random user name
	username := integration.GenerateRandomID()

	testAttributes := []msp.Attribute{
		{
			Name:  integration.GenerateRandomID(),
			Value: fmt.Sprintf("%s:ecert", integration.GenerateRandomID()),
			ECert: true,
		},
		{
			Name:  integration.GenerateRandomID(),
			Value: fmt.Sprintf("%s:ecert", integration.GenerateRandomID()),
			ECert: true,
		},
	}

	req := &msp.IdentityRequest{
		ID:          username,
		Affiliation: "org2",
		Type:        IdentityTypeUser,
		Attributes:  testAttributes,
		// Affiliation and ID are mandatory. "org1" and "org2" are hardcoded as CA defaults
		// See https://github.com/hyperledger/fabric-ca/blob/release/cmd/fabric-ca-server/config.go
	}

	// Create new identity
	newIdentity, err := mspClient.CreateIdentity(req)
	if err != nil {
		t.Fatalf("Create identity failed: %v", err)
	}

	if newIdentity.Secret == "" {
		t.Fatal("Secret should have been generated")
	}

	identity, err := mspClient.GetIdentity(username)
	if err != nil {
		t.Fatalf("get identity failed: %v", err)
	}

	t.Logf("Get Identity: [%v]:", identity)

	if !verifyIdentity(req, identity) {
		t.Fatalf("verify identity failed req=[%v]; resp=[%v] ", req, identity)
	}

	// Enroll the new user
	err = mspClient.Enroll(username, msp.WithSecret(newIdentity.Secret))
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

func TestUpdateIdentity(t *testing.T) {

	mspClient := setupClient(t)

	// Generate a random user name
	username := integration.GenerateRandomID()

	testAttributes := []msp.Attribute{
		{
			Name:  integration.GenerateRandomID(),
			Value: fmt.Sprintf("%s:ecert", integration.GenerateRandomID()),
			ECert: true,
		},
		{
			Name:  integration.GenerateRandomID(),
			Value: fmt.Sprintf("%s:ecert", integration.GenerateRandomID()),
			ECert: true,
		},
	}

	req := &msp.IdentityRequest{
		ID:          username,
		Affiliation: "org2",
		Type:        IdentityTypeUser,
		Attributes:  testAttributes,
		// Affiliation and ID are mandatory. "org1" and "org2" are hardcoded as CA defaults
		// See https://github.com/hyperledger/fabric-ca/blob/release/cmd/fabric-ca-server/config.go
	}

	// Create new identity
	newIdentity, err := mspClient.CreateIdentity(req)
	if err != nil {
		t.Fatalf("Create identity failed: %v", err)
	}

	// Update secret
	req.Secret = "top-secret"

	identity, err := mspClient.ModifyIdentity(req)
	if err != nil {
		t.Fatalf("modify identity failed: %v", err)
	}

	if identity.Secret != "top-secret" {
		t.Fatalf("update identity failed: %v", err)
	}

	// Enroll the new user with old secret
	err = mspClient.Enroll(username, msp.WithSecret(newIdentity.Secret))
	if err == nil {
		t.Fatalf("Enroll should have failed since secret has been updated")
	}

	// Enroll the new user with updated secret
	err = mspClient.Enroll(username, msp.WithSecret(identity.Secret))
	if err != nil {
		t.Fatalf("Enroll failed: %v", err)
	}

	removed, err := mspClient.RemoveIdentity(&msp.RemoveIdentityRequest{ID: username})
	if err != nil {
		t.Fatalf("remove identity failed: %v", err)
	}

	t.Logf("Removed identity [%v]", removed)

	// Test enroll with deleted identity
	err = mspClient.Enroll(username, msp.WithSecret(identity.Secret))
	if err == nil {
		t.Fatal("Enroll should have failed since identity has been deleted")
	}
}

func setupClient(t *testing.T) *msp.Client {

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

	return mspClient

}

func verifyIdentity(req *msp.IdentityRequest, res *msp.IdentityResponse) bool {
	if req.Affiliation != res.Affiliation || req.Type != res.Type {
		return false
	}

	// TODO: Verify attributes
	return true
}
