/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"regexp"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

type eventType string
type event interface{}

type registerEvent struct {
	respch chan<- *fab.RegistrationResponse
}

type registerChannelEvent struct {
	registerEvent
	eventTypes []eventType
	reg        *channelRegistration
}

type registerFilteredBlockEvent struct {
	registerEvent
	reg *filteredBlockRegistration
}

type registerCCEvent struct {
	registerEvent
	reg *ccRegistration
}

type registerTxStatusEvent struct {
	registerEvent
	reg *txRegistration
}

type connectionResponse struct {
	err error
}

type connectEvent struct {
	respch chan<- *connectionResponse
}

type disconnectEvent struct {
	respch chan<- *connectionResponse
}

type unregisterEvent struct {
	reg fab.Registration
}

type unregisterChannelEvent struct {
	registerEvent
}

func newConnectEvent(respch chan<- *connectionResponse) *connectEvent {
	return &connectEvent{respch: respch}
}

func newDisconnectEvent(respch chan<- *connectionResponse) *disconnectEvent {
	return &disconnectEvent{respch: respch}
}

func newRegisterChannelEvent(eventTypes []eventType, respch chan<- *fab.RegistrationResponse) *registerChannelEvent {
	return &registerChannelEvent{
		reg:           &channelRegistration{},
		registerEvent: registerEvent{respch: respch},
		eventTypes:    eventTypes,
	}
}

func newUnregisterChannelEvent(respch chan<- *fab.RegistrationResponse) *unregisterChannelEvent {
	return &unregisterChannelEvent{
		registerEvent: registerEvent{respch: respch},
	}
}

func newRegisterFilteredBlockEvent(eventch chan<- *fab.FilteredBlockEvent, respch chan<- *fab.RegistrationResponse) *registerFilteredBlockEvent {
	return &registerFilteredBlockEvent{
		reg:           &filteredBlockRegistration{eventch: eventch},
		registerEvent: registerEvent{respch: respch},
	}
}

func newUnregisterEvent(reg fab.Registration) *unregisterEvent {
	return &unregisterEvent{
		reg: reg,
	}
}

func newRegisterCCEvent(ccID, eventFilter string, eventRegExp *regexp.Regexp, eventch chan<- *fab.CCEvent, respch chan<- *fab.RegistrationResponse) *registerCCEvent {
	return &registerCCEvent{
		reg: &ccRegistration{
			ccID:        ccID,
			eventFilter: eventFilter,
			eventRegExp: eventRegExp,
			eventch:     eventch,
		},
		registerEvent: registerEvent{respch: respch},
	}
}

func newCCEvent(chaincodeID, eventName, txID string) *fab.CCEvent {
	return &fab.CCEvent{
		ChaincodeID: chaincodeID,
		EventName:   eventName,
		TxID:        txID,
	}
}

func newRegisterTxStatusEvent(txID string, eventch chan<- *fab.TxStatusEvent, respch chan<- *fab.RegistrationResponse) *registerTxStatusEvent {
	return &registerTxStatusEvent{
		reg:           &txRegistration{txID: txID, eventch: eventch},
		registerEvent: registerEvent{respch: respch},
	}
}

func newTxStatusEvent(txID string, txValidationCode pb.TxValidationCode) *fab.TxStatusEvent {
	return &fab.TxStatusEvent{
		TxID:             txID,
		TxValidationCode: txValidationCode,
	}
}
