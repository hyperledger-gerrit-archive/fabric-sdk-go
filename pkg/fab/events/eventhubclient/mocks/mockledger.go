/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"

	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// BlockEventFactory creates Block events for the Event Hub
var BlockEventFactory = func(block servicemocks.Block) servicemocks.BlockEvent {
	fmt.Printf("Block: %#v\v", block)
	b, ok := block.(*servicemocks.BlockWrapper)
	if !ok {
		panic(fmt.Sprintf("Invalid block type: %T", block))
	}
	return NewBlockEvent(b.Block())
}

// FilteredBlockEventFactory creates Filtered Block events for the Event Hub
var FilteredBlockEventFactory = func(block servicemocks.Block) servicemocks.BlockEvent {
	fmt.Printf("Block: %#v\v", block)
	b, ok := block.(*servicemocks.FilteredBlockWrapper)
	if !ok {
		panic(fmt.Sprintf("Invalid block type: %T", block))
	}
	return NewFilteredBlockEvent(b.Block())
}

// MockLedger is a mock ledger that stores blocks sequentially
type MockLedger struct {
	servicemocks.MockLedger
}

// NewMockLedger creates a new MockLedger
func NewMockLedger(eventFactory servicemocks.EventFactory) *MockLedger {
	return &MockLedger{
		MockLedger: *servicemocks.NewMockLedger(eventFactory),
	}
}

// NewBlock stores a new block on the ledger
func (l *MockLedger) NewBlock(channelID string, transactions ...*servicemocks.TxInfo) {
	l.Lock()
	defer l.Unlock()
	l.Store(servicemocks.NewBlockWrapper(l.BlockProducer().NewBlock(channelID, transactions...)))
}

// NewFilteredBlock stores a new filtered block on the ledger
func (l *MockLedger) NewFilteredBlock(channelID string, filteredTx ...*pb.FilteredTransaction) {
	l.Lock()
	defer l.Unlock()
	l.Store(servicemocks.NewFilteredBlockWrapper(l.BlockProducer().NewFilteredBlock(channelID, filteredTx...)))
}
