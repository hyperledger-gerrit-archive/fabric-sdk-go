/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

// KeyValueStore is a generic key-value store interface.
type KeyValueStore interface {
	/**
	 * Get the value associated with the key.
	 * If not found returns (nil, nil)
	 *
	 * @param {interface{}} key
	 * @returns {interface{}}
	 */
	Value(key interface{}) (interface{}, error)

	/**
	 * Set the value associated with the key.
	 * If the value is nil, the entry is removed from the storage.
	 *
	 * @param {interface{}} key
	 * @param {interface{}} value to save
	 */
	SetValue(key interface{}, value interface{}) error
}
