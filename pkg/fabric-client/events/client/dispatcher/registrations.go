/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

// ConnectionReg is a connection registration
type ConnectionReg struct {
	Eventch chan<- *apifabclient.ConnectionEvent
}
