/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"
	"sync"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// Block is n abstract block
type Block interface {
	Number() uint64
	SetNumber(blockNum uint64)
}

// BlockEvent is an abstract event
type BlockEvent interface{}

// Consumer is a consumer of a BlockEvent
type Consumer chan interface{}

// EventFactory creates block events
type EventFactory func(block Block) BlockEvent

// BlockEventFactory creates block events
var BlockEventFactory = func(block Block) BlockEvent {
	b, ok := block.(*BlockWrapper)
	if !ok {
		panic(fmt.Sprintf("Invalid block type: %T", block))
	}
	return b.Block()
}

// FilteredBlockEventFactory creates filtered block events
var FilteredBlockEventFactory = func(block Block) BlockEvent {
	b, ok := block.(*FilteredBlockWrapper)
	if !ok {
		panic(fmt.Sprintf("Invalid block type: %T", block))
	}
	return b.Block()
}

// MockLedger is a mock ledger that stores blocks sequentially
type MockLedger struct {
	sync.RWMutex
	blockProducer *BlockProducer
	consumers     []Consumer
	blocks        []Block
	eventFactory  EventFactory
}

// NewMockLedger creates a new MockLedger
func NewMockLedger(eventFactory EventFactory) *MockLedger {
	return &MockLedger{
		eventFactory:  eventFactory,
		blockProducer: NewBlockProducer(),
	}
}

// BlockProducer returns the block producer
func (l *MockLedger) BlockProducer() *BlockProducer {
	return l.blockProducer
}

// Register registers an event consumer
func (l *MockLedger) Register(consumer Consumer) {
	fmt.Printf("MockLedger - Registering consumer\n")
	l.Lock()
	defer l.Unlock()
	l.consumers = append(l.consumers, consumer)
}

// Unregister unregisters the given consumer
func (l *MockLedger) Unregister(Consumer Consumer) {
	l.Lock()
	defer l.Unlock()

	for i, p := range l.consumers {
		if p == Consumer {
			fmt.Printf("MockLedger - Unregistering consumer %d\n", i)
			if i != 0 {
				l.consumers = l.consumers[1:]
			}
			l.consumers = l.consumers[1:]
			break
		}
	}
}

// NewBlock stores a new block on the ledger
func (l *MockLedger) NewBlock(channelID string, transactions ...*TxInfo) {
	l.Lock()
	defer l.Unlock()
	l.Store(NewBlockWrapper(l.blockProducer.NewBlock(channelID, transactions...)))
}

// NewFilteredBlock stores a new filtered block on the ledger
func (l *MockLedger) NewFilteredBlock(channelID string, filteredTx ...*pb.FilteredTransaction) {
	l.Lock()
	defer l.Unlock()
	l.Store(NewFilteredBlockWrapper(l.blockProducer.NewFilteredBlock(channelID, filteredTx...)))
}

// Store stores the given block to the ledger
func (l *MockLedger) Store(block Block) {
	fmt.Printf("MockLedger - Storing block #%d\n", block.Number())
	l.blocks = append(l.blocks, block)

	for i, p := range l.consumers {
		blockEvent := l.eventFactory(block)
		fmt.Printf("MockLedger - Notifying Consumer %d with block %d - Block Event: %#v\n", i, block.Number(), blockEvent)
		p <- blockEvent
	}
}

// SendFrom sends block events to all registered consumers from the
// given block number
func (l *MockLedger) SendFrom(blockNum uint64) {
	l.RLock()
	defer l.RUnlock()

	if blockNum >= uint64(len(l.blocks)) {
		return
	}

	for _, block := range l.blocks[blockNum:] {
		fmt.Printf("Delivering block number: %d\n", block.Number())
		for i, p := range l.consumers {
			fmt.Printf("MockLedger - Notifying Consumer %d with block %d\n", i, block.Number())
			p <- l.eventFactory(block)
		}
	}
}
