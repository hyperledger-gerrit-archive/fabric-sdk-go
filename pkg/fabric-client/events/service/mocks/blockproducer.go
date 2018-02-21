/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"sync"

	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// BlockProducer is a block BlockProducer that ensures the block
// number is set sequencially
type BlockProducer struct {
	sync.Mutex
	blockNum uint64
}

// NewBlockProducer returns a new block producer
func NewBlockProducer() *BlockProducer {
	return &BlockProducer{}
}

// NewBlock returns a new block
func (p *BlockProducer) NewBlock(channelID string, transactions ...*TxInfo) *cb.Block {
	p.Lock()
	defer p.Unlock()

	block := NewBlock(channelID, transactions...)
	block.Header.Number = p.blockNum
	p.blockNum++
	return block
}

// NewFilteredBlock returns a new filtered block
func (p *BlockProducer) NewFilteredBlock(channelID string, filteredTx ...*pb.FilteredTransaction) *pb.FilteredBlock {
	p.Lock()
	defer p.Unlock()

	block := NewFilteredBlock(channelID, filteredTx...)
	block.Number = p.blockNum
	p.blockNum++
	return block
}
