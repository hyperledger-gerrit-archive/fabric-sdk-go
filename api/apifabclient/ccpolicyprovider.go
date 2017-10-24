/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

// CCPolicyProvider is used to discover peers on the network
type CCPolicyProvider interface {
	NewCCPolicyService(client FabricClient) (CCPolicyService, error)
}

// CCPolicyService retrieves policy for the given chaincode ID on the given channel
type CCPolicyService interface {
	GetChaincodePolicy(channelID string, chaincodeID string) (*common.SignaturePolicyEnvelope, error)
}
