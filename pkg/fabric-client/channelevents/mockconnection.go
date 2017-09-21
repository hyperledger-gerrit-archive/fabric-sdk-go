// +build channelevents

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

// Operation is the operation being performed
type Operation string

// Result is the result to take for a given operation
type Result string

const (
	// RegisterChannel is the Register Channel operation (used in the OperationMap)
	RegisterChannel Operation = "reg-channel"

	// UnregisterChannel is the Unregister Channel operation (used in the OperationMap)
	UnregisterChannel Operation = "unreg-channel"

	// SucceedResult indicates that the operation should succeed
	SucceedResult Result = "succeed"

	// FailResult indicates that the operation should fail
	FailResult Result = "fail"

	// NoOpResult indicates that the operation should perform no operation
	NoOpResult Result = "no-op"

	// InvalidChannelResult indicates that the operation should use an invalid channel ID
	InvalidChannelResult Result = "invalid-channel"
)

// MockConnection is a fake connection used for unit testing
type MockConnection struct {
	Operations       OperationMap
	AuthorizedEvents []string
	rcvch            chan interface{}
}

// NewMockConnection returns a new MockConnection
func NewMockConnection(opts ...MockConnOpt) *MockConnection {
	conn := &MockConnection{
		rcvch:      make(chan interface{}),
		Operations: make(map[Operation]ResultDesc),
	}
	for _, opt := range opts {
		opt.Apply(conn)
	}
	return conn
}

// Close implements of the Connection interface
func (c *MockConnection) Close() {
	fmt.Printf("MockConnection.Close\n")
	if c.rcvch != nil {
		close(c.rcvch)
	}
}

// Send implements of the Connection interface
func (c *MockConnection) Send(emsg *pb.Event) error {
	fmt.Printf("MockConnection.Send - %v\n", emsg)
	if c.rcvch == nil {
		return errors.New("mock connection not initialized")
	}

	switch e := emsg.Event.(type) {
	case *pb.Event_RegisterChannel:
		result, ok := c.Operations[RegisterChannel]
		if !ok || result.Result != NoOpResult {
			channelID := e.RegisterChannel.ChannelIds[0]
			if result.Result == InvalidChannelResult {
				channelID = "invalid"
			}
			c.ProduceEvent(c.newRegisterChannelResponse(channelID))
		}
	case *pb.Event_DeregisterChannel:
		result, ok := c.Operations[UnregisterChannel]
		if !ok || result.Result != NoOpResult {
			channelID := e.DeregisterChannel.ChannelIds[0]
			if result.Result == InvalidChannelResult {
				channelID = "invalid"
			}
			c.ProduceEvent(c.newDeregisterChannelResponse(channelID))
		}
	default:
		return errors.Errorf("unsupported event type [%s]", reflect.TypeOf(e))
	}
	return nil
}

// Disconnect implements of the Connection interface
func (c *MockConnection) Disconnect(err error) {
	c.ProduceEvent(&disconnectedEvent{err: err})
}

// Receive implements of the Connection interface
func (c *MockConnection) Receive(eventch chan<- interface{}) {
	for {
		fmt.Printf("MockConnection listening for events...\n")
		e, ok := <-c.rcvch
		if !ok {
			break
		}

		fmt.Printf("MockConnection received event: %s. Sending to channel...\n", e)
		eventch <- e
		fmt.Printf("...sent.\n")
	}
	fmt.Printf("MockConnection exiting listener\n")
}

// ProduceEvent allows a unit test to inject an event into the Channel Event Client
func (c *MockConnection) ProduceEvent(event interface{}) {
	fmt.Printf("MockConnection.ProduceEvent: %s\n", event)
	go func() {
		c.rcvch <- event
	}()
}

func (c *MockConnection) newRegisterChannelResponse(channelID string) *pb.Event_ChannelServiceResponse {
	success := true
	errMsg := ""
	result, ok := c.Operations[RegisterChannel]
	if ok {
		success = result.Result == SucceedResult
		errMsg = result.ErrMsg
	}

	return &pb.Event_ChannelServiceResponse{
		ChannelServiceResponse: &pb.ChannelServiceResponse{
			Success: success,
			Action:  "RegisterChannel",
			ChannelServiceResults: []*pb.ChannelServiceResult{
				&pb.ChannelServiceResult{
					ChannelId:        channelID,
					ErrorMsg:         errMsg,
					AuthorizedEvents: c.AuthorizedEvents,
				},
			},
		},
	}
}

func (c *MockConnection) newDeregisterChannelResponse(channelID string) *pb.Event_ChannelServiceResponse {
	success := true
	errMsg := ""
	result, ok := c.Operations[UnregisterChannel]
	if ok {
		success = result.Result == SucceedResult
		errMsg = result.ErrMsg
	}

	return &pb.Event_ChannelServiceResponse{
		ChannelServiceResponse: &pb.ChannelServiceResponse{
			Success: success,
			Action:  "DeregisterChannel",
			ChannelServiceResults: []*pb.ChannelServiceResult{
				&pb.ChannelServiceResult{
					ChannelId: channelID,
					ErrorMsg:  errMsg,
				},
			},
		},
	}
}

// MockConnectionProviderFactory creates various mock Connection Providers
type MockConnectionProviderFactory struct {
	connection *MockConnection
	mtx        sync.RWMutex
}

func NewMockConnectionProviderFactory() *MockConnectionProviderFactory {
	return &MockConnectionProviderFactory{}
}

// NewMockConnectionProviderFactory returns a new MockConnectionProviderFactory
func (cp *MockConnectionProviderFactory) Connection() *MockConnection {
	cp.mtx.RLock()
	defer cp.mtx.RUnlock()
	return cp.connection
}

// Provider returns a connection provider that always returns the given connection
func (cp *MockConnectionProviderFactory) Provider(conn *MockConnection) ConnectionProvider {
	return func(apifabclient.FabricClient, *apiconfig.PeerConfig) (Connection, error) {
		return conn, nil
	}
}

// FlakeyProvider returns a connection provider that returns a connection according to the given
// connection attempt results. The results tell the connection provider whether or not to fail
// to return a connection; whath authorization to give the connection; etc.
func (cp *MockConnectionProviderFactory) FlakeyProvider(connAttemptResults ConnectAttemptResults) ConnectionProvider {
	var connectAttempt ConnectionAttempt
	return func(apifabclient.FabricClient, *apiconfig.PeerConfig) (Connection, error) {
		connectAttempt++

		result, ok := connAttemptResults[connectAttempt]
		if !ok {
			fmt.Printf("***** Failing to create a mock connection...")
			return nil, errors.New("simulating failed connection attempt")
		}

		fmt.Printf("***** Creating mock connection...\n")

		cp.mtx.Lock()
		defer cp.mtx.Unlock()
		cp.connection = NewMockConnection(NewAuthorizedEventsOpt(result.AuthorizedEvents...))
		return cp.connection, nil
	}
}

// ConnectionAttempt specifies the number of connection attempts
type ConnectionAttempt uint

// ConnectResult contains the data to use for the N'th connection attempt
type ConnectResult struct {
	Attempt          ConnectionAttempt
	AuthorizedEvents []string
}

// NewConnectResult returns a new ConnectResult
func NewConnectResult(attempt ConnectionAttempt, authorizedEvents ...string) ConnectResult {
	return ConnectResult{Attempt: attempt, AuthorizedEvents: authorizedEvents}
}

// ConnectAttemptResults maps a connection attemp to the connection result to use
type ConnectAttemptResults map[ConnectionAttempt]ConnectResult

// NewConnectResults returns a new ConnectAttemptResults
func NewConnectResults(results ...ConnectResult) ConnectAttemptResults {
	mapResults := make(map[ConnectionAttempt]ConnectResult)
	for _, r := range results {
		mapResults[r.Attempt] = r
	}
	return mapResults
}

// ResultDesc describes the result of a operation and optional error string to use
type ResultDesc struct {
	Result Result
	ErrMsg string
}

// OperationMap maps a Operation to a ResultDesc
type OperationMap map[Operation]ResultDesc

// MockConnOpt applies an option to a MockConnection
type MockConnOpt interface {
	// Apply applies the option to the MockConnection
	Apply(conn *MockConnection)
}

// AuthorizedEventsOpt is a connection option that applies authorized events to the MockConnection
type AuthorizedEventsOpt struct {
	AuthorizedEvents []string
}

// Apply applies the option to the MockConnection
func (o *AuthorizedEventsOpt) Apply(conn *MockConnection) {
	conn.AuthorizedEvents = o.AuthorizedEvents
}

// NewAuthorizedEventsOpt returns a new AuthorizedEventsOpt
func NewAuthorizedEventsOpt(authorizedEvents ...string) *AuthorizedEventsOpt {
	return &AuthorizedEventsOpt{AuthorizedEvents: authorizedEvents}
}

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

// OperationResultsOpt is a connection option that indicates what to do for each operation
type OperationResultsOpt struct {
	Operations OperationMap
}

// Apply applies the option to the MockConnection
func (o *OperationResultsOpt) Apply(conn *MockConnection) {
	conn.Operations = o.Operations
}

// NewResultsOpt returns a new OperationResultsOpt
func NewResultsOpt(funcResults ...*OperationResult) *OperationResultsOpt {
	opt := &OperationResultsOpt{Operations: make(map[Operation]ResultDesc)}
	for _, fr := range funcResults {
		opt.Operations[fr.Operation] = ResultDesc{Result: fr.Result, ErrMsg: fr.ErrMessage}
	}
	return opt
}
