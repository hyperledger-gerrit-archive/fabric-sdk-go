/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// Event is an event that's sent to the dispatcher. This includes client registration
// requests or events that come from an event producer.
type Event interface{}

// RegisterEvent is the base for all registration events.
type RegisterEvent struct {
	RespCh chan<- *apifabclient.RegistrationResponse
}

// StopEvent tells the dispatcher to stop processing
type StopEvent struct {
	RespCh chan<- error
}

// RegisterBlockEvent registers for block events
type RegisterBlockEvent struct {
	RegisterEvent
	Reg *BlockReg
}

// RegisterFilteredBlockEvent registers for filtered block events
type RegisterFilteredBlockEvent struct {
	RegisterEvent
	Reg *FilteredBlockReg
}

// RegisterChaincodeEvent registers for chaincode events
type RegisterChaincodeEvent struct {
	RegisterEvent
	Reg *ChaincodeReg
}

// RegisterTxStatusEvent registers for transaction status events
type RegisterTxStatusEvent struct {
	RegisterEvent
	Reg *TxStatusReg
}

// UnregisterEvent unregisters a registration
type UnregisterEvent struct {
	Reg apifabclient.Registration
}

// NewRegisterBlockEvent creates a new RegisterBlockEvent
func NewRegisterBlockEvent(filter apifabclient.BlockFilter, eventch chan<- *apifabclient.BlockEvent, respch chan<- *apifabclient.RegistrationResponse) *RegisterBlockEvent {
	return &RegisterBlockEvent{
		Reg:           &BlockReg{Filter: filter, Eventch: eventch},
		RegisterEvent: RegisterEvent{RespCh: respch},
	}
}

// NewRegisterFilteredBlockEvent creates a new RegisterFilterBlockEvent
func NewRegisterFilteredBlockEvent(eventch chan<- *apifabclient.FilteredBlockEvent, respch chan<- *apifabclient.RegistrationResponse) *RegisterFilteredBlockEvent {
	return &RegisterFilteredBlockEvent{
		Reg:           &FilteredBlockReg{Eventch: eventch},
		RegisterEvent: RegisterEvent{RespCh: respch},
	}
}

// NewUnregisterEvent creates a new UnregisterEvent
func NewUnregisterEvent(reg apifabclient.Registration) *UnregisterEvent {
	return &UnregisterEvent{
		Reg: reg,
	}
}

// NewRegisterChaincodeEvent creates a new RegisterChaincodeEvent
func NewRegisterChaincodeEvent(ccID, eventFilter string, eventch chan<- *apifabclient.CCEvent, respch chan<- *apifabclient.RegistrationResponse) *RegisterChaincodeEvent {
	return &RegisterChaincodeEvent{
		Reg: &ChaincodeReg{
			ChaincodeID: ccID,
			EventFilter: eventFilter,
			Eventch:     eventch,
		},
		RegisterEvent: RegisterEvent{RespCh: respch},
	}
}

// NewRegisterTxStatusEvent creates a new RegisterTxStatusEvent
func NewRegisterTxStatusEvent(txID string, eventch chan<- *apifabclient.TxStatusEvent, respch chan<- *apifabclient.RegistrationResponse) *RegisterTxStatusEvent {
	return &RegisterTxStatusEvent{
		Reg:           &TxStatusReg{TxID: txID, Eventch: eventch},
		RegisterEvent: RegisterEvent{RespCh: respch},
	}
}

// NewChaincodeEvent creates a new ChaincodeEvent
func NewChaincodeEvent(chaincodeID, eventName, txID string) *apifabclient.CCEvent {
	return &apifabclient.CCEvent{
		ChaincodeID: chaincodeID,
		EventName:   eventName,
		TxID:        txID,
	}
}

// NewTxStatusEvent creates a new TxStatusEvent
func NewTxStatusEvent(txID string, txValidationCode pb.TxValidationCode) *apifabclient.TxStatusEvent {
	return &apifabclient.TxStatusEvent{
		TxID:             txID,
		TxValidationCode: txValidationCode,
	}
}

// NewStopEvent creates a new StopEvent
func NewStopEvent(respch chan<- error) *StopEvent {
	return &StopEvent{
		RespCh: respch,
	}
}

// SuccessResponse returns a success response
func SuccessResponse(reg apifabclient.Registration) *apifabclient.RegistrationResponse {
	return &apifabclient.RegistrationResponse{Reg: reg}
}

// ErrorResponse returns an error response
func ErrorResponse(err error) *apifabclient.RegistrationResponse {
	return &apifabclient.RegistrationResponse{Err: err}
}

// UnregResponse returns an unregister response
func UnregResponse(err error) *apifabclient.RegistrationResponse {
	return &apifabclient.RegistrationResponse{Err: err}
}
