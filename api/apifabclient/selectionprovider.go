/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

// SelectionProvider is used to select peers for endorsement
type SelectionProvider interface {
	NewSelectionService(ccPolicyService CCPolicyService) (SelectionService, error)
}

// SelectionService selects peers for endorsement and commit events
type SelectionService interface {
	// GetEndorsersForChaincode returns a set of peers that should satisfy the endorsement
	// policies of all of the given chaincodes
	GetEndorsersForChaincode(channelID string, channelPeers []Peer, chaincodeIDs ...string) ([]Peer, error)
}
