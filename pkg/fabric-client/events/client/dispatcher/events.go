/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/service/dispatcher"
)

// RegisterConnectionEvent is a request to register for connection events
type RegisterConnectionEvent struct {
	esdispatcher.RegisterEvent
	Reg *ConnectionReg
}

// NewRegisterConnectionEvent creates a new RegisterConnectionEvent
func NewRegisterConnectionEvent(eventch chan<- *apifabclient.ConnectionEvent, respch chan<- *apifabclient.RegistrationResponse) *RegisterConnectionEvent {
	return &RegisterConnectionEvent{
		Reg:           &ConnectionReg{Eventch: eventch},
		RegisterEvent: esdispatcher.RegisterEvent{RespCh: respch},
	}
}

// ConnectedEvent indicates that the client has connected to the server
type ConnectedEvent struct {
}

// NewConnectedEvent creates a new ConnectedEvent
func NewConnectedEvent() *ConnectedEvent {
	return &ConnectedEvent{}
}

// DisconnectedEvent indicates that the client has disconnected from the server
type DisconnectedEvent struct {
	Err error
}

// NewDisconnectedEvent creates a new DisconnectedEvent
func NewDisconnectedEvent(err error) *DisconnectedEvent {
	return &DisconnectedEvent{Err: err}
}

// ConnectEvent is a request to connect to the server
type ConnectEvent struct {
	Respch       chan<- *ConnectionResponse
	FromBlockNum uint64
}

// NewConnectEvent creates a new ConnectEvent
func NewConnectEvent(respch chan<- *ConnectionResponse) *ConnectEvent {
	return &ConnectEvent{Respch: respch}
}

// ConnectionResponse is the response of the connect request
type ConnectionResponse struct {
	Err error
}

// DisconnectEvent is a request to disconnect to the server
type DisconnectEvent struct {
	Respch chan<- *ConnectionResponse
}

// NewDisconnectEvent creates a new DisconnectEvent
func NewDisconnectEvent(respch chan<- *ConnectionResponse) *DisconnectEvent {
	return &DisconnectEvent{Respch: respch}
}
