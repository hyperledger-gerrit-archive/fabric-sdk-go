/*
Copyright IBM Corp, SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	ehpb "github.com/hyperledger/fabric/protos/peer"
)

//EventsClient holds the stream and adapter for consumer to work with
type EventsClient interface {
	RegisterAsync(ies []*ehpb.Interest) error
	UnregisterAsync(ies []*ehpb.Interest) error
	Unregister(ies []*ehpb.Interest) error
	Recv() (*ehpb.Event, error)
	Start() error
	Stop() error
}
