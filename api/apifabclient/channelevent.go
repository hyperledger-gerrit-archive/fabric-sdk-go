/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// FilteredBlockEvent contains the data for a filtered block event
type FilteredBlockEvent struct {
	FilteredBlock *pb.FilteredBlock
}

// TxStatusEvent contains the data for a transaction status event
type TxStatusEvent struct {
	TxID             string
	TxValidationCode pb.TxValidationCode
}

// CCEvent contains the data for a chaincocde event
type CCEvent struct {
	TxID        string
	ChaincodeID string
	EventName   string
}

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

	// RegisterFilteredBlockEvent registers for filtered block events. If the client is not authorized to receive
	// filtered block events then an error is returned.
	// - Returns the registration and a channel that is used to receive events
	// NOTE: It is recommended that the event be handled in a separate Go routine so as not to block other events.
	RegisterFilteredBlockEvent() (Registration, <-chan *FilteredBlockEvent, error)

	// RegisterChaincodeEvent registers for chaincode events. If the client is not authorized to receive
	// chaincode events then an error is returned.
	// - ccID is the chaincode ID for which events are to be received
	// - eventFilter is the chaincode event filter (regular expression) for which events are to be received
	RegisterChaincodeEvent(ccID, eventFilter string) (Registration, <-chan *CCEvent, error)

	// RegisterTxStatusEvent registers for transaction status events. If the client is not authorized to receive
	// transaction status events then an error is returned.
	// - txID is the transaction ID for which events are to be received
	RegisterTxStatusEvent(txID string) (Registration, <-chan *TxStatusEvent, error)

	// Unregister unregisters the given registration.
	// - reg is the registration handle that was returned from one of the RegisterXXX functions
	Unregister(reg Registration)
}
