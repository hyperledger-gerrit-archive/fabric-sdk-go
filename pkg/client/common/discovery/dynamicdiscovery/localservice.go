/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"sync"
	"time"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	reqContext "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/pkg/errors"
)

// LocalService implements a dynamic Discovery Service that queries
// Fabric's Discovery service for the peers that are in the local MSP.
type LocalService struct {
	ctx             contextAPI.Local
	responseTimeout time.Duration
	lock            sync.RWMutex
	discClient      discoverClient
	peersRef        *lazyref.Reference
}

// newLocalService creates a Local Discovery Service to query the list of member peers in the local MSP.
func newLocalService(options options) *LocalService {
	logger.Debugf("Creating new dynamic discovery service with cache refresh interval %s", options.refreshInterval)

	s := &LocalService{
		responseTimeout: options.responseTimeout,
	}
	s.peersRef = lazyref.New(
		func() (interface{}, error) {
			return s.queryPeers()
		},
		lazyref.WithRefreshInterval(lazyref.InitOnFirstAccess, options.refreshInterval),
	)
	return s
}

// Initialize initializes the service with local context
func (s *LocalService) Initialize(ctx contextAPI.Local) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.ctx != nil {
		// Already initialized
		logger.Debugf("Already initialized with context: %#v", s.ctx)
		return nil
	}

	discoverClient, err := clientProvider(ctx)
	if err != nil {
		return errors.Wrapf(err, "error creating discover client")
	}

	logger.Debugf("Initializing with context: %#v", ctx)
	s.ctx = ctx
	s.discClient = discoverClient
	return nil
}

// Close stops the lazyref background refresh
func (s *LocalService) Close() {
	logger.Debugf("Closing peers ref...")
	s.peersRef.Close()
}

// GetPeers returns the available peers in the local MSP
func (s *LocalService) GetPeers() ([]fab.Peer, error) {
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

func (s *LocalService) context() contextAPI.Local {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ctx
}

func (s *LocalService) discoverClient() discoverClient {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.discClient
}

func (s *LocalService) queryPeers() ([]fab.Peer, error) {
	logger.Debugf("Refreshing local peers from discovery service...")

	ctx := s.context()
	if ctx == nil {
		return nil, errors.Errorf("the service has not been initialized")
	}

	target, err := s.getTarget(ctx)
	if err != nil {
		return nil, err
	}

	reqCtx, cancel := reqContext.NewRequest(ctx, reqContext.WithTimeout(s.responseTimeout))
	defer cancel()

	req := discclient.NewRequest().AddLocalPeersQuery()
	responses, err := s.discoverClient().Send(reqCtx, req, *target)
	if err != nil {
		return nil, errors.Wrapf(err, "error calling discover service send")
	}
	if len(responses) == 0 {
		return nil, errors.Wrapf(err, "expecting 1 response from discover service send but got none")
	}

	response := responses[0]
	endpoints, err := response.ForLocal().Peers()
	if err != nil {
		return nil, errors.Wrapf(err, "error getting peers from discovery response")
	}

	return s.filterLocalMSP(asPeers(ctx, endpoints)), nil
}

func (s *LocalService) getTarget(ctx contextAPI.Client) (*fab.PeerConfig, error) {
	peers, err := ctx.EndpointConfig().NetworkPeers()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get peer configs")
	}
	mspID := ctx.Identifier().MSPID
	for _, p := range peers {
		// Need to go to a peer with the local MSPID, otherwise the request will be rejected
		if p.MSPID == mspID {
			return &p.PeerConfig, nil
		}
	}
	return nil, errors.Errorf("no bootstrap peers configured for MSP [%s]", mspID)
}

// Even though the local peer query should only return peers in the local
// MSP, this function double checks and logs a warning if this is not the case.
func (s *LocalService) filterLocalMSP(peers []fab.Peer) []fab.Peer {
	localMSPID := s.ctx.Identifier().MSPID
	var filteredPeers []fab.Peer
	for _, p := range peers {
		if p.MSPID() != localMSPID {
			logger.Warnf("Peer [%s] is not part of the local MSP [%s] but in MSP [%s]", p.URL(), localMSPID, p.MSPID())
		} else {
			filteredPeers = append(filteredPeers, p)
		}
	}
	return filteredPeers
}
