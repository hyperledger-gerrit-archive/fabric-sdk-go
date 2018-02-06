/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blocklist

import (
	"sync"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// Filter is a discovery filter that blocklists certain peers that are
// known to be down for the configured amount of time
type Filter struct {
	// blocklistURLs contains a map of peer URLs as keys and timestamps as values
	// peers are expired from the blocklist based on these timestamps
	blocklistURLs  sync.Map
	expiryInterval time.Duration
}

// New creates a new blocklist filter with the given expiry interval
func New(expire time.Duration) *Filter {
	return &Filter{expiryInterval: expire}
}

// Accept returns whether or not to Accept a peer as a canditate for endorsement
func (b *Filter) Accept(peer apifabclient.Peer) bool {
	peerAddress := urlutil.ToAddress(peer.URL())
	value, ok := b.blocklistURLs.Load(peerAddress)
	if ok {
		timeAdded, ok := value.(time.Time)
		if ok && timeAdded.Add(b.expiryInterval).After(time.Now()) {
			logger.Infof("Rejecting peer %s", peer.URL())
			return false
		}
		b.blocklistURLs.Delete(peerAddress)
	}

	return true
}

// Blocklist the given peer URL
func (b *Filter) Blocklist(err error) {
	s, ok := status.FromError(err)
	if !ok {
		return
	}
	if ok, peerURL := required(s); ok && peerURL != "" {
		logger.Infof("Blocklisting peer %s", peerURL)
		b.blocklistURLs.Store(peerURL, time.Now())
	}
}

// required decides whether the given status error warrants a blocklist
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
