/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicdiscovery

import (
	"context"
	"sync"
	"time"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabdiscovery "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazyref"
	"github.com/pkg/errors"
)

type discoveryClient interface {
	Send(ctx context.Context, req *discclient.Request, targets ...fab.PeerConfig) ([]fabdiscovery.Response, error)
}

// clientProvider is overridden by unit tests
var clientProvider = func(ctx contextAPI.Client) (discoveryClient, error) {
	return fabdiscovery.New(ctx)
}

// service implements a dynamic Discovery Service that queries
// Fabric's Discovery service for information about the peers that
// are currently joined to the given channel.
type service struct {
	responseTimeout time.Duration
	lock            sync.RWMutex
	ctx             contextAPI.Client
	discClient      discoveryClient
	peersRef        *lazyref.Reference
}

type queryPeers func() ([]fab.Peer, error)

func newService(config fab.EndpointConfig, query queryPeers, opts ...coptions.Opt) *service {
	options := options{}
	coptions.Apply(&options, opts)

	if options.refreshInterval == 0 {
		options.refreshInterval = config.Timeout(fab.DiscoveryServiceRefresh)
	}

	if options.responseTimeout == 0 {
		options.responseTimeout = config.Timeout(fab.DiscoveryResponse)
	}

	logger.Debugf("Cache refresh interval: %s", options.refreshInterval)
	logger.Debugf("Deliver service response timeout: %s", options.responseTimeout)

	return &service{
		responseTimeout: options.responseTimeout,
		peersRef: lazyref.New(
			func() (interface{}, error) {
				return query()
			},
			lazyref.WithRefreshInterval(lazyref.InitOnFirstAccess, options.refreshInterval),
		),
	}
}

// initialize initializes the service with client context
func (s *service) initialize(ctx contextAPI.Client) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.ctx != nil {
		// Already initialized
		logger.Debugf("Already initialized with context: %#v", s.ctx)
		return nil
	}

	discoveryClient, err := clientProvider(ctx)
	if err != nil {
		return errors.Wrap(err, "error creating discover client")
	}

	logger.Debugf("Initializing with context: %#v", ctx)
	s.ctx = ctx
	s.discClient = discoveryClient
	return nil
}

// Close stops the lazyref background refresh
func (s *service) Close() {
	logger.Debug("Closing peers ref...")
	s.peersRef.Close()
}

// GetPeers returns the available peers
func (s *service) GetPeers() ([]fab.Peer, error) {
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

func (s *service) context() contextAPI.Client {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ctx
}

func (s *service) discoveryClient() discoveryClient {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.discClient
}
