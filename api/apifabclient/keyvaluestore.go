/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

// KeyValueStore ...
/**
 * Abstract class for a Key-Value store. The Chain class uses this store
 * to save sensitive information such as authenticated user's private keys,
 * certificates, etc.
 *
 */
type KeyValueStore interface {
	/**
	 * Get the value associated with name.
	 *
	 * @param {interface{}} key
	 * @returns {interface{}}
	 */
	Value(key interface{}) (interface{}, error)

	/**
	 * Set the value associated with name.
	 * @param {interface{}} key
	 * @param {interface{}} value to save
	 */
	SetValue(key interface{}, value interface{}) error
}
