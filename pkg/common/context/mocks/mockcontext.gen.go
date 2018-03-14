// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/hyperledger/fabric-sdk-go/pkg/common/context (interfaces: Providers,Client)

// Package mock_context is a generated GoMock package.
package mock_context

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	core "github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	fab "github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	msp "github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
)

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

// ChannelProvider mocks base method
func (m *MockProviders) ChannelProvider() fab.ChannelProvider {
	ret := m.ctrl.Call(m, "ChannelProvider")
	ret0, _ := ret[0].(fab.ChannelProvider)
	return ret0
}

// ChannelProvider indicates an expected call of ChannelProvider
func (mr *MockProvidersMockRecorder) ChannelProvider() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChannelProvider", reflect.TypeOf((*MockProviders)(nil).ChannelProvider))
}

// Config mocks base method
func (m *MockProviders) Config() core.Config {
	ret := m.ctrl.Call(m, "Config")
	ret0, _ := ret[0].(core.Config)
	return ret0
}

// Config indicates an expected call of Config
func (mr *MockProvidersMockRecorder) Config() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Config", reflect.TypeOf((*MockProviders)(nil).Config))
}

// CryptoSuite mocks base method
func (m *MockProviders) CryptoSuite() core.CryptoSuite {
	ret := m.ctrl.Call(m, "CryptoSuite")
	ret0, _ := ret[0].(core.CryptoSuite)
	return ret0
}

// CryptoSuite indicates an expected call of CryptoSuite
func (mr *MockProvidersMockRecorder) CryptoSuite() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CryptoSuite", reflect.TypeOf((*MockProviders)(nil).CryptoSuite))
}

// DiscoveryProvider mocks base method
func (m *MockProviders) DiscoveryProvider() fab.DiscoveryProvider {
	ret := m.ctrl.Call(m, "DiscoveryProvider")
	ret0, _ := ret[0].(fab.DiscoveryProvider)
	return ret0
}

// DiscoveryProvider indicates an expected call of DiscoveryProvider
func (mr *MockProvidersMockRecorder) DiscoveryProvider() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DiscoveryProvider", reflect.TypeOf((*MockProviders)(nil).DiscoveryProvider))
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

// InfraProvider mocks base method
func (m *MockProviders) InfraProvider() fab.InfraProvider {
	ret := m.ctrl.Call(m, "InfraProvider")
	ret0, _ := ret[0].(fab.InfraProvider)
	return ret0
}

// InfraProvider indicates an expected call of InfraProvider
func (mr *MockProvidersMockRecorder) InfraProvider() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InfraProvider", reflect.TypeOf((*MockProviders)(nil).InfraProvider))
}

// SelectionProvider mocks base method
func (m *MockProviders) SelectionProvider() fab.SelectionProvider {
	ret := m.ctrl.Call(m, "SelectionProvider")
	ret0, _ := ret[0].(fab.SelectionProvider)
	return ret0
}

// SelectionProvider indicates an expected call of SelectionProvider
func (mr *MockProvidersMockRecorder) SelectionProvider() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectionProvider", reflect.TypeOf((*MockProviders)(nil).SelectionProvider))
}

// SigningManager mocks base method
func (m *MockProviders) SigningManager() core.SigningManager {
	ret := m.ctrl.Call(m, "SigningManager")
	ret0, _ := ret[0].(core.SigningManager)
	return ret0
}

// SigningManager indicates an expected call of SigningManager
func (mr *MockProvidersMockRecorder) SigningManager() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SigningManager", reflect.TypeOf((*MockProviders)(nil).SigningManager))
}

// UserStore mocks base method
func (m *MockProviders) UserStore() msp.UserStore {
	ret := m.ctrl.Call(m, "UserStore")
	ret0, _ := ret[0].(msp.UserStore)
	return ret0
}

// UserStore indicates an expected call of UserStore
func (mr *MockProvidersMockRecorder) UserStore() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UserStore", reflect.TypeOf((*MockProviders)(nil).UserStore))
}

// MockClient is a mock of Client interface
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// ChannelProvider mocks base method
func (m *MockClient) ChannelProvider() fab.ChannelProvider {
	ret := m.ctrl.Call(m, "ChannelProvider")
	ret0, _ := ret[0].(fab.ChannelProvider)
	return ret0
}

// ChannelProvider indicates an expected call of ChannelProvider
func (mr *MockClientMockRecorder) ChannelProvider() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChannelProvider", reflect.TypeOf((*MockClient)(nil).ChannelProvider))
}

// Config mocks base method
func (m *MockClient) Config() core.Config {
	ret := m.ctrl.Call(m, "Config")
	ret0, _ := ret[0].(core.Config)
	return ret0
}

// Config indicates an expected call of Config
func (mr *MockClientMockRecorder) Config() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Config", reflect.TypeOf((*MockClient)(nil).Config))
}

// CryptoSuite mocks base method
func (m *MockClient) CryptoSuite() core.CryptoSuite {
	ret := m.ctrl.Call(m, "CryptoSuite")
	ret0, _ := ret[0].(core.CryptoSuite)
	return ret0
}

// CryptoSuite indicates an expected call of CryptoSuite
func (mr *MockClientMockRecorder) CryptoSuite() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CryptoSuite", reflect.TypeOf((*MockClient)(nil).CryptoSuite))
}

// DiscoveryProvider mocks base method
func (m *MockClient) DiscoveryProvider() fab.DiscoveryProvider {
	ret := m.ctrl.Call(m, "DiscoveryProvider")
	ret0, _ := ret[0].(fab.DiscoveryProvider)
	return ret0
}

// DiscoveryProvider indicates an expected call of DiscoveryProvider
func (mr *MockClientMockRecorder) DiscoveryProvider() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DiscoveryProvider", reflect.TypeOf((*MockClient)(nil).DiscoveryProvider))
}

// IdentityManager mocks base method
func (m *MockClient) IdentityManager(arg0 string) (msp.IdentityManager, bool) {
	ret := m.ctrl.Call(m, "IdentityManager", arg0)
	ret0, _ := ret[0].(msp.IdentityManager)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// IdentityManager indicates an expected call of IdentityManager
func (mr *MockClientMockRecorder) IdentityManager(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IdentityManager", reflect.TypeOf((*MockClient)(nil).IdentityManager), arg0)
}

// InfraProvider mocks base method
func (m *MockClient) InfraProvider() fab.InfraProvider {
	ret := m.ctrl.Call(m, "InfraProvider")
	ret0, _ := ret[0].(fab.InfraProvider)
	return ret0
}

// InfraProvider indicates an expected call of InfraProvider
func (mr *MockClientMockRecorder) InfraProvider() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InfraProvider", reflect.TypeOf((*MockClient)(nil).InfraProvider))
}

// MSPID mocks base method
func (m *MockClient) MSPID() string {
	ret := m.ctrl.Call(m, "MSPID")
	ret0, _ := ret[0].(string)
	return ret0
}

// MSPID indicates an expected call of MSPID
func (mr *MockClientMockRecorder) MSPID() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MSPID", reflect.TypeOf((*MockClient)(nil).MSPID))
}

// PrivateKey mocks base method
func (m *MockClient) PrivateKey() core.Key {
	ret := m.ctrl.Call(m, "PrivateKey")
	ret0, _ := ret[0].(core.Key)
	return ret0
}

// PrivateKey indicates an expected call of PrivateKey
func (mr *MockClientMockRecorder) PrivateKey() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PrivateKey", reflect.TypeOf((*MockClient)(nil).PrivateKey))
}

// SelectionProvider mocks base method
func (m *MockClient) SelectionProvider() fab.SelectionProvider {
	ret := m.ctrl.Call(m, "SelectionProvider")
	ret0, _ := ret[0].(fab.SelectionProvider)
	return ret0
}

// SelectionProvider indicates an expected call of SelectionProvider
func (mr *MockClientMockRecorder) SelectionProvider() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectionProvider", reflect.TypeOf((*MockClient)(nil).SelectionProvider))
}

// SerializedIdentity mocks base method
func (m *MockClient) SerializedIdentity() ([]byte, error) {
	ret := m.ctrl.Call(m, "SerializedIdentity")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SerializedIdentity indicates an expected call of SerializedIdentity
func (mr *MockClientMockRecorder) SerializedIdentity() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SerializedIdentity", reflect.TypeOf((*MockClient)(nil).SerializedIdentity))
}

// SigningManager mocks base method
func (m *MockClient) SigningManager() core.SigningManager {
	ret := m.ctrl.Call(m, "SigningManager")
	ret0, _ := ret[0].(core.SigningManager)
	return ret0
}

// SigningManager indicates an expected call of SigningManager
func (mr *MockClientMockRecorder) SigningManager() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SigningManager", reflect.TypeOf((*MockClient)(nil).SigningManager))
}

// UserStore mocks base method
func (m *MockClient) UserStore() msp.UserStore {
	ret := m.ctrl.Call(m, "UserStore")
	ret0, _ := ret[0].(msp.UserStore)
	return ret0
}

// UserStore indicates an expected call of UserStore
func (mr *MockClientMockRecorder) UserStore() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UserStore", reflect.TypeOf((*MockClient)(nil).UserStore))
}
