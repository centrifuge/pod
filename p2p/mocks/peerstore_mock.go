// Code generated by mockery v2.13.0-beta.1. DO NOT EDIT.

package mocks

import (
	context "context"

	crypto "github.com/libp2p/go-libp2p-core/crypto"
	mock "github.com/stretchr/testify/mock"

	multiaddr "github.com/multiformats/go-multiaddr"

	peer "github.com/libp2p/go-libp2p-core/peer"

	time "time"
)

// PeerstoreMock is an autogenerated mock type for the Peerstore type
type PeerstoreMock struct {
	mock.Mock
}

// AddAddr provides a mock function with given fields: p, addr, ttl
func (_m *PeerstoreMock) AddAddr(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration) {
	_m.Called(p, addr, ttl)
}

// AddAddrs provides a mock function with given fields: p, addrs, ttl
func (_m *PeerstoreMock) AddAddrs(p peer.ID, addrs []multiaddr.Multiaddr, ttl time.Duration) {
	_m.Called(p, addrs, ttl)
}

// AddPrivKey provides a mock function with given fields: _a0, _a1
func (_m *PeerstoreMock) AddPrivKey(_a0 peer.ID, _a1 crypto.PrivKey) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(peer.ID, crypto.PrivKey) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AddProtocols provides a mock function with given fields: _a0, _a1
func (_m *PeerstoreMock) AddProtocols(_a0 peer.ID, _a1 ...string) error {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(peer.ID, ...string) error); ok {
		r0 = rf(_a0, _a1...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AddPubKey provides a mock function with given fields: _a0, _a1
func (_m *PeerstoreMock) AddPubKey(_a0 peer.ID, _a1 crypto.PubKey) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(peer.ID, crypto.PubKey) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AddrStream provides a mock function with given fields: _a0, _a1
func (_m *PeerstoreMock) AddrStream(_a0 context.Context, _a1 peer.ID) <-chan multiaddr.Multiaddr {
	ret := _m.Called(_a0, _a1)

	var r0 <-chan multiaddr.Multiaddr
	if rf, ok := ret.Get(0).(func(context.Context, peer.ID) <-chan multiaddr.Multiaddr); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan multiaddr.Multiaddr)
		}
	}

	return r0
}

// Addrs provides a mock function with given fields: p
func (_m *PeerstoreMock) Addrs(p peer.ID) []multiaddr.Multiaddr {
	ret := _m.Called(p)

	var r0 []multiaddr.Multiaddr
	if rf, ok := ret.Get(0).(func(peer.ID) []multiaddr.Multiaddr); ok {
		r0 = rf(p)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]multiaddr.Multiaddr)
		}
	}

	return r0
}

// ClearAddrs provides a mock function with given fields: p
func (_m *PeerstoreMock) ClearAddrs(p peer.ID) {
	_m.Called(p)
}

// Close provides a mock function with given fields:
func (_m *PeerstoreMock) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FirstSupportedProtocol provides a mock function with given fields: _a0, _a1
func (_m *PeerstoreMock) FirstSupportedProtocol(_a0 peer.ID, _a1 ...string) (string, error) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 string
	if rf, ok := ret.Get(0).(func(peer.ID, ...string) string); ok {
		r0 = rf(_a0, _a1...)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(peer.ID, ...string) error); ok {
		r1 = rf(_a0, _a1...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: p, key
func (_m *PeerstoreMock) Get(p peer.ID, key string) (interface{}, error) {
	ret := _m.Called(p, key)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(peer.ID, string) interface{}); ok {
		r0 = rf(p, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(peer.ID, string) error); ok {
		r1 = rf(p, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetProtocols provides a mock function with given fields: _a0
func (_m *PeerstoreMock) GetProtocols(_a0 peer.ID) ([]string, error) {
	ret := _m.Called(_a0)

	var r0 []string
	if rf, ok := ret.Get(0).(func(peer.ID) []string); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(peer.ID) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LatencyEWMA provides a mock function with given fields: _a0
func (_m *PeerstoreMock) LatencyEWMA(_a0 peer.ID) time.Duration {
	ret := _m.Called(_a0)

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func(peer.ID) time.Duration); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// PeerInfo provides a mock function with given fields: _a0
func (_m *PeerstoreMock) PeerInfo(_a0 peer.ID) peer.AddrInfo {
	ret := _m.Called(_a0)

	var r0 peer.AddrInfo
	if rf, ok := ret.Get(0).(func(peer.ID) peer.AddrInfo); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(peer.AddrInfo)
	}

	return r0
}

// Peers provides a mock function with given fields:
func (_m *PeerstoreMock) Peers() peer.IDSlice {
	ret := _m.Called()

	var r0 peer.IDSlice
	if rf, ok := ret.Get(0).(func() peer.IDSlice); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(peer.IDSlice)
		}
	}

	return r0
}

// PeersWithAddrs provides a mock function with given fields:
func (_m *PeerstoreMock) PeersWithAddrs() peer.IDSlice {
	ret := _m.Called()

	var r0 peer.IDSlice
	if rf, ok := ret.Get(0).(func() peer.IDSlice); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(peer.IDSlice)
		}
	}

	return r0
}

// PeersWithKeys provides a mock function with given fields:
func (_m *PeerstoreMock) PeersWithKeys() peer.IDSlice {
	ret := _m.Called()

	var r0 peer.IDSlice
	if rf, ok := ret.Get(0).(func() peer.IDSlice); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(peer.IDSlice)
		}
	}

	return r0
}

// PrivKey provides a mock function with given fields: _a0
func (_m *PeerstoreMock) PrivKey(_a0 peer.ID) crypto.PrivKey {
	ret := _m.Called(_a0)

	var r0 crypto.PrivKey
	if rf, ok := ret.Get(0).(func(peer.ID) crypto.PrivKey); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(crypto.PrivKey)
		}
	}

	return r0
}

// PubKey provides a mock function with given fields: _a0
func (_m *PeerstoreMock) PubKey(_a0 peer.ID) crypto.PubKey {
	ret := _m.Called(_a0)

	var r0 crypto.PubKey
	if rf, ok := ret.Get(0).(func(peer.ID) crypto.PubKey); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(crypto.PubKey)
		}
	}

	return r0
}

// Put provides a mock function with given fields: p, key, val
func (_m *PeerstoreMock) Put(p peer.ID, key string, val interface{}) error {
	ret := _m.Called(p, key, val)

	var r0 error
	if rf, ok := ret.Get(0).(func(peer.ID, string, interface{}) error); ok {
		r0 = rf(p, key, val)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RecordLatency provides a mock function with given fields: _a0, _a1
func (_m *PeerstoreMock) RecordLatency(_a0 peer.ID, _a1 time.Duration) {
	_m.Called(_a0, _a1)
}

// RemovePeer provides a mock function with given fields: _a0
func (_m *PeerstoreMock) RemovePeer(_a0 peer.ID) {
	_m.Called(_a0)
}

// RemoveProtocols provides a mock function with given fields: _a0, _a1
func (_m *PeerstoreMock) RemoveProtocols(_a0 peer.ID, _a1 ...string) error {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(peer.ID, ...string) error); ok {
		r0 = rf(_a0, _a1...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetAddr provides a mock function with given fields: p, addr, ttl
func (_m *PeerstoreMock) SetAddr(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration) {
	_m.Called(p, addr, ttl)
}

// SetAddrs provides a mock function with given fields: p, addrs, ttl
func (_m *PeerstoreMock) SetAddrs(p peer.ID, addrs []multiaddr.Multiaddr, ttl time.Duration) {
	_m.Called(p, addrs, ttl)
}

// SetProtocols provides a mock function with given fields: _a0, _a1
func (_m *PeerstoreMock) SetProtocols(_a0 peer.ID, _a1 ...string) error {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(peer.ID, ...string) error); ok {
		r0 = rf(_a0, _a1...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SupportsProtocols provides a mock function with given fields: _a0, _a1
func (_m *PeerstoreMock) SupportsProtocols(_a0 peer.ID, _a1 ...string) ([]string, error) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []string
	if rf, ok := ret.Get(0).(func(peer.ID, ...string) []string); ok {
		r0 = rf(_a0, _a1...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(peer.ID, ...string) error); ok {
		r1 = rf(_a0, _a1...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateAddrs provides a mock function with given fields: p, oldTTL, newTTL
func (_m *PeerstoreMock) UpdateAddrs(p peer.ID, oldTTL time.Duration, newTTL time.Duration) {
	_m.Called(p, oldTTL, newTTL)
}

type NewPeerstoreMockT interface {
	mock.TestingT
	Cleanup(func())
}

// NewPeerstoreMock creates a new instance of PeerstoreMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewPeerstoreMock(t NewPeerstoreMockT) *PeerstoreMock {
	mock := &PeerstoreMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
