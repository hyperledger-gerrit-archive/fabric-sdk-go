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

// ChannelEventClient is a client that connects to a peer and receives channel events,
// such as filtered block, chaincode, and transaction status events.
type ChannelEventClient interface {
	// Connect connects to the peer and registers for channel events on a particular channel.
	Connect() error

	// Disconnect disconnects from the peer. Once this function is invoked the client may no longer be used.
	Disconnect()
}
