/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api/core"

// SigningManager signs object with provided key
type SigningManager interface {
	Sign([]byte, core.Key) ([]byte, error)
}
