/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

type event interface{}

type registerEvent struct {
	respch chan<- *fab.RegistrationResponse
}

type registerChannelEvent struct {
	registerEvent
	reg *channelRegistration
}

type registerBlockEvent struct {
	registerEvent
	reg *blockRegistration
}

type registerCCEvent struct {
	registerEvent
	reg *ccRegistration
}

type registerTxStatusEvent struct {
	registerEvent
	reg *txRegistration
}

type registerConnectionEvent struct {
	registerEvent
	reg *connectionRegistration
}

type disconnectedEvent struct {
	err error
}

type connectedEvent struct {
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
	registerEvent
	reg fab.Registration
}

type unregisterChannelEvent struct {
	registerEvent
}
