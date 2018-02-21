/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/client/lbp"
	eventservice "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events/service"
)

// Opts contains the options for the event client
type Opts struct {
	eventservice.Opts

	// From dispatcher.Opts
	LBP lbp.LoadBalancePolicy

	Reconn                  bool
	MaxConnAttempts         uint
	MaxReconnAttempts       uint
	ReconnInitialDelay      time.Duration
	TimeBetweenConnAttempts time.Duration
	ConnEventCh             chan *apifabclient.ConnectionEvent
	RespTimeout             time.Duration
}

// DefaultOpts returns client options set to default values
func DefaultOpts() *Opts {
	return &Opts{
		Opts:                    *eventservice.DefaultOpts(),
		LBP:                     lbp.NewRoundRobin(),
		Reconn:                  true,
		MaxConnAttempts:         1,
		MaxReconnAttempts:       0, // Try forever
		ReconnInitialDelay:      0,
		TimeBetweenConnAttempts: 5 * time.Second,
		RespTimeout:             5 * time.Second,
	}
}

// LoadBalancePolicy returns the load-balance policy to use when
// choosing an event server endpoint from a set of endpoints
func (o *Opts) LoadBalancePolicy() lbp.LoadBalancePolicy {
	return o.LBP
}

// Reconnect indicates whether the client should automatically attempt to reconnect
// to the serever after a connection has been lost
func (o *Opts) Reconnect() bool {
	return o.Reconn
}

// MaxConnectAttempts is the maximum number of times that the client will attempt
// to connect to the server. If set to 0 then the client will try until it is stopped.
func (o *Opts) MaxConnectAttempts() uint {
	return o.MaxConnAttempts
}

// MaxReconnectAttempts is the maximum number of times that the client will attempt
// to reconnect to the server after a connection has been lost. If set to 0 then the
// client will try until it is stopped.
func (o *Opts) MaxReconnectAttempts() uint {
	return o.MaxReconnAttempts
}

// ReconnectInitialDelay is the initial delay before attempting to reconnect.
func (o *Opts) ReconnectInitialDelay() time.Duration {
	return o.ReconnInitialDelay
}

// TimeBetweenConnectAttempts is the time between connection attempts.
func (o *Opts) TimeBetweenConnectAttempts() time.Duration {
	return o.TimeBetweenConnAttempts
}

// ConnectEventCh is the channel that is to receive connection events, i.e. when the client connects and/or
// disconnects from the channel event service.
func (o *Opts) ConnectEventCh() chan *apifabclient.ConnectionEvent {
	return o.ConnEventCh
}

// ResponseTimeout is the timeout when waiting for a response from the event server
func (o *Opts) ResponseTimeout() time.Duration {
	return o.RespTimeout
}
