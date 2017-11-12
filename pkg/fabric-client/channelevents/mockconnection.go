// +build channelevents

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channelevents

import (
	"reflect"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
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

	// NoOpResult indicates that the operation should be ignored (i.e. just do nothing)
	// This should result in the client timing out waiting for a response.
	NoOpResult Result = "no-op"

	// InvalidChannelResult indicates that the operation should use an invalid channel ID
	InvalidChannelResult Result = "invalid-channel"
)

// MockConnection is a fake connection used for unit testing
type MockConnection struct {
	Operations       OperationMap
	AuthorizedEvents []eventType
	rcvch            chan interface{}
}

// NewMockConnection returns a new MockConnection using the given options
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

// Close implements the Connection interface
func (c *MockConnection) Close() {
	if c.rcvch != nil {
		close(c.rcvch)
		c.rcvch = nil
	}
}

// Send implements the Connection interface
func (c *MockConnection) Send(emsg *pb.ChannelServiceRequest) error {
	if c.rcvch == nil {
		return errors.New("mock connection not initialized")
	}

	switch e := emsg.Request.(type) {
	case *pb.ChannelServiceRequest_RegisterChannel:
		result, ok := c.Operations[RegisterChannel]
		if !ok || result.Result != NoOpResult {
			channelID := e.RegisterChannel.ChannelIds[0]
			if result.Result == InvalidChannelResult {
				channelID = "invalid"
			}
			c.ProduceEvent(c.newRegisterChannelResponse(channelID))
		}
	case *pb.ChannelServiceRequest_DeregisterChannel:
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

// Disconnect implements the Connection interface
func (c *MockConnection) Disconnect(err error) {
}

// Receive implements the Connection interface
func (c *MockConnection) Receive(eventch chan<- interface{}) {
	for {
		e, ok := <-c.rcvch
		if !ok {
			break
		}

		eventch <- e
	}
}

// ProduceEvent allows a unit test to inject an event
func (c *MockConnection) ProduceEvent(event interface{}) {
	go func() {
		c.rcvch <- event
	}()
}

func (c *MockConnection) newRegisterChannelResponse(channelID string) *pb.ChannelServiceResponse_Result {
	success := true
	errMsg := ""
	result, ok := c.Operations[RegisterChannel]
	if ok {
		success = result.Result == SucceedResult
		errMsg = result.ErrMsg
	}

	return &pb.ChannelServiceResponse_Result{
		Result: &pb.ChannelServiceResult{
			Success: success,
			Action:  "RegisterChannel",
			ChannelResults: []*pb.ChannelResult{
				&pb.ChannelResult{
					ChannelId:        channelID,
					ErrorMsg:         errMsg,
					RegisteredEvents: asStrings(c.AuthorizedEvents),
				},
			},
		},
	}
}

func asStrings(authEvents []eventType) []string {
	ret := make([]string, len(authEvents))
	for i, et := range authEvents {
		ret[i] = string(et)
	}
	return ret
}

func (c *MockConnection) newDeregisterChannelResponse(channelID string) *pb.ChannelServiceResponse_Result {
	success := true
	errMsg := ""
	result, ok := c.Operations[UnregisterChannel]
	if ok {
		success = result.Result == SucceedResult
		errMsg = result.ErrMsg
	}

	return &pb.ChannelServiceResponse_Result{
		Result: &pb.ChannelServiceResult{
			Success: success,
			Action:  "DeregisterChannel",
			ChannelResults: []*pb.ChannelResult{
				&pb.ChannelResult{
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
	return func(string, apifabclient.FabricClient, *apiconfig.PeerConfig) (Connection, error) {
		return conn, nil
	}
}

// ResultDesc describes the result of an operation and optional error string
type ResultDesc struct {
	Result Result
	ErrMsg string
}

// OperationMap maps an Operation to a ResultDesc
type OperationMap map[Operation]ResultDesc

// MockConnOpt applies an option to a MockConnection
type MockConnOpt interface {
	// Apply applies the option to the MockConnection
	Apply(conn *MockConnection)
}

// AuthorizedEventsOpt is a connection option that applies authorized events to the MockConnection
type AuthorizedEventsOpt struct {
	AuthorizedEvents []eventType
}

// Apply applies the option to the MockConnection
func (o *AuthorizedEventsOpt) Apply(conn *MockConnection) {
	conn.AuthorizedEvents = o.AuthorizedEvents
}

// NewAuthorizedEventsOpt returns a new AuthorizedEventsOpt
func NewAuthorizedEventsOpt(authorizedEvents ...eventType) *AuthorizedEventsOpt {
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
