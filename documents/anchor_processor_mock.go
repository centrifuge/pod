// Code generated by mockery v2.13.0-beta.1. DO NOT EDIT.

package documents

import (
	context "context"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	mock "github.com/stretchr/testify/mock"

	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"

	types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// AnchorProcessorMock is an autogenerated mock type for the AnchorProcessor type
type AnchorProcessorMock struct {
	mock.Mock
}

// AnchorDocument provides a mock function with given fields: ctx, doc
func (_m *AnchorProcessorMock) AnchorDocument(ctx context.Context, doc Document) error {
	ret := _m.Called(ctx, doc)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, Document) error); ok {
		r0 = rf(ctx, doc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PreAnchorDocument provides a mock function with given fields: ctx, doc
func (_m *AnchorProcessorMock) PreAnchorDocument(ctx context.Context, doc Document) error {
	ret := _m.Called(ctx, doc)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, Document) error); ok {
		r0 = rf(ctx, doc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PrepareForAnchoring provides a mock function with given fields: ctx, doc
func (_m *AnchorProcessorMock) PrepareForAnchoring(ctx context.Context, doc Document) error {
	ret := _m.Called(ctx, doc)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, Document) error); ok {
		r0 = rf(ctx, doc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PrepareForSignatureRequests provides a mock function with given fields: ctx, doc
func (_m *AnchorProcessorMock) PrepareForSignatureRequests(ctx context.Context, doc Document) error {
	ret := _m.Called(ctx, doc)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, Document) error); ok {
		r0 = rf(ctx, doc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RequestDocumentWithAccessToken provides a mock function with given fields: ctx, granterDID, tokenIdentifier, documentIdentifier, delegatingDocumentIdentifier
func (_m *AnchorProcessorMock) RequestDocumentWithAccessToken(ctx context.Context, granterDID *types.AccountID, tokenIdentifier []byte, documentIdentifier []byte, delegatingDocumentIdentifier []byte) (*p2ppb.GetDocumentResponse, error) {
	ret := _m.Called(ctx, granterDID, tokenIdentifier, documentIdentifier, delegatingDocumentIdentifier)

	var r0 *p2ppb.GetDocumentResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.AccountID, []byte, []byte, []byte) *p2ppb.GetDocumentResponse); ok {
		r0 = rf(ctx, granterDID, tokenIdentifier, documentIdentifier, delegatingDocumentIdentifier)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*p2ppb.GetDocumentResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.AccountID, []byte, []byte, []byte) error); ok {
		r1 = rf(ctx, granterDID, tokenIdentifier, documentIdentifier, delegatingDocumentIdentifier)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RequestSignatures provides a mock function with given fields: ctx, doc
func (_m *AnchorProcessorMock) RequestSignatures(ctx context.Context, doc Document) error {
	ret := _m.Called(ctx, doc)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, Document) error); ok {
		r0 = rf(ctx, doc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Send provides a mock function with given fields: ctx, cd, recipient
func (_m *AnchorProcessorMock) Send(ctx context.Context, cd *coredocumentpb.CoreDocument, recipient *types.AccountID) error {
	ret := _m.Called(ctx, cd, recipient)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *coredocumentpb.CoreDocument, *types.AccountID) error); ok {
		r0 = rf(ctx, cd, recipient)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendDocument provides a mock function with given fields: ctx, doc
func (_m *AnchorProcessorMock) SendDocument(ctx context.Context, doc Document) error {
	ret := _m.Called(ctx, doc)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, Document) error); ok {
		r0 = rf(ctx, doc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type NewAnchorProcessorMockT interface {
	mock.TestingT
	Cleanup(func())
}

// NewAnchorProcessorMock creates a new instance of AnchorProcessorMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAnchorProcessorMock(t NewAnchorProcessorMockT) *AnchorProcessorMock {
	mock := &AnchorProcessorMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
