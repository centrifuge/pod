// Code generated by mockery v2.13.0-beta.1. DO NOT EDIT.

package mocks

import (
	context "context"

	connmgr "github.com/libp2p/go-libp2p-core/connmgr"

	event "github.com/libp2p/go-libp2p-core/event"

	mock "github.com/stretchr/testify/mock"

	multiaddr "github.com/multiformats/go-multiaddr"

	network "github.com/libp2p/go-libp2p-core/network"

	peer "github.com/libp2p/go-libp2p-core/peer"

	peerstore "github.com/libp2p/go-libp2p-core/peerstore"

	protocol "github.com/libp2p/go-libp2p-core/protocol"
)

// HostMock is an autogenerated mock type for the Host type
type HostMock struct {
	mock.Mock
}

// Addrs provides a mock function with given fields:
func (_m *HostMock) Addrs() []multiaddr.Multiaddr {
	ret := _m.Called()

	var r0 []multiaddr.Multiaddr
	if rf, ok := ret.Get(0).(func() []multiaddr.Multiaddr); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]multiaddr.Multiaddr)
		}
	}

	return r0
}

// Close provides a mock function with given fields:
func (_m *HostMock) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ConnManager provides a mock function with given fields:
func (_m *HostMock) ConnManager() connmgr.ConnManager {
	ret := _m.Called()

	var r0 connmgr.ConnManager
	if rf, ok := ret.Get(0).(func() connmgr.ConnManager); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(connmgr.ConnManager)
		}
	}

	return r0
}

// Connect provides a mock function with given fields: ctx, pi
func (_m *HostMock) Connect(ctx context.Context, pi peer.AddrInfo) error {
	ret := _m.Called(ctx, pi)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, peer.AddrInfo) error); ok {
		r0 = rf(ctx, pi)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EventBus provides a mock function with given fields:
func (_m *HostMock) EventBus() event.Bus {
	ret := _m.Called()

	var r0 event.Bus
	if rf, ok := ret.Get(0).(func() event.Bus); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(event.Bus)
		}
	}

	return r0
}

// ID provides a mock function with given fields:
func (_m *HostMock) ID() peer.ID {
	ret := _m.Called()

	var r0 peer.ID
	if rf, ok := ret.Get(0).(func() peer.ID); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(peer.ID)
	}

	return r0
}

// Mux provides a mock function with given fields:
func (_m *HostMock) Mux() protocol.Switch {
	ret := _m.Called()

	var r0 protocol.Switch
	if rf, ok := ret.Get(0).(func() protocol.Switch); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(protocol.Switch)
		}
	}

	return r0
}

// Network provides a mock function with given fields:
func (_m *HostMock) Network() network.Network {
	ret := _m.Called()

	var r0 network.Network
	if rf, ok := ret.Get(0).(func() network.Network); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(network.Network)
		}
	}

	return r0
}

// NewStream provides a mock function with given fields: ctx, p, pids
func (_m *HostMock) NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
	_va := make([]interface{}, len(pids))
	for _i := range pids {
		_va[_i] = pids[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, p)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 network.Stream
	if rf, ok := ret.Get(0).(func(context.Context, peer.ID, ...protocol.ID) network.Stream); ok {
		r0 = rf(ctx, p, pids...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(network.Stream)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, peer.ID, ...protocol.ID) error); ok {
		r1 = rf(ctx, p, pids...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Peerstore provides a mock function with given fields:
func (_m *HostMock) Peerstore() peerstore.Peerstore {
	ret := _m.Called()

	var r0 peerstore.Peerstore
	if rf, ok := ret.Get(0).(func() peerstore.Peerstore); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(peerstore.Peerstore)
		}
	}

	return r0
}

// RemoveStreamHandler provides a mock function with given fields: pid
func (_m *HostMock) RemoveStreamHandler(pid protocol.ID) {
	_m.Called(pid)
}

// SetStreamHandler provides a mock function with given fields: pid, handler
func (_m *HostMock) SetStreamHandler(pid protocol.ID, handler network.StreamHandler) {
	_m.Called(pid, handler)
}

// SetStreamHandlerMatch provides a mock function with given fields: _a0, _a1, _a2
func (_m *HostMock) SetStreamHandlerMatch(_a0 protocol.ID, _a1 func(string) bool, _a2 network.StreamHandler) {
	_m.Called(_a0, _a1, _a2)
}

type NewHostMockT interface {
	mock.TestingT
	Cleanup(func())
}

// NewHostMock creates a new instance of HostMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewHostMock(t NewHostMockT) *HostMock {
	mock := &HostMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
