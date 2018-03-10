// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp (interfaces: CAClient,IdentityManager,Providers)

// Package mock_msp is a generated GoMock package.
package mock_msp

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	msp "github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
)

// MockCAClient is a mock of CAClient interface
type MockCAClient struct {
	ctrl     *gomock.Controller
	recorder *MockCAClientMockRecorder
}

// MockCAClientMockRecorder is the mock recorder for MockCAClient
type MockCAClientMockRecorder struct {
	mock *MockCAClient
}

// NewMockCAClient creates a new mock instance
func NewMockCAClient(ctrl *gomock.Controller) *MockCAClient {
	mock := &MockCAClient{ctrl: ctrl}
	mock.recorder = &MockCAClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCAClient) EXPECT() *MockCAClientMockRecorder {
	return m.recorder
}

// CAName mocks base method
func (m *MockCAClient) CAName() string {
	ret := m.ctrl.Call(m, "CAName")
	ret0, _ := ret[0].(string)
	return ret0
}

// CAName indicates an expected call of CAName
func (mr *MockCAClientMockRecorder) CAName() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CAName", reflect.TypeOf((*MockCAClient)(nil).CAName))
}

// Enroll mocks base method
func (m *MockCAClient) Enroll(arg0, arg1 string) error {
	ret := m.ctrl.Call(m, "Enroll", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Enroll indicates an expected call of Enroll
func (mr *MockCAClientMockRecorder) Enroll(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enroll", reflect.TypeOf((*MockCAClient)(nil).Enroll), arg0, arg1)
}

// Reenroll mocks base method
func (m *MockCAClient) Reenroll(arg0 string) error {
	ret := m.ctrl.Call(m, "Reenroll", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reenroll indicates an expected call of Reenroll
func (mr *MockCAClientMockRecorder) Reenroll(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reenroll", reflect.TypeOf((*MockCAClient)(nil).Reenroll), arg0)
}

// Register mocks base method
func (m *MockCAClient) Register(arg0 *msp.RegistrationRequest) (string, error) {
	ret := m.ctrl.Call(m, "Register", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Register indicates an expected call of Register
func (mr *MockCAClientMockRecorder) Register(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockCAClient)(nil).Register), arg0)
}

// Revoke mocks base method
func (m *MockCAClient) Revoke(arg0 *msp.RevocationRequest) (*msp.RevocationResponse, error) {
	ret := m.ctrl.Call(m, "Revoke", arg0)
	ret0, _ := ret[0].(*msp.RevocationResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Revoke indicates an expected call of Revoke
func (mr *MockCAClientMockRecorder) Revoke(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Revoke", reflect.TypeOf((*MockCAClient)(nil).Revoke), arg0)
}

// MockIdentityManager is a mock of IdentityManager interface
type MockIdentityManager struct {
	ctrl     *gomock.Controller
	recorder *MockIdentityManagerMockRecorder
}

// MockIdentityManagerMockRecorder is the mock recorder for MockIdentityManager
type MockIdentityManagerMockRecorder struct {
	mock *MockIdentityManager
}

// NewMockIdentityManager creates a new mock instance
func NewMockIdentityManager(ctrl *gomock.Controller) *MockIdentityManager {
	mock := &MockIdentityManager{ctrl: ctrl}
	mock.recorder = &MockIdentityManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIdentityManager) EXPECT() *MockIdentityManagerMockRecorder {
	return m.recorder
}

// GetSigningIdentity mocks base method
func (m *MockIdentityManager) GetSigningIdentity(arg0 string) (*msp.SigningIdentity, error) {
	ret := m.ctrl.Call(m, "GetSigningIdentity", arg0)
	ret0, _ := ret[0].(*msp.SigningIdentity)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSigningIdentity indicates an expected call of GetSigningIdentity
func (mr *MockIdentityManagerMockRecorder) GetSigningIdentity(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSigningIdentity", reflect.TypeOf((*MockIdentityManager)(nil).GetSigningIdentity), arg0)
}

// GetUser mocks base method
func (m *MockIdentityManager) GetUser(arg0 string) (msp.User, error) {
	ret := m.ctrl.Call(m, "GetUser", arg0)
	ret0, _ := ret[0].(msp.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser
func (mr *MockIdentityManagerMockRecorder) GetUser(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockIdentityManager)(nil).GetUser), arg0)
}

// MockProviders is a mock of Providers interface
type MockProviders struct {
	ctrl     *gomock.Controller
	recorder *MockProvidersMockRecorder
}

// MockProvidersMockRecorder is the mock recorder for MockProviders
type MockProvidersMockRecorder struct {
	mock *MockProviders
}

// NewMockProviders creates a new mock instance
func NewMockProviders(ctrl *gomock.Controller) *MockProviders {
	mock := &MockProviders{ctrl: ctrl}
	mock.recorder = &MockProvidersMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockProviders) EXPECT() *MockProvidersMockRecorder {
	return m.recorder
}

// IdentityManager mocks base method
func (m *MockProviders) IdentityManager(arg0 string) (msp.IdentityManager, bool) {
	ret := m.ctrl.Call(m, "IdentityManager", arg0)
	ret0, _ := ret[0].(msp.IdentityManager)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// IdentityManager indicates an expected call of IdentityManager
func (mr *MockProvidersMockRecorder) IdentityManager(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IdentityManager", reflect.TypeOf((*MockProviders)(nil).IdentityManager), arg0)
}
