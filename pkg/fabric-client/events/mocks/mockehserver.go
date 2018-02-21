/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"io"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// MockEventhubServer is a mock event hub server
type MockEventhubServer struct {
}

// NewMockEventhubServer returns a new MockEventhubServer
func NewMockEventhubServer() *MockEventhubServer {
	return new(MockEventhubServer)
}

// Chat starts a listener on the given chat stream
func (s *MockEventhubServer) Chat(srv pb.Events_ChatServer) error {
	for {
		signedEvt, err := srv.Recv()
		if err == io.EOF || signedEvt == nil {
			break
		}
	}
	return nil
}
