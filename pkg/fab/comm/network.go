/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

// NetworkPeerConfig fetches the peer configuration based on a key (name or URL).
func NetworkPeerConfig(cfg fab.EndpointConfig, key string) (*fab.NetworkPeer, error) {
	peerCfg, ok, err := cfg.PeerConfig(key)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get network peer config")
	}
	if !ok {
		return nil, errors.New("no matching peer found")
	}

	// find MSP ID
	networkPeers, err := cfg.NetworkPeers()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to load network peer config")
	}

	var mspID string
	for _, peer := range networkPeers {
		if peer.URL == peerCfg.URL { // need to use the looked-up URL due to matching
			mspID = peer.MSPID
			break
		}
	}

	np := fab.NetworkPeer{
		PeerConfig: *peerCfg,
		MSPID:      mspID,
	}

	return &np, nil
}

// SearchPeerConfigFromURL searches for the peer configuration based on a URL.
func SearchPeerConfigFromURL(cfg fab.EndpointConfig, url string) (*fab.PeerConfig, error) {
	peerCfg, ok, err := cfg.PeerConfig(url)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get peer config from [%s]", url)
	}
	if ok {
		return peerCfg, nil
	}

	//If the given url is already parsed URL through entity matcher, then 'cfg.PeerConfig()'
	//may return NoMatchingPeerEntity error. So retry with network peer URLs
	networkPeers, err := cfg.NetworkPeers()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to load network peer config")
	}

	for _, peer := range networkPeers {
		if peer.URL == url {
			return &peer.PeerConfig, nil
		}
	}

	return nil, errors.Errorf("unable to get peerconfig for given url : %s", url)
}
