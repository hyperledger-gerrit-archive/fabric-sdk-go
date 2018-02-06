/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blacklist

import (
	"sync"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// Filter is a discovery filter that blacklists certain peers that are
// known to be down for the configured amount of time
type Filter struct {
	// blacklistURLs contains a map of peer URLs as keys and timestamps as values
	// peers are expired from the blacklist based on these timestamps
	blacklistURLs  sync.Map
	expiryInterval time.Duration
}

// New creates a new blacklist filter with the given expiry interval
func New(expire time.Duration) *Filter {
	return &Filter{expiryInterval: expire}
}

// Accept returns whether or not to Accept a peer as a canditate for endorsement
func (b *Filter) Accept(peer apifabclient.Peer) bool {
	peerAddress := urlutil.ToAddress(peer.URL())
	value, ok := b.blacklistURLs.Load(peerAddress)
	if ok {
		timeAdded, ok := value.(time.Time)
		if ok && timeAdded.Add(b.expiryInterval).After(time.Now()) {
			logger.Infof("Rejecting peer %s", peer.URL())
			return false
		}
		b.blacklistURLs.Delete(peerAddress)
	}

	return true
}

// Blacklist the given peer URL
func (b *Filter) Blacklist(err error) {
	s, ok := status.FromError(err)
	if !ok {
		return
	}
	if ok, peerURL := required(s); ok && peerURL != "" {
		logger.Infof("Blacklisting peer %s", peerURL)
		b.blacklistURLs.Store(peerURL, time.Now())
	}
}

// required decides whether the given status error warrants a blacklist
// on the peer causing the error
func required(s *status.Status) (bool, string) {
	if s.Group == status.EndorserClientStatus && s.Code == status.ConnectionFailed.ToInt32() {
		return true, peerURLFromConnectionFailedStatus(s.Details)
	}
	return false, ""
}

// peerURLFromConnectionFailedStatus extracts the peer url from the status error
// details
func peerURLFromConnectionFailedStatus(details []interface{}) string {
	if len(details) != 0 {
		url, ok := details[0].(string)
		if ok {
			return urlutil.ToAddress(url)
		}
	}
	return ""
}
