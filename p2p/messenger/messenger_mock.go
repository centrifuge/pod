// Code generated by mockery v2.13.0-beta.1. DO NOT EDIT.

package messenger

import (
	context "context"

	peer "github.com/libp2p/go-libp2p-core/peer"
	mock "github.com/stretchr/testify/mock"

	protocol "github.com/libp2p/go-libp2p-core/protocol"

	protocolpb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
)

// MessengerMock is an autogenerated mock type for the Messenger type
type MessengerMock struct {
	mock.Mock
}

// Init provides a mock function with given fields: protocols
func (_m *MessengerMock) Init(protocols ...protocol.ID) {
	_va := make([]interface{}, len(protocols))
	for _i := range protocols {
		_va[_i] = protocols[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// SendMessage provides a mock function with given fields: ctx, peerID, mes, protocolID
func (_m *MessengerMock) SendMessage(ctx context.Context, peerID peer.ID, mes *protocolpb.P2PEnvelope, protocolID protocol.ID) (*protocolpb.P2PEnvelope, error) {
	ret := _m.Called(ctx, peerID, mes, protocolID)

	var r0 *protocolpb.P2PEnvelope
	if rf, ok := ret.Get(0).(func(context.Context, peer.ID, *protocolpb.P2PEnvelope, protocol.ID) *protocolpb.P2PEnvelope); ok {
		r0 = rf(ctx, peerID, mes, protocolID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*protocolpb.P2PEnvelope)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, peer.ID, *protocolpb.P2PEnvelope, protocol.ID) error); ok {
		r1 = rf(ctx, peerID, mes, protocolID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewMessengerMockT interface {
	mock.TestingT
	Cleanup(func())
}

// NewMessengerMock creates a new instance of MessengerMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMessengerMock(t NewMessengerMockT) *MessengerMock {
	mock := &MessengerMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
