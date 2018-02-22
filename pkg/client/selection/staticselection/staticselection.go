/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticselection

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// SelectionProvider implements selection provider
type SelectionProvider struct {
	config apiconfig.Config
}

// NewSelectionProvider returns static selection provider
func NewSelectionProvider(config apiconfig.Config) (*SelectionProvider, error) {
	return &SelectionProvider{config: config}, nil
}

// selectionService implements static selection service
type selectionService struct {
}

// NewSelectionService creates a static selection service
func (p *SelectionProvider) NewSelectionService(channelID string) (context.SelectionService, error) {
	return &selectionService{}, nil
}

func (s *selectionService) GetEndorsersForChaincode(channelPeers []context.Peer,
	chaincodeIDs ...string) ([]context.Peer, error) {

	if len(chaincodeIDs) == 0 {
		return nil, errors.New("no chaincode IDs provided")
	}

	return channelPeers, nil
}
