/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"sync"
	"time"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	commonContext "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	reqContext "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/pkg/errors"
)

// Service implements a dynamic Discovery Service that queries
// Fabric's Discovery service for information about the peers that
// are currently joined to the given channel.
type Service struct {
	lock      sync.RWMutex
	chContext commonContext.Channel
	peersRef  *lazyref.Reference
}

// NewService creates a Discovery Service to query the list of member peers on a given channel.
func newService(refreshInterval time.Duration) *Service {
	if refreshInterval == 0 {
		refreshInterval = defaultCacheRefreshInterval
	}

	logger.Debugf("Creating new dynamic discovery service with cache refresh interval %s", refreshInterval)

	s := &Service{}
	s.peersRef = lazyref.New(
		func() (interface{}, error) {
			return s.queryPeers()
		},
		lazyref.WithRefreshInterval(lazyref.InitOnFirstAccess, refreshInterval),
	)
	return s
}

// Initialize initializes the service with channel context
func (s *Service) Initialize(ctx contextAPI.Channel) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.chContext != nil {
		// Already initialized
		logger.Debugf("Already initialized with context: %#v", s.chContext)
		return nil
	}

	logger.Debugf("Initializing with context: %#v", ctx)
	s.chContext = ctx
	return nil
}

// Close will close the cache
func (s *Service) Close() {
	s.peersRef.Close()
}

// GetPeers will invoke the membership snap for the specified channelID to retrieve the list of peers
func (s *Service) GetPeers() ([]fab.Peer, error) {
	refValue, err := s.peersRef.Get()
	if err != nil {
		return nil, err
	}
	peers, ok := refValue.([]fab.Peer)
	if !ok {
		return nil, errors.New("get peersRef didn't return Peer type")
	}
	return peers, nil
}

func (s *Service) channelContext() commonContext.Channel {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.chContext
}

func (s *Service) queryPeers() ([]fab.Peer, error) {
	channelContext := s.channelContext()

	service, err := channelContext.InfraProvider().CreateDiscoverService(channelContext, channelContext.ChannelID())
	if err != nil {
		return nil, err
	}

	reqCtx, cancel := reqContext.NewRequest(channelContext, reqContext.WithTimeout(10*time.Second)) // FIXME: Make configurable
	defer cancel()

	req := discclient.NewRequest().OfChannel(channelContext.ChannelID()).AddPeersQuery()

	resp, err := service.Send(reqCtx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "error calling discover service send")
	}

	endpoints, err := resp.ForChannel(channelContext.ChannelID()).Peers()
	if err != nil {
		return nil, errors.Wrapf(err, "error getting peers from discovery response")
	}

	logger.Infof("Got %d endpoints from discover service", len(endpoints))

	var peers []fab.Peer
	for _, endpoint := range endpoints {
		url := endpoint.AliveMessage.GetAliveMsg().Membership.Endpoint

		logger.Infof("Adding endpoint [%s]", url)

		peerConfig, err := channelContext.EndpointConfig().PeerConfigByURL(url)
		if err != nil {
			return nil, errors.Wrapf(err, "error getting peer config for url [%s]", url)
		}
		peer, err := channelContext.InfraProvider().CreatePeerFromConfig(&fabApi.NetworkPeer{PeerConfig: *peerConfig, MSPID: endpoint.MSPID})
		if err != nil {
			return nil, errors.WithMessage(err, "error creating new peer")
		}
		peers = append(peers, peer)
	}

	return peers, nil
}
