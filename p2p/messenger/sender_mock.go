// Code generated by mockery v2.13.0-beta.1. DO NOT EDIT.

package messenger

import (
	context "context"

	protocolpb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	mock "github.com/stretchr/testify/mock"
)

// MessageSenderMock is an autogenerated mock type for the MessageSender type
type MessageSenderMock struct {
	mock.Mock
}

// Prepare provides a mock function with given fields:
func (_m *MessageSenderMock) Prepare() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendMessage provides a mock function with given fields: ctx, pmes
func (_m *MessageSenderMock) SendMessage(ctx context.Context, pmes *protocolpb.P2PEnvelope) (*protocolpb.P2PEnvelope, error) {
	ret := _m.Called(ctx, pmes)

	var r0 *protocolpb.P2PEnvelope
	if rf, ok := ret.Get(0).(func(context.Context, *protocolpb.P2PEnvelope) *protocolpb.P2PEnvelope); ok {
		r0 = rf(ctx, pmes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*protocolpb.P2PEnvelope)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *protocolpb.P2PEnvelope) error); ok {
		r1 = rf(ctx, pmes)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewMessageSenderMockT interface {
	mock.TestingT
	Cleanup(func())
}

// NewMessageSenderMock creates a new instance of MessageSenderMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMessageSenderMock(t NewMessageSenderMockT) *MessageSenderMock {
	mock := &MessageSenderMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
