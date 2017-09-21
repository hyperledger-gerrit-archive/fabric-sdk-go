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
	event chan<- *fab.RegistrationResponse
}

type connectionRegistration struct {
	event chan<- *fab.ConnectionEvent
}

type blockRegistration struct {
	event chan<- *fab.BlockEvent
}

type ccRegistration struct {
	ccID        string
	eventFilter string
	eventRegExp *regexp.Regexp
	event       chan<- *fab.CCEvent
}

type txRegistration struct {
	txID  string
	event chan<- *fab.TxStatusEvent
}
