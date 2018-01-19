/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"testing"

	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
)

const (
	identityOptConfigFile = "testdata/test.yaml"
	validIdentityOptUser  = "User1"
	validIdentityOptOrg   = "Org2"
)

func TestFromNameValid(t *testing.T) {
	c, err := configImpl.FromFile(identityOptConfigFile)
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	opts := identityOptions{}
	opt := FromName(validIdentityOptUser)
	err = opt(&opts, sdk, validIdentityOptOrg)
	if err != nil {
		t.Fatalf("Expected no error from opt, but got %v", err)
	}
	if opts.identity == nil {
		t.Fatal("Expected identity to be populated")
	}
}

func TestFromNameInvalid(t *testing.T) {
	c, err := configImpl.FromFile(identityOptConfigFile)
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	opts := identityOptions{}
	opt := FromName("notarealuser")
	err = opt(&opts, sdk, validIdentityOptOrg)
	if err == nil {
		t.Fatal("Expected error from opt")
	}
	if opts.identity != nil {
		t.Fatal("Expected identity to not be populated")
	}
}

func TestFromIdentity(t *testing.T) {
	c, err := configImpl.FromFile(identityOptConfigFile)
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(c)
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	identity, err := sdk.NewPreEnrolledUser(validIdentityOptOrg, validIdentityOptUser)
	if err != nil {
		t.Fatalf("Unexpected error loading identity: %v", err)
	}

	opts := identityOptions{}
	opt := FromIdentity(identity)
	err = opt(&opts, sdk, validIdentityOptOrg)
	if err != nil {
		t.Fatalf("Expected no error from opt, but got %v", err)
	}
	if opts.identity != identity {
		t.Fatal("Expected identity to be populated")
	}
}
