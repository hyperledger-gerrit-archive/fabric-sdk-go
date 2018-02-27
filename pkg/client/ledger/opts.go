/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ledger

import "github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"

const (
	minMatches = 1
	maxTargets = 1
)

//RequestOption func for each Opts argument
type RequestOption func(opts *Opts) error

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*Client) error

// WithDefaultTargetFilter option to configure new
func WithDefaultTargetFilter(filter TargetFilter) ClientOption {
	return func(rmc *Client) error {
		rmc.filter = filter
		return nil
	}
}

// TargetFilter allows for filtering target peers
type TargetFilter interface {
	// Accept returns true if peer should be included in the list of target peers
	Accept(peer fab.Peer) bool
}

//Opts contains options for operations performed by LedgerClient
type Opts struct {
	Targets      []fab.Peer   // target peers
	TargetFilter TargetFilter // target filter
	MaxTargets   int          // maximum number of targets to select
	MinMatches   int          // min number of targets that have to agree on result
}

//WithTargets encapsulates fab.Peer targets to ledger RequestOption
func WithTargets(targets ...fab.Peer) RequestOption {
	return func(opts *Opts) error {
		opts.Targets = targets
		return nil
	}
}

//WithTargetFilter encapsulates TargetFilter targets to ledger RequestOption
func WithTargetFilter(targetFilter TargetFilter) RequestOption {
	return func(opts *Opts) error {
		opts.TargetFilter = targetFilter
		return nil
	}
}
