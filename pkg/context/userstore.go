/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/pkg/errors"
)

var (
	// ErrUserNotFound indicates the user was not found
	ErrUserNotFound = errors.New("user not found")
)

// UserKey is a lookup key in UserStore
type UserKey struct {
	MspID string
	Name  string
}

// UserStore is responsible for User persistence
type UserStore interface {
	Store(User) error
	Load(UserKey) (User, error)
}
