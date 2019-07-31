// +build unit

package documents

import (
	"context"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/jobs"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Exists(accountID, id []byte) bool {
	args := m.Called(accountID, id)
	return args.Get(0).(bool)
}

func (m *MockRepository) Get(accountID, id []byte) (Model, error) {
	args := m.Called(accountID, id)
	doc, _ := args.Get(0).(Model)
	return doc, args.Error(0)
}

func (m *MockRepository) Create(accountID, id []byte, model Model) error {
	args := m.Called(accountID, id)
	return args.Error(0)
}

func (m *MockRepository) Update(accountID, id []byte, model Model) error {
	args := m.Called(accountID, id)
	return args.Error(0)
}

func (m *MockRepository) Register(model Model) {
	m.Called(model)
	return
}

func (m *MockRepository) GetLatest(accountID, docID []byte) (Model, error) {
	args := m.Called(accountID, docID)
	doc, _ := args.Get(0).(Model)
	return doc, args.Error(1)
}

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

func TestService_Validate(t *testing.T) {
	r := NewServiceRegistry()
	scheme := "invoice"
	srv := new(MockService)
	srv.On("Validate", mock.Anything, mock.Anything).Return(nil)
	err := r.Register(scheme, srv)
	assert.NoError(t, err)

	// unsupported svc schema
	m := new(mockModel)
	m.On("Scheme", mock.Anything).Return("some scheme")
	s := service{registry: r}
	err = s.Validate(context.Background(), m)
	assert.Error(t, err)

	// version fetch error
	m = new(mockModel)
	id := utils.RandomSlice(32)
	m.On("ID", mock.Anything).Return(id)
	m.On("Scheme", mock.Anything).Return("invoice")
	err = s.Validate(context.Background(), m)
	assert.Error(t, err)

	// create validation error, already anchored
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	m = new(mockModel)
	nid := utils.RandomSlice(32)
	m.On("ID", mock.Anything).Return(id)
	m.On("CurrentVersion").Return(id)
	m.On("NextVersion").Return(nid)
	m.On("PreviousVersion").Return(nid)
	m.On("Scheme", mock.Anything).Return("invoice")
	mr := new(MockRepository)
	mr.On("GetLatest", mock.Anything, mock.Anything).Return(nil, ErrDocumentVersionNotFound)
	s.repo = mr
	anchorRepo := new(mockRepo)
	anchorRepo.On("GetAnchorData", mock.Anything).Return(utils.RandomSlice(32), time.Now(), nil)
	s.anchorRepo = anchorRepo
	err = s.Validate(ctxh, m)
	assert.Error(t, err)

	// create validation success
	anchorRepo = new(mockRepo)
	anchorRepo.On("GetAnchorData", mock.Anything).Return(id, time.Now(), errors.New("anchor data missing"))
	s.anchorRepo = anchorRepo
	err = s.Validate(ctxh, m)
	assert.NoError(t, err)

	// Update validation error, already anchored
	m1 := new(mockModel)
	nid1 := utils.RandomSlice(32)
	m1.On("ID", mock.Anything).Return(id)
	m1.On("CurrentVersion").Return(nid)
	m1.On("NextVersion").Return(nid1)
	m1.On("PreviousVersion").Return(id)
	m1.On("Scheme", mock.Anything).Return("invoice")
	mr = new(MockRepository)
	mr.On("GetLatest", mock.Anything, mock.Anything).Return(m, nil)
	s.repo = mr
	anchorRepo = new(mockRepo)
	anchorRepo.On("GetAnchorData", mock.Anything).Return(utils.RandomSlice(32), time.Now(), nil)
	s.anchorRepo = anchorRepo
	err = s.Validate(ctxh, m1)
	assert.Error(t, err)

	// update validation success
	anchorRepo = new(mockRepo)
	anchorRepo.On("GetAnchorData", mock.Anything).Return(id, time.Now(), errors.New("anchor data missing"))
	s.anchorRepo = anchorRepo
	err = s.Validate(ctxh, m1)
	assert.NoError(t, err)

	// specific document validation error
	r = NewServiceRegistry()
	srv = new(MockService)
	srv.On("Validate", mock.Anything, mock.Anything).Return(errors.New("specific document error"))
	err = r.Register(scheme, srv)
	assert.NoError(t, err)
	s.registry = r
	err = s.Validate(ctxh, m1)
	assert.Error(t, err)
}

func TestService_Commit(t *testing.T) {
	r := NewServiceRegistry()
	scheme := "invoice"
	srv := new(MockService)
	srv.On("Validate", mock.Anything, mock.Anything).Return(nil)
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

	// Fail validation
	nid := utils.RandomSlice(32)
	m.On("CurrentVersion").Return(id)
	m.On("NextVersion").Return(nid)
	m.On("PreviousVersion").Return(nid)
	mr := new(MockRepository)
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
