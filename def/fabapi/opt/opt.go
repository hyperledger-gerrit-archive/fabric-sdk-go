/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package opt

// SDKOpts provides bootstrap setup
type SDKOpts struct {
	//ConfigFile to load from a predefined path
	ConfigFile string
	//ConfigBytes to load from an bytes array
	ConfigBytes []byte
}

// ConfigOpts provides setup parameters for Config
type ConfigOpts struct { // TODO (moved ConfigFile to SDKOpts to make setup easier for API consumer)
}

// StateStoreOpts provides setup parameters for KeyValueStore
type StateStoreOpts struct {
	Path string
}
