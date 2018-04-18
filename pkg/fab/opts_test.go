/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

func TestCreateCustomEndpoingConfig(t *testing.T) {
	mtc := MockTimeoutConfigImpl{}
	mmi := MockMspIDImpl{}
	endpoingConfingOption := BuildConfigEndpointFromOptions(mtc, mmi)
	var eco *EndpointConfigOptions
	var ok bool
	if eco, ok = endpoingConfingOption.(*EndpointConfigOptions); !ok {
		t.Fatalf("BuildConfigEndpointFromOptions did not return a Options instance %+T. OK? %b", endpoingConfingOption, ok)
	}

	if eco.timeout == nil {
		t.Fatalf("EndpointConfig was supposed to have Timeout function overriden from Options but was not %+v", eco)
	}
}

type MockTimeoutConfig interface {
	Timeout(fab.TimeoutType) time.Duration
}

type MockTimeoutConfigImpl struct {
}

func (M *MockTimeoutConfigImpl) Timeout(timeoutType fab.TimeoutType) time.Duration {
	return 10 * time.Second
}

type MockMspID interface {
	MSPID(org string) (string, error)
}

type MockMspIDImpl struct {
}

func (M *MockMspIDImpl) MSPID(org string) (string, error) {
	return "", nil
}
