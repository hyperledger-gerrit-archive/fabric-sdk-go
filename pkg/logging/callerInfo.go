/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package logging

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
)

type callerInfo struct {
	sync.RWMutex
	showcaller map[apilogging.Level]bool
}

func (l *callerInfo) ShowCallerInfo(level apilogging.Level) {
	l.RLock()
	defer l.RUnlock()
	if l.showcaller == nil {
		l.showcaller = make(map[apilogging.Level]bool)
	}
	l.showcaller[level] = true
}

func (l *callerInfo) HideCallerInfo(level apilogging.Level) {
	l.Lock()
	defer l.Unlock()
	if l.showcaller == nil {
		l.showcaller = make(map[apilogging.Level]bool)
	}
	l.showcaller[level] = false
}

func (l *callerInfo) IsCallerInfoEnabled(level apilogging.Level) bool {
	showcaller, exists := l.showcaller[level]
	if exists == false {
		//If no configuration exists, then return false
		return false
	}
	return showcaller
}
