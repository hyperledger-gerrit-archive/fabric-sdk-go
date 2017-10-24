/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package noopccpolicy

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// DefaultCCPolicyProvider implements default chaincode policy provider without chaincode policy service
type DefaultCCPolicyProvider struct {
}

// NewCCPolicyProvider returns default policy provider with no chaincode policy service
func NewCCPolicyProvider(config apiconfig.Config) (*DefaultCCPolicyProvider, error) {
	return &DefaultCCPolicyProvider{}, nil
}

// NewCCPolicyService returns nil chaincode policy service
func (dp *DefaultCCPolicyProvider) NewCCPolicyService(client fab.FabricClient) (fab.CCPolicyService, error) {
	return nil, nil
}
