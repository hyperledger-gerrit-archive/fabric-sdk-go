/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events"
	evclient "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/endpoint"
	ehclient "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/eventhubclient"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

func TestEventHubClientBlockEvents(t *testing.T) {
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
		ehclient.WithBlockEvents(),
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

	testEventClient(t, testSetup, chainCodeID, true, client)
}

func testEventClient(t *testing.T, testSetup integration.BaseSetupImpl, chainCodeID string, blockEvents bool, client fab.EventClient) {
	// Invoke the chaincode before registering for events to ensure that the chaincode has been instantiated;
	// otherwise we may receive a block/filtered block event from the instantiate (and our test will fail due to too many events)
	tpResponses, prop, err := integration.CreateAndSendTransactionProposal(
		testSetup.Transactor,
		chainCodeID,
		"invoke",
		[][]byte{
			[]byte("invoke"),
			[]byte("SEVERE"),
		},
		testSetup.Targets,
		nil,
	)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v", err)
	}

	txID := string(prop.TxnID)

	var wg sync.WaitGroup
	var numExpected uint32

	var breg fab.Registration
	var beventch <-chan *fab.BlockEvent
	if blockEvents {
		breg, beventch, err = client.RegisterBlockEvent()
		if err != nil {
			t.Fatalf("Error registering for block events: %s", err)
		}
		defer client.Unregister(breg)
		numExpected++
		wg.Add(1)
	}

	fbreg, fbeventch, err := client.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("Error registering for filtered block events: %s", err)
	}
	defer client.Unregister(fbreg)
	numExpected++
	wg.Add(1)

	ccreg, cceventch, err := client.RegisterChaincodeEvent(chainCodeID, ".*")
	if err != nil {
		t.Fatalf("Error registering for filtered block events: %s", err)
	}
	defer client.Unregister(ccreg)
	numExpected++
	wg.Add(1)

	txReg, txstatusch, err := client.RegisterTxStatusEvent(txID)
	if err != nil {
		t.Fatalf("Error registering for Tx Status event: %s", err)
	}
	defer client.Unregister(txReg)
	numExpected++
	wg.Add(1)

	var numReceived uint32

	if beventch != nil {
		go func() {
			defer wg.Done()
			select {
			case event, ok := <-beventch:
				if !ok {
					t.Fatalf("unexpected closed channel while waiting for Tx Status event")
				}
				t.Logf("Received block event: %#v", event)
				if event.Block == nil {
					t.Fatalf("Expecting block in block event but got nil")
				}
				atomic.AddUint32(&numReceived, 1)
			case <-time.After(5 * time.Second):
			}
		}()
	}

	go func() {
		defer wg.Done()
		select {
		case event, ok := <-fbeventch:
			if !ok {
				t.Fatalf("unexpected closed channel while waiting for Tx Status event")
			}
			t.Logf("Received filtered block event: %#v", event)
			if event.FilteredBlock == nil || len(event.FilteredBlock.FilteredTx) == 0 {
				t.Fatalf("Expecting one transaction in filtered block but got none")
			}
			filteredTx := event.FilteredBlock.FilteredTx[0]
			if filteredTx.Txid != txID {
				t.Fatalf("Expecting filtered transaction to contain TxID [%s] but got TxID [%s]", txID, filteredTx.Txid)
			}
			atomic.AddUint32(&numReceived, 1)
		case <-time.After(5 * time.Second):
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case event, ok := <-cceventch:
			if !ok {
				t.Fatalf("unexpected closed channel while waiting for Tx Status event")
			}
			t.Logf("Received chaincode event: %#v", event)
			if event.ChaincodeID != chainCodeID {
				t.Fatalf("Expecting event for CC ID [%s] but got event for CC ID [%s]", chainCodeID, event.ChaincodeID)
			}
			atomic.AddUint32(&numReceived, 1)
		case <-time.After(5 * time.Second):
			return
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case txStatus, ok := <-txstatusch:
			if !ok {
				t.Fatalf("unexpected closed channel while waiting for Tx Status event")
			}
			t.Logf("Received Tx Status event: %#v", txStatus)
			if txStatus.TxID != txID {
				t.Fatalf("Expecting event for TxID [%s] but got event for TxID [%s]", txID, txStatus.TxID)
			}
			atomic.AddUint32(&numReceived, 1)
		case <-time.After(5 * time.Second):
			return
		}
	}()

	// Commit the transaction to generate events
	_, err = integration.CreateAndSendTransaction(testSetup.Transactor, prop, tpResponses)
	if err != nil {
		t.Fatalf("First invoke failed err: %v", err)
	}

	wg.Wait()

	if numReceived != numExpected {
		t.Fatalf("expecting %d events but received %d", numExpected, numReceived)
	}
}

type eventHubDiscoveryService struct {
	target fab.DiscoveryService
}

func (s *eventHubDiscoveryService) GetPeers() ([]fab.Peer, error) {
	var eventEndpoints []fab.Peer

	peers, err := s.target.GetPeers()
	if err != nil {
		return nil, err
	}

	for _, peer := range peers {
		eventURL, err := s.getEventURL(peer)
		if err != nil {
			return nil, err
		}
		eventEndpoints = append(eventEndpoints,
			&endpoint.EventEndpoint{
				Peer:   peer,
				EvtURL: eventURL,
			},
		)
	}

	return eventEndpoints, nil
}

func (s *eventHubDiscoveryService) getEventURL(peer fab.Peer) (string, error) {
	url := peer.URL()
	i := strings.LastIndex(url, ":")
	if i < 0 {
		return "", errors.Errorf("Invalid peer URL: %s", url)
	}
	// FIXME: The eventhub port should come from config once
	// config has reg-exp matching on peer URL. Hard code it for now.
	return fmt.Sprintf("%s:%d", url[0:i], 7053), nil
}
