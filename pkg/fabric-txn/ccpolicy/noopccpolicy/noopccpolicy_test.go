/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package noopccpolicy

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
)

func TestNoopCCPolicy(t *testing.T) {

	config, err := config.InitConfig("../../../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	ccPolicyProvider, err := NewCCPolicyProvider(config)
	if err != nil {
		t.Fatalf("Failed to setup cc policy provider: %s", err)
	}

	ccPolicyService, err := ccPolicyProvider.NewCCPolicyService(nil)
	if err != nil {
		t.Fatalf("Failed to setup cc policy service: %s", err)
	}

	if ccPolicyService != nil {
		t.Fatalf("Noop cc policy provider should have returned nil service")
	}

}
