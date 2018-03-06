// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api (interfaces: CoreProviderFactory,ServiceProviderFactory)

// Package mock_api is a generated GoMock package.
package mock_api

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	context "github.com/hyperledger/fabric-sdk-go/pkg/common/context"
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
func (m *MockCoreProviderFactory) CreateFabricProvider(arg0 context.Providers) (fab.InfraProvider, error) {
	ret := m.ctrl.Call(m, "CreateFabricProvider", arg0)
	ret0, _ := ret[0].(fab.InfraProvider)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateFabricProvider indicates an expected call of CreateFabricProvider
func (mr *MockCoreProviderFactoryMockRecorder) CreateFabricProvider(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateFabricProvider", reflect.TypeOf((*MockCoreProviderFactory)(nil).CreateFabricProvider), arg0)
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
func (m *MockServiceProviderFactory) CreateDiscoveryProvider(arg0 core.Config, arg1 fab.InfraProvider) (fab.DiscoveryProvider, error) {
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
