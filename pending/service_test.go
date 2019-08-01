// +build unit

package pending

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
	Repository
}

func (m *mockRepo) Get(accID, id []byte) (documents.Model, error) {
	args := m.Called(accID, id)
	doc, _ := args.Get(0).(documents.Model)
	return doc, args.Error(1)
}

func (m *mockRepo) Delete(accID, id []byte) error {
	args := m.Called(accID, id)
	return args.Error(0)
}

func (m *mockRepo) Create(accID, id []byte, doc documents.Model) error {
	args := m.Called(accID, id, doc)
	return args.Error(0)
}

func (m *mockRepo) Update(accID, id []byte, doc documents.Model) error {
	args := m.Called(accID, id, doc)
	return args.Error(0)
}

func TestService_Commit(t *testing.T) {
	s := service{}

	// missing did
	ctx := context.Background()
	docID := utils.RandomSlice(32)
	_, err := s.Commit(ctx, docID)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrDIDMissingFromContext, err))

	// missing model
	ctx = testingconfig.CreateAccountContext(t, cfg)
	repo := new(mockRepo)
	repo.On("Get", did[:], docID).Return(nil, errors.New("not found")).Once()
	s.pendingRepo = repo
	_, err = s.Commit(ctx, docID)
	assert.Error(t, err)

	// failed commit
	doc := new(documents.MockModel)
	repo.On("Get", did[:], docID).Return(doc, nil)
	docSrv := new(testingdocuments.MockService)
	docSrv.On("Commit", ctx, doc).Return(nil, errors.New("failed to commit")).Once()
	s.docSrv = docSrv
	_, err = s.Commit(ctx, docID)
	assert.Error(t, err)

	// success
	jobID := jobs.NewJobID()
	repo.On("Delete", did[:], docID).Return(nil)
	docSrv.On("Commit", ctx, doc).Return(jobID, nil)
	jid, err := s.Commit(ctx, docID)
	assert.NoError(t, err)
	assert.Equal(t, jobID, jid)
	docSrv.AssertExpectations(t)
	doc.AssertExpectations(t)
}

func TestService_Create(t *testing.T) {
	s := service{}

	// missing did
	ctx := context.Background()
	payload := documents.UpdatePayload{}
	_, err := s.Create(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrDIDMissingFromContext, err))

	// derive failed
	ctx = testingconfig.CreateAccountContext(t, cfg)
	docSrv := new(testingdocuments.MockService)
	docSrv.On("Derive", ctx, payload).Return(nil, errors.New("failed to derive")).Once()
	s.docSrv = docSrv
	_, err = s.Create(ctx, payload)
	assert.Error(t, err)

	// already existing document
	payload.DocumentID = utils.RandomSlice(32)
	repo := new(mockRepo)
	repo.On("Get", did[:], payload.DocumentID).Return(new(documents.MockModel), nil).Once()
	s.pendingRepo = repo
	_, err = s.Create(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrPendingDocumentExists, err))

	// success
	repo.On("Get", did[:], payload.DocumentID).Return(nil, errors.New("missing")).Once()
	doc := new(documents.MockModel)
	doc.On("ID").Return(payload.DocumentID).Once()
	repo.On("Create", did[:], payload.DocumentID, doc).Return(nil).Once()
	docSrv.On("Derive", ctx, payload).Return(doc, nil).Once()
	gdoc, err := s.Create(ctx, payload)
	assert.NoError(t, err)
	assert.Equal(t, doc, gdoc)
	doc.AssertExpectations(t)
	docSrv.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestService_Update(t *testing.T) {
	s := service{}

	// missing did
	ctx := context.Background()
	payload := documents.UpdatePayload{}
	_, err := s.Update(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrDIDMissingFromContext, err))

	// document doesnt exist yet
	repo := new(mockRepo)
	repo.On("Get", did[:], payload.DocumentID).Return(nil, errors.New("not found")).Once()
	s.pendingRepo = repo
	ctx = testingconfig.CreateAccountContext(t, cfg)
	_, err = s.Update(ctx, payload)
	assert.Error(t, err)

	// wrong status
	oldModel := new(documents.MockModel)
	oldModel.On("GetStatus").Return(documents.Committing).Once()
	repo.On("Get", did[:], payload.DocumentID).Return(oldModel, nil).Once()
	_, err = s.Update(ctx, payload)
	assert.Error(t, err)

	// Patch error
	oldModel.On("GetStatus").Return(documents.Pending)
	oldModel.On("Patch", payload).Return(nil, errors.New("error patching")).Once()
	repo.On("Get", did[:], payload.DocumentID).Return(oldModel, nil)
	_, err = s.Update(ctx, payload)
	assert.Error(t, err)

	// Success
	oldModel.On("ID").Return(payload.DocumentID).Once()
	oldModel.On("Patch", payload).Return(oldModel, nil).Once()
	repo.On("Update", did[:], payload.DocumentID, oldModel).Return(nil).Once()
	_, err = s.Update(ctx, payload)
	assert.NoError(t, err)

}
