// +build unit

package coreapi

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_CreateDocument(t *testing.T) {
	docSrv := new(testingdocuments.MockService)
	srv := Service{docSrv: docSrv}
	m := new(testingdocuments.MockModel)
	docSrv.On("CreateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	nm, _, err := srv.CreateDocument(context.Background(), documents.CreatePayload{})
	assert.NoError(t, err)
	assert.Equal(t, m, nm)
}

func TestService_UpdateDocument(t *testing.T) {
	docSrv := new(testingdocuments.MockService)
	srv := Service{docSrv: docSrv}
	m := new(testingdocuments.MockModel)
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	nm, _, err := srv.UpdateDocument(context.Background(), documents.UpdatePayload{})
	assert.NoError(t, err)
	assert.Equal(t, m, nm)
}
