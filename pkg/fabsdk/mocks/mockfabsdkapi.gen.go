// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api (interfaces: CoreProviderFactory,ServiceProviderFactory,SessionClientFactory)

// Package mock_api is a generated GoMock package.
package mock_api

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	channel "github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	context "github.com/hyperledger/fabric-sdk-go/pkg/context"
	core "github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	fab "github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
)

// MockCoreProviderFactory is a mock of CoreProviderFactory interface
type MockCoreProviderFactory struct {
	ctrl     *gomock.Controller
	recorder *MockCoreProviderFactoryMockRecorder
}

// MockCoreProviderFactoryMockRecorder is the mock recorder for MockCoreProviderFactory
type MockCoreProviderFactoryMockRecorder struct {
	mock *MockCoreProviderFactory
}

// NewMockCoreProviderFactory creates a new mock instance
func NewMockCoreProviderFactory(ctrl *gomock.Controller) *MockCoreProviderFactory {
	mock := &MockCoreProviderFactory{ctrl: ctrl}
	mock.recorder = &MockCoreProviderFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCoreProviderFactory) EXPECT() *MockCoreProviderFactoryMockRecorder {
	return m.recorder
}

// CreateCryptoSuiteProvider mocks base method
func (m *MockCoreProviderFactory) CreateCryptoSuiteProvider(arg0 core.Config) (core.CryptoSuite, error) {
	ret := m.ctrl.Call(m, "CreateCryptoSuiteProvider", arg0)
	ret0, _ := ret[0].(core.CryptoSuite)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateCryptoSuiteProvider indicates an expected call of CreateCryptoSuiteProvider
func (mr *MockCoreProviderFactoryMockRecorder) CreateCryptoSuiteProvider(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCryptoSuiteProvider", reflect.TypeOf((*MockCoreProviderFactory)(nil).CreateCryptoSuiteProvider), arg0)
}

// CreateFabricProvider mocks base method
func (m *MockCoreProviderFactory) CreateFabricProvider(arg0 core.Providers) (fab.InfraProvider, error) {
	ret := m.ctrl.Call(m, "CreateFabricProvider", arg0)
	ret0, _ := ret[0].(fab.InfraProvider)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateFabricProvider indicates an expected call of CreateFabricProvider
func (mr *MockCoreProviderFactoryMockRecorder) CreateFabricProvider(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateFabricProvider", reflect.TypeOf((*MockCoreProviderFactory)(nil).CreateFabricProvider), arg0)
}

// CreateIdentityManager mocks base method
func (m *MockCoreProviderFactory) CreateIdentityManager(arg0 string, arg1 core.KVStore, arg2 core.CryptoSuite, arg3 core.Config) (core.IdentityManager, error) {
	ret := m.ctrl.Call(m, "CreateIdentityManager", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(core.IdentityManager)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateIdentityManager indicates an expected call of CreateIdentityManager
func (mr *MockCoreProviderFactoryMockRecorder) CreateIdentityManager(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateIdentityManager", reflect.TypeOf((*MockCoreProviderFactory)(nil).CreateIdentityManager), arg0, arg1, arg2, arg3)
}

// CreateSigningManager mocks base method
func (m *MockCoreProviderFactory) CreateSigningManager(arg0 core.CryptoSuite, arg1 core.Config) (core.SigningManager, error) {
	ret := m.ctrl.Call(m, "CreateSigningManager", arg0, arg1)
	ret0, _ := ret[0].(core.SigningManager)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateSigningManager indicates an expected call of CreateSigningManager
func (mr *MockCoreProviderFactoryMockRecorder) CreateSigningManager(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSigningManager", reflect.TypeOf((*MockCoreProviderFactory)(nil).CreateSigningManager), arg0, arg1)
}

// CreateStateStoreProvider mocks base method
func (m *MockCoreProviderFactory) CreateStateStoreProvider(arg0 core.Config) (core.KVStore, error) {
	ret := m.ctrl.Call(m, "CreateStateStoreProvider", arg0)
	ret0, _ := ret[0].(core.KVStore)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateStateStoreProvider indicates an expected call of CreateStateStoreProvider
func (mr *MockCoreProviderFactoryMockRecorder) CreateStateStoreProvider(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateStateStoreProvider", reflect.TypeOf((*MockCoreProviderFactory)(nil).CreateStateStoreProvider), arg0)
}

// MockServiceProviderFactory is a mock of ServiceProviderFactory interface
type MockServiceProviderFactory struct {
	ctrl     *gomock.Controller
	recorder *MockServiceProviderFactoryMockRecorder
}

// MockServiceProviderFactoryMockRecorder is the mock recorder for MockServiceProviderFactory
type MockServiceProviderFactoryMockRecorder struct {
	mock *MockServiceProviderFactory
}

// NewMockServiceProviderFactory creates a new mock instance
func NewMockServiceProviderFactory(ctrl *gomock.Controller) *MockServiceProviderFactory {
	mock := &MockServiceProviderFactory{ctrl: ctrl}
	mock.recorder = &MockServiceProviderFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockServiceProviderFactory) EXPECT() *MockServiceProviderFactoryMockRecorder {
	return m.recorder
}

// CreateDiscoveryProvider mocks base method
func (m *MockServiceProviderFactory) CreateDiscoveryProvider(arg0 core.Config, arg1 api0.FabricProvider) (fab.DiscoveryProvider, error) {
	ret := m.ctrl.Call(m, "CreateDiscoveryProvider", arg0, arg1)
	ret0, _ := ret[0].(fab.DiscoveryProvider)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateDiscoveryProvider indicates an expected call of CreateDiscoveryProvider
func (mr *MockServiceProviderFactoryMockRecorder) CreateDiscoveryProvider(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateDiscoveryProvider", reflect.TypeOf((*MockServiceProviderFactory)(nil).CreateDiscoveryProvider), arg0, arg1)
}

// CreateSelectionProvider mocks base method
func (m *MockServiceProviderFactory) CreateSelectionProvider(arg0 core.Config) (fab.SelectionProvider, error) {
	ret := m.ctrl.Call(m, "CreateSelectionProvider", arg0)
	ret0, _ := ret[0].(fab.SelectionProvider)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateSelectionProvider indicates an expected call of CreateSelectionProvider
func (mr *MockServiceProviderFactoryMockRecorder) CreateSelectionProvider(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSelectionProvider", reflect.TypeOf((*MockServiceProviderFactory)(nil).CreateSelectionProvider), arg0)
}

// MockSessionClientFactory is a mock of SessionClientFactory interface
type MockSessionClientFactory struct {
	ctrl     *gomock.Controller
	recorder *MockSessionClientFactoryMockRecorder
}

// MockSessionClientFactoryMockRecorder is the mock recorder for MockSessionClientFactory
type MockSessionClientFactoryMockRecorder struct {
	mock *MockSessionClientFactory
}

// NewMockSessionClientFactory creates a new mock instance
func NewMockSessionClientFactory(ctrl *gomock.Controller) *MockSessionClientFactory {
	mock := &MockSessionClientFactory{ctrl: ctrl}
	mock.recorder = &MockSessionClientFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSessionClientFactory) EXPECT() *MockSessionClientFactoryMockRecorder {
	return m.recorder
}

// CreateChannelClient mocks base method
func (m *MockSessionClientFactory) CreateChannelClient(arg0 context.Providers, arg1 context.Session, arg2 string, arg3 fab.TargetFilter) (*channel.Client, error) {
	ret := m.ctrl.Call(m, "CreateChannelClient", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*channel.Client)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateChannelClient indicates an expected call of CreateChannelClient
func (mr *MockSessionClientFactoryMockRecorder) CreateChannelClient(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateChannelClient", reflect.TypeOf((*MockSessionClientFactory)(nil).CreateChannelClient), arg0, arg1, arg2, arg3)
}
