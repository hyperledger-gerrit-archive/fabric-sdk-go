/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defcore

import (
	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
)

//LoggerProvider returns logging provider for SDK logger
func LoggerProvider() apilogging.LoggerProvider {
	return deflogger.LoggerProvider()
}
