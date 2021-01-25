// +build unit integration

package jobsv2

import (
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/gocelery/v2"
	"github.com/stretchr/testify/mock"
)

func (b Bootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	return b.Bootstrap(ctx)
}

func (b Bootstrapper) TestTearDown() error {
	return nil
}

type MockDispatcher struct {
	mock.Mock
	Dispatcher
}

func (m *MockDispatcher) Job(acc identity.DID, jobID []byte) (*gocelery.Job, error) {
	args := m.Called(acc, jobID)
	job, _ := args.Get(0).(*gocelery.Job)
	return job, args.Error(1)
}
