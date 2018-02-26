/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/pkg/errors"
)

const (
	clientConfigFile     = "testdata/test.yaml"
	clientValidAdmin     = "Admin"
	clientValidUser      = "User1"
	clientValidExtraOrg  = "OrgX"
	clientValidExtraUser = "OrgXUser"
)

type testProvs context.Providers

func TestClientWithContext(t *testing.T) {
	sdk, err := New(configImpl.FromFile(clientConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}
	ctxIdentity, err := sdk.NewUser("Org2", "User1")
	if err != nil {
		t.Fatalf("Context identity error %v", err)
	}
	fmt.Printf("***contextIdentity %v\n", ctxIdentity)
	client, err := sdk.Context(context.WithIdentity(ctxIdentity))
	if err != nil {
		t.Fatalf("Expected no error from Client, but got %v", err)
	}
	if client == nil {
		t.Fatalf("Expected client to be configured")
	}
	// npf := pf.NewProviderFactory()
	// fp, _ := (npf.CreateFabricProvider(sdk.fabContext()))
	// fmt.Printf("%v %v\n", npf, fp)
	client, err = sdk.Context(context.WithIdentity(ctxIdentity), context.WithProvider(sdk.context()))
	if err != nil {
		t.Fatalf("Expected no error from Client, but got %v", err)
	}
	if client == nil {
		t.Fatalf("Expected client to be configured")
	}

}

func TestNewGoodClientOpt(t *testing.T) {
	sdk, err := New(configImpl.FromFile(clientConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(WithUser(clientValidUser), goodClientOpt()).ResourceMgmt()
	if err != nil {
		t.Fatalf("Expected no error from Client, but got %v", err)
	}
}

func TestFromConfigGoodClientOpt(t *testing.T) {
	c, err := configImpl.FromFile(clientConfigFile)()
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}

	sdk, err := New(WithConfig(c))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(WithUser(clientValidUser), goodClientOpt()).ResourceMgmt()
	if err != nil {
		t.Fatalf("Expected no error from Client, but got %v", err)
	}
}

func goodClientOpt() ContextOption {
	return func(opts *contextOptions) error {
		return nil
	}
}

func TestNewBadClientOpt(t *testing.T) {
	sdk, err := New(configImpl.FromFile(clientConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(WithUser(clientValidUser), badClientOpt()).ResourceMgmt()
	if err == nil {
		t.Fatal("Expected error from Client")
	}
}

func badClientOpt() ContextOption {
	return func(opts *contextOptions) error {
		return errors.New("Bad Opt")
	}
}

func TestClient(t *testing.T) {
	sdk, err := New(configImpl.FromFile(clientConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(WithUser(clientValidUser)).ResourceMgmt()
	if err != nil {
		t.Fatalf("Expected no error from Client, but got %v", err)
	}
}

func TestWithOrg(t *testing.T) {
	sdk, err := New(configImpl.FromFile(clientConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(WithUser("notarealuser"), WithOrg(clientValidExtraOrg)).ResourceMgmt()
	if err == nil {
		t.Fatal("Expected error from Client")
	}

	_, err = sdk.NewClient(WithUser(clientValidExtraUser), WithOrg(clientValidExtraOrg)).ResourceMgmt()
	if err != nil {
		t.Fatalf("Expected no error from Client, but got %v", err)
	}
}

func TestWithFilter(t *testing.T) {
	tf := mockTargetFilter{}
	opt := WithTargetFilter(&tf)

	opts := clientOptions{}
	err := opt(&opts)
	if err != nil {
		t.Fatalf("Expected no error from option, but got %v", err)
	}

	if opts.targetFilter != &tf {
		t.Fatalf("Expected target filter to be set in opts")
	}
}

func TestWithConfig(t *testing.T) {
	c, err := configImpl.FromFile(clientConfigFile)()
	if err != nil {
		t.Fatalf("Unexpected error from config: %v", err)
	}
	opt := withConfig(c)

	opts := contextOptions{}
	err = opt(&opts)
	if err != nil {
		t.Fatalf("Expected no error from option, but got %v", err)
	}

	if opts.config != c {
		t.Fatalf("Expected config to be set in opts")
	}
}

func TestNoIdentity(t *testing.T) {
	sdk, err := New(configImpl.FromFile(clientConfigFile))
	if err != nil {
		t.Fatalf("Expected no error from New, but got %v", err)
	}

	_, err = sdk.NewClient(noopIdentityOpt(), goodClientOpt()).ResourceMgmt()
	if err == nil {
		t.Fatal("Expected error from Client")
	}
}

func noopIdentityOpt() IdentityOption {
	return func(o *identityOptions, sdk *FabricSDK, orgName string) error {
		return nil
	}
}

type mockTargetFilter struct{}

func (f *mockTargetFilter) Accept(peer fab.Peer) bool {
	return false
}

func goodClientOptFromCtx() context.ClientOption {
	return func(o *context.ClientOptions) error {
		return nil
	}
}

func badClientOptFromCtx() context.ClientOption {
	return func(o *context.ClientOptions) error {
		return errors.New("Bad Opt")
	}
}

func goodCoreOption() context.ProviderOption {
	return func(o *context.ProviderOptions) error {
		return nil
	}
}

func badCoreOption() context.ProviderOption {
	return func(o *context.ProviderOptions) error {
		return errors.New("Bad Core Opt")
	}
}
