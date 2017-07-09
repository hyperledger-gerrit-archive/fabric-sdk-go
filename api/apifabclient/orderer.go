/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	"github.com/hyperledger/fabric/protos/common"
)

// Orderer The Orderer class represents a node in the target blockchain network from which
// HFC can receive a block of transactions, or send a transaction to be orderered
type Orderer interface {
	Broadcaster
	URL() string
	SendDeliver(envelope *SignedEnvelope) (chan *common.Block, chan error)
}

// Broadcaster represents a node in the target blockchain network
// to which HFC can send a signed envelope
type Broadcaster interface {
	SendBroadcast(envelope *SignedEnvelope) (*common.Status, error)
}

// A SignedEnvelope can can be sent to an orderer for broadcasting
type SignedEnvelope struct {
	Payload   []byte
	Signature []byte
}
