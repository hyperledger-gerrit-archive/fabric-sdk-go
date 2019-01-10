// +build pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/metrics"
)

// GetMetrics not used in this mockcontext
func (c *MockChannelContext) GetMetrics() *metrics.ClientMetrics {
	return &metrics.ClientMetrics{}
}

// GetMetrics not used in this mockcontext
func (c *MockProviderContext) GetMetrics() *metrics.ClientMetrics {
	return &metrics.ClientMetrics{}
}
