/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package logging

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
)

type callerInfoKey struct {
	module string
	level  apilogging.Level
}

type callerInfo struct {
	sync.RWMutex
	showcaller map[callerInfoKey]bool
}

func (l *callerInfo) ShowCallerInfo(module string, level apilogging.Level) {
	l.Lock()
	defer l.Unlock()
	if l.showcaller == nil {
		l.showcaller = l.getDefaultCallerInfoSetting()
	}
	l.showcaller[callerInfoKey{module, level}] = true
}

func (l *callerInfo) HideCallerInfo(module string, level apilogging.Level) {
	l.Lock()
	defer l.Unlock()
	if l.showcaller == nil {
		l.showcaller = l.getDefaultCallerInfoSetting()
	}
	l.showcaller[callerInfoKey{module, level}] = false
}

func (l *callerInfo) IsCallerInfoEnabled(module string, level apilogging.Level) bool {
	l.RLock()
	defer l.RUnlock()
	showcaller, exists := l.showcaller[callerInfoKey{module, level}]
	if exists == false {
		//If no callerinfo setting exists, then look for default
		showcaller, exists = l.showcaller[callerInfoKey{"", level}]
		if exists == false {
			return true
		}
	}
	return showcaller
}

//getDefaultCallerInfoSetting default setting for callerinfo
func (l *callerInfo) getDefaultCallerInfoSetting() map[callerInfoKey]bool {
	return map[callerInfoKey]bool{
		callerInfoKey{"", CRITICAL}: true,
		callerInfoKey{"", ERROR}:    true,
		callerInfoKey{"", WARNING}:  true,
		callerInfoKey{"", INFO}:     true,
		callerInfoKey{"", DEBUG}:    true,
	}
}
