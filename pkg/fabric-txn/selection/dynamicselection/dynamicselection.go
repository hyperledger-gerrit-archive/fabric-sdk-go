/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dynamicselection

import (
	"fmt"
	"sort"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/selection/dynamicselection/pgresolver"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// ChannelUser contains user(identity) info to be used for specific channel
type ChannelUser struct {
	ChannelID string
	UserName  string
	OrgName   string
}

// SelectionProvider implements selection provider
type SelectionProvider struct {
	config apiconfig.Config
	users  []ChannelUser
	sdk    *fabapi.FabricSDK
}

// NewSelectionProvider returns dynamic selection provider
func NewSelectionProvider(config apiconfig.Config, users []ChannelUser) (*SelectionProvider, error) {
	return &SelectionProvider{config: config, users: users}, nil
}

type selectionService struct {
	channelID        string
	mutex            sync.RWMutex
	pgResolvers      map[string]pgresolver.PeerGroupResolver
	pgLBP            pgresolver.LoadBalancePolicy
	ccPolicyProvider CCPolicyProvider
}

// Initialize allow for initializing providers
func (p *SelectionProvider) Initialize(sdk *fabapi.FabricSDK) error {
	p.sdk = sdk
	return nil
}

// NewSelectionService creates a selection service
func (p *SelectionProvider) NewSelectionService(channelID string) (fab.SelectionService, error) {
	if channelID == "" {
		return nil, errors.New("Must provide channel ID")
	}

	var channelUser *ChannelUser
	for _, p := range p.users {
		if p.ChannelID == channelID {
			channelUser = &p
		}
	}

	if channelUser == nil {
		return nil, errors.New("Must provide user for channel")
	}

	ccPolicyProvider, err := newCCPolicyProvider(p.sdk, channelID, channelUser.UserName, channelUser.OrgName)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to create cc policy provider")
	}

	return &selectionService{
		channelID:        channelID,
		pgResolvers:      make(map[string]pgresolver.PeerGroupResolver),
		pgLBP:            pgresolver.NewRandomLBP(),
		ccPolicyProvider: ccPolicyProvider,
	}, nil
}

func (s *selectionService) GetEndorsersForChaincode(channelPeers []fab.Peer,
	chaincodeIDs ...string) ([]fab.Peer, error) {

	if len(chaincodeIDs) == 0 {
		return nil, errors.New("no chaincode IDs provided")
	}

	if len(channelPeers) == 0 {
		return nil, errors.New("Must provide at least one channel peer")
	}

	resolver, err := s.getPeerGroupResolver(channelPeers, chaincodeIDs)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("Error getting peer group resolver for chaincodes [%v] on channel [%s]", chaincodeIDs, s.channelID))
	}
	return resolver.Resolve().Peers(), nil
}

func (s *selectionService) getPeerGroupResolver(channelPeers []fab.Peer, chaincodeIDs []string) (pgresolver.PeerGroupResolver, error) {
	key := newResolverKey(s.channelID, chaincodeIDs...)

	s.mutex.RLock()
	resolver := s.pgResolvers[key.String()]
	s.mutex.RUnlock()

	if resolver == nil {
		var err error
		if resolver, err = s.createPGResolver(channelPeers, key); err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("unable to create new peer group resolver for chaincode(s) [%v] on channel [%s]", chaincodeIDs, s.channelID))
		}
	}
	return resolver, nil
}

func (s *selectionService) createPGResolver(channelPeers []fab.Peer, key *resolverKey) (pgresolver.PeerGroupResolver, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	resolver := s.pgResolvers[key.String()]
	if resolver != nil {
		// Already cached
		return resolver, nil
	}

	// Retrieve the signature policies for all of the chaincodes
	var policyGroups []pgresolver.Group
	for _, ccID := range key.chaincodeIDs {
		policyGroup, err := s.getPolicyGroupForCC(key.channelID, ccID, channelPeers)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("error retrieving signature policy for chaincode [%s] on channel [%s]", ccID, key.channelID))
		}
		policyGroups = append(policyGroups, policyGroup)
	}

	// Perform an 'and' operation on all of the peer groups
	aggregatePolicyGroup, err := pgresolver.NewGroupOfGroups(policyGroups).Nof(int32(len(policyGroups)))
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("error computing signature policy for chaincode(s) [%v] on channel [%s]", key.chaincodeIDs, key.channelID))
	}

	// Create the resolver
	if resolver, err = pgresolver.NewPeerGroupResolver(aggregatePolicyGroup, s.pgLBP); err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("error creating peer group resolver for chaincodes [%v] on channel [%s]", key.chaincodeIDs, key.channelID))
	}

	s.pgResolvers[key.String()] = resolver

	return resolver, nil
}

func (s *selectionService) getPolicyGroupForCC(channelID string, ccID string, channelPeers []fab.Peer) (pgresolver.Group, error) {
	sigPolicyEnv, err := s.ccPolicyProvider.GetChaincodePolicy(ccID)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("error querying chaincode [%s] on channel [%s]", ccID, channelID))
	}

	return pgresolver.NewSignaturePolicyCompiler(
		func(mspID string) []fab.Peer {
			return s.getAvailablePeers(channelPeers, mspID)
		}).Compile(sigPolicyEnv)
}

func (s *selectionService) getAvailablePeers(channelPeers []fab.Peer, mspID string) []fab.Peer {
	var peers []fab.Peer
	for _, peer := range channelPeers {
		if string(peer.MSPID()) == mspID {
			peers = append(peers, peer)
		}
	}

	str := ""
	for i, peer := range peers {
		str += peer.URL()
		if i+1 < len(peers) {
			str += ","
		}
	}
	logger.Debugf("Available peers:\n%s\n", str)

	return peers
}

type resolverKey struct {
	channelID    string
	chaincodeIDs []string
	key          string
}

func (k *resolverKey) String() string {
	return k.key
}

func newResolverKey(channelID string, chaincodeIDs ...string) *resolverKey {
	arr := chaincodeIDs[:]
	sort.Strings(arr)

	key := channelID + "-"
	for i, s := range arr {
		key += s
		if i+1 < len(arr) {
			key += ":"
		}
	}
	return &resolverKey{channelID: channelID, chaincodeIDs: arr, key: key}
}
