/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package status

import (
	"strconv"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	grpcCodes "google.golang.org/grpc/codes"
)

// Code represents a status code
type Code uint32

const (
	// OK is returned on success.
	OK Code = 0

	// Unknown represents status codes that are uncategorized or unknown to the SDK
	Unknown Code = 1

	// ConnectionFailed is returned when a network connection attempt from the SDK fails
	ConnectionFailed Code = 2

	// EndorsementMismatch is returned when there is a mismatch in endorsements received by the SDK
	EndorsementMismatch Code = 3

	// EmptyCert is return when an empty cert is returned
	EmptyCert Code = 4

	// Timeout operation timed out
	Timeout Code = 5

	// NoPeersFound No peers were discovered/configured
	NoPeersFound Code = 6

	// MultipleErrors multiple errors occurred
	MultipleErrors Code = 7

	// SignatureVerificationFailed is when signature fails verification
	SignatureVerificationFailed Code = 8

	// MissingEndorsement is if an endoresement is missing
	MissingEndorsement Code = 9

	// ChaincodeError is for errors returned by Chaincode
	ChaincodeError Code = 10

	// NoMatchingCertificateAuthorityEntity is if entityMatchers are unable to find any matchingCertificateAuthority
	NoMatchingCertificateAuthorityEntity Code = 21

	// NoMatchingPeerEntity is if entityMatchers are unable to find any matchingPeer
	NoMatchingPeerEntity Code = 22

	// NoMatchingOrdererEntity is if entityMatchers are unable to find any matchingOrderer
	NoMatchingOrdererEntity Code = 23

	// PrematureChaincodeExecution indicates that an attempt was made to invoke a chaincode that's
	// in the process of being launched.
	PrematureChaincodeExecution Code = 24
)

// CodeName maps the codes in this packages to human-readable strings
var CodeName = map[int32]string{
	0:  "OK",
	1:  "UNKNOWN",
	2:  "CONNECTION_FAILED",
	3:  "ENDORSEMENT_MISMATCH",
	4:  "EMPTY_CERT",
	5:  "TIMEOUT",
	6:  "NO_PEERS_FOUND",
	7:  "MULTIPLE_ERRORS",
	8:  "SIGNATURE_VERIFICATION_FAILED",
	9:  "MISSING_ENDORSEMENT",
	10: "CHAINCODE_ERROR",
	21: "NO_MATCHING_CERTIFICATE_AUTHORITY_ENTITY",
	22: "NO_MATCHING_PEER_ENTITY",
	23: "NO_MATCHING_ORDERER_ENTITY",
	24: "PREMATURE_CHAINCODE_EXECUTION",
}

// ToInt32 cast to int32
func (c Code) ToInt32() int32 {
	return int32(c)
}

// String representation of the code
func (c Code) String() string {
	if s, ok := CodeName[c.ToInt32()]; ok {
		return s
	}
	return strconv.Itoa(int(c))
}

// ToSDKStatusCode cast to fabric-sdk-go status code
func ToSDKStatusCode(c int32) Code {
	return Code(c)
}

// ToGRPCStatusCode cast to gRPC status code
func ToGRPCStatusCode(c int32) grpcCodes.Code {
	return grpcCodes.Code(c)
}

// ToPeerStatusCode cast to peer status
func ToPeerStatusCode(c int32) common.Status {
	return ToFabricCommonStatusCode(c)
}

// ToOrdererStatusCode cast to peer status
func ToOrdererStatusCode(c int32) common.Status {
	return ToFabricCommonStatusCode(c)
}

// ToFabricCommonStatusCode cast to common.Status
func ToFabricCommonStatusCode(c int32) common.Status {
	return common.Status(c)
}

// ToTransactionValidationCode cast to transaction validation status code
func ToTransactionValidationCode(c int32) pb.TxValidationCode {
	return pb.TxValidationCode(c)
}
