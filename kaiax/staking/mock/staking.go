// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/kaiachain/kaia/kaiax/staking (interfaces: StakingModule)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/kaiachain/kaia/blockchain/types"
	common "github.com/kaiachain/kaia/common"
	staking "github.com/kaiachain/kaia/kaiax/staking"
	rpc "github.com/kaiachain/kaia/networks/rpc"
)

// MockStakingModule is a mock of StakingModule interface.
type MockStakingModule struct {
	ctrl     *gomock.Controller
	recorder *MockStakingModuleMockRecorder
}

// MockStakingModuleMockRecorder is the mock recorder for MockStakingModule.
type MockStakingModuleMockRecorder struct {
	mock *MockStakingModule
}

// NewMockStakingModule creates a new mock instance.
func NewMockStakingModule(ctrl *gomock.Controller) *MockStakingModule {
	mock := &MockStakingModule{ctrl: ctrl}
	mock.recorder = &MockStakingModuleMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStakingModule) EXPECT() *MockStakingModuleMockRecorder {
	return m.recorder
}

// APIs mocks base method.
func (m *MockStakingModule) APIs() []rpc.API {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "APIs")
	ret0, _ := ret[0].([]rpc.API)
	return ret0
}

// APIs indicates an expected call of APIs.
func (mr *MockStakingModuleMockRecorder) APIs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "APIs", reflect.TypeOf((*MockStakingModule)(nil).APIs))
}

// GetStakingInfo mocks base method.
func (m *MockStakingModule) GetStakingInfo(arg0 uint64) (*staking.StakingInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStakingInfo", arg0)
	ret0, _ := ret[0].(*staking.StakingInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStakingInfo indicates an expected call of GetStakingInfo.
func (mr *MockStakingModuleMockRecorder) GetStakingInfo(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStakingInfo", reflect.TypeOf((*MockStakingModule)(nil).GetStakingInfo), arg0)
}

// RewindDelete mocks base method.
func (m *MockStakingModule) RewindDelete(arg0 common.Hash, arg1 uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RewindDelete", arg0, arg1)
}

// RewindDelete indicates an expected call of RewindDelete.
func (mr *MockStakingModuleMockRecorder) RewindDelete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RewindDelete", reflect.TypeOf((*MockStakingModule)(nil).RewindDelete), arg0, arg1)
}

// RewindTo mocks base method.
func (m *MockStakingModule) RewindTo(arg0 *types.Block) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RewindTo", arg0)
}

// RewindTo indicates an expected call of RewindTo.
func (mr *MockStakingModuleMockRecorder) RewindTo(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RewindTo", reflect.TypeOf((*MockStakingModule)(nil).RewindTo), arg0)
}

// Start mocks base method.
func (m *MockStakingModule) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockStakingModuleMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockStakingModule)(nil).Start))
}

// Stop mocks base method.
func (m *MockStakingModule) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockStakingModuleMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockStakingModule)(nil).Stop))
}