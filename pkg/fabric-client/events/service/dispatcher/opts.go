/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import "time"

// Opts contains options for the event service dispatcher
type Opts struct {
	EvtConsumerBufferSize uint
	EvtConsumerTimeout    time.Duration
}

// DefaultOpts returns default options for the event service dispatcher
func DefaultOpts() *Opts {
	return &Opts{
		EvtConsumerBufferSize: 100,
		EvtConsumerTimeout:    500 * time.Millisecond,
	}
}

// EventConsumerBufferSize is the size of the registered consumer's event channel.
func (o *Opts) EventConsumerBufferSize() uint {
	return o.EvtConsumerBufferSize
}

// EventConsumerTimeout is the timeout when sending events to a registered consumer.
// If < 0, if buffer full, unblocks immediately and does not send.
// If 0, if buffer full, will block and guarantee the event will be sent out.
// If > 0, if buffer full, blocks util timeout.
func (o *Opts) EventConsumerTimeout() time.Duration {
	return o.EvtConsumerTimeout
}
