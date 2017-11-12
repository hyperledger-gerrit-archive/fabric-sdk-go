/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

// ConnectionEvent is sent when the client disconnects or reconnects to the peer
type ConnectionEvent struct {
	Connected bool
	Err       error
}

// Registration is a handle that is returned from a successful RegisterXXXEvent.
// This handle should be used in Unregister in order to unregister the event.
type Registration interface{}

// RegistrationResponse is the response that is returned for any register/unregister event.
// For a successful registration, the registration handle is set. This handle should be used
// in a subsequent Unregister request. If an error occurs then the error is set.
type RegistrationResponse struct {
	// Reg is a handle to the registration
	Reg Registration

	// Err contains the error if registration is unsuccessful
	Err error
}

// ChannelEventClient is a client that connects to a peer and receives channel events,
// such as filtered block, chaincode, and transaction status events.
type ChannelEventClient interface {
	// Connect connects to the peer and registers for channel events on a particular channel.
	Connect() error

	// Disconnect disconnects from the peer. Once this function is invoked the client may no longer be used.
	Disconnect()
}
