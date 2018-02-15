/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/chclient"
)

// SessionContext primarily represents the session and identity context
type SessionContext interface {
	apifabclient.IdentityContext
}

// Client represents the Client APIs supported by the SDK
type Client interface {
	Channel(id string) (chclient.ChannelClient, error)
}
