/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package status defines metadata for errors returned by fabric-sdk-go. This
// information may be used by SDK users to make decisions about how to handle
// certain error conditions.
package status

// Group of status to help users infer status codes from various components
type Group int32

const (
	// UnknownStatus unknown status group
	UnknownStatus Group = iota

	// TransportStatus defines the status returned by the transport layer of
	// the connections made by fabric-sdk-go

	// GRPCTransportStatus is the status associated with requests made over
	// gRPC connections
	GRPCTransportStatus
	// HTTPTransportStatus is the status associated with requests made over HTTP
	// connections
	HTTPTransportStatus

	// ServerStatus defines the status returned by various servers that fabric-sdk-go
	// is a client to

	// EndorserServerStatus status returned by the endorser server
	EndorserServerStatus
	// EventServerStatus status returned by the eventhub
	EventServerStatus
	// OrdererServerStatus status returned by the ordering service
	OrdererServerStatus
	// FabricCAServerStatus status returned by the Fabric CA server
	FabricCAServerStatus

	// ClientStatus defines the status from responses inferred by fabric-sdk-go.
	// This could be a result of response validation performed by the SDK - for example,
	// a client status could be produced by validating endorsements

	// EndorserClientStatus status returned from the endorser client
	EndorserClientStatus
	// OrdererClientStatus status returned from the orderer client
	OrdererClientStatus
	// ClientStatus is a generic client status
	ClientStatus
)

// GroupName maps the groups in this packages to human-readable strings
var GroupName = map[int32]string{
	0: "Unknown",
	1: "gRPC Transport Status",
	2: "HTTP Transport Status",
	3: "Endorser Server Status",
	4: "Event Server Status",
	5: "Orderer Server Status",
	6: "Fabric CA Server Status",
	7: "Endorser Client Status",
	8: "Orderer Client Status",
	9: "Client Status",
}

func (g Group) String() string {
	if s, ok := GroupName[int32(g)]; ok {
		return s
	}
	return UnknownStatus.String()
}
