/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deliverclient

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
)

type params struct {
	connProvider api.ConnectionProvider
	seekType     seek.Type
	fromBlock    uint64
	respTimeout  time.Duration
}

func defaultParams() *params {
	return &params{
		connProvider: deliverFilteredProvider,
		seekType:     seek.Newest,
		respTimeout:  5 * time.Second,
	}
}

func (p *params) PermitBlockEvents() {
	logger.Debugf("PermitBlockEvents")
	p.connProvider = deliverProvider
}

// SetConnectionProvider is only used in unit tests
func (p *params) SetConnectionProvider(connProvider api.ConnectionProvider) {
	logger.Debugf("ConnectionProvider: %#v", connProvider)
	p.connProvider = connProvider
}

func (p *params) SetFromBlock(value uint64) {
	logger.Debugf("FromBlock: %d", value)
	p.fromBlock = value
}

func (p *params) SetSeekType(value seek.Type) {
	logger.Debugf("SeekType: %s", value)
	p.seekType = value
}

func (p *params) SetResponseTimeout(value time.Duration) {
	logger.Debugf("ResponseTimeout: %s", value)
	p.respTimeout = value
}
