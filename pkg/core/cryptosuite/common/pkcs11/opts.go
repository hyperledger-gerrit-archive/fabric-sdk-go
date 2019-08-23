/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkcs11

import "time"

const (
	defaultSessionCacheSize         = 10
	defaultOpenSessionRetryAttempts = 10
	defaultOpenSessionRetryInterval = 10 * time.Second
)

//ctxOpts options for conext handler
type ctxOpts struct {
	//sessionCacheSize size of session cache pool
	sessionCacheSize int
	//openSessionRetry number of retry attempts for open session logic
	openSessionRetryAttempts int
	//openSessionRetryInterval time interval to wait before retrying to open session
	openSessionRetryInterval time.Duration
}

//Options for PKCS11 ContextHandle
type Options func(opts *ctxOpts)

func getCtxOpts(opts ...Options) ctxOpts {
	ctxOpts := ctxOpts{}
	for _, option := range opts {
		option(&ctxOpts)
	}

	if ctxOpts.sessionCacheSize == 0 {
		ctxOpts.sessionCacheSize = defaultSessionCacheSize
	}

	if ctxOpts.openSessionRetryAttempts == 0 {
		ctxOpts.openSessionRetryAttempts = defaultOpenSessionRetryAttempts
	}

	if ctxOpts.openSessionRetryInterval == 0 {
		ctxOpts.openSessionRetryInterval = defaultOpenSessionRetryInterval
	}

	return ctxOpts
}

//WithSessionCacheSize size of session cache pool
func WithSessionCacheSize(size int) Options {
	return func(o *ctxOpts) {
		o.sessionCacheSize = size
	}
}

//WithOpenSessionRetry number of retry for open session logic
func WithOpenSessionRetry(count int, interval time.Duration) Options {
	return func(o *ctxOpts) {
		o.openSessionRetryAttempts = count
		o.openSessionRetryInterval = interval
	}
}
