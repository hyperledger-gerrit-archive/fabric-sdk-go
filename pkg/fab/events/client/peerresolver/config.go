/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peerresolver

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/lbp"
)

var logger = logging.NewLogger("fabsdk/fab")

// GetBalancer returns the configured load balancer
func GetBalancer(config fab.EventServiceConfig) lbp.LoadBalancePolicy {
	switch config.PeerBalancer() {
	case fab.RoundRobin:
		logger.Infof("Using round-robin load balancer.")
		return lbp.NewRoundRobin()
	default:
		logger.Infof("Using random load balancer.")
		return lbp.NewRandom()
	}
}
