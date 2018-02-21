/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/client/lbp"
	esdispatcher "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/service/dispatcher"
)

// Opts contains the options for the event client dispatcher
type Opts struct {
	esdispatcher.Opts
	LBP lbp.LoadBalancePolicy
}

// DefaultOpts returns default options for the event client dispatcher
func DefaultOpts() *Opts {
	return &Opts{
		Opts: *esdispatcher.DefaultOpts(),
		LBP:  lbp.NewRoundRobin(),
	}
}

// LoadBalancePolicy returns the load-balance policy to use when
// choosing an event endpoint from a set of endpoints
func (o *Opts) LoadBalancePolicy() lbp.LoadBalancePolicy {
	return o.LBP
}
