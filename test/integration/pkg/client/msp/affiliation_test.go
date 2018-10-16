/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

func TestAffiliation(t *testing.T) {
	mspClient, sdk := setupClient(t)
	defer integration.CleanupUserData(t, sdk)

	req := &msp.AffiliationRequest{
		Name:   "com.example",
		Force:  true,
		CAName: "ca.org1.example.com",
	}

	addResp, err := mspClient.AddAffiliation(req)
	if err != nil {
		t.Fatalf("Add affiliation failed: %s", err)
	}

	if addResp.Name == "" {
		t.Fatalf("Name should not be empty")
	}

	affiliation, err := mspClient.GetAffiliation(req.Name)
	if err != nil {
		t.Fatalf("Get affiliation failed: %s", err)
	}

	t.Logf("Get Affiliation: [%v]", affiliation)

	if affiliation.Name != req.Name {
		t.Fatalf("verify affiliation name failed req.Name=[%s]; resp.Name=[%s]", req.Name, affiliation.Name)
	}

	modifyResp, err := mspClient.ModifyAffiliation(&msp.ModifyAffiliationRequest{AffiliationRequest: req, NewName: "org.example"})
	if err != nil {
		t.Fatalf("Modify affiliation failed: %s", err)
	}
	if modifyResp.Name != "org.example" {
		t.Fatal("New name should be org.example")
	}

	removeResp, err := mspClient.RemoveAffiliation(&msp.AffiliationRequest{Name: "org.example", Force: true})
	if err != nil {
		t.Fatalf("Remove affiliation failed: %s", err)
	}
	if removeResp.Name != "org.example" {
		t.Fatal("Name should be org.example")
	}

	_, err = mspClient.GetAffiliation("org.example")
	if err == nil || !strings.Contains(err.Error(), "no rows in result set") {
		t.Fatal("Should have failed to get affiliation due to missing name org.example")
	}
}

func TestGetAllAffiliations(t *testing.T) {
	mspClient, sdk := setupClient(t)
	defer integration.CleanupUserData(t, sdk)

	req1 := &msp.AffiliationRequest{Name: "org.org1.hyperledger", Force: true, CAName: "ca.org1.example.com"}
	affiliation, err := mspClient.AddAffiliation(req1)
	if err != nil {
		t.Fatalf("Add affiliation failed: %s", err)
	}
	t.Logf("First affiliation created: [%v]", affiliation)

	req2 := &msp.AffiliationRequest{Name: "org.org2.hyperledger", Force: true, CAName: "ca.org1.example.com"}
	affiliation, err = mspClient.AddAffiliation(req2)
	if err != nil {
		t.Fatalf("Add affiliation failed: %s", err)
	}
	t.Logf("Second affiliation created: [%v]", affiliation)

	affiliations, err := mspClient.GetAllAffiliations(msp.WithCA("ca.org1.example.com"))
	if err != nil {
		t.Fatalf("Retrieve affiliations failed: %s", err)
	}

	results := make([]string, 0)
	for _, affi := range affiliations.Affiliations {
		t.Logf("Affiliation: %v", affi)
		if affi.Name == "org" {
			for _, affii := range affi.Affiliations {
				if affii.Name == "org.org1" || affii.Name == "org.org2" {
					for _, affiii := range affii.Affiliations {
						results = append(results, affiii.Name)
					}
				}
			}
		}
	}

	if len(results) != 2 {
		t.Fatal("There should be 2 organizations")
	}
	if !containsAffiliations(results, "org.org1.hyperledger", "org.org2.hyperledger") {
		t.Fatal("Unable to retrieve newly added affiliation")
	}
}

func containsAffiliations(affiliations []string, requests ...string) bool {
	for _, request := range requests {
		if !containsAffiliation(affiliations, request) {
			return false
		}
	}
	return true
}

func containsAffiliation(affiliations []string, reqeust string) bool {
	for _, affiliation := range affiliations {
		if affiliation == reqeust {
			return true
		}
	}
	return false
}
