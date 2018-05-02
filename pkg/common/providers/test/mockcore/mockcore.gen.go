// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core (interfaces: CryptoSuiteConfig,ConfigBackend,Providers)

package mockcore

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	core "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
)

// MockCryptoSuiteConfig is a mock of CryptoSuiteConfig interface
type MockCryptoSuiteConfig struct {
	ctrl     *gomock.Controller
	recorder *MockCryptoSuiteConfigMockRecorder
}

// MockCryptoSuiteConfigMockRecorder is the mock recorder for MockCryptoSuiteConfig
type MockCryptoSuiteConfigMockRecorder struct {
	mock *MockCryptoSuiteConfig
}

// NewMockCryptoSuiteConfig creates a new mock instance
func NewMockCryptoSuiteConfig(ctrl *gomock.Controller) *MockCryptoSuiteConfig {
	mock := &MockCryptoSuiteConfig{ctrl: ctrl}
	mock.recorder = &MockCryptoSuiteConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockCryptoSuiteConfig) EXPECT() *MockCryptoSuiteConfigMockRecorder {
	return _m.recorder
}

// IsSecurityEnabled mocks base method
func (_m *MockCryptoSuiteConfig) IsSecurityEnabled() bool {
	ret := _m.ctrl.Call(_m, "IsSecurityEnabled")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsSecurityEnabled indicates an expected call of IsSecurityEnabled
func (_mr *MockCryptoSuiteConfigMockRecorder) IsSecurityEnabled() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "IsSecurityEnabled", reflect.TypeOf((*MockCryptoSuiteConfig)(nil).IsSecurityEnabled))
}

// KeyStorePath mocks base method
func (_m *MockCryptoSuiteConfig) KeyStorePath() string {
	ret := _m.ctrl.Call(_m, "KeyStorePath")
	ret0, _ := ret[0].(string)
	return ret0
}

// KeyStorePath indicates an expected call of KeyStorePath
func (_mr *MockCryptoSuiteConfigMockRecorder) KeyStorePath() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "KeyStorePath", reflect.TypeOf((*MockCryptoSuiteConfig)(nil).KeyStorePath))
}

// SecurityAlgorithm mocks base method
func (_m *MockCryptoSuiteConfig) SecurityAlgorithm() string {
	ret := _m.ctrl.Call(_m, "SecurityAlgorithm")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityAlgorithm indicates an expected call of SecurityAlgorithm
func (_mr *MockCryptoSuiteConfigMockRecorder) SecurityAlgorithm() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SecurityAlgorithm", reflect.TypeOf((*MockCryptoSuiteConfig)(nil).SecurityAlgorithm))
}

// SecurityLevel mocks base method
func (_m *MockCryptoSuiteConfig) SecurityLevel() int {
	ret := _m.ctrl.Call(_m, "SecurityLevel")
	ret0, _ := ret[0].(int)
	return ret0
}

// SecurityLevel indicates an expected call of SecurityLevel
func (_mr *MockCryptoSuiteConfigMockRecorder) SecurityLevel() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SecurityLevel", reflect.TypeOf((*MockCryptoSuiteConfig)(nil).SecurityLevel))
}

// SecurityProvider mocks base method
func (_m *MockCryptoSuiteConfig) SecurityProvider() string {
	ret := _m.ctrl.Call(_m, "SecurityProvider")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityProvider indicates an expected call of SecurityProvider
func (_mr *MockCryptoSuiteConfigMockRecorder) SecurityProvider() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SecurityProvider", reflect.TypeOf((*MockCryptoSuiteConfig)(nil).SecurityProvider))
}

// SecurityProviderAddress mocks base method
func (_m *MockCryptoSuiteConfig) SecurityProviderAddress() string {
	ret := _m.ctrl.Call(_m, "SecurityProviderAddress")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityProviderAddress indicates an expected call of SecurityProviderAddress
func (_mr *MockCryptoSuiteConfigMockRecorder) SecurityProviderAddress() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SecurityProviderAddress", reflect.TypeOf((*MockCryptoSuiteConfig)(nil).SecurityProviderAddress))
}

// SecurityProviderLabel mocks base method
func (_m *MockCryptoSuiteConfig) SecurityProviderLabel() string {
	ret := _m.ctrl.Call(_m, "SecurityProviderLabel")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityProviderLabel indicates an expected call of SecurityProviderLabel
func (_mr *MockCryptoSuiteConfigMockRecorder) SecurityProviderLabel() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SecurityProviderLabel", reflect.TypeOf((*MockCryptoSuiteConfig)(nil).SecurityProviderLabel))
}

// SecurityProviderLibPath mocks base method
func (_m *MockCryptoSuiteConfig) SecurityProviderLibPath() string {
	ret := _m.ctrl.Call(_m, "SecurityProviderLibPath")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityProviderLibPath indicates an expected call of SecurityProviderLibPath
func (_mr *MockCryptoSuiteConfigMockRecorder) SecurityProviderLibPath() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SecurityProviderLibPath", reflect.TypeOf((*MockCryptoSuiteConfig)(nil).SecurityProviderLibPath))
}

// SecurityProviderPin mocks base method
func (_m *MockCryptoSuiteConfig) SecurityProviderPin() string {
	ret := _m.ctrl.Call(_m, "SecurityProviderPin")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityProviderPin indicates an expected call of SecurityProviderPin
func (_mr *MockCryptoSuiteConfigMockRecorder) SecurityProviderPin() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SecurityProviderPin", reflect.TypeOf((*MockCryptoSuiteConfig)(nil).SecurityProviderPin))
}

// SecurityProviderToken mocks base method
func (_m *MockCryptoSuiteConfig) SecurityProviderToken() string {
	ret := _m.ctrl.Call(_m, "SecurityProviderToken")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityProviderToken indicates an expected call of SecurityProviderToken
func (_mr *MockCryptoSuiteConfigMockRecorder) SecurityProviderToken() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SecurityProviderToken", reflect.TypeOf((*MockCryptoSuiteConfig)(nil).SecurityProviderToken))
}

// SoftVerify mocks base method
func (_m *MockCryptoSuiteConfig) SoftVerify() bool {
	ret := _m.ctrl.Call(_m, "SoftVerify")
	ret0, _ := ret[0].(bool)
	return ret0
}

// SoftVerify indicates an expected call of SoftVerify
func (_mr *MockCryptoSuiteConfigMockRecorder) SoftVerify() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SoftVerify", reflect.TypeOf((*MockCryptoSuiteConfig)(nil).SoftVerify))
}

// MockConfigBackend is a mock of ConfigBackend interface
type MockConfigBackend struct {
	ctrl     *gomock.Controller
	recorder *MockConfigBackendMockRecorder
}

// MockConfigBackendMockRecorder is the mock recorder for MockConfigBackend
type MockConfigBackendMockRecorder struct {
	mock *MockConfigBackend
}

// NewMockConfigBackend creates a new mock instance
func NewMockConfigBackend(ctrl *gomock.Controller) *MockConfigBackend {
	mock := &MockConfigBackend{ctrl: ctrl}
	mock.recorder = &MockConfigBackendMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockConfigBackend) EXPECT() *MockConfigBackendMockRecorder {
	return _m.recorder
}

// Lookup mocks base method
func (_m *MockConfigBackend) Lookup(_param0 string) (interface{}, bool) {
	ret := _m.ctrl.Call(_m, "Lookup", _param0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// Lookup indicates an expected call of Lookup
func (_mr *MockConfigBackendMockRecorder) Lookup(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Lookup", reflect.TypeOf((*MockConfigBackend)(nil).Lookup), arg0)
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
func (_m *MockProviders) EXPECT() *MockProvidersMockRecorder {
	return _m.recorder
}

// CryptoSuite mocks base method
func (_m *MockProviders) CryptoSuite() core.CryptoSuite {
	ret := _m.ctrl.Call(_m, "CryptoSuite")
	ret0, _ := ret[0].(core.CryptoSuite)
	return ret0
}

// CryptoSuite indicates an expected call of CryptoSuite
func (_mr *MockProvidersMockRecorder) CryptoSuite() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CryptoSuite", reflect.TypeOf((*MockProviders)(nil).CryptoSuite))
}

// SigningManager mocks base method
func (_m *MockProviders) SigningManager() core.SigningManager {
	ret := _m.ctrl.Call(_m, "SigningManager")
	ret0, _ := ret[0].(core.SigningManager)
	return ret0
}

// SigningManager indicates an expected call of SigningManager
func (_mr *MockProvidersMockRecorder) SigningManager() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SigningManager", reflect.TypeOf((*MockProviders)(nil).SigningManager))
}
