/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/client/lbp"
	"github.com/hyperledger/fabric-sdk-go/pkg/options"
)

type params struct {
	loadBalancePolicy lbp.LoadBalancePolicy
}

func defaultParams() *params {
	return &params{
		loadBalancePolicy: lbp.NewRoundRobin(),
	}
}

// WithLoadBalancePolicy sets the load-balance policy to use when
// choosing an event endpoint from a set of endpoints
func WithLoadBalancePolicy(value lbp.LoadBalancePolicy) options.Opt {
	return func(p options.Params) {
		if setter, ok := p.(loadBalancePolicySetter); ok {
			logger.Debugf("Applying option LoadBalancePolicy: %#v", value)
			setter.SetBalancePolicy(value)
		}
	}
}

type loadBalancePolicySetter interface {
	SetBalancePolicy(value lbp.LoadBalancePolicy)
}

func (p *params) SetBalancePolicy(value lbp.LoadBalancePolicy) {
	p.loadBalancePolicy = value
}
