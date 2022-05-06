package v3

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/mock"
)

// UniquesAPIMock is an autogenerated mock type for the UniquesAPI type
type UniquesAPIMock struct {
	mock.Mock
}

// CreateClass provides a mock function with given fields: ctx, classID
func (_m *UniquesAPIMock) CreateClass(ctx context.Context, classID types.U64) (*centchain.ExtrinsicInfo, error) {
	ret := _m.Called(ctx, classID)

	var r0 *centchain.ExtrinsicInfo
	if rf, ok := ret.Get(0).(func(context.Context, types.U64) *centchain.ExtrinsicInfo); ok {
		r0 = rf(ctx, classID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*centchain.ExtrinsicInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.U64) error); ok {
		r1 = rf(ctx, classID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetClassDetails provides a mock function with given fields: ctx, classID
func (_m *UniquesAPIMock) GetClassDetails(ctx context.Context, classID types.U64) (*types.ClassDetails, error) {
	ret := _m.Called(ctx, classID)

	var r0 *types.ClassDetails
	if rf, ok := ret.Get(0).(func(context.Context, types.U64) *types.ClassDetails); ok {
		r0 = rf(ctx, classID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.ClassDetails)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.U64) error); ok {
		r1 = rf(ctx, classID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetInstanceDetails provides a mock function with given fields: ctx, classID, instanceID
func (_m *UniquesAPIMock) GetInstanceDetails(ctx context.Context, classID types.U64, instanceID types.U128) (*types.InstanceDetails, error) {
	ret := _m.Called(ctx, classID, instanceID)

	var r0 *types.InstanceDetails
	if rf, ok := ret.Get(0).(func(context.Context, types.U64, types.U128) *types.InstanceDetails); ok {
		r0 = rf(ctx, classID, instanceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.InstanceDetails)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.U64, types.U128) error); ok {
		r1 = rf(ctx, classID, instanceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MintInstance provides a mock function with given fields: ctx, classID, instanceID, owner
func (_m *UniquesAPIMock) MintInstance(ctx context.Context, classID types.U64, instanceID types.U128, owner types.AccountID) (*centchain.ExtrinsicInfo, error) {
	ret := _m.Called(ctx, classID, instanceID, owner)

	var r0 *centchain.ExtrinsicInfo
	if rf, ok := ret.Get(0).(func(context.Context, types.U64, types.U128, types.AccountID) *centchain.ExtrinsicInfo); ok {
		r0 = rf(ctx, classID, instanceID, owner)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*centchain.ExtrinsicInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.U64, types.U128, types.AccountID) error); ok {
		r1 = rf(ctx, classID, instanceID, owner)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewUniquesAPIMock creates a new instance of UniquesAPIMock. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewUniquesAPIMock(t testing.TB) *UniquesAPIMock {
	mock := &UniquesAPIMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// ServiceMock is an autogenerated mock type for the Service type
type ServiceMock struct {
	mock.Mock
}

// MintNFT provides a mock function with given fields: ctx, req
func (_m *ServiceMock) MintNFT(ctx context.Context, req *MintNFTRequest) (*MintNFTResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *MintNFTResponse
	if rf, ok := ret.Get(0).(func(context.Context, *MintNFTRequest) *MintNFTResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*MintNFTResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *MintNFTRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OwnerOf provides a mock function with given fields: ctx, req
func (_m *ServiceMock) OwnerOf(ctx context.Context, req *OwnerOfRequest) (*OwnerOfResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *OwnerOfResponse
	if rf, ok := ret.Get(0).(func(context.Context, *OwnerOfRequest) *OwnerOfResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*OwnerOfResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *OwnerOfRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewServiceMock creates a new instance of ServiceMock. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewServiceMock(t testing.TB) *ServiceMock {
	mock := &ServiceMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
