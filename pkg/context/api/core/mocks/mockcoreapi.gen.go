// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/hyperledger/fabric-sdk-go/pkg/context/api/core (interfaces: Config,Providers)

// Package mock_core is a generated GoMock package.
package mock_core

import (
	tls "crypto/tls"
	x509 "crypto/x509"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	core "github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// MockConfig is a mock of Config interface
type MockConfig struct {
	ctrl     *gomock.Controller
	recorder *MockConfigMockRecorder
}

// MockConfigMockRecorder is the mock recorder for MockConfig
type MockConfigMockRecorder struct {
	mock *MockConfig
}

// NewMockConfig creates a new mock instance
func NewMockConfig(ctrl *gomock.Controller) *MockConfig {
	mock := &MockConfig{ctrl: ctrl}
	mock.recorder = &MockConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConfig) EXPECT() *MockConfigMockRecorder {
	return m.recorder
}

// CAClientCertPath mocks base method
func (m *MockConfig) CAClientCertPath(arg0 string) (string, error) {
	ret := m.ctrl.Call(m, "CAClientCertPath", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAClientCertPath indicates an expected call of CAClientCertPath
func (mr *MockConfigMockRecorder) CAClientCertPath(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CAClientCertPath", reflect.TypeOf((*MockConfig)(nil).CAClientCertPath), arg0)
}

// CAClientCertPem mocks base method
func (m *MockConfig) CAClientCertPem(arg0 string) (string, error) {
	ret := m.ctrl.Call(m, "CAClientCertPem", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAClientCertPem indicates an expected call of CAClientCertPem
func (mr *MockConfigMockRecorder) CAClientCertPem(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CAClientCertPem", reflect.TypeOf((*MockConfig)(nil).CAClientCertPem), arg0)
}

// CAClientKeyPath mocks base method
func (m *MockConfig) CAClientKeyPath(arg0 string) (string, error) {
	ret := m.ctrl.Call(m, "CAClientKeyPath", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAClientKeyPath indicates an expected call of CAClientKeyPath
func (mr *MockConfigMockRecorder) CAClientKeyPath(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CAClientKeyPath", reflect.TypeOf((*MockConfig)(nil).CAClientKeyPath), arg0)
}

// CAClientKeyPem mocks base method
func (m *MockConfig) CAClientKeyPem(arg0 string) (string, error) {
	ret := m.ctrl.Call(m, "CAClientKeyPem", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAClientKeyPem indicates an expected call of CAClientKeyPem
func (mr *MockConfigMockRecorder) CAClientKeyPem(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CAClientKeyPem", reflect.TypeOf((*MockConfig)(nil).CAClientKeyPem), arg0)
}

// CAConfig mocks base method
func (m *MockConfig) CAConfig(arg0 string) (*core.CAConfig, error) {
	ret := m.ctrl.Call(m, "CAConfig", arg0)
	ret0, _ := ret[0].(*core.CAConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAConfig indicates an expected call of CAConfig
func (mr *MockConfigMockRecorder) CAConfig(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CAConfig", reflect.TypeOf((*MockConfig)(nil).CAConfig), arg0)
}

// CAKeyStorePath mocks base method
func (m *MockConfig) CAKeyStorePath() string {
	ret := m.ctrl.Call(m, "CAKeyStorePath")
	ret0, _ := ret[0].(string)
	return ret0
}

// CAKeyStorePath indicates an expected call of CAKeyStorePath
func (mr *MockConfigMockRecorder) CAKeyStorePath() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CAKeyStorePath", reflect.TypeOf((*MockConfig)(nil).CAKeyStorePath))
}

// CAServerCertPaths mocks base method
func (m *MockConfig) CAServerCertPaths(arg0 string) ([]string, error) {
	ret := m.ctrl.Call(m, "CAServerCertPaths", arg0)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAServerCertPaths indicates an expected call of CAServerCertPaths
func (mr *MockConfigMockRecorder) CAServerCertPaths(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CAServerCertPaths", reflect.TypeOf((*MockConfig)(nil).CAServerCertPaths), arg0)
}

// CAServerCertPems mocks base method
func (m *MockConfig) CAServerCertPems(arg0 string) ([]string, error) {
	ret := m.ctrl.Call(m, "CAServerCertPems", arg0)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAServerCertPems indicates an expected call of CAServerCertPems
func (mr *MockConfigMockRecorder) CAServerCertPems(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CAServerCertPems", reflect.TypeOf((*MockConfig)(nil).CAServerCertPems), arg0)
}

// ChannelConfig mocks base method
func (m *MockConfig) ChannelConfig(arg0 string) (*core.ChannelConfig, error) {
	ret := m.ctrl.Call(m, "ChannelConfig", arg0)
	ret0, _ := ret[0].(*core.ChannelConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ChannelConfig indicates an expected call of ChannelConfig
func (mr *MockConfigMockRecorder) ChannelConfig(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChannelConfig", reflect.TypeOf((*MockConfig)(nil).ChannelConfig), arg0)
}

// ChannelOrderers mocks base method
func (m *MockConfig) ChannelOrderers(arg0 string) ([]core.OrdererConfig, error) {
	ret := m.ctrl.Call(m, "ChannelOrderers", arg0)
	ret0, _ := ret[0].([]core.OrdererConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ChannelOrderers indicates an expected call of ChannelOrderers
func (mr *MockConfigMockRecorder) ChannelOrderers(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChannelOrderers", reflect.TypeOf((*MockConfig)(nil).ChannelOrderers), arg0)
}

// ChannelPeers mocks base method
func (m *MockConfig) ChannelPeers(arg0 string) ([]core.ChannelPeer, error) {
	ret := m.ctrl.Call(m, "ChannelPeers", arg0)
	ret0, _ := ret[0].([]core.ChannelPeer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ChannelPeers indicates an expected call of ChannelPeers
func (mr *MockConfigMockRecorder) ChannelPeers(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChannelPeers", reflect.TypeOf((*MockConfig)(nil).ChannelPeers), arg0)
}

// Client mocks base method
func (m *MockConfig) Client() (*core.ClientConfig, error) {
	ret := m.ctrl.Call(m, "Client")
	ret0, _ := ret[0].(*core.ClientConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Client indicates an expected call of Client
func (mr *MockConfigMockRecorder) Client() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Client", reflect.TypeOf((*MockConfig)(nil).Client))
}

// CredentialStorePath mocks base method
func (m *MockConfig) CredentialStorePath() string {
	ret := m.ctrl.Call(m, "CredentialStorePath")
	ret0, _ := ret[0].(string)
	return ret0
}

// CredentialStorePath indicates an expected call of CredentialStorePath
func (mr *MockConfigMockRecorder) CredentialStorePath() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CredentialStorePath", reflect.TypeOf((*MockConfig)(nil).CredentialStorePath))
}

// CryptoConfigPath mocks base method
func (m *MockConfig) CryptoConfigPath() string {
	ret := m.ctrl.Call(m, "CryptoConfigPath")
	ret0, _ := ret[0].(string)
	return ret0
}

// CryptoConfigPath indicates an expected call of CryptoConfigPath
func (mr *MockConfigMockRecorder) CryptoConfigPath() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CryptoConfigPath", reflect.TypeOf((*MockConfig)(nil).CryptoConfigPath))
}

// Ephemeral mocks base method
func (m *MockConfig) Ephemeral() bool {
	ret := m.ctrl.Call(m, "Ephemeral")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Ephemeral indicates an expected call of Ephemeral
func (mr *MockConfigMockRecorder) Ephemeral() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ephemeral", reflect.TypeOf((*MockConfig)(nil).Ephemeral))
}

// IsSecurityEnabled mocks base method
func (m *MockConfig) IsSecurityEnabled() bool {
	ret := m.ctrl.Call(m, "IsSecurityEnabled")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsSecurityEnabled indicates an expected call of IsSecurityEnabled
func (mr *MockConfigMockRecorder) IsSecurityEnabled() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsSecurityEnabled", reflect.TypeOf((*MockConfig)(nil).IsSecurityEnabled))
}

// KeyStorePath mocks base method
func (m *MockConfig) KeyStorePath() string {
	ret := m.ctrl.Call(m, "KeyStorePath")
	ret0, _ := ret[0].(string)
	return ret0
}

// KeyStorePath indicates an expected call of KeyStorePath
func (mr *MockConfigMockRecorder) KeyStorePath() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KeyStorePath", reflect.TypeOf((*MockConfig)(nil).KeyStorePath))
}

// MspID mocks base method
func (m *MockConfig) MspID(arg0 string) (string, error) {
	ret := m.ctrl.Call(m, "MspID", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MspID indicates an expected call of MspID
func (mr *MockConfigMockRecorder) MspID(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MspID", reflect.TypeOf((*MockConfig)(nil).MspID), arg0)
}

// NetworkConfig mocks base method
func (m *MockConfig) NetworkConfig() (*core.NetworkConfig, error) {
	ret := m.ctrl.Call(m, "NetworkConfig")
	ret0, _ := ret[0].(*core.NetworkConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NetworkConfig indicates an expected call of NetworkConfig
func (mr *MockConfigMockRecorder) NetworkConfig() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NetworkConfig", reflect.TypeOf((*MockConfig)(nil).NetworkConfig))
}

// NetworkPeers mocks base method
func (m *MockConfig) NetworkPeers() ([]core.NetworkPeer, error) {
	ret := m.ctrl.Call(m, "NetworkPeers")
	ret0, _ := ret[0].([]core.NetworkPeer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NetworkPeers indicates an expected call of NetworkPeers
func (mr *MockConfigMockRecorder) NetworkPeers() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NetworkPeers", reflect.TypeOf((*MockConfig)(nil).NetworkPeers))
}

// OrdererConfig mocks base method
func (m *MockConfig) OrdererConfig(arg0 string) (*core.OrdererConfig, error) {
	ret := m.ctrl.Call(m, "OrdererConfig", arg0)
	ret0, _ := ret[0].(*core.OrdererConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OrdererConfig indicates an expected call of OrdererConfig
func (mr *MockConfigMockRecorder) OrdererConfig(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OrdererConfig", reflect.TypeOf((*MockConfig)(nil).OrdererConfig), arg0)
}

// OrderersConfig mocks base method
func (m *MockConfig) OrderersConfig() ([]core.OrdererConfig, error) {
	ret := m.ctrl.Call(m, "OrderersConfig")
	ret0, _ := ret[0].([]core.OrdererConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OrderersConfig indicates an expected call of OrderersConfig
func (mr *MockConfigMockRecorder) OrderersConfig() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OrderersConfig", reflect.TypeOf((*MockConfig)(nil).OrderersConfig))
}

// PeerConfig mocks base method
func (m *MockConfig) PeerConfig(arg0, arg1 string) (*core.PeerConfig, error) {
	ret := m.ctrl.Call(m, "PeerConfig", arg0, arg1)
	ret0, _ := ret[0].(*core.PeerConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PeerConfig indicates an expected call of PeerConfig
func (mr *MockConfigMockRecorder) PeerConfig(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PeerConfig", reflect.TypeOf((*MockConfig)(nil).PeerConfig), arg0, arg1)
}

// PeerMspID mocks base method
func (m *MockConfig) PeerMspID(arg0 string) (string, error) {
	ret := m.ctrl.Call(m, "PeerMspID", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PeerMspID indicates an expected call of PeerMspID
func (mr *MockConfigMockRecorder) PeerMspID(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PeerMspID", reflect.TypeOf((*MockConfig)(nil).PeerMspID), arg0)
}

// PeersConfig mocks base method
func (m *MockConfig) PeersConfig(arg0 string) ([]core.PeerConfig, error) {
	ret := m.ctrl.Call(m, "PeersConfig", arg0)
	ret0, _ := ret[0].([]core.PeerConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PeersConfig indicates an expected call of PeersConfig
func (mr *MockConfigMockRecorder) PeersConfig(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PeersConfig", reflect.TypeOf((*MockConfig)(nil).PeersConfig), arg0)
}

// RandomOrdererConfig mocks base method
func (m *MockConfig) RandomOrdererConfig() (*core.OrdererConfig, error) {
	ret := m.ctrl.Call(m, "RandomOrdererConfig")
	ret0, _ := ret[0].(*core.OrdererConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RandomOrdererConfig indicates an expected call of RandomOrdererConfig
func (mr *MockConfigMockRecorder) RandomOrdererConfig() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RandomOrdererConfig", reflect.TypeOf((*MockConfig)(nil).RandomOrdererConfig))
}

// SecurityAlgorithm mocks base method
func (m *MockConfig) SecurityAlgorithm() string {
	ret := m.ctrl.Call(m, "SecurityAlgorithm")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityAlgorithm indicates an expected call of SecurityAlgorithm
func (mr *MockConfigMockRecorder) SecurityAlgorithm() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SecurityAlgorithm", reflect.TypeOf((*MockConfig)(nil).SecurityAlgorithm))
}

// SecurityLevel mocks base method
func (m *MockConfig) SecurityLevel() int {
	ret := m.ctrl.Call(m, "SecurityLevel")
	ret0, _ := ret[0].(int)
	return ret0
}

// SecurityLevel indicates an expected call of SecurityLevel
func (mr *MockConfigMockRecorder) SecurityLevel() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SecurityLevel", reflect.TypeOf((*MockConfig)(nil).SecurityLevel))
}

// SecurityProvider mocks base method
func (m *MockConfig) SecurityProvider() string {
	ret := m.ctrl.Call(m, "SecurityProvider")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityProvider indicates an expected call of SecurityProvider
func (mr *MockConfigMockRecorder) SecurityProvider() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SecurityProvider", reflect.TypeOf((*MockConfig)(nil).SecurityProvider))
}

// SecurityProviderLabel mocks base method
func (m *MockConfig) SecurityProviderLabel() string {
	ret := m.ctrl.Call(m, "SecurityProviderLabel")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityProviderLabel indicates an expected call of SecurityProviderLabel
func (mr *MockConfigMockRecorder) SecurityProviderLabel() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SecurityProviderLabel", reflect.TypeOf((*MockConfig)(nil).SecurityProviderLabel))
}

// SecurityProviderLibPath mocks base method
func (m *MockConfig) SecurityProviderLibPath() string {
	ret := m.ctrl.Call(m, "SecurityProviderLibPath")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityProviderLibPath indicates an expected call of SecurityProviderLibPath
func (mr *MockConfigMockRecorder) SecurityProviderLibPath() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SecurityProviderLibPath", reflect.TypeOf((*MockConfig)(nil).SecurityProviderLibPath))
}

// SecurityProviderPin mocks base method
func (m *MockConfig) SecurityProviderPin() string {
	ret := m.ctrl.Call(m, "SecurityProviderPin")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityProviderPin indicates an expected call of SecurityProviderPin
func (mr *MockConfigMockRecorder) SecurityProviderPin() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SecurityProviderPin", reflect.TypeOf((*MockConfig)(nil).SecurityProviderPin))
}

// SetTLSCACertPool mocks base method
func (m *MockConfig) SetTLSCACertPool(arg0 *x509.CertPool) {
	m.ctrl.Call(m, "SetTLSCACertPool", arg0)
}

// SetTLSCACertPool indicates an expected call of SetTLSCACertPool
func (mr *MockConfigMockRecorder) SetTLSCACertPool(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTLSCACertPool", reflect.TypeOf((*MockConfig)(nil).SetTLSCACertPool), arg0)
}

// SoftVerify mocks base method
func (m *MockConfig) SoftVerify() bool {
	ret := m.ctrl.Call(m, "SoftVerify")
	ret0, _ := ret[0].(bool)
	return ret0
}

// SoftVerify indicates an expected call of SoftVerify
func (mr *MockConfigMockRecorder) SoftVerify() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SoftVerify", reflect.TypeOf((*MockConfig)(nil).SoftVerify))
}

// TLSCACertPool mocks base method
func (m *MockConfig) TLSCACertPool(arg0 ...*x509.Certificate) (*x509.CertPool, error) {
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "TLSCACertPool", varargs...)
	ret0, _ := ret[0].(*x509.CertPool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TLSCACertPool indicates an expected call of TLSCACertPool
func (mr *MockConfigMockRecorder) TLSCACertPool(arg0 ...interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TLSCACertPool", reflect.TypeOf((*MockConfig)(nil).TLSCACertPool), arg0...)
}

// TLSClientCerts mocks base method
func (m *MockConfig) TLSClientCerts() ([]tls.Certificate, error) {
	ret := m.ctrl.Call(m, "TLSClientCerts")
	ret0, _ := ret[0].([]tls.Certificate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TLSClientCerts indicates an expected call of TLSClientCerts
func (mr *MockConfigMockRecorder) TLSClientCerts() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TLSClientCerts", reflect.TypeOf((*MockConfig)(nil).TLSClientCerts))
}

// Timeout mocks base method
func (m *MockConfig) Timeout(arg0 core.TimeoutType) time.Duration {
	ret := m.ctrl.Call(m, "Timeout", arg0)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// Timeout indicates an expected call of Timeout
func (mr *MockConfigMockRecorder) Timeout(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Timeout", reflect.TypeOf((*MockConfig)(nil).Timeout), arg0)
}

// TimeoutOrDefault mocks base method
func (m *MockConfig) TimeoutOrDefault(arg0 core.TimeoutType) time.Duration {
	ret := m.ctrl.Call(m, "TimeoutOrDefault", arg0)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// TimeoutOrDefault indicates an expected call of TimeoutOrDefault
func (mr *MockConfigMockRecorder) TimeoutOrDefault(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TimeoutOrDefault", reflect.TypeOf((*MockConfig)(nil).TimeoutOrDefault), arg0)
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

// IdentityManager mocks base method
func (m *MockProviders) IdentityManager() core.IdentityManager {
	ret := m.ctrl.Call(m, "IdentityManager")
	ret0, _ := ret[0].(core.IdentityManager)
	return ret0
}

// IdentityManager indicates an expected call of IdentityManager
func (mr *MockProvidersMockRecorder) IdentityManager() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IdentityManager", reflect.TypeOf((*MockProviders)(nil).IdentityManager))
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

// StateStore mocks base method
func (m *MockProviders) StateStore() core.KVStore {
	ret := m.ctrl.Call(m, "StateStore")
	ret0, _ := ret[0].(core.KVStore)
	return ret0
}

// StateStore indicates an expected call of StateStore
func (mr *MockProvidersMockRecorder) StateStore() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StateStore", reflect.TypeOf((*MockProviders)(nil).StateStore))
}
