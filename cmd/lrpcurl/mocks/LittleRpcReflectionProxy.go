// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import (
	server "github.com/nyan233/littlerpc/server"
	mock "github.com/stretchr/testify/mock"
)

// LittleRpcReflectionProxy is an autogenerated mock type for the LittleRpcReflectionProxy type
type LittleRpcReflectionProxy struct {
	mock.Mock
}

// AllCodec provides a mock function with given fields:
func (_m *LittleRpcReflectionProxy) AllCodec() ([]string, error) {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AllInstance provides a mock function with given fields:
func (_m *LittleRpcReflectionProxy) AllInstance() (map[string]string, error) {
	ret := _m.Called()

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func() map[string]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AllPacker provides a mock function with given fields:
func (_m *LittleRpcReflectionProxy) AllPacker() ([]string, error) {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MethodArgumentType provides a mock function with given fields: serviceName
func (_m *LittleRpcReflectionProxy) MethodArgumentType(serviceName string) ([]*server.ArgumentType, error) {
	ret := _m.Called(serviceName)

	var r0 []*server.ArgumentType
	if rf, ok := ret.Get(0).(func(string) []*server.ArgumentType); ok {
		r0 = rf(serviceName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*server.ArgumentType)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(serviceName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MethodTable provides a mock function with given fields: sourceName
func (_m *LittleRpcReflectionProxy) MethodTable(sourceName string) (*server.MethodTable, error) {
	ret := _m.Called(sourceName)

	var r0 *server.MethodTable
	if rf, ok := ret.Get(0).(func(string) *server.MethodTable); ok {
		r0 = rf(sourceName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*server.MethodTable)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(sourceName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewLittleRpcReflectionProxy interface {
	mock.TestingT
	Cleanup(func())
}

// NewLittleRpcReflectionProxy creates a new instance of LittleRpcReflectionProxy. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewLittleRpcReflectionProxy(t mockConstructorTestingTNewLittleRpcReflectionProxy) *LittleRpcReflectionProxy {
	mock := &LittleRpcReflectionProxy{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
