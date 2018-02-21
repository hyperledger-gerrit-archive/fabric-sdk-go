/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package service

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/service/dispatcher"
)

// Opts contains options for the event servce
type Opts struct {
	dispatcher.Opts
}

// DefaultOpts returns default event service
func DefaultOpts() *Opts {
	return &Opts{
		Opts: *dispatcher.DefaultOpts(),
	}
}
