// +build prerelease

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"testing"
	"time"

	evclient "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client"
	clientdisp "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient"
	ehclient "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/eventhubclient"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

func TestEventHubClientFilteredBlockEvents(t *testing.T) {
	chainCodeID := integration.GenerateRandomID()
	testSetup := initializeTests(t, chainCodeID)

	discoveryService, err := testSetup.SDK.Context().DiscoveryProvider().NewDiscoveryService(testSetup.ChannelID)
	if err != nil {
		t.Fatalf("Error creating discovery service: %s", err)
	}

	discoveryService = &eventHubDiscoveryService{
		target: discoveryService,
	}

	ctx := events.Context{
		ProviderContext: testSetup.SDK.Context(),
		IdentityContext: testSetup.Identity,
	}

	client, err := ehclient.New(ctx, channelID, discoveryService,
		evclient.WithResponseTimeout(5*time.Second),
		evclient.WithMaxReconnectAttempts(1),
	)
	if err != nil {
		t.Fatalf("Error creating event hub client: %s", err)
	}
	if err := client.Connect(); err != nil {
		t.Fatalf("Error connecting event hub client: %s", err)
	}
	defer client.Close()

	testEventClient(t, testSetup, chainCodeID, false, client)
}

func TestDeliverClient(t *testing.T) {
	chainCodeID := integration.GenerateRandomID()
	testSetup := initializeTests(t, chainCodeID)

	discoveryService, err := testSetup.SDK.Context().DiscoveryProvider().NewDiscoveryService(testSetup.ChannelID)
	if err != nil {
		t.Fatalf("Error creating discovery service: %s", err)
	}

	ctx := events.Context{
		ProviderContext: testSetup.SDK.Context(),
		IdentityContext: testSetup.Identity,
	}

	t.Run("Filtered Block Events", func(t *testing.T) {
		client, err := deliverclient.New(ctx, channelID, discoveryService,
			evclient.WithMaxConnectAttempts(1),
			evclient.WithMaxReconnectAttempts(1),
		)
		if err != nil {
			t.Fatalf("Error creating deliver client: %s", err)
		}
		if err := client.Connect(); err != nil {
			t.Fatalf("Error connecting deliver client: %s", err)
		}
		defer client.Close()

		testEventClient(t, testSetup, chainCodeID, false, client)
	})

	t.Run("Block Events", func(t *testing.T) {
		client, err := deliverclient.New(ctx, channelID, discoveryService,
			deliverclient.WithBlockEvents(),
			evclient.WithMaxConnectAttempts(1),
			evclient.WithMaxReconnectAttempts(1),
		)
		if err != nil {
			t.Fatalf("Error creating deliver client: %s", err)
		}
		if err := client.Connect(); err != nil {
			t.Fatalf("Error connecting deliver client: %s", err)
		}
		defer client.Close()

		testEventClient(t, testSetup, chainCodeID, true, client)
	})
}

func TestDeliverClientForbidden(t *testing.T) {
	chainCodeID := integration.GenerateRandomID()
	testSetup := initializeTests(t, chainCodeID)

	discoveryService, err := testSetup.SDK.Context().DiscoveryProvider().NewDiscoveryService(testSetup.ChannelID)
	if err != nil {
		t.Fatalf("Error creating discovery service: %s", err)
	}

	ctx := events.Context{
		ProviderContext: testSetup.SDK.Context(),
		IdentityContext: &identityContext{
			mspID:      "invalid",
			identity:   []byte("invalid"),
			privateKey: testSetup.Identity.PrivateKey(),
		},
	}

	conneventch := make(chan *clientdisp.ConnectionEvent)
	var client fab.EventClient
	client, err = deliverclient.New(ctx, channelID, discoveryService,
		deliverclient.WithBlockEvents(),
		evclient.WithConnectionEvent(conneventch),
	)
	if err != nil {
		t.Fatalf("Error creating deliver client: %s", err)
	}
	if err := client.Connect(); err != nil {
		t.Fatalf("Error connecting deliver client: %s", err)
	}
	defer client.Close()

	for {
		select {
		case event := <-conneventch:
			if event.Connected {
				t.Logf("Got connected event")
			} else {
				t.Logf("Got disconnected event with error [%s]", event.Err)
				return
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("Timeout waiting for disconnected event")
		}
	}
}

type identityContext struct {
	mspID      string
	identity   []byte
	privateKey core.Key
}

func (c *identityContext) MspID() string {
	return c.mspID
}

func (c *identityContext) Identity() ([]byte, error) {
	return c.identity, nil
}

func (c *identityContext) PrivateKey() core.Key {
	return c.privateKey
}
