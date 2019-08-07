// +build integration unit

package pending

import (
	"context"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/stretchr/testify/mock"
)

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (Bootstrapper) TestTearDown() error {
	return nil
}

type MockService struct {
	mock.Mock
	Service
}

func (m *MockService) Create(ctx context.Context, payload documents.UpdatePayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	doc, _ := args.Get(0).(documents.Model)
	return doc, args.Error(1)
}

func (m *MockService) Update(ctx context.Context, payload documents.UpdatePayload) (documents.Model, error) {
	args := m.Called(ctx, payload)
	doc, _ := args.Get(0).(documents.Model)
	return doc, args.Error(1)
}

func (m *MockService) Commit(ctx context.Context, docID []byte) (documents.Model, jobs.JobID, error) {
	args := m.Called(ctx, docID)
	doc, _ := args.Get(0).(documents.Model)
	jobID, _ := args.Get(1).(jobs.JobID)
	return doc, jobID, args.Error(2)
}

func (m *MockService) Get(ctx context.Context, docID []byte, st documents.Status) (documents.Model, error) {
	args := m.Called(ctx, docID, st)
	doc, _ := args.Get(0).(documents.Model)
	return doc, args.Error(1)
}
