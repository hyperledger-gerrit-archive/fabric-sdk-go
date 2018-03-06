/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defca

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/capvdr"
)

func TestProviderFactory(t *testing.T) {
	factory := NewProviderFactory()
	ctx := mocks.NewMockProviderContext()

	fabricProvider, err := factory.CreateCAProvider(ctx)
	if err != nil {
		t.Fatalf("Unexpected error creating fabric provider %v", err)
	}

	_, ok := fabricProvider.(*capvdr.CAProvider)
	if !ok {
		t.Fatalf("Unexpected fabric provider created")
	}
}
