/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package completion

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
)

var logger = logging.NewLogger("fabsdk/util")

type Handle interface {
	Closed() <-chan bool
	Done()
}

type Handler struct {
	done    bool
	mutex   sync.Mutex
	wg      sync.WaitGroup
	handles []*CompletionHandle
}

func New() *Handler {
	return &Handler{}
}

func (c *Handler) Register() (Handle, error) {
	logger.Infof("Registering new handle...")

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.done {
		return nil, errors.New("already done")
	}

	h := newHandle(&c.wg)
	c.handles = append(c.handles, h)
	c.wg.Add(1)

	return h, nil
}

func (c *Handler) Done() {
	logger.Infof("Handler done - notifying all handles...")

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.done {
		// Already done
		return
	}

	for _, h := range c.handles {
		logger.Infof("... notifying handle...")
		h.notify()
	}

	logger.Infof("... waiting for all handles to finish...")
	c.wg.Wait()
	logger.Infof("... all handles are done.")
}

type CompletionHandle struct {
	done    int32
	eventch chan bool
	wg      *sync.WaitGroup
}

// NewHandle returns a completion handle. The caller must
// call Close when it is done with the handle.
func NewHandle() *CompletionHandle {
	var wg sync.WaitGroup
	wg.Add(1)
	return newHandle(&wg)
}

func newHandle(wg *sync.WaitGroup) *CompletionHandle {
	return &CompletionHandle{
		eventch: make(chan bool, 1),
		wg:      wg,
	}
}

// Close notifies the resource that the handle is closed
// and waits for the resource to finish handling the event.
func (h *CompletionHandle) Close() {
	h.notify()
	h.wg.Wait()
}

// Closed returns a channel that the resource may listen
// to for closed events.
func (h *CompletionHandle) Closed() <-chan bool {
	return h.eventch
}

// Done is called by the resources when it has finished.
// Calling Done() multiple times has no effect.
func (h *CompletionHandle) Done() {
	if atomic.CompareAndSwapInt32(&h.done, 0, 1) {
		logger.Infof("Handler is done")
		h.wg.Done()
	}
}

func (h *CompletionHandle) notify() {
	if !h.isDone() {
		logger.Infof("Notifying handler of done...")
		h.eventch <- true
		logger.Infof("... handler was notified of done...")
	}
}

func (h *CompletionHandle) isDone() bool {
	return atomic.LoadInt32(&h.done) == 1
}
