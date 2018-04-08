/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"crypto/sha256"
	"strconv"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// EventsCacheKey holds a key for the events provider cache
type EventsCacheKey struct {
	CacheKey
	opts []options.Opt
}

type params struct {
	permitBlockEvents bool
}

func defaultParams() *params {
	return &params{}
}

func (p *params) PermitBlockEvents() {
	p.permitBlockEvents = true
}

type permitBlockEventsSetter interface {
	PermitBlockEvents()
}

func (p *params) getOptKey() string {
	//	Construct opts portion
	optKey := "blockEvents:" + strconv.FormatBool(p.permitBlockEvents)
	return optKey
}

// NewEventsCacheKey returns a new CacheKey
func NewEventsCacheKey(ctx fab.ClientContext, chConfig fab.ChannelCfg, opts ...options.Opt) (*EventsCacheKey, error) {
	identity, err := ctx.Serialize()
	if err != nil {
		return nil, err
	}

	params := defaultParams()
	options.Apply(params, opts)

	h := sha256.New()
	h.Write(append(identity, []byte(params.getOptKey())...)) // nolint
	hash := h.Sum([]byte(chConfig.ID()))

	return &EventsCacheKey{
		CacheKey: CacheKey{
			key:      string(hash),
			context:  ctx,
			chConfig: chConfig,
		},
		opts: opts,
	}, nil
}

// Opts returns the options to use for creating events service
func (k *EventsCacheKey) Opts() []options.Opt {
	return k.opts
}
