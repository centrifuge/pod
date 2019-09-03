// +build unit

package documents

import (
	"context"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Validate(t *testing.T) {
	r := NewServiceRegistry()
	scheme := "invoice"
	srv := new(MockService)
	srv.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := r.Register(scheme, srv)
	assert.NoError(t, err)

	// unsupported svc schema
	m := new(mockModel)
	m.On("Scheme", mock.Anything).Return("some scheme")
	s := service{registry: r}
	err = s.Validate(context.Background(), m, nil)
	assert.Error(t, err)

	// create validation error, already anchored
	id := utils.RandomSlice(32)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	m = new(mockModel)
	nid := utils.RandomSlice(32)
	m.On("ID", mock.Anything).Return(id)
	m.On("CurrentVersion").Return(id)
	m.On("NextVersion").Return(nid)
	m.On("PreviousVersion").Return(nid)
	m.On("Scheme", mock.Anything).Return("invoice")
	anchorRepo := new(mockRepo)
	anchorRepo.On("GetAnchorData", mock.Anything).Return(utils.RandomSlice(32), time.Now(), nil)
	s.anchorRepo = anchorRepo
	err = s.Validate(ctxh, m, nil)
	assert.Error(t, err)

	// create validation success
	anchorRepo = new(mockRepo)
	anchorRepo.On("GetAnchorData", mock.Anything).Return(id, time.Now(), errors.New("anchor data missing"))
	s.anchorRepo = anchorRepo
	err = s.Validate(ctxh, m, nil)
	assert.NoError(t, err)

	// Update validation error, already anchored
	m1 := new(mockModel)
	nid1 := utils.RandomSlice(32)
	m1.On("ID", mock.Anything).Return(id)
	m1.On("CurrentVersion").Return(nid)
	m1.On("NextVersion").Return(nid1)
	m1.On("PreviousVersion").Return(id)
	m1.On("Scheme", mock.Anything).Return("invoice")
	anchorRepo = new(mockRepo)
	anchorRepo.On("GetAnchorData", mock.Anything).Return(utils.RandomSlice(32), time.Now(), nil)
	s.anchorRepo = anchorRepo
	err = s.Validate(ctxh, m1, m)
	assert.Error(t, err)

	// update validation success
	anchorRepo = new(mockRepo)
	anchorRepo.On("GetAnchorData", mock.Anything).Return(id, time.Now(), errors.New("anchor data missing"))
	s.anchorRepo = anchorRepo
	err = s.Validate(ctxh, m1, m)
	assert.NoError(t, err)

	// specific document validation error
	r = NewServiceRegistry()
	srv = new(MockService)
	srv.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("specific document error"))
	err = r.Register(scheme, srv)
	assert.NoError(t, err)
	s.registry = r
	err = s.Validate(ctxh, m1, m)
	assert.Error(t, err)
}

func TestService_Commit(t *testing.T) {
	r := NewServiceRegistry()
	scheme := "invoice"
	srv := new(MockService)
	srv.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := r.Register(scheme, srv)
	assert.NoError(t, err)
	s := service{registry: r}
	m := new(mockModel)
	id := utils.RandomSlice(32)
	m.On("ID", mock.Anything).Return(id)
	m.On("Scheme", mock.Anything).Return("invoice")

	// Account ID not set
	_, err = s.Commit(context.Background(), m)
	assert.Error(t, err)

	// db error when fetching
	mr := new(MockRepository)
	mr.On("GetLatest", mock.Anything, mock.Anything).Return(nil, errors.New("some db error")).Once()
	s.repo = mr
	_, err = s.Commit(context.Background(), m)
	assert.Error(t, err)

	// Fail validation
	nid := utils.RandomSlice(32)
	m.On("CurrentVersion").Return(id)
	m.On("NextVersion").Return(nid)
	m.On("PreviousVersion").Return(nid)
	mr = new(MockRepository)
	mr.On("GetLatest", mock.Anything, mock.Anything).Return(nil, ErrDocumentVersionNotFound)
	s.repo = mr
	anchorRepo := new(mockRepo)
	anchorRepo.On("GetAnchorData", mock.Anything).Return(utils.RandomSlice(32), time.Now(), nil)
	s.anchorRepo = anchorRepo
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, err = s.Commit(ctxh, m)
	assert.Error(t, err)

	// Error create model
	anchorRepo = new(mockRepo)
	anchorRepo.On("GetAnchorData", mock.Anything).Return(nil, time.Now(), errors.New("anchor data missing"))
	s.anchorRepo = anchorRepo
	m.On("SetStatus", mock.Anything).Return(nil)
	mr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(ErrDocumentPersistence)
	_, err = s.Commit(ctxh, m)
	assert.Error(t, err)

	// Error anchoring
	jobMan := &testingjobs.MockJobManager{}
	jobMan.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), make(chan error), errors.New("error anchoring"))
	s.jobManager = jobMan
	mr = new(MockRepository)
	mr.On("GetLatest", mock.Anything, mock.Anything).Return(nil, ErrDocumentVersionNotFound)
	mr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.repo = mr
	_, err = s.Commit(ctxh, m)
	assert.Error(t, err)

	// Commit success
	jobMan = &testingjobs.MockJobManager{}
	jobMan.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), make(chan error), nil)
	s.jobManager = jobMan
	_, err = s.Commit(ctxh, m)
	assert.NoError(t, err)
}

func TestService_Derive(t *testing.T) {
	scheme := "invoice"
	payload := UpdatePayload{CreatePayload: CreatePayload{Scheme: scheme}}
	s := service{}

	// missing account ctx
	ctx := context.Background()
	_, err := s.Derive(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrDocumentConfigAccountID, err))

	// unknown scheme
	ctx = testingconfig.CreateAccountContext(t, cfg)
	s.registry = NewServiceRegistry()
	_, err = s.Derive(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrDocumentSchemeUnknown, err))

	// derive failed
	doc := new(MockModel)
	docSrv := new(MockService)
	docSrv.On("New", scheme).Return(doc, nil)
	doc.On("DeriveFromCreatePayload", mock.Anything, mock.Anything).Return(errors.New("derive failed")).Once()
	assert.NoError(t, s.registry.Register(scheme, docSrv))
	_, err = s.Derive(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrDocumentInvalid, err))

	// create successful
	doc.On("DeriveFromCreatePayload", mock.Anything, mock.Anything).Return(nil).Once()
	gdoc, err := s.Derive(ctx, payload)
	assert.NoError(t, err)
	assert.Equal(t, doc, gdoc)

	// missing old version
	docID := utils.RandomSlice(32)
	repo := new(MockRepository)
	repo.On("GetLatest", did[:], docID).Return(nil, ErrDocumentNotFound).Once()
	s.repo = repo
	payload.DocumentID = docID
	_, err = s.Derive(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrDocumentNotFound, err))

	// invalid type
	doc.On("Scheme").Return("invalid").Once()
	repo.On("GetLatest", did[:], docID).Return(doc, nil)
	_, err = s.Derive(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrDocumentInvalidType, err))

	// DeriveFromUpdatePayload failed
	doc.On("Scheme").Return(scheme)
	doc.On("DeriveFromUpdatePayload", mock.Anything, payload).Return(nil, ErrDocumentInvalid).Once()
	_, err = s.Derive(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrDocumentInvalid, err))

	// success
	doc.On("DeriveFromUpdatePayload", mock.Anything, payload).Return(doc, nil).Once()
	gdoc, err = s.Derive(ctx, payload)
	assert.NoError(t, err)
	assert.Equal(t, gdoc, doc)
	doc.AssertExpectations(t)
	repo.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}
