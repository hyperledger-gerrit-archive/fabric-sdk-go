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

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"

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

// Group of status to help users infer status codes from various components
type Group string

const (
	// TransportStatus defines the status returned by the transport layer of
	// the connections made by fabric-sdk-go

	// GRPCTransportStatus is the status associated with requests made over
	// gRPC connections
	GRPCTransportStatus Group = "gRPC Transport Status"
	// HTTPTransportStatus is the status associated with requests made over HTTP
	// connections
	HTTPTransportStatus Group = "HTTP Transport Status"

	// ServerStatus defines the status returned by various servers that fabric-sdk-go
	// is a client to

	// EndorserServerStatus status returned by the endorser server
	EndorserServerStatus Group = "Endorser Server Status"
	// EventServerStatus status returned by the eventhub
	EventServerStatus Group = "Event Server Status"
	// OrdererServerStatus status returned by the ordering service
	OrdererServerStatus Group = "Orderer Server Status"
	// FabricCAServerStatus status returned by the Fabric CA server
	FabricCAServerStatus Group = "Fabric CA Server Status"

	// ClientStatus defines the status from responses inferred by fabric-sdk-go.
	// This could be a result of response validation performed by the SDK - for example,
	// a client status could be produced by validating endorsements

	// EndorserClientStatus status returned from the endorser client
	EndorserClientStatus Group = "Endorser Client Status"
	// OrdererClientStatus status returned from the orderer client
	OrdererClientStatus Group = "Orderer Client Status"
)

// FromError returns a Status representing err if it was produced from this
// package, otherwise it returns nil, false.
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

	return nil, false
}

func (s *Status) Error() string {
	return fmt.Sprintf("%s Code: (%d) %s. Description: %s", string(s.Group), s.Code, s.codeString(), s.Message)
}

func (s *Status) codeString() string {
	switch s.Group {
	case GRPCTransportStatus:
		return ToGRPCStatusCode(s.Code).String()
	case EndorserServerStatus, OrdererServerStatus:
		return ToFabricCommonStatusCode(s.Code).String()
	case EventServerStatus:
		return ToTransactionValidationCode(s.Code).String()
	case EndorserClientStatus, OrdererClientStatus:
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
