/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

type event interface{}

type connectionResponse struct {
	err error
}

type connectEvent struct {
	respch chan<- *connectionResponse
}

type disconnectEvent struct {
	respch chan<- *connectionResponse
}

func newConnectEvent(respch chan<- *connectionResponse) *connectEvent {
	return &connectEvent{respch: respch}
}

func newDisconnectEvent(respch chan<- *connectionResponse) *disconnectEvent {
	return &disconnectEvent{respch: respch}
}
