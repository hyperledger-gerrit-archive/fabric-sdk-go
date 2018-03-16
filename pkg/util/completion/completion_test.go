/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package completion

import (
	"fmt"
	"testing"
	"time"
)

func TestHandle(t *testing.T) {
	handle := NewHandle()

	go func() {
		defer handle.Done()

		select {
		case <-handle.Closed():
			fmt.Printf("Done!\n")
		}
	}()

	time.Sleep(time.Second)
	handle.Close()
	fmt.Printf("Handler is done!\n")
}

func TestHandler(t *testing.T) {
	c := New()

	for i := 0; i < 10; i++ {
		num := i
		go func() {
			handle, err := c.Register()
			if err != nil {
				// Log error
				return
			}
			defer handle.Done()

			select {
			case <-handle.Closed():
				time.Sleep(time.Duration(num*250) * time.Millisecond)
				fmt.Printf("%d is done!\n", num)
			}
		}()
	}

	time.Sleep(2 * time.Second)
	c.Done()
	fmt.Printf("Everyone is done!\n")
}
