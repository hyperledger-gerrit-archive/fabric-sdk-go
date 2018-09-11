/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

// MockPeer contains mock PeerState
type MockPeer struct {
	*fabmocks.MockPeer
	blockHeight uint64
	chaincodes  []MockChaincodeInfo
	lock        sync.RWMutex
}

// MockPeerOpt is a mock peer option
type MockPeerOpt func(*MockPeer)

// WithMSP sets the MSP ID of the mock peer
func WithMSP(mspID string) MockPeerOpt {
	return func(p *MockPeer) {
		p.SetMSPID(mspID)
	}
}

// WithBlockHeight sets the block height of the mock peer
func WithBlockHeight(blockHeight uint64) MockPeerOpt {
	return func(p *MockPeer) {
		p.blockHeight = blockHeight
	}
}

// NewMockStatefulPeer returns a new MockPeer with the given options
func NewMockStatefulPeer(name, url string, opts ...MockPeerOpt) *MockPeer {
	p := &MockPeer{
		MockPeer: fabmocks.NewMockPeer(name, url),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// NewMockPeer returns a new MockPeer
// Deprecated: This function will be deprecated in the future. Use NewMockStatefulPeer instead.
func NewMockPeer(name, url string, blockHeight uint64) *MockPeer {
	return &MockPeer{
		MockPeer:    fabmocks.NewMockPeer(name, url),
		blockHeight: blockHeight,
	}
}

// BlockHeight returns the block height
func (p *MockPeer) BlockHeight() uint64 {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.blockHeight
}

// SetBlockHeight sets the block height
func (p *MockPeer) SetBlockHeight(blockHeight uint64) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.blockHeight = blockHeight
}

// Chaincodes returns the chaincodes info
func (p *MockPeer) Chaincodes() []fab.ChaincodeInfo {
	cs := make([]fab.ChaincodeInfo, 0, len(p.chaincodes))
	for _, i := range p.chaincodes {
		cs = append(cs, i)
	}
	return cs
}

// SetChaincodeInfo sets the chaincodes info
func (p *MockPeer) SetChaincodeInfo(i []MockChaincodeInfo) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.chaincodes = i
}

// MockChaincodeInfo contains mock fab.ChaincodeInfo
type MockChaincodeInfo struct {
	name    string
	version string
}

// Name returns the chaincode name
func (i MockChaincodeInfo) Name() string {
	return i.name
}

// Version returns the chaincode version
func (i MockChaincodeInfo) Version() string {
	return i.version
}
