// +build unit

package generic

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/testingutils"
	testinganchors "github.com/centrifuge/go-centrifuge/testingutils/anchors"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_CreateModel(t *testing.T) {
	payload := documents.CreatePayload{}
	srv := service{}

	// empty context
	_, _, err := srv.CreateModel(context.Background(), payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentConfigAccountID, err))

	// success
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	payload.Data = validData(t)
	srv.repo = testRepo()
	jm := testingjobs.MockJobManager{}
	jm.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), make(chan error), nil)
	srv.jobManager = jm
	m, _, err := srv.CreateModel(ctxh, payload)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	jm.AssertExpectations(t)
}

func getServiceWithMockedLayers() (testingcommons.MockIdentityService, documents.Service) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(did.ToAddress().Bytes(), nil)
	idService := testingcommons.MockIdentityService{}
	idService.On("IsSignedWithPurpose", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Once()
	queueSrv := new(testingutils.MockQueue)
	queueSrv.On("EnqueueJob", mock.Anything, mock.Anything).Return(&gocelery.AsyncResult{}, nil)

	repo := testRepo()
	anchorRepo := &testinganchors.MockAnchorRepo{}
	anchorRepo.On("GetAnchorData", mock.Anything).Return(nil, errors.New("missing"))
	docSrv := documents.DefaultService(cfg, repo, anchorRepo, documents.NewServiceRegistry(), &idService, nil, nil)
	return idService, DefaultService(
		docSrv,
		repo,
		queueSrv,
		ctx[jobs.BootstrappedService].(jobs.Manager), anchorRepo)
}

func TestService_UpdateModel(t *testing.T) {
	payload := documents.UpdatePayload{}
	srv := service{}

	// empty context
	_, _, err := srv.UpdateModel(context.Background(), payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentConfigAccountID, err))

	// missing id
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, srvr := getServiceWithMockedLayers()
	srv = srvr.(service)
	payload.DocumentID = utils.RandomSlice(32)
	_, _, err = srv.UpdateModel(ctxh, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// payload invalid
	g, _ := createCDWithEmbeddedGeneric(t)
	err = testRepo().Create(did[:], g.ID(), g)
	assert.NoError(t, err)
	payload.DocumentID = g.ID()

	// failed validations
	payload.Data = validData(t)
	dr := anchors.RandomDocumentRoot()
	anchorRepo := new(testinganchors.MockAnchorRepo)
	anchorRepo.On("GetAnchorData", mock.Anything).Return(dr, nil)
	oldRepo := srv.anchorRepo
	srv.anchorRepo = anchorRepo
	_, _, err = srv.UpdateModel(ctxh, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), documents.ErrDocumentIDReused.Error())
	anchorRepo.AssertExpectations(t)

	// Success
	jm := testingjobs.MockJobManager{}
	jm.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), make(chan error), nil)
	srv.jobManager = jm
	srv.anchorRepo = oldRepo
	m, _, err := srv.UpdateModel(ctxh, payload)
	assert.NoError(t, err)
	assert.Equal(t, m.ID(), g.ID())
	assert.Equal(t, m.CurrentVersion(), g.NextVersion())
	jm.AssertExpectations(t)
}

func TestService_Update(t *testing.T) {
	_, srv := getServiceWithMockedLayers()
	gsrv := srv.(service)
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// empty context
	_, _, _, err := srv.Update(context.Background(), nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentConfigAccountID, err))

	// missing last version
	g, _ := createCDWithEmbeddedGeneric(t)
	_, _, _, err = gsrv.Update(ctxh, g)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))
	assert.NoError(t, testRepo().Create(did[:], g.CurrentVersion(), g))

	// validations failed
	err = g.(*Generic).PrepareNewVersion(g, documents.CollaboratorsAccess{}, nil)
	dr := anchors.RandomDocumentRoot()
	anchorRepo := new(testinganchors.MockAnchorRepo)
	anchorRepo.On("GetAnchorData", mock.Anything).Return(dr, nil)
	oldRepo := gsrv.anchorRepo
	gsrv.anchorRepo = anchorRepo
	_, _, _, err = gsrv.Update(ctxh, g)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), documents.ErrDocumentIDReused.Error())
	anchorRepo.AssertExpectations(t)
	gsrv.anchorRepo = oldRepo

	// success
	jm := testingjobs.MockJobManager{}
	jm.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), make(chan error), nil)
	gsrv.jobManager = jm
	m, _, _, err := gsrv.Update(ctxh, g)
	assert.NoError(t, err)
	g, err = testRepo().Get(did[:], m.PreviousVersion())
	assert.NoError(t, err)
	assert.Equal(t, m.ID(), g.ID())
	assert.Equal(t, m.CurrentVersion(), g.NextVersion())
	jm.AssertExpectations(t)
}
