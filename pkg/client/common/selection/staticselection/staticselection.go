/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticselection

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/client")

// SelectionProvider implements selection provider
type SelectionProvider struct {
	config core.Config
}

// New returns static selection provider
func New(config core.Config) (*SelectionProvider, error) {
	return &SelectionProvider{config: config}, nil
}

// selectionService implements static selection service
type selectionService struct {
}

// CreateSelectionService creates a static selection service
func (p *SelectionProvider) CreateSelectionService(channelID string) (fab.SelectionService, error) {
	return &selectionService{}, nil
}

func (s *selectionService) GetEndorsersForChaincode(channelPeers []fab.Peer,
	chaincodeIDs ...string) ([]fab.Peer, error) {

	if len(chaincodeIDs) == 0 {
		return nil, errors.New("no chaincode IDs provided")
	}

	return channelPeers, nil
}
