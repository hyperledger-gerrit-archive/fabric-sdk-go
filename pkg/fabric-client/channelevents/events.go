/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

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
