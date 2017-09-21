/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"regexp"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

type channelRegistration struct {
	respch chan<- *fab.RegistrationResponse
}

type connectionRegistration struct {
	eventch chan<- *fab.ConnectionEvent
}

type blockRegistration struct {
	eventch chan<- *fab.BlockEvent
}

type ccRegistration struct {
	ccID        string
	eventFilter string
	eventRegExp *regexp.Regexp
	eventch     chan<- *fab.CCEvent
}

type txRegistration struct {
	txID    string
	eventch chan<- *fab.TxStatusEvent
}
