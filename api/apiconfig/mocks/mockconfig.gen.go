// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/hyperledger/fabric-sdk-go/api/apiconfig (interfaces: Config)

package mock_apiconfig

import (
	tls "crypto/tls"
	x509 "crypto/x509"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	apiconfig "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
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
func (_m *MockConfig) EXPECT() *MockConfigMockRecorder {
	return _m.recorder
}

// CAClientCertPath mocks base method
func (_m *MockConfig) CAClientCertPath(_param0 string) (string, error) {
	ret := _m.ctrl.Call(_m, "CAClientCertPath", _param0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAClientCertPath indicates an expected call of CAClientCertPath
func (_mr *MockConfigMockRecorder) CAClientCertPath(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CAClientCertPath", reflect.TypeOf((*MockConfig)(nil).CAClientCertPath), arg0)
}

// CAClientCertPem mocks base method
func (_m *MockConfig) CAClientCertPem(_param0 string) (string, error) {
	ret := _m.ctrl.Call(_m, "CAClientCertPem", _param0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAClientCertPem indicates an expected call of CAClientCertPem
func (_mr *MockConfigMockRecorder) CAClientCertPem(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CAClientCertPem", reflect.TypeOf((*MockConfig)(nil).CAClientCertPem), arg0)
}

// CAClientKeyPath mocks base method
func (_m *MockConfig) CAClientKeyPath(_param0 string) (string, error) {
	ret := _m.ctrl.Call(_m, "CAClientKeyPath", _param0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAClientKeyPath indicates an expected call of CAClientKeyPath
func (_mr *MockConfigMockRecorder) CAClientKeyPath(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CAClientKeyPath", reflect.TypeOf((*MockConfig)(nil).CAClientKeyPath), arg0)
}

// CAClientKeyPem mocks base method
func (_m *MockConfig) CAClientKeyPem(_param0 string) (string, error) {
	ret := _m.ctrl.Call(_m, "CAClientKeyPem", _param0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAClientKeyPem indicates an expected call of CAClientKeyPem
func (_mr *MockConfigMockRecorder) CAClientKeyPem(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CAClientKeyPem", reflect.TypeOf((*MockConfig)(nil).CAClientKeyPem), arg0)
}

// CAConfig mocks base method
func (_m *MockConfig) CAConfig(_param0 string) (*apiconfig.CAConfig, error) {
	ret := _m.ctrl.Call(_m, "CAConfig", _param0)
	ret0, _ := ret[0].(*apiconfig.CAConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAConfig indicates an expected call of CAConfig
func (_mr *MockConfigMockRecorder) CAConfig(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CAConfig", reflect.TypeOf((*MockConfig)(nil).CAConfig), arg0)
}

// CAKeyStorePath mocks base method
func (_m *MockConfig) CAKeyStorePath() string {
	ret := _m.ctrl.Call(_m, "CAKeyStorePath")
	ret0, _ := ret[0].(string)
	return ret0
}

// CAKeyStorePath indicates an expected call of CAKeyStorePath
func (_mr *MockConfigMockRecorder) CAKeyStorePath() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CAKeyStorePath", reflect.TypeOf((*MockConfig)(nil).CAKeyStorePath))
}

// CAServerCertPaths mocks base method
func (_m *MockConfig) CAServerCertPaths(_param0 string) ([]string, error) {
	ret := _m.ctrl.Call(_m, "CAServerCertPaths", _param0)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAServerCertPaths indicates an expected call of CAServerCertPaths
func (_mr *MockConfigMockRecorder) CAServerCertPaths(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CAServerCertPaths", reflect.TypeOf((*MockConfig)(nil).CAServerCertPaths), arg0)
}

// CAServerCertPems mocks base method
func (_m *MockConfig) CAServerCertPems(_param0 string) ([]string, error) {
	ret := _m.ctrl.Call(_m, "CAServerCertPems", _param0)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAServerCertPems indicates an expected call of CAServerCertPems
func (_mr *MockConfigMockRecorder) CAServerCertPems(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CAServerCertPems", reflect.TypeOf((*MockConfig)(nil).CAServerCertPems), arg0)
}

// ChannelConfig mocks base method
func (_m *MockConfig) ChannelConfig(_param0 string) (*apiconfig.ChannelConfig, error) {
	ret := _m.ctrl.Call(_m, "ChannelConfig", _param0)
	ret0, _ := ret[0].(*apiconfig.ChannelConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ChannelConfig indicates an expected call of ChannelConfig
func (_mr *MockConfigMockRecorder) ChannelConfig(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "ChannelConfig", reflect.TypeOf((*MockConfig)(nil).ChannelConfig), arg0)
}

// ChannelOrderers mocks base method
func (_m *MockConfig) ChannelOrderers(_param0 string) ([]apiconfig.OrdererConfig, error) {
	ret := _m.ctrl.Call(_m, "ChannelOrderers", _param0)
	ret0, _ := ret[0].([]apiconfig.OrdererConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ChannelOrderers indicates an expected call of ChannelOrderers
func (_mr *MockConfigMockRecorder) ChannelOrderers(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "ChannelOrderers", reflect.TypeOf((*MockConfig)(nil).ChannelOrderers), arg0)
}

// ChannelPeers mocks base method
func (_m *MockConfig) ChannelPeers(_param0 string) ([]apiconfig.ChannelPeer, error) {
	ret := _m.ctrl.Call(_m, "ChannelPeers", _param0)
	ret0, _ := ret[0].([]apiconfig.ChannelPeer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ChannelPeers indicates an expected call of ChannelPeers
func (_mr *MockConfigMockRecorder) ChannelPeers(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "ChannelPeers", reflect.TypeOf((*MockConfig)(nil).ChannelPeers), arg0)
}

// Client mocks base method
func (_m *MockConfig) Client() (*apiconfig.ClientConfig, error) {
	ret := _m.ctrl.Call(_m, "Client")
	ret0, _ := ret[0].(*apiconfig.ClientConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Client indicates an expected call of Client
func (_mr *MockConfigMockRecorder) Client() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Client", reflect.TypeOf((*MockConfig)(nil).Client))
}

// CryptoConfigPath mocks base method
func (_m *MockConfig) CryptoConfigPath() string {
	ret := _m.ctrl.Call(_m, "CryptoConfigPath")
	ret0, _ := ret[0].(string)
	return ret0
}

// CryptoConfigPath indicates an expected call of CryptoConfigPath
func (_mr *MockConfigMockRecorder) CryptoConfigPath() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CryptoConfigPath", reflect.TypeOf((*MockConfig)(nil).CryptoConfigPath))
}

// Ephemeral mocks base method
func (_m *MockConfig) Ephemeral() bool {
	ret := _m.ctrl.Call(_m, "Ephemeral")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Ephemeral indicates an expected call of Ephemeral
func (_mr *MockConfigMockRecorder) Ephemeral() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Ephemeral", reflect.TypeOf((*MockConfig)(nil).Ephemeral))
}

// IsSecurityEnabled mocks base method
func (_m *MockConfig) IsSecurityEnabled() bool {
	ret := _m.ctrl.Call(_m, "IsSecurityEnabled")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsSecurityEnabled indicates an expected call of IsSecurityEnabled
func (_mr *MockConfigMockRecorder) IsSecurityEnabled() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "IsSecurityEnabled", reflect.TypeOf((*MockConfig)(nil).IsSecurityEnabled))
}

// KeyStorePath mocks base method
func (_m *MockConfig) KeyStorePath() string {
	ret := _m.ctrl.Call(_m, "KeyStorePath")
	ret0, _ := ret[0].(string)
	return ret0
}

// KeyStorePath indicates an expected call of KeyStorePath
func (_mr *MockConfigMockRecorder) KeyStorePath() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "KeyStorePath", reflect.TypeOf((*MockConfig)(nil).KeyStorePath))
}

// MspID mocks base method
func (_m *MockConfig) MspID(_param0 string) (string, error) {
	ret := _m.ctrl.Call(_m, "MspID", _param0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MspID indicates an expected call of MspID
func (_mr *MockConfigMockRecorder) MspID(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "MspID", reflect.TypeOf((*MockConfig)(nil).MspID), arg0)
}

// NetworkConfig mocks base method
func (_m *MockConfig) NetworkConfig() (*apiconfig.NetworkConfig, error) {
	ret := _m.ctrl.Call(_m, "NetworkConfig")
	ret0, _ := ret[0].(*apiconfig.NetworkConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NetworkConfig indicates an expected call of NetworkConfig
func (_mr *MockConfigMockRecorder) NetworkConfig() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "NetworkConfig", reflect.TypeOf((*MockConfig)(nil).NetworkConfig))
}

// NetworkPeers mocks base method
func (_m *MockConfig) NetworkPeers() ([]apiconfig.NetworkPeer, error) {
	ret := _m.ctrl.Call(_m, "NetworkPeers")
	ret0, _ := ret[0].([]apiconfig.NetworkPeer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NetworkPeers indicates an expected call of NetworkPeers
func (_mr *MockConfigMockRecorder) NetworkPeers() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "NetworkPeers", reflect.TypeOf((*MockConfig)(nil).NetworkPeers))
}

// OrdererConfig mocks base method
func (_m *MockConfig) OrdererConfig(_param0 string) (*apiconfig.OrdererConfig, error) {
	ret := _m.ctrl.Call(_m, "OrdererConfig", _param0)
	ret0, _ := ret[0].(*apiconfig.OrdererConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OrdererConfig indicates an expected call of OrdererConfig
func (_mr *MockConfigMockRecorder) OrdererConfig(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "OrdererConfig", reflect.TypeOf((*MockConfig)(nil).OrdererConfig), arg0)
}

// OrderersConfig mocks base method
func (_m *MockConfig) OrderersConfig() ([]apiconfig.OrdererConfig, error) {
	ret := _m.ctrl.Call(_m, "OrderersConfig")
	ret0, _ := ret[0].([]apiconfig.OrdererConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OrderersConfig indicates an expected call of OrderersConfig
func (_mr *MockConfigMockRecorder) OrderersConfig() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "OrderersConfig", reflect.TypeOf((*MockConfig)(nil).OrderersConfig))
}

// PeerConfig mocks base method
func (_m *MockConfig) PeerConfig(_param0 string, _param1 string) (*apiconfig.PeerConfig, error) {
	ret := _m.ctrl.Call(_m, "PeerConfig", _param0, _param1)
	ret0, _ := ret[0].(*apiconfig.PeerConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PeerConfig indicates an expected call of PeerConfig
func (_mr *MockConfigMockRecorder) PeerConfig(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "PeerConfig", reflect.TypeOf((*MockConfig)(nil).PeerConfig), arg0, arg1)
}

// PeerMspID mocks base method
func (_m *MockConfig) PeerMspID(_param0 string) (string, error) {
	ret := _m.ctrl.Call(_m, "PeerMspID", _param0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PeerMspID indicates an expected call of PeerMspID
func (_mr *MockConfigMockRecorder) PeerMspID(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "PeerMspID", reflect.TypeOf((*MockConfig)(nil).PeerMspID), arg0)
}

// PeersConfig mocks base method
func (_m *MockConfig) PeersConfig(_param0 string) ([]apiconfig.PeerConfig, error) {
	ret := _m.ctrl.Call(_m, "PeersConfig", _param0)
	ret0, _ := ret[0].([]apiconfig.PeerConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PeersConfig indicates an expected call of PeersConfig
func (_mr *MockConfigMockRecorder) PeersConfig(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "PeersConfig", reflect.TypeOf((*MockConfig)(nil).PeersConfig), arg0)
}

// RandomOrdererConfig mocks base method
func (_m *MockConfig) RandomOrdererConfig() (*apiconfig.OrdererConfig, error) {
	ret := _m.ctrl.Call(_m, "RandomOrdererConfig")
	ret0, _ := ret[0].(*apiconfig.OrdererConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RandomOrdererConfig indicates an expected call of RandomOrdererConfig
func (_mr *MockConfigMockRecorder) RandomOrdererConfig() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "RandomOrdererConfig", reflect.TypeOf((*MockConfig)(nil).RandomOrdererConfig))
}

// SecurityAlgorithm mocks base method
func (_m *MockConfig) SecurityAlgorithm() string {
	ret := _m.ctrl.Call(_m, "SecurityAlgorithm")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityAlgorithm indicates an expected call of SecurityAlgorithm
func (_mr *MockConfigMockRecorder) SecurityAlgorithm() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SecurityAlgorithm", reflect.TypeOf((*MockConfig)(nil).SecurityAlgorithm))
}

// SecurityLevel mocks base method
func (_m *MockConfig) SecurityLevel() int {
	ret := _m.ctrl.Call(_m, "SecurityLevel")
	ret0, _ := ret[0].(int)
	return ret0
}

// SecurityLevel indicates an expected call of SecurityLevel
func (_mr *MockConfigMockRecorder) SecurityLevel() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SecurityLevel", reflect.TypeOf((*MockConfig)(nil).SecurityLevel))
}

// SecurityProvider mocks base method
func (_m *MockConfig) SecurityProvider() string {
	ret := _m.ctrl.Call(_m, "SecurityProvider")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityProvider indicates an expected call of SecurityProvider
func (_mr *MockConfigMockRecorder) SecurityProvider() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SecurityProvider", reflect.TypeOf((*MockConfig)(nil).SecurityProvider))
}

// SecurityProviderLabel mocks base method
func (_m *MockConfig) SecurityProviderLabel() string {
	ret := _m.ctrl.Call(_m, "SecurityProviderLabel")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityProviderLabel indicates an expected call of SecurityProviderLabel
func (_mr *MockConfigMockRecorder) SecurityProviderLabel() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SecurityProviderLabel", reflect.TypeOf((*MockConfig)(nil).SecurityProviderLabel))
}

// SecurityProviderLibPath mocks base method
func (_m *MockConfig) SecurityProviderLibPath() string {
	ret := _m.ctrl.Call(_m, "SecurityProviderLibPath")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityProviderLibPath indicates an expected call of SecurityProviderLibPath
func (_mr *MockConfigMockRecorder) SecurityProviderLibPath() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SecurityProviderLibPath", reflect.TypeOf((*MockConfig)(nil).SecurityProviderLibPath))
}

// SecurityProviderPin mocks base method
func (_m *MockConfig) SecurityProviderPin() string {
	ret := _m.ctrl.Call(_m, "SecurityProviderPin")
	ret0, _ := ret[0].(string)
	return ret0
}

// SecurityProviderPin indicates an expected call of SecurityProviderPin
func (_mr *MockConfigMockRecorder) SecurityProviderPin() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SecurityProviderPin", reflect.TypeOf((*MockConfig)(nil).SecurityProviderPin))
}

// SetTLSCACertPool mocks base method
func (_m *MockConfig) SetTLSCACertPool(_param0 *x509.CertPool) {
	_m.ctrl.Call(_m, "SetTLSCACertPool", _param0)
}

// SetTLSCACertPool indicates an expected call of SetTLSCACertPool
func (_mr *MockConfigMockRecorder) SetTLSCACertPool(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SetTLSCACertPool", reflect.TypeOf((*MockConfig)(nil).SetTLSCACertPool), arg0)
}

// SoftVerify mocks base method
func (_m *MockConfig) SoftVerify() bool {
	ret := _m.ctrl.Call(_m, "SoftVerify")
	ret0, _ := ret[0].(bool)
	return ret0
}

// SoftVerify indicates an expected call of SoftVerify
func (_mr *MockConfigMockRecorder) SoftVerify() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SoftVerify", reflect.TypeOf((*MockConfig)(nil).SoftVerify))
}

// TLSCACertPool mocks base method
func (_m *MockConfig) TLSCACertPool(_param0 string) (*x509.CertPool, error) {
	ret := _m.ctrl.Call(_m, "TLSCACertPool", _param0)
	ret0, _ := ret[0].(*x509.CertPool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TLSCACertPool indicates an expected call of TLSCACertPool
func (_mr *MockConfigMockRecorder) TLSCACertPool(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "TLSCACertPool", reflect.TypeOf((*MockConfig)(nil).TLSCACertPool), arg0)
}

// TLSClientCerts mocks base method
func (_m *MockConfig) TLSClientCerts() ([]tls.Certificate, error) {
	ret := _m.ctrl.Call(_m, "TLSClientCerts")
	ret0, _ := ret[0].([]tls.Certificate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TLSClientCerts indicates an expected call of TLSClientCerts
func (_mr *MockConfigMockRecorder) TLSClientCerts() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "TLSClientCerts", reflect.TypeOf((*MockConfig)(nil).TLSClientCerts))
}

// TimeoutOrDefault mocks base method
func (_m *MockConfig) TimeoutOrDefault(_param0 apiconfig.TimeoutType) time.Duration {
	ret := _m.ctrl.Call(_m, "TimeoutOrDefault", _param0)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// TimeoutOrDefault indicates an expected call of TimeoutOrDefault
func (_mr *MockConfigMockRecorder) TimeoutOrDefault(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "TimeoutOrDefault", reflect.TypeOf((*MockConfig)(nil).TimeoutOrDefault), arg0)
}
