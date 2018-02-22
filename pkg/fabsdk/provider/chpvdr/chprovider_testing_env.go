// +build testing

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import "github.com/hyperledger/fabric-sdk-go/pkg/context"

// SetChannelConfig allows setting channel configuration.
// This method is intended to enable tests and should not be called.
func (cp *ChannelProvider) SetChannelConfig(cfg context.ChannelCfg) {
	cp.chCfgMap.Store(cfg.Name(), cfg)
}
