/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabapi

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// NewSDK wraps the NewSDK func moved to the pkg folder.
// Notice: this wrapper is deprecated and will be removed.
func NewSDK(options Options) (*fabsdk.FabricSDK, error) {
	opts := NewOptions(options)
	sdk, err := fabsdk.NewSDK(opts)
	if err != nil {
		return nil, err
	}

	logger.Info("fabapi.NewSDK is depecated - please use fabsdk.NewSDK")

	return sdk, nil
}
