// +build unit

package documents

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Derive(t *testing.T) {
	r := NewServiceRegistry()
	scheme := "invoice"
	srv := new(MockService)
	srv.On("Derive", mock.Anything, mock.Anything).Return(new(mockModel), nil)
	err := r.Register(scheme, srv)
	assert.NoError(t, err)

	// missing service
	payload := UpdatePayload{CreatePayload: CreatePayload{Scheme: "some scheme"}}
	s := service{registry: r}
	_, err = s.Derive(context.Background(), payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrDocumentSchemeUnknown, err))

	// success
	payload.Scheme = scheme
	m, err := s.Derive(context.Background(), payload)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	srv.AssertExpectations(t)
}
