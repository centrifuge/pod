// +build integration unit

package pending

import (
	"context"

	"github.com/centrifuge/go-centrifuge/documents"
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
