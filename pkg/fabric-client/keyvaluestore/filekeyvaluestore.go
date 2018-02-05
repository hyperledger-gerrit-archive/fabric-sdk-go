/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package keyvaluestore

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// KeySerializer converts a key to a unique fila path
type KeySerializer func(storePath string, key interface{}) (string, error)

// ValueSerializer serializes a value into a byte array
type ValueSerializer func(value interface{}) ([]byte, error)

// ValueDeserializer de-serializes a value from a byte array
type ValueDeserializer func(value []byte) (interface{}, error)

// FileKeyValueStore stores each value into a separate file.
// KeySerializer maps a key to a unique file path (raletive to the store path)
// ValueSerializer and ValueDeserializer serializes/de-serializes a value
// to and from a byte array that is stored in the path derived from the key.
type FileKeyValueStore struct {
	path              string
	keySerializer     KeySerializer
	valueSerializer   ValueSerializer
	valueDeserializer ValueDeserializer
}

// FileKeyValueStoreOptions allow overriding store defaults
type FileKeyValueStoreOptions struct {
	// Store path, mandatory
	Path string
	// Optional. If not provided, default key serializer is used.
	KeySerializer KeySerializer
	// Optional. If not provided, default value serializer is used.
	ValueSerializer ValueSerializer
	// Optional. If not provided, default value de-serializer is used.
	ValueDeserializer ValueDeserializer
}

// Default key serializer
func defaultKeySerializer(storePath string, key interface{}) (string, error) {
	keyString, ok := key.(string)
	if !ok {
		return "", errors.New("converting key to string failed")
	}
	return path.Join(storePath, keyString), nil
}

// Default value serializer
func defaultValueSerializer(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	valueBytes, ok := value.([]byte)
	if !ok {
		return nil, errors.New("converting value to byte array failed")
	}
	return valueBytes, nil
}

// Default value de-serializer
func defaultValueDeserializer(value []byte) (interface{}, error) {
	return value, nil
}

// GetPath returns the store path
func (fkvs *FileKeyValueStore) GetPath() string {
	return fkvs.path
}

// NewFileKeyValueStore ...
func NewFileKeyValueStore(opts *FileKeyValueStoreOptions) (*FileKeyValueStore, error) {
	if opts == nil {
		return nil, errors.New("FileKeyValueStoreOptions is nil")
	}
	if opts.Path == "" {
		return nil, errors.New("FileKeyValueStore path is empty")
	}
	if opts.KeySerializer == nil {
		opts.KeySerializer = defaultKeySerializer
	}
	if opts.ValueSerializer == nil {
		opts.ValueSerializer = defaultValueSerializer
	}
	if opts.ValueDeserializer == nil {
		opts.ValueDeserializer = defaultValueDeserializer
	}
	return &FileKeyValueStore{
		path:              opts.Path,
		keySerializer:     opts.KeySerializer,
		valueSerializer:   opts.ValueSerializer,
		valueDeserializer: opts.ValueDeserializer,
	}, nil
}

// Value ...
/**
 * Get the value associated with key.
 * @param {interface{}} key
 * @returns {interface{}} value
 */
func (fkvs *FileKeyValueStore) Value(key interface{}) (interface{}, error) {
	file, err := fkvs.keySerializer(fkvs.path, key)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, nil
	}
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return fkvs.valueDeserializer(bytes)
}

// SetValue ...
/**
 * Set the value associated with key.
 * @param {interface{}} key
 * @param {interface{}} value
 */
func (fkvs *FileKeyValueStore) SetValue(key interface{}, value interface{}) error {
	if key == nil {
		return errors.New("key is nil")
	}
	file, err := fkvs.keySerializer(fkvs.path, key)
	if err != nil {
		return err
	}
	if value == nil {
		_, err := os.Stat(file)
		if err != nil {
			if !os.IsNotExist(err) {
				return errors.Wrapf(err, "stat dir failed")
			}
			// Doesn't exist, OK
			return nil
		}
		return os.Remove(file)
	}
	valueBytes, err := fkvs.valueSerializer(value)
	if err != nil {
		return err
	}
	os.MkdirAll(path.Dir(file), 0700)
	return ioutil.WriteFile(file, valueBytes, 0600)
}
