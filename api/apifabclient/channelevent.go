/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// BlockEvent contains the data for the block event
type BlockEvent struct {
	Block *common.Block
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

// ChannelEventClient is a client that is used to connect to a peer and receive channel events,
// such as block, chaincode, and transaction status events.
type ChannelEventClient interface {
	// Connect connects to the peer and registers for channel events on a particular channel.
	Connect() error

	// Disconnect disconnects from the peer. Once this function is invoke the client may no longer be used.
	Disconnect()

	// RegisterBlockEvent registers for block events. If the client is not authorized to receive
	// block events then an error is returned.
	// - eventch is the Go channel to which events are sent. Note that the events should be processed
	//         in a separate Go routine so that the event dispatcher is not blocked.
	// Example:
	//
	RegisterBlockEvent(eventch chan<- *BlockEvent) (Registration, error)

	// RegisterChaincodeEvent registers for chaincode events. If the client is not authorized to receive
	// chaincode events then an error is returned.
	// - ccID is the chaincode ID for which events are to be received
	// - eventFilter is the chaincode event name for which events are to be received
	// - eventch is the Go channel to which events are sent. Note that the events should be processed
	//         in a separate Go routine so that the event dispatcher is not blocked.
	RegisterChaincodeEvent(ccID, eventFilter string, eventch chan<- *CCEvent) (Registration, error)

	// RegisterTxStatusEvent registers for transaction status events. If the client is not authorized to receive
	// transaction status events then an error is returned.
	// - txID is the transaction ID for which events are to be received
	// - eventch is the Go channel to which events are sent. Note that the events should be processed in a separate Go routine so that the event dispatcher is not blocked.
	RegisterTxStatusEvent(txID string, eventch chan<- *TxStatusEvent) (Registration, error)

	// Unregister unregisters the given registration.
	// - reg is the registration handle that was returned from one of the RegisterXXX functions
	Unregister(reg Registration) error
}
