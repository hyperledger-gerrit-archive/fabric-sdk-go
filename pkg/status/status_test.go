/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package status

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/stretchr/testify/assert"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func TestStatusConstructors(t *testing.T) {
	s := New(EndorserClientStatus, ConnectionFailed.ToInt32(), "test", nil)
	assert.NotNil(t, s, "Expected status to be constructed")
	assert.EqualValues(t, ConnectionFailed, ToSDKStatusCode(s.Code))
	assert.Equal(t, EndorserClientStatus, s.Group)
	assert.Equal(t, "test", s.Message, "Expected test message")

	s = NewFromGRPCStatus(nil)
	assert.Nil(t, s)
	s = NewFromGRPCStatus(grpcstatus.New(grpccodes.DeadlineExceeded, "test"))
	assert.NotNil(t, s, "Expected status to be constructed")
	assert.EqualValues(t, grpccodes.DeadlineExceeded, ToGRPCStatusCode(s.Code))
	assert.Equal(t, GRPCTransportStatus, s.Group)
	assert.Equal(t, "test", s.Message, "Expected test message")

	s = NewFromProposalResponse(nil, "")
	assert.Nil(t, s)
	s = NewFromProposalResponse(&pb.ProposalResponse{
		Response: &pb.Response{
			Status:  int32(common.Status_BAD_REQUEST),
			Message: "test",
		}}, "localhost")
	assert.NotNil(t, s, "Expected status to be constructed")
	assert.EqualValues(t, common.Status_BAD_REQUEST, ToPeerStatusCode(s.Code))
	assert.Equal(t, EndorserServerStatus, s.Group)
	assert.Equal(t, "test", s.Message, "Expected test message")
	assert.Equal(t, "localhost", s.Details[0].(string))
}

func TestFromError(t *testing.T) {
	s := New(EndorserClientStatus, ConnectionFailed.ToInt32(), "test", nil)
	derivedStatus, ok := FromError(s)
	assert.True(t, ok)
	assert.Equal(t, s, derivedStatus)

	// Test unwrap
	s1 := errors.Wrap(s, "test")
	derivedStatus, ok = FromError(s1)
	assert.True(t, ok)
	assert.Equal(t, s, derivedStatus)

	s, ok = FromError(nil)
	assert.True(t, ok)
	assert.EqualValues(t, OK.ToInt32(), s.Code)

	s, ok = FromError(fmt.Errorf("Test"))
	assert.False(t, ok)
}

func TestStatusToError(t *testing.T) {
	s := New(EndorserClientStatus, ConnectionFailed.ToInt32(), "test", nil)
	assert.Equal(t, "Endorser Client Status Code: 2. Message: test", s.Error())
}

func TestStatuCodeConversion(t *testing.T) {
	c := ToOrdererStatusCode(int32(common.Status_FORBIDDEN))
	assert.EqualValues(t, c, common.Status_FORBIDDEN)

	c1 := ToTransactionValidationCode(int32(pb.TxValidationCode_BAD_COMMON_HEADER))
	assert.EqualValues(t, c1, pb.TxValidationCode_BAD_COMMON_HEADER)
}
