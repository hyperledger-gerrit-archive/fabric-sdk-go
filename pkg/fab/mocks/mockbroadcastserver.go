/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/golang/protobuf/proto"
	po "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// TestBlock is a test block
var TestBlock = &po.DeliverResponse{
	Type: &po.DeliverResponse_Block{
		Block: &common.Block{
			Data: &common.BlockData{
				Data: [][]byte{[]byte("test")},
			},
		},
	},
}

var broadcastResponseSuccess = &po.BroadcastResponse{Status: common.Status_SUCCESS}
var broadcastResponseError = &po.BroadcastResponse{Status: common.Status_INTERNAL_SERVER_ERROR}

// MockBroadcastServer mock broadcast server
type MockBroadcastServer struct {
	DeliverError                 error
	BroadcastInternalServerError bool
	DeliverResponse              *po.DeliverResponse
	BroadcastError               error
	BroadcastCustomResponse      *po.BroadcastResponse
	Creds                        credentials.TransportCredentials
	srv                          *grpc.Server
	wg                           sync.WaitGroup
	// Use the MockBroadCastServer with either a common.Block or a pb.FilteredBlock channel (do not set both)
	Deliveries         chan *common.Block
	FilteredDeliveries chan *pb.FilteredBlock
	blkNum             uint64
}

// Broadcast mock broadcast
func (m *MockBroadcastServer) Broadcast(server po.AtomicBroadcast_BroadcastServer) error {
	res, err := server.Recv()
	if err == io.EOF {
		return nil
	}

	pl := &common.Payload{}
	err = proto.Unmarshal(res.Payload, pl)
	if err != nil {
		return err
	}
	chdr := &common.ChannelHeader{}
	err = proto.Unmarshal(pl.Header.ChannelHeader, chdr)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	if m.BroadcastError != nil {
		return m.BroadcastError
	}

	if m.BroadcastInternalServerError {
		return server.Send(broadcastResponseError)
	}

	if m.BroadcastCustomResponse != nil {
		return server.Send(m.BroadcastCustomResponse)
	}

	err = server.Send(broadcastResponseSuccess)
	if m.Deliveries != nil {
		block := mocks.NewBlock(chdr.ChannelId,
			mocks.NewTransaction(chdr.TxId, pb.TxValidationCode_VALID, common.HeaderType_MESSAGE),
		)
		// m.blkNum is used by FilteredBlock only

		m.Deliveries <- block
	} else if m.FilteredDeliveries != nil {
		filteredBlock := mocks.NewFilteredBlock(chdr.ChannelId,
			mocks.NewFilteredTx(chdr.TxId, pb.TxValidationCode_VALID),
		)
		// increase m.blkNum to mock adding of filtered blocks to the ledger
		m.blkNum++
		filteredBlock.Number = m.blkNum

		m.FilteredDeliveries <- filteredBlock

	}
	return err
}

// Deliver mock deliver
func (m *MockBroadcastServer) Deliver(server po.AtomicBroadcast_DeliverServer) error {
	if m.DeliverError != nil {
		return m.DeliverError
	}

	if m.DeliverResponse != nil {
		if _, err := server.Recv(); err != nil {
			return err
		}
		if err := server.SendMsg(m.DeliverResponse); err != nil {
			return err
		}
		return nil
	}

	if _, err := server.Recv(); err != nil {
		return err
	}
	if err := server.Send(TestBlock); err != nil {
		return err
	}

	return nil
}

// Start the mock broadcast server
func (m *MockBroadcastServer) Start(address string) string {
	if m.srv != nil {
		panic("MockBroadcastServer already started")
	}

	// pass in TLS creds if present
	if m.Creds != nil {
		m.srv = grpc.NewServer(grpc.Creds(m.Creds))
	} else {
		m.srv = grpc.NewServer()
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		panic(fmt.Sprintf("Error starting BroadcastServer %s", err))
	}
	addr := lis.Addr().String()

	test.Logf("Starting MockEventServer [%s]", addr)
	po.RegisterAtomicBroadcastServer(m.srv, m)
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := m.srv.Serve(lis); err != nil {
			test.Logf("StartMockBroadcastServer failed [%s]", err)
		}
	}()

	return addr
}

// Stop the mock broadcast server and wait for completion.
func (m *MockBroadcastServer) Stop() {
	if m.srv == nil {
		panic("MockBroadcastServer not started")
	}

	m.srv.Stop()
	m.wg.Wait()
	m.srv = nil
}
