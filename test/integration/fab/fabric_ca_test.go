/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const (
	org1Name = "Org1"
	org2Name = "Org2"
)

func TestEnrollOrg2(t *testing.T) {

	sdk, err := fabsdk.New(config.FromFile(sdkConfigFile))
	if err != nil {
		t.Fatalf("SDK init failed: %v", err)
	}

	im, err := sdk.NewIdentityManager(org2Name)
	if err != nil {
		t.Fatalf("NewIdentityManager return error: %v", err)
	}

	err = im.Enroll("admin", "adminpw")
	if err != nil {
		t.Fatalf("Enroll returned error: %v", err)
	}
}
