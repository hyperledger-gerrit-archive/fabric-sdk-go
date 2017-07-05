// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/hyperledger/fabric-sdk-go/api/apifabca (interfaces: FabricCAClient)

package mock_apifabca

import (
	gomock "github.com/golang/mock/gomock"
	apifabca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	bccsp "github.com/hyperledger/fabric/bccsp"
)

// MockFabricCAClient is a mock of FabricCAClient interface
type MockFabricCAClient struct {
	ctrl     *gomock.Controller
	recorder *MockFabricCAClientMockRecorder
}

// MockFabricCAClientMockRecorder is the mock recorder for MockFabricCAClient
type MockFabricCAClientMockRecorder struct {
	mock *MockFabricCAClient
}

// NewMockFabricCAClient creates a new mock instance
func NewMockFabricCAClient(ctrl *gomock.Controller) *MockFabricCAClient {
	mock := &MockFabricCAClient{ctrl: ctrl}
	mock.recorder = &MockFabricCAClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockFabricCAClient) EXPECT() *MockFabricCAClientMockRecorder {
	return _m.recorder
}

// Enroll mocks base method
func (_m *MockFabricCAClient) Enroll(_param0 string, _param1 string) (bccsp.Key, []byte, error) {
	ret := _m.ctrl.Call(_m, "Enroll", _param0, _param1)
	ret0, _ := ret[0].(bccsp.Key)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Enroll indicates an expected call of Enroll
func (_mr *MockFabricCAClientMockRecorder) Enroll(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Enroll", arg0, arg1)
}

// GetCAName mocks base method
func (_m *MockFabricCAClient) GetCAName() string {
	ret := _m.ctrl.Call(_m, "GetCAName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetCAName indicates an expected call of GetCAName
func (_mr *MockFabricCAClientMockRecorder) GetCAName() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetCAName")
}

// Reenroll mocks base method
func (_m *MockFabricCAClient) Reenroll(_param0 apifabca.User) (bccsp.Key, []byte, error) {
	ret := _m.ctrl.Call(_m, "Reenroll", _param0)
	ret0, _ := ret[0].(bccsp.Key)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Reenroll indicates an expected call of Reenroll
func (_mr *MockFabricCAClientMockRecorder) Reenroll(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Reenroll", arg0)
}

// Register mocks base method
func (_m *MockFabricCAClient) Register(_param0 apifabca.User, _param1 *apifabca.RegistrationRequest) (string, error) {
	ret := _m.ctrl.Call(_m, "Register", _param0, _param1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Register indicates an expected call of Register
func (_mr *MockFabricCAClientMockRecorder) Register(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Register", arg0, arg1)
}

// Revoke mocks base method
func (_m *MockFabricCAClient) Revoke(_param0 apifabca.User, _param1 *apifabca.RevocationRequest) error {
	ret := _m.ctrl.Call(_m, "Revoke", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Revoke indicates an expected call of Revoke
func (_mr *MockFabricCAClientMockRecorder) Revoke(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Revoke", arg0, arg1)
}
