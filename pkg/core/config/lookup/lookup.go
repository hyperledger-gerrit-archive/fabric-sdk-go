/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lookup

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
)

//New providers lookup wrapper around given backend
func New(coreBackend core.ConfigBackend) *ConfigLookup {
	return &ConfigLookup{backend: coreBackend}
}

//unmarshalOpts opts for unmarshal key function
type unmarshalOpts struct {
	hookFunc mapstructure.DecodeHookFunc
}

// UnmarshalOption describes a functional parameter unmarshaling
type UnmarshalOption func(o *unmarshalOpts)

// WithUnmarshalHookFunction provides an option to pass Custom Decode Hook Func
// for unmarshaling
func WithUnmarshalHookFunction(hookFunction mapstructure.DecodeHookFunc) UnmarshalOption {
	return func(o *unmarshalOpts) {
		o.hookFunc = hookFunction
	}
}

//ConfigLookup is wrapper for core.ConfigBackend which performs key lookup and unmarshalling
type ConfigLookup struct {
	backend core.ConfigBackend
}

//Lookup returns value for given key
func (c *ConfigLookup) Lookup(key string) (interface{}, bool) {
	return c.backend.Lookup(key)
}

//GetBool returns bool value for given key
func (c *ConfigLookup) GetBool(key string) bool {
	value, ok := c.backend.Lookup(key)
	if !ok {
		return false
	}
	return cast.ToBool(value)
}

//GetString returns string value for given key
func (c *ConfigLookup) GetString(key string) string {
	value, ok := c.backend.Lookup(key)
	if !ok {
		return ""
	}
	return cast.ToString(value)
}

//GetInt returns int value for given key
func (c *ConfigLookup) GetInt(key string) int {
	value, ok := c.backend.Lookup(key)
	if !ok {
		return 0
	}
	return cast.ToInt(value)
}

//GetDuration returns time.Duration value for given key
func (c *ConfigLookup) GetDuration(key string) time.Duration {
	value, ok := c.backend.Lookup(key)
	if !ok {
		return 0
	}
	return cast.ToDuration(value)
}

//UnmarshalKey unmarshals value for given key to rawval type
func (c *ConfigLookup) UnmarshalKey(key string, rawVal interface{}, opts ...UnmarshalOption) error {
	value, ok := c.backend.Lookup(key)
	if !ok {
		return nil
	}

	//mandatory hook func
	hookFn := mapstructure.StringToTimeDurationHookFunc()

	//check for opts
	unmarshalOpts := unmarshalOpts{}
	for _, param := range opts {
		param(&unmarshalOpts)
	}

	//compose multiple hook funcs to one if found in opts
	if unmarshalOpts.hookFunc != nil {
		hookFn = mapstructure.ComposeDecodeHookFunc(hookFn, unmarshalOpts.hookFunc)
	}

	//build decoder
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: hookFn,
		Result:     rawVal,
	})
	if err != nil {
		return err
	}

	//decode
	return decoder.Decode(value)
}
