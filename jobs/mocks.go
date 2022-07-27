//go:build unit || integration
// +build unit integration

package jobs

import (
	"context"
	"sync"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	"github.com/stretchr/testify/mock"
)

func (b Bootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	return b.Bootstrap(ctx)
}

func (b Bootstrapper) TestTearDown() error {
	return nil
}

type MockResult struct {
	mock.Mock
	Result
}

// DispatcherMock is an autogenerated mock type for the Dispatcher type
type DispatcherMock struct {
	mock.Mock
}

// Dispatch provides a mock function with given fields: accountID, job
func (_m *DispatcherMock) Dispatch(accountID *types.AccountID, job *gocelery.Job) (Result, error) {
	ret := _m.Called(accountID, job)

	var r0 Result
	if rf, ok := ret.Get(0).(func(*types.AccountID, *gocelery.Job) Result); ok {
		r0 = rf(accountID, job)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Result)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*types.AccountID, *gocelery.Job) error); ok {
		r1 = rf(accountID, job)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Job provides a mock function with given fields: accountID, jobID
func (_m *DispatcherMock) Job(accountID *types.AccountID, jobID gocelery.JobID) (*gocelery.Job, error) {
	ret := _m.Called(accountID, jobID)

	var r0 *gocelery.Job
	if rf, ok := ret.Get(0).(func(*types.AccountID, gocelery.JobID) *gocelery.Job); ok {
		r0 = rf(accountID, jobID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gocelery.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*types.AccountID, gocelery.JobID) error); ok {
		r1 = rf(accountID, jobID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Name provides a mock function with given fields:
func (_m *DispatcherMock) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// RegisterRunner provides a mock function with given fields: name, runner
func (_m *DispatcherMock) RegisterRunner(name string, runner gocelery.Runner) bool {
	ret := _m.Called(name, runner)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, gocelery.Runner) bool); ok {
		r0 = rf(name, runner)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// RegisterRunnerFunc provides a mock function with given fields: name, runnerFunc
func (_m *DispatcherMock) RegisterRunnerFunc(name string, runnerFunc gocelery.RunnerFunc) bool {
	ret := _m.Called(name, runnerFunc)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, gocelery.RunnerFunc) bool); ok {
		r0 = rf(name, runnerFunc)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Result provides a mock function with given fields: accountID, jobID
func (_m *DispatcherMock) Result(accountID *types.AccountID, jobID gocelery.JobID) (Result, error) {
	ret := _m.Called(accountID, jobID)

	var r0 Result
	if rf, ok := ret.Get(0).(func(*types.AccountID, gocelery.JobID) Result); ok {
		r0 = rf(accountID, jobID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Result)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*types.AccountID, gocelery.JobID) error); ok {
		r1 = rf(accountID, jobID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Start provides a mock function with given fields: ctx, wg, startupErr
func (_m *DispatcherMock) Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error) {
	_m.Called(ctx, wg, startupErr)
}

// NewDispatcherMock creates a new instance of DispatcherMock. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewDispatcherMock(t testing.TB) *DispatcherMock {
	mock := &DispatcherMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
