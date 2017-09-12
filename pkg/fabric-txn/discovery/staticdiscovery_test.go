/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

func TestDiscovery(t *testing.T) {

	testChannel, err := setupTestChannel()
	if err != nil {
		t.Fatalf("Failed to setup test channel: %s", err)
	}

	discoveryProvider, err := NewDiscoveryProvider(&fcmocks.MockConfig{})
	if err != nil {
		t.Fatalf("Failed to  setup discovery provider: %s", err)
	}

	discoveryService, err := discoveryProvider.NewDiscoveryService(testChannel)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	peers, err := discoveryService.GetPeers("testCC")
	if err != nil {
		t.Fatalf("Failed to get peers from discovery service: %s", err)
	}

	// TODO: This will change when channel configuration is added (new config)
	expectedNumOfPeeers := 0
	if len(peers) != expectedNumOfPeeers {
		t.Fatalf("Expecting %d, got %d peers", expectedNumOfPeeers, len(peers))
	}

}

func setupTestChannel() (*channel.Channel, error) {
	client := setupTestClient()
	return channel.NewChannel("testChannel", client)
}

func setupTestClient() *fcmocks.MockClient {
	client := fcmocks.NewMockClient()
	user := fcmocks.NewMockUser("test")
	cryptoSuite := &fcmocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetUserContext(user)
	client.SetCryptoSuite(cryptoSuite)
	return client
}
