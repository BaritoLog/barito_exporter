// Code generated by MockGen. DO NOT EDIT.
// Source: ./appgroup/appgroup.go

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockAppGroup is a mock of AppGroup interface
type MockAppGroup struct {
	ctrl     *gomock.Controller
	recorder *MockAppGroupMockRecorder
}

// MockAppGroupMockRecorder is the mock recorder for MockAppGroup
type MockAppGroupMockRecorder struct {
	mock *MockAppGroup
}

// NewMockAppGroup creates a new mock instance
func NewMockAppGroup(ctrl *gomock.Controller) *MockAppGroup {
	mock := &MockAppGroup{ctrl: ctrl}
	mock.recorder = &MockAppGroupMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAppGroup) EXPECT() *MockAppGroupMockRecorder {
	return m.recorder
}

// RefreshMetadata mocks base method
func (m *MockAppGroup) RefreshMetadata() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RefreshMetadata")
	ret0, _ := ret[0].(error)
	return ret0
}

// RefreshMetadata indicates an expected call of RefreshMetadata
func (mr *MockAppGroupMockRecorder) RefreshMetadata() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshMetadata", reflect.TypeOf((*MockAppGroup)(nil).RefreshMetadata))
}

// GetName mocks base method
func (m *MockAppGroup) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName
func (mr *MockAppGroupMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockAppGroup)(nil).GetName))
}

// GetClusterName mocks base method
func (m *MockAppGroup) GetClusterName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClusterName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetClusterName indicates an expected call of GetClusterName
func (mr *MockAppGroupMockRecorder) GetClusterName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClusterName", reflect.TypeOf((*MockAppGroup)(nil).GetClusterName))
}

// GetSecret mocks base method
func (m *MockAppGroup) GetSecret() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSecret")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetSecret indicates an expected call of GetSecret
func (mr *MockAppGroupMockRecorder) GetSecret() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSecret", reflect.TypeOf((*MockAppGroup)(nil).GetSecret))
}

// GetListES mocks base method
func (m *MockAppGroup) GetListES() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetListES")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetListES indicates an expected call of GetListES
func (mr *MockAppGroupMockRecorder) GetListES() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetListES", reflect.TypeOf((*MockAppGroup)(nil).GetListES))
}

// GetListKafka mocks base method
func (m *MockAppGroup) GetListKafka() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetListKafka")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetListKafka indicates an expected call of GetListKafka
func (mr *MockAppGroupMockRecorder) GetListKafka() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetListKafka", reflect.TypeOf((*MockAppGroup)(nil).GetListKafka))
}

// GetKibanaHost mocks base method
func (m *MockAppGroup) GetKibanaHost() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetKibanaHost")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetKibanaHost indicates an expected call of GetKibanaHost
func (mr *MockAppGroupMockRecorder) GetKibanaHost() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetKibanaHost", reflect.TypeOf((*MockAppGroup)(nil).GetKibanaHost))
}
