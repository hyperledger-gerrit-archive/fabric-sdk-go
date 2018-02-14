/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	idmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/idmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/api/core/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/chclient"
)

// SessionContext primarily represents the session and identity context
type SessionContext interface {
	identity.Context
}

// Client represents the Client APIs supported by the SDK
type Client interface {
	IdentityMgmt() (idmgmt.IdentityMgmtClient, error)
	ChannelMgmt() (chmgmt.ChannelMgmtClient, error)
	ResourceMgmt() (resmgmt.ResourceMgmtClient, error)
	Channel(id string) (chclient.ChannelClient, error)
}
