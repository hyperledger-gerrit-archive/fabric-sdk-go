/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabapi

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
)

func TestNewDefaultSDK(t *testing.T) {

	setup := Options{
		ConfigFile: "../../test/fixtures/config/invalid.yaml",
		StateStoreOpts: opt.StateStoreOpts{
			Path: "/tmp/state",
		},
	}

	_, err := NewSDK(setup)
	if err == nil {
		t.Fatalf("Should have failed for invalid config file")
	}

	setup.ConfigFile = "../../test/fixtures/config/config_test.yaml"

	sdk, err := NewSDK(setup)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	_, err = sdk.NewChannelClient("mychannel", "User1")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

}

func TestNewDefaultTwoValidSDK(t *testing.T) {
	setup := Options{
		ConfigFile: "../../test/fixtures/config/config_test.yaml",
		StateStoreOpts: opt.StateStoreOpts{
			Path: "/tmp/state",
		},
	}

	sdk1, err := NewSDK(setup)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	setup.ConfigFile = "../../test/fixtures/config/scenarios/config_org2_test.yaml"
	sdk2, err := NewSDK(setup)
	if err != nil {
		t.Fatalf("Error initializing SDK: %s", err)
	}

	client1, err := sdk1.configProvider.Client()
	if err != nil {
		t.Fatalf("Error getting client from config: %s", err)
	}

	if client1.Organization != "Org1" {
		t.Fatalf("Unexpected org in config: %s", client1.Organization)
	}

	client2, err := sdk2.configProvider.Client()
	if err != nil {
		t.Fatalf("Error getting client from config: %s", err)
	}

	if client2.Organization != "Org2" {
		t.Fatalf("Unexpected org in config: %s", client1.Organization)
	}
}
