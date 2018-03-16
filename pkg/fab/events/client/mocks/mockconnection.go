/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"sync"
	"sync/atomic"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/api"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

// Operation is the operation being performed
type Operation string

// Result is the result to take for a given operation
type Result string

const (
	// SucceedResult indicates that the operation should succeed
	SucceedResult Result = "succeed"

	// FailResult indicates that the operation should fail
	FailResult Result = "fail"

	// NoOpResult indicates that the operation should be ignored (i.e. just do nothing)
	// This should result in the client timing out waiting for a response.
	NoOpResult Result = "no-op"
)

// Attempt specifies the number of connection attempts
type Attempt uint

const (
	// FirstAttempt is the first attempt
	FirstAttempt Attempt = 1
	// SecondAttempt is the second attempt
	SecondAttempt Attempt = 2
	// ThirdAttempt is the third attempt
	ThirdAttempt Attempt = 3
	// FourthAttempt is the fourth attempt
	FourthAttempt Attempt = 4
)

// Outcome is the outcome of the attempt
type Outcome string

const (
	// ReconnectedOutcome means that the client reconnect
	ReconnectedOutcome Outcome = "reconnected"
	// ClosedOutcome means that the client was closed
	ClosedOutcome Outcome = "closed"
	// TimedOutOutcome means that the client timed out
	TimedOutOutcome Outcome = "timeout"
	// ConnectedOutcome means that the client connect
	ConnectedOutcome Outcome = "connected"
	// ErrorOutcome means that the operation resulted in an error
	ErrorOutcome Outcome = "error"
)

// Connection extends Connection and adds functions
// to allow simulating certain situations
type Connection interface {
	api.Connection

	// ProduceEvent send the given event to the event channel
	ProduceEvent(event interface{})
	// Result returns the result for the given operation
	Result(operation Operation) (ResultDesc, bool)
	// Ledger returns the mock ledger
	Ledger() servicemocks.Ledger
}

// ConnectionFactory creates a new mock connection
type ConnectionFactory func(opts ...Opt) Connection

// MockConnection is a mock connection used for unit testing
type MockConnection struct {
	producer   *servicemocks.MockProducer
	operations OperationMap
	producerch <-chan interface{}
	rcvch      chan interface{}
	closed     int32
}

// Opts contains mock connection options
type Opts struct {
	Ledger     servicemocks.Ledger
	Operations OperationMap
	Factory    ConnectionFactory
}

// NewMockConnection returns a new MockConnection using the given options
func NewMockConnection(opts ...Opt) *MockConnection {
	copts := &Opts{}
	for _, opt := range opts {
		opt(copts)
	}

	operations := copts.Operations
	if operations == nil {
		operations = make(map[Operation]ResultDesc)
	}

	if copts.Ledger == nil {
		panic("ledger is nil")
	}

	producer := servicemocks.NewMockProducer(copts.Ledger)

	c := &MockConnection{
		producer:   producer,
		producerch: producer.Register(),
		rcvch:      make(chan interface{}),
		operations: operations,
	}
	return c
}

// Close implements the MockConnection interface
func (c *MockConnection) Close() {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		// Already closed
		return
	}

	c.producer.Close()
	close(c.rcvch)
}

// Closed return true if the connection is closed
func (c *MockConnection) Closed() bool {
	return atomic.LoadInt32(&c.closed) == 1
}

// Receive implements the MockConnection interface
func (c *MockConnection) Receive(eventch chan<- interface{}) {
	for {
		select {
		case e, ok := <-c.producerch:
			if !ok {
				return
			}
			eventch <- e
		case e, ok := <-c.rcvch:
			if !ok {
				return
			}
			eventch <- e
		}
	}
}

// ProduceEvent send the given event to the event channel
func (c *MockConnection) ProduceEvent(event interface{}) {
	go func() {
		c.rcvch <- event
	}()
}

// Result returns the result for the given operation
func (c *MockConnection) Result(operation Operation) (ResultDesc, bool) {
	op, ok := c.operations[operation]
	return op, ok
}

// Ledger returns the mock ledger
func (c *MockConnection) Ledger() servicemocks.Ledger {
	return c.producer.Ledger()
}

// ProviderFactory creates various mock MockConnection Providers
type ProviderFactory struct {
	connection Connection
	mtx        sync.RWMutex
}

// NewProviderFactory returns a new ProviderFactory
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

// Connection returns the connection
func (cp *ProviderFactory) Connection() Connection {
	cp.mtx.RLock()
	defer cp.mtx.RUnlock()
	return cp.connection
}

// Provider returns a connection provider that always returns the given connection
func (cp *ProviderFactory) Provider(conn Connection) api.ConnectionProvider {
	return func(context.Client, fab.ChannelCfg, fab.Peer) (api.Connection, error) {
		return conn, nil
	}
}

// FlakeyProvider creates a connection provider that returns a connection according to the given
// connection attempt results. The results tell the connection provider whether or not to fail,
// to return a connection, what authorization to give the connection, etc.
func (cp *ProviderFactory) FlakeyProvider(connAttemptResults ConnectAttemptResults, opts ...Opt) api.ConnectionProvider {
	var connectAttempt Attempt
	return func(context.Client, fab.ChannelCfg, fab.Peer) (api.Connection, error) {
		connectAttempt++

		_, ok := connAttemptResults[connectAttempt]
		if !ok {
			return nil, errors.New("simulating failed connection attempt")
		}

		cp.mtx.Lock()
		defer cp.mtx.Unlock()

		copts := &Opts{}
		for _, opt := range opts {
			opt(copts)
		}
		factory := copts.Factory
		if factory == nil {
			cp.connection = NewMockConnection(opts...)
		} else {
			cp.connection = factory(opts...)
		}

		return cp.connection, nil
	}
}

// ConnectResult contains the data to use for the N'th connection attempt
type ConnectResult struct {
	Attempt Attempt
	Result  Result
}

// NewConnectResult returns a new ConnectResult
func NewConnectResult(attempt Attempt, result Result) ConnectResult {
	return ConnectResult{Attempt: attempt, Result: result}
}

// ConnectAttemptResults maps a connection attempt to a connection result
type ConnectAttemptResults map[Attempt]ConnectResult

// NewConnectResults returns a new ConnectAttemptResults
func NewConnectResults(results ...ConnectResult) ConnectAttemptResults {
	mapResults := make(map[Attempt]ConnectResult)
	for _, r := range results {
		mapResults[r.Attempt] = r
	}
	return mapResults
}

// ResultDesc describes the result of an operation and optional error string
type ResultDesc struct {
	Result Result
	ErrMsg string
}

// OperationMap maps an Operation to a ResultDesc
type OperationMap map[Operation]ResultDesc

// Opt applies an option
type Opt func(opts *Opts)

// OperationResult contains the result of an operation
type OperationResult struct {
	Operation  Operation
	Result     Result
	ErrMessage string
}

// NewResult returns a new OperationResult
func NewResult(operation Operation, result Result, errMsg ...string) *OperationResult {
	msg := ""
	if len(errMsg) > 0 {
		msg = errMsg[0]
	}
	return &OperationResult{
		Operation:  operation,
		Result:     result,
		ErrMessage: msg,
	}
}

// WithLedger provides the mock connection with a ledger
func WithLedger(ledger servicemocks.Ledger) Opt {
	return func(opts *Opts) {
		opts.Ledger = ledger
	}
}

// WithResults specifies the results for one or more operations
func WithResults(funcResults ...*OperationResult) Opt {
	return func(opts *Opts) {
		opts.Operations = make(map[Operation]ResultDesc)
		for _, fr := range funcResults {
			opts.Operations[fr.Operation] = ResultDesc{Result: fr.Result, ErrMsg: fr.ErrMessage}
		}
	}
}

// WithFactory specifies the connection factory for creating new mock connections
func WithFactory(factory ConnectionFactory) Opt {
	return func(opts *Opts) {
		opts.Factory = factory
	}
}

func newDeliverStatusResponse(status cb.Status) *pb.DeliverResponse_Status {
	return &pb.DeliverResponse_Status{
		Status: status,
	}
}