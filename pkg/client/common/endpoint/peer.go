/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endpoint

import (
	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/client")

// PeerFromDiscoveryClient returns a fab.Peer that carries state information (implements fab.PeerState interface)
func PeerFromDiscoveryClient(ctx contextAPI.Client, endpoint *discclient.Peer) (fab.Peer, error) {
	url := endpoint.AliveMessage.GetAliveMsg().Membership.Endpoint

	logger.Debugf("Adding endpoint [%s]", url)

	peerConfig, found := ctx.EndpointConfig().PeerConfig(url)
	if !found {
		return nil, errors.Errorf("peer config not found for [%s]", url)
	}

	var chaincodes []fab.ChaincodeInfo
	var blockHeight uint64
	if endpoint.StateInfoMessage != nil {
		properties := endpoint.StateInfoMessage.GetStateInfo().GetProperties()
		chaincodes = make([]fab.ChaincodeInfo, 0, len(properties.GetChaincodes()))
		for _, c := range properties.Chaincodes {
			chaincodes = append(chaincodes, chaincodeInfo{
				name:    c.Name,
				version: c.Version,
			})
		}
		blockHeight = properties.GetLedgerHeight()
	}

	peer, err := ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: *peerConfig, MSPID: endpoint.MSPID})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create peer config for [%s]", url)
	}

	return &peerEndpoint{
		Peer:        peer,
		blockHeight: blockHeight,
		chaincodes:  chaincodes,
	}, nil
}

// PeersFromDiscoveryClient returns a []fab.Peer that carries state information (implements fab.PeerState interface)
func PeersFromDiscoveryClient(ctx contextAPI.Client, endpoints []*discclient.Peer) []fab.Peer {
	var peers []fab.Peer
	for _, endpoint := range endpoints {
		peer, err := PeerFromDiscoveryClient(ctx, endpoint)
		if err != nil {
			logger.Debugf(err.Error())
			continue
		}
		peers = append(peers, peer)
	}
	return peers
}

type peerEndpoint struct {
	fab.Peer
	blockHeight uint64
	chaincodes  []fab.ChaincodeInfo
}

func (p peerEndpoint) BlockHeight() uint64 {
	return p.blockHeight
}

func (p peerEndpoint) Chaincodes() []fab.ChaincodeInfo {
	return p.chaincodes
}

type chaincodeInfo struct {
	name    string
	version string
}

func (c chaincodeInfo) Name() string {
	return c.name
}

func (c chaincodeInfo) Version() string {
	return c.version
}
