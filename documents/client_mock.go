// Code generated by mockery v2.13.0-beta.1. DO NOT EDIT.

package documents

import (
	context "context"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	mock "github.com/stretchr/testify/mock"

	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"

	types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// ClientMock is an autogenerated mock type for the Client type
type ClientMock struct {
	mock.Mock
}

// GetDocumentRequest provides a mock function with given fields: ctx, documentOwner, in
func (_m *ClientMock) GetDocumentRequest(ctx context.Context, documentOwner *types.AccountID, in *p2ppb.GetDocumentRequest) (*p2ppb.GetDocumentResponse, error) {
	ret := _m.Called(ctx, documentOwner, in)

	var r0 *p2ppb.GetDocumentResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.AccountID, *p2ppb.GetDocumentRequest) *p2ppb.GetDocumentResponse); ok {
		r0 = rf(ctx, documentOwner, in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*p2ppb.GetDocumentResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.AccountID, *p2ppb.GetDocumentRequest) error); ok {
		r1 = rf(ctx, documentOwner, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSignaturesForDocument provides a mock function with given fields: ctx, model
func (_m *ClientMock) GetSignaturesForDocument(ctx context.Context, model Document) ([]*coredocumentpb.Signature, []error, error) {
	ret := _m.Called(ctx, model)

	var r0 []*coredocumentpb.Signature
	if rf, ok := ret.Get(0).(func(context.Context, Document) []*coredocumentpb.Signature); ok {
		r0 = rf(ctx, model)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*coredocumentpb.Signature)
		}
	}

	var r1 []error
	if rf, ok := ret.Get(1).(func(context.Context, Document) []error); ok {
		r1 = rf(ctx, model)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]error)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, Document) error); ok {
		r2 = rf(ctx, model)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// SendAnchoredDocument provides a mock function with given fields: ctx, receiverID, in
func (_m *ClientMock) SendAnchoredDocument(ctx context.Context, receiverID *types.AccountID, in *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error) {
	ret := _m.Called(ctx, receiverID, in)

	var r0 *p2ppb.AnchorDocumentResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.AccountID, *p2ppb.AnchorDocumentRequest) *p2ppb.AnchorDocumentResponse); ok {
		r0 = rf(ctx, receiverID, in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*p2ppb.AnchorDocumentResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.AccountID, *p2ppb.AnchorDocumentRequest) error); ok {
		r1 = rf(ctx, receiverID, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewClientMockT interface {
	mock.TestingT
	Cleanup(func())
}

// NewClientMock creates a new instance of ClientMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewClientMock(t NewClientMockT) *ClientMock {
	mock := &ClientMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
