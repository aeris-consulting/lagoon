// Code generated by MockGen. DO NOT EDIT.
// Source: datasource/datasource.go

// Package datasource is a generated GoMock package.
package datasource

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockSingleValue is a mock of SingleValue interface
type MockSingleValue struct {
	ctrl     *gomock.Controller
	recorder *MockSingleValueMockRecorder
}

// MockSingleValueMockRecorder is the mock recorder for MockSingleValue
type MockSingleValueMockRecorder struct {
	mock *MockSingleValue
}

// NewMockSingleValue creates a new mock instance
func NewMockSingleValue(ctrl *gomock.Controller) *MockSingleValue {
	mock := &MockSingleValue{ctrl: ctrl}
	mock.recorder = &MockSingleValueMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSingleValue) EXPECT() *MockSingleValueMockRecorder {
	return m.recorder
}

// MockVendor is a mock of Vendor interface
type MockVendor struct {
	ctrl     *gomock.Controller
	recorder *MockVendorMockRecorder
}

// MockVendorMockRecorder is the mock recorder for MockVendor
type MockVendorMockRecorder struct {
	mock *MockVendor
}

// NewMockVendor creates a new mock instance
func NewMockVendor(ctrl *gomock.Controller) *MockVendor {
	mock := &MockVendor{ctrl: ctrl}
	mock.recorder = &MockVendorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockVendor) EXPECT() *MockVendorMockRecorder {
	return m.recorder
}

// Accept mocks base method
func (m *MockVendor) Accept(source *DataSourceDescriptor) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Accept", source)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Accept indicates an expected call of Accept
func (mr *MockVendorMockRecorder) Accept(source interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Accept", reflect.TypeOf((*MockVendor)(nil).Accept), source)
}

// CreateDataSource mocks base method
func (m *MockVendor) CreateDataSource(source *DataSourceDescriptor) (DataSource, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateDataSource", source)
	ret0, _ := ret[0].(DataSource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateDataSource indicates an expected call of CreateDataSource
func (mr *MockVendorMockRecorder) CreateDataSource(source interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateDataSource", reflect.TypeOf((*MockVendor)(nil).CreateDataSource), source)
}

// MockDataSource is a mock of DataSource interface
type MockDataSource struct {
	ctrl     *gomock.Controller
	recorder *MockDataSourceMockRecorder
}

// MockDataSourceMockRecorder is the mock recorder for MockDataSource
type MockDataSourceMockRecorder struct {
	mock *MockDataSource
}

// NewMockDataSource creates a new mock instance
func NewMockDataSource(ctrl *gomock.Controller) *MockDataSource {
	mock := &MockDataSource{ctrl: ctrl}
	mock.recorder = &MockDataSourceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDataSource) EXPECT() *MockDataSourceMockRecorder {
	return m.recorder
}

// Open mocks base method
func (m *MockDataSource) Open() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Open")
	ret0, _ := ret[0].(error)
	return ret0
}

// Open indicates an expected call of Open
func (mr *MockDataSourceMockRecorder) Open() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Open", reflect.TypeOf((*MockDataSource)(nil).Open))
}

// Close mocks base method
func (m *MockDataSource) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close
func (mr *MockDataSourceMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockDataSource)(nil).Close))
}

// ListEntryPoints mocks base method
func (m *MockDataSource) ListEntryPoints(filter string, entrypoints chan<- DataBatch, minTreeLevel, maxTreeLevel uint) (ActionStatus, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEntryPoints", filter, entrypoints, minTreeLevel, maxTreeLevel)
	ret0, _ := ret[0].(ActionStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEntryPoints indicates an expected call of ListEntryPoints
func (mr *MockDataSourceMockRecorder) ListEntryPoints(filter, entrypoints, minTreeLevel, maxTreeLevel interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEntryPoints", reflect.TypeOf((*MockDataSource)(nil).ListEntryPoints), filter, entrypoints, minTreeLevel, maxTreeLevel)
}

// GetEntryPointInfos mocks base method
func (m *MockDataSource) GetEntryPointInfos(entryPointValue EntryPoint) (EntryPointInfos, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEntryPointInfos", entryPointValue)
	ret0, _ := ret[0].(EntryPointInfos)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEntryPointInfos indicates an expected call of GetEntryPointInfos
func (mr *MockDataSourceMockRecorder) GetEntryPointInfos(entryPointValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEntryPointInfos", reflect.TypeOf((*MockDataSource)(nil).GetEntryPointInfos), entryPointValue)
}

// GetContent mocks base method
func (m *MockDataSource) GetContent(entryPointValue EntryPoint, filter string, content chan<- DataBatch) (ActionStatus, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContent", entryPointValue, filter, content)
	ret0, _ := ret[0].(ActionStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContent indicates an expected call of GetContent
func (mr *MockDataSourceMockRecorder) GetContent(entryPointValue, filter, content interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContent", reflect.TypeOf((*MockDataSource)(nil).GetContent), entryPointValue, filter, content)
}

// DeleteEntrypoint mocks base method
func (m *MockDataSource) DeleteEntrypoint(entryPointValue EntryPoint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteEntrypoint", entryPointValue)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteEntrypoint indicates an expected call of DeleteEntrypoint
func (mr *MockDataSourceMockRecorder) DeleteEntrypoint(entryPointValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEntrypoint", reflect.TypeOf((*MockDataSource)(nil).DeleteEntrypoint), entryPointValue)
}

// DeleteEntrypointChildren mocks base method
func (m *MockDataSource) DeleteEntrypointChildren(entryPointValue EntryPoint, errorChannel chan<- error) (ActionStatus, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteEntrypointChildren", entryPointValue, errorChannel)
	ret0, _ := ret[0].(ActionStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteEntrypointChildren indicates an expected call of DeleteEntrypointChildren
func (mr *MockDataSourceMockRecorder) DeleteEntrypointChildren(entryPointValue, errorChannel interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEntrypointChildren", reflect.TypeOf((*MockDataSource)(nil).DeleteEntrypointChildren), entryPointValue, errorChannel)
}

// Consume mocks base method
func (m *MockDataSource) Consume(entryPointValue EntryPoint, values chan<- DataBatch, filter Filter, fromBeginning bool) (ActionStatus, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Consume", entryPointValue, values, filter, fromBeginning)
	ret0, _ := ret[0].(ActionStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Consume indicates an expected call of Consume
func (mr *MockDataSourceMockRecorder) Consume(entryPointValue, values, filter, fromBeginning interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Consume", reflect.TypeOf((*MockDataSource)(nil).Consume), entryPointValue, values, filter, fromBeginning)
}

// ExecuteCommand mocks base method
func (m *MockDataSource) ExecuteCommand(args []interface{}, nodeID string) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExecuteCommand", args, nodeID)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExecuteCommand indicates an expected call of ExecuteCommand
func (mr *MockDataSourceMockRecorder) ExecuteCommand(args, nodeID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExecuteCommand", reflect.TypeOf((*MockDataSource)(nil).ExecuteCommand), args, nodeID)
}

// GetInfos mocks base method
func (m *MockDataSource) GetInfos() (Cluster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInfos")
	ret0, _ := ret[0].(Cluster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInfos indicates an expected call of GetInfos
func (mr *MockDataSourceMockRecorder) GetInfos() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInfos", reflect.TypeOf((*MockDataSource)(nil).GetInfos))
}

// GetStatus mocks base method
func (m *MockDataSource) GetStatus() (ClusterState, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStatus")
	ret0, _ := ret[0].(ClusterState)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStatus indicates an expected call of GetStatus
func (mr *MockDataSourceMockRecorder) GetStatus() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatus", reflect.TypeOf((*MockDataSource)(nil).GetStatus))
}
