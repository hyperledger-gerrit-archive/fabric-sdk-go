/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package status defines metadata for errors returned by fabric-sdk-go. This
// information may be used by SDK users to make decisions about how to handle
// certain error conditions.
package status

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/util/errors/multi"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	grpcstatus "google.golang.org/grpc/status"
)

// Status provides additional information about an unsuccessful operation
// performed by fabric-sdk-go. Essentially, this object contains metadata about
// an error returned by the SDK.
type Status struct {
	// Group status group
	Group Group
	// Code status code
	Code int32
	// Message status message
	Message string
	// Details any additional status details
	Details []interface{}
}

// FromError returns a Status representing err if available,
// otherwise it returns nil, false.
func FromError(err error) (s *Status, ok bool) {
	if err == nil {
		return &Status{Code: int32(OK)}, true
	}
	if s, ok := err.(*Status); ok {
		return s, true
	}
	unwrappedErr := errors.Cause(err)
	if s, ok := unwrappedErr.(*Status); ok {
		return s, true
	}
	if m, ok := unwrappedErr.(multi.Errors); ok {
		return New(ClientStatus, MultipleErrors.ToInt32(), m.Error(), nil), true
	}

	return nil, false
}

func (s *Status) Error() string {
	return fmt.Sprintf("%s Code: (%d) %s. Description: %s", s.Group.String(), s.Code, s.codeString(), s.Message)
}

func (s *Status) codeString() string {
	switch s.Group {
	case GRPCTransportStatus:
		return ToGRPCStatusCode(s.Code).String()
	case EndorserServerStatus, OrdererServerStatus:
		return ToFabricCommonStatusCode(s.Code).String()
	case EventServerStatus:
		return ToTransactionValidationCode(s.Code).String()
	case EndorserClientStatus, OrdererClientStatus, ClientStatus:
		return ToSDKStatusCode(s.Code).String()
	default:
		return Unknown.String()
	}
}

// New returns a Status with the given parameters
func New(group Group, code int32, msg string, details []interface{}) *Status {
	return &Status{Group: group, Code: code, Message: msg, Details: details}
}

// NewFromProposalResponse creates a status created from the given ProposalResponse
func NewFromProposalResponse(res *pb.ProposalResponse, endorser string) *Status {
	if res == nil {
		return nil
	}
	details := []interface{}{endorser, res.Response.Payload}

	return New(EndorserServerStatus, res.Response.Status, res.Response.Message, details)
}

// NewFromGRPCStatus new Status from gRPC status response
func NewFromGRPCStatus(s *grpcstatus.Status) *Status {
	if s == nil {
		return nil
	}
	details := make([]interface{}, len(s.Proto().Details))
	for i, detail := range s.Proto().Details {
		details[i] = detail
	}

	return &Status{Group: GRPCTransportStatus, Code: s.Proto().Code,
		Message: s.Message(), Details: details}
}

// ChaincodeStatus is for extracting Code and message from chaincode GRPC errors
type ChaincodeStatus struct {
	Code    int
	Message string
}

// NewFromExtractedChaincodeError returns Status when an error occurs in GRPC Transport
func NewFromExtractedChaincodeError(code int, message string) *Status {
	status := &ChaincodeStatus{Code: code, Message: message}

	return &Status{Group: GRPCTransportStatus, Code: ChaincodeError.ToInt32(),
		Message: message, Details: []interface{}{status}}
}

// WithStack creates a new error with status
func WithStack(group Group, code int32, message string, details []interface{}) error {
	return errors.WithStack(New(group, code, message, details))
}
