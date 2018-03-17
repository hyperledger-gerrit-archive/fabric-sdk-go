/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chconfig

import (
	"crypto/sha256"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"

	"github.com/pkg/errors"
)

// Provider provides ChannelConfig
type Provider func(channelID string) (fab.ChannelConfig, error)

// CacheKey channel config reference cache key
type CacheKey interface {
	lazycache.Key
	Context() fab.ClientContext
	ChannelID() string
}

// CacheKey holds a key for the provider cache
type cacheKey struct {
	key       string
	context   fab.ClientContext
	channelID string
}

// NewCacheKey returns a new CacheKey
func NewCacheKey(ctx fab.ClientContext, channel string) (CacheKey, error) {
	identity, err := ctx.Serialize()
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	h.Write(identity)
	hash := h.Sum([]byte(channel))

	return &cacheKey{
		key:       string(hash),
		context:   ctx,
		channelID: channel,
	}, nil
}

// NewRefCache a cache of channel config references that refreshed with the
// given interval
func NewRefCache(refresh time.Duration, pvdr Provider) *lazycache.Cache {
	initializer := func(key lazycache.Key) (interface{}, error) {
		ck, ok := key.(CacheKey)
		if !ok {
			return nil, errors.New("Unexpected cache key")
		}
		return NewRef(refresh, pvdr, ck.ChannelID(), ck.Context()), nil
	}

	return lazycache.New("Channel_Cfg_Cache", initializer)
}

// String returns the key as a string
func (k *cacheKey) String() string {
	return k.key
}

// Context returns the Context
func (k *cacheKey) Context() fab.ClientContext {
	return k.context
}

// ChannelID returns the channelID
func (k *cacheKey) ChannelID() string {
	return k.channelID
}
