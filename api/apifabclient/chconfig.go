/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	msp "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
)

// ChannelConfig allows for interaction with peer regarding channel configuration
type ChannelConfig interface {

	// Query channel configuration
	Query() (ChannelCfg, error)
}

// ChannelCfg contains channel configuration
type ChannelCfg interface {
	Name() string
	Msps() []*msp.MSPConfig
	AnchorPeers() []*OrgAnchorPeer
	Orderers() []string
	Versions() *Versions
}

// ChannelMemberID helps identify a channel's members
type ChannelMemberID interface {
	// Validate if the given ID was issued by the channel's members
	Validate(serializedID []byte) error
	// Verify the given signature
	Verify(serializedID []byte, msg []byte, sig []byte) error
}

// Versions ...
type Versions struct {
	ReadSet  *common.ConfigGroup
	WriteSet *common.ConfigGroup
	Channel  *common.ConfigGroup
}
