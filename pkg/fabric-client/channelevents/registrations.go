/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

type channelRegistration struct {
	respch chan<- *fab.RegistrationResponse
}

type connectionRegistration struct {
	eventch chan<- *fab.ConnectionEvent
}

type filteredBlockRegistration struct {
	eventch chan<- *fab.FilteredBlockEvent
}
