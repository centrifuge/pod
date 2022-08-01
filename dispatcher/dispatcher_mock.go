package dispatcher

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
)

type DispatcherMock[T any] struct {
	mock.Mock
}

func (_m *DispatcherMock[T]) Dispatch(ctx context.Context, t T) error {
	ret := _m.Called(ctx, t)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, T) error); ok {
		r0 = rf(ctx, t)
	} else {
		r0 = ret.Get(0).(error)
	}

	return r0
}

func (_m *DispatcherMock[T]) Subscribe(ctx context.Context) (chan T, error) {
	ret := _m.Called(ctx)

	var r0 chan T
	if rf, ok := ret.Get(0).(func(context.Context) chan T); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan T)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *DispatcherMock[T]) Unsubscribe(c chan T) error {
	ret := _m.Called(c)

	var r0 error
	if rf, ok := ret.Get(0).(func(chan T) error); ok {
		r0 = rf(c)
	} else {
		r0 = ret.Get(0).(error)
	}

	return r0
}

func (_m *DispatcherMock[T]) Stop() {
	_m.Called()
}

// NewDispatcherMock creates a new instance of DispatcherMock. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewDispatcherMock[T any](t testing.TB) *DispatcherMock[T] {
	mock := &DispatcherMock[T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
