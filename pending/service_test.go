//go:build unit
// +build unit

package pending

import (
	"context"
	"io/ioutil"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"

	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
	Repository
}

func (m *mockRepo) Get(accID, id []byte) (documents.Document, error) {
	args := m.Called(accID, id)
	doc, _ := args.Get(0).(documents.Document)
	return doc, args.Error(1)
}

func (m *mockRepo) Delete(accID, id []byte) error {
	args := m.Called(accID, id)
	return args.Error(0)
}

func (m *mockRepo) Create(accID, id []byte, doc documents.Document) error {
	args := m.Called(accID, id, doc)
	return args.Error(0)
}

func (m *mockRepo) Update(accID, id []byte, doc documents.Document) error {
	args := m.Called(accID, id, doc)
	return args.Error(0)
}

func TestService_Commit(t *testing.T) {
	s := service{}

	// missing did
	ctx := context.Background()
	docID := utils.RandomSlice(32)
	_, _, err := s.Commit(ctx, docID)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// missing model
	ctx = testingconfig.CreateAccountContext(t, cfg)
	repo := new(mockRepo)
	repo.On("Get", did[:], docID).Return(nil, errors.New("not found")).Once()
	s.pendingRepo = repo
	_, _, err = s.Commit(ctx, docID)
	assert.Error(t, err)

	// failed commit
	doc := new(documents.MockModel)
	repo.On("Get", did[:], docID).Return(doc, nil)
	docSrv := new(testingdocuments.MockService)
	docSrv.On("Commit", ctx, doc).Return(nil, errors.New("failed to commit")).Once()
	s.docSrv = docSrv
	_, _, err = s.Commit(ctx, docID)
	assert.Error(t, err)

	// success
	jobID := gocelery.JobID(utils.RandomSlice(32))
	repo.On("Delete", did[:], docID).Return(nil)
	docSrv.On("Commit", ctx, doc).Return(jobID, nil)
	m, jid, err := s.Commit(ctx, docID)
	assert.NoError(t, err)
	assert.Equal(t, jobID, jid)
	assert.NotNil(t, m)
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
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

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

func TestService_Get(t *testing.T) {
	// not pending document
	st := documents.Committed
	s := service{}
	ctx := context.Background()
	docID := utils.RandomSlice(32)
	docSrv := new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", docID).Return(new(documents.MockModel), nil).Once()
	s.docSrv = docSrv
	doc, err := s.Get(ctx, docID, st)
	assert.NoError(t, err)
	assert.NotNil(t, doc)

	// pending doc
	// missing did from context
	st = documents.Pending
	_, err = s.Get(ctx, docID, st)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// success
	repo := new(mockRepo)
	repo.On("Get", did[:], docID).Return(doc, nil).Once()
	s.pendingRepo = repo
	ctx = testingconfig.CreateAccountContext(t, cfg)
	gdoc, err := s.Get(ctx, docID, st)
	assert.NoError(t, err)
	assert.Equal(t, doc, gdoc)
	docSrv.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestService_GetVersion(t *testing.T) {
	// found in docService
	doc := new(documents.MockModel)
	docID, versionID := utils.RandomSlice(32), utils.RandomSlice(32)
	s := service{}
	ctx := context.Background()
	docSrv := new(testingdocuments.MockService)
	docSrv.On("GetVersion", docID, versionID).Return(doc, nil).Once()
	s.docSrv = docSrv
	gdoc, err := s.GetVersion(ctx, docID, versionID)
	assert.NoError(t, err)
	assert.Equal(t, doc, gdoc)

	// not found in docService
	// ctx with no did
	docSrv.On("GetVersion", docID, versionID).Return(nil, documents.ErrDocumentNotFound)
	_, err = s.GetVersion(ctx, docID, versionID)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// different current version
	ctx = testingconfig.CreateAccountContext(t, cfg)
	repo := new(mockRepo)
	repo.On("Get", did[:], docID).Return(doc, nil)
	doc.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	s.pendingRepo = repo
	_, err = s.GetVersion(ctx, docID, versionID)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// successful retrieval
	doc.On("CurrentVersion").Return(versionID).Once()
	gdoc, err = s.GetVersion(ctx, docID, versionID)
	assert.NoError(t, err)
	assert.Equal(t, doc, gdoc)
	docSrv.AssertExpectations(t)
	repo.AssertExpectations(t)
	doc.AssertExpectations(t)
}

func TestService_Update(t *testing.T) {
	s := service{}

	// missing did
	ctx := context.Background()
	payload := documents.UpdatePayload{}
	_, err := s.Update(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// document doesnt exist yet
	repo := new(mockRepo)
	repo.On("Get", did[:], payload.DocumentID).Return(nil, errors.New("not found")).Once()
	s.pendingRepo = repo
	ctx = testingconfig.CreateAccountContext(t, cfg)
	_, err = s.Update(ctx, payload)
	assert.Error(t, err)

	// Patch error
	oldModel := new(documents.MockModel)
	oldModel.On("Patch", payload).Return(errors.New("error patching")).Once()
	repo.On("Get", did[:], payload.DocumentID).Return(oldModel, nil)
	_, err = s.Update(ctx, payload)
	assert.Error(t, err)

	// Success
	oldModel.On("ID").Return(payload.DocumentID).Once()
	oldModel.On("Patch", payload).Return(nil).Once()
	repo.On("Update", did[:], payload.DocumentID, oldModel).Return(nil).Once()
	_, err = s.Update(ctx, payload)
	assert.NoError(t, err)
}

func TestService_AddSignedAttribute(t *testing.T) {
	s := service{}
	label := "signed_attribute"
	value := utils.RandomSlice(32)
	docID := utils.RandomSlice(32)
	versionID := utils.RandomSlice(32)

	// missing did
	ctx := context.Background()
	_, err := s.AddSignedAttribute(ctx, docID, label, value, documents.AttrBytes)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// missing document
	ctx = testingconfig.CreateAccountContext(t, cfg)
	prepo := new(mockRepo)
	prepo.On("Get", did[:], docID).Return(nil, errors.New("Missing")).Once()
	s.pendingRepo = prepo
	_, err = s.AddSignedAttribute(ctx, docID, label, value, documents.AttrBytes)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// failed to get new attribute
	doc := new(documents.MockModel)
	doc.On("ID").Return(docID)
	doc.On("CurrentVersion").Return(versionID)
	prepo.On("Get", did[:], docID).Return(doc, nil)
	_, err = s.AddSignedAttribute(ctx, docID, "", value, documents.AttrBytes)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrEmptyAttrLabel, err))

	// failed to add attribute to document
	doc.On("AddAttributes", mock.Anything, false, mock.Anything).Return(errors.New("failed to add")).Once()
	_, err = s.AddSignedAttribute(ctx, docID, label, value, documents.AttrBytes)
	assert.Error(t, err)

	// success
	doc.On("AddAttributes", mock.Anything, false, mock.Anything).Return(nil).Once()
	prepo.On("Update", did[:], docID, doc).Return(nil).Once()
	doc1, err := s.AddSignedAttribute(ctx, docID, label, value, documents.AttrBytes)
	assert.NoError(t, err)
	assert.Equal(t, doc, doc1)
	prepo.AssertExpectations(t)
	doc.AssertExpectations(t)
}

func TestService_RemoveCollaborators(t *testing.T) {
	s := service{}

	// missing did from context
	ctx := context.Background()
	docID := utils.RandomSlice(32)
	_, err := s.RemoveCollaborators(ctx, docID, nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// missing doc
	ctx = testingconfig.CreateAccountContext(t, cfg)
	repo := new(mockRepo)
	repo.On("Get", did[:], docID).Return(nil, errors.New("failed")).Once()
	s.pendingRepo = repo
	_, err = s.RemoveCollaborators(ctx, docID, nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// failed to remove collaborators
	d := new(documents.MockModel)
	d.On("RemoveCollaborators", mock.Anything).Return(errors.New("failed")).Once()
	repo.On("Get", did[:], docID).Return(d, nil)
	_, err = s.RemoveCollaborators(ctx, docID, []identity.DID{testingidentity.GenerateRandomDID()})
	assert.Error(t, err)

	// success
	d.On("RemoveCollaborators", mock.Anything).Return(nil)
	repo.On("Update", did[:], docID, d).Return(nil).Once()
	d1, err := s.RemoveCollaborators(ctx, docID, []identity.DID{testingidentity.GenerateRandomDID()})
	assert.NoError(t, err)
	assert.Equal(t, d, d1)
	repo.AssertExpectations(t)
	d.AssertExpectations(t)
}

func TestService_AddRole(t *testing.T) {
	s := service{}
	key := "roleKey"
	collabs := []identity.DID{testingidentity.GenerateRandomDID()}

	// missing did from context
	ctx := context.Background()
	docID := utils.RandomSlice(32)
	_, err := s.AddRole(ctx, docID, key, collabs)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// missing doc
	ctx = testingconfig.CreateAccountContext(t, cfg)
	repo := new(mockRepo)
	repo.On("Get", did[:], docID).Return(nil, errors.New("failed")).Once()
	s.pendingRepo = repo
	_, err = s.AddRole(ctx, docID, key, collabs)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// failed to add role
	d := new(documents.MockModel)
	repo.On("Get", did[:], docID).Return(d, nil).Twice()
	d.On("AddRole", key, collabs).Return(nil, errors.New("failed to add role")).Once()
	_, err = s.AddRole(ctx, docID, key, collabs)
	assert.Error(t, err)

	// success
	d.On("AddRole", key, collabs).Return(new(coredocumentpb.Role), nil).Once()
	repo.On("Update", did[:], docID, d).Return(nil).Once()
	_, err = s.AddRole(ctx, docID, key, collabs)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	d.AssertExpectations(t)
}

func TestService_GetRole(t *testing.T) {
	s := service{}
	key := utils.RandomSlice(32)

	// missing did from context
	ctx := context.Background()
	docID := utils.RandomSlice(32)
	_, err := s.GetRole(ctx, docID, key)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// missing doc from both the states
	ctx = testingconfig.CreateAccountContext(t, cfg)
	repo := new(mockRepo)
	repo.On("Get", did[:], docID).Return(nil, errors.New("failed")).Twice()
	docSrv := new(documents.MockService)
	docSrv.On("GetCurrentVersion", ctx, docID).Return(nil, errors.New("failed")).Once()
	s.pendingRepo = repo
	s.docSrv = docSrv
	_, err = s.GetRole(ctx, docID, key)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// missing document from the pending version
	d := new(documents.MockModel)
	docSrv.On("GetCurrentVersion", ctx, docID).Return(d, nil).Once()
	d.On("GetRole", key).Return(new(coredocumentpb.Role), nil).Twice()
	_, err = s.GetRole(ctx, docID, key)
	assert.NoError(t, err)

	// document from the pending version
	repo.On("Get", did[:], docID).Return(d, nil).Once()
	_, err = s.GetRole(ctx, docID, key)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	docSrv.AssertExpectations(t)
	d.AssertExpectations(t)
}

func TestService_UpdateRole(t *testing.T) {
	s := service{}
	key := utils.RandomSlice(32)
	collabs := []identity.DID{testingidentity.GenerateRandomDID()}

	// missing did from context
	ctx := context.Background()
	docID := utils.RandomSlice(32)
	_, err := s.UpdateRole(ctx, docID, key, collabs)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// missing doc
	ctx = testingconfig.CreateAccountContext(t, cfg)
	repo := new(mockRepo)
	repo.On("Get", did[:], docID).Return(nil, errors.New("failed")).Once()
	s.pendingRepo = repo
	_, err = s.UpdateRole(ctx, docID, key, collabs)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// failed to patch role
	d := new(documents.MockModel)
	d.On("UpdateRole", key, collabs).Return(nil, errors.New("failed to update roles")).Once()
	repo.On("Get", did[:], docID).Return(d, nil).Twice()
	_, err = s.UpdateRole(ctx, docID, key, collabs)
	assert.Error(t, err)

	// success
	d.On("UpdateRole", key, collabs).Return(new(coredocumentpb.Role), nil).Once()
	repo.On("Update", did[:], docID, d).Return(nil).Once()
	_, err = s.UpdateRole(ctx, docID, key, collabs)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	d.AssertExpectations(t)
}

func TestService_AddTransitionRules(t *testing.T) {
	s := service{}
	ctx := context.Background()
	docID := utils.RandomSlice(32)
	addRules := AddTransitionRules{AttributeRules: []AttributeRule{
		{
			RoleID:   utils.RandomSlice(32),
			KeyLabel: "",
		},
	}}

	// missing did from context
	_, err := s.AddTransitionRules(ctx, docID, addRules)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// missing doc
	ctx = testingconfig.CreateAccountContext(t, cfg)
	repo := new(mockRepo)
	repo.On("Get", did[:], docID).Return(nil, errors.New("failed")).Once()
	s.pendingRepo = repo
	_, err = s.AddTransitionRules(ctx, docID, addRules)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// empty label
	d := new(documents.MockModel)
	repo.On("Get", did[:], docID).Return(d, nil).Times(3)
	_, err = s.AddTransitionRules(ctx, docID, addRules)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrEmptyAttrLabel, err))

	// failed to add repo
	addRules.AttributeRules[0].KeyLabel = "test"
	d.On("AddTransitionRuleForAttribute", addRules.AttributeRules[0].RoleID.Bytes(), mock.Anything).Return(
		nil, errors.New("failed to create rule")).Once()
	_, err = s.AddTransitionRules(ctx, docID, addRules)
	assert.Error(t, err)

	// success
	d.On("AddTransitionRuleForAttribute", addRules.AttributeRules[0].RoleID.Bytes(), mock.Anything).Return(
		new(coredocumentpb.TransitionRule), nil).Once()
	repo.On("Update", did[:], docID, d).Return(nil).Once()
	_, err = s.AddTransitionRules(ctx, docID, addRules)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	d.AssertExpectations(t)
}

func TestService_GetTransitionRule(t *testing.T) {
	s := service{}
	ctx := context.Background()
	docID := utils.RandomSlice(32)
	ruleID := utils.RandomSlice(32)

	// missing did from context
	_, err := s.GetTransitionRule(ctx, docID, ruleID)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// missing doc from both the states
	ctx = testingconfig.CreateAccountContext(t, cfg)
	repo := new(mockRepo)
	repo.On("Get", did[:], docID).Return(nil, errors.New("failed")).Twice()
	docSrv := new(documents.MockService)
	docSrv.On("GetCurrentVersion", ctx, docID).Return(nil, errors.New("failed")).Once()
	s.pendingRepo = repo
	s.docSrv = docSrv
	_, err = s.GetTransitionRule(ctx, docID, ruleID)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// missing document from the pending version
	d := new(documents.MockModel)
	docSrv.On("GetCurrentVersion", ctx, docID).Return(d, nil).Once()
	d.On("GetTransitionRule", ruleID).Return(new(coredocumentpb.TransitionRule), nil).Twice()
	_, err = s.GetTransitionRule(ctx, docID, ruleID)
	assert.NoError(t, err)

	// document from the pending version
	repo.On("Get", did[:], docID).Return(d, nil).Once()
	_, err = s.GetTransitionRule(ctx, docID, ruleID)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	docSrv.AssertExpectations(t)
	d.AssertExpectations(t)
}

func TestService_DeleteTransitionRules(t *testing.T) {
	s := service{}
	ctx := context.Background()
	docID := utils.RandomSlice(32)
	ruleID := utils.RandomSlice(32)

	// missing did from context
	err := s.DeleteTransitionRule(ctx, docID, ruleID)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// missing doc
	ctx = testingconfig.CreateAccountContext(t, cfg)
	repo := new(mockRepo)
	repo.On("Get", did[:], docID).Return(nil, errors.New("failed")).Once()
	s.pendingRepo = repo
	err = s.DeleteTransitionRule(ctx, docID, ruleID)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// empty label
	d := new(documents.MockModel)
	repo.On("Get", did[:], docID).Return(d, nil).Times(2)
	d.On("DeleteTransitionRule", ruleID).Return(documents.ErrTransitionRuleMissing).Once()
	err = s.DeleteTransitionRule(ctx, docID, ruleID)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrTransitionRuleMissing, err))

	// success
	d.On("DeleteTransitionRule", ruleID).Return(nil).Once()
	repo.On("Update", did[:], docID, d).Return(nil).Once()
	err = s.DeleteTransitionRule(ctx, docID, ruleID)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	d.AssertExpectations(t)
}

func Test_AddTransitionRule_ComputeFields(t *testing.T) {
	s := service{}
	ctx := context.Background()
	docID := utils.RandomSlice(32)
	addRules := AddTransitionRules{ComputeFieldsRules: []ComputeFieldsRule{
		{
			WASM:                 utils.RandomSlice(32),
			AttributeLabels:      []string{"test"},
			TargetAttributeLabel: "result",
		},
	}}

	// missing did from context
	_, err := s.AddTransitionRules(ctx, docID, addRules)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// missing doc
	ctx = testingconfig.CreateAccountContext(t, cfg)
	repo := new(mockRepo)
	repo.On("Get", did[:], docID).Return(nil, errors.New("failed")).Once()
	s.pendingRepo = repo
	_, err = s.AddTransitionRules(ctx, docID, addRules)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// failed to add compute rule
	d := new(documents.MockModel)
	repo.On("Get", did[:], docID).Return(d, nil).Times(2)
	attr := addRules.ComputeFieldsRules[0]
	d.On("AddComputeFieldsRule", attr.WASM.Bytes(), attr.AttributeLabels, attr.TargetAttributeLabel).Return(nil, documents.ErrComputeFieldsInvalidWASM).Once()
	_, err = s.AddTransitionRules(ctx, docID, addRules)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrComputeFieldsInvalidWASM, err))

	// success
	wasm, err := ioutil.ReadFile("../testingutils/compute_fields/simple_average.wasm")
	assert.NoError(t, err)
	attr.WASM = wasm
	addRules.ComputeFieldsRules[0] = attr
	d.On("AddComputeFieldsRule", wasm, attr.AttributeLabels, attr.TargetAttributeLabel).Return(new(coredocumentpb.TransitionRule), nil).Once()
	repo.On("Update", did[:], docID, d).Return(nil).Once()
	rule, err := s.AddTransitionRules(ctx, docID, addRules)
	assert.NoError(t, err)
	assert.NotNil(t, rule)
	repo.AssertExpectations(t)
	d.AssertExpectations(t)
}

func TestService_AddAttributes(t *testing.T) {
	s := service{}
	ctx := context.Background()
	docID := utils.RandomSlice(32)
	attr, err := documents.NewStringAttribute("test", documents.AttrInt256, "1000")
	assert.NoError(t, err)
	attrs := []documents.Attribute{attr}

	// missing did from context
	_, err = s.AddAttributes(ctx, docID, attrs)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// missing doc
	ctx = testingconfig.CreateAccountContext(t, cfg)
	repo := new(mockRepo)
	repo.On("Get", did[:], docID).Return(nil, errors.New("failed")).Once()
	s.pendingRepo = repo
	_, err = s.AddAttributes(ctx, docID, attrs)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// add attributes fails
	d := new(documents.MockModel)
	repo.On("Get", did[:], docID).Return(d, nil).Times(2)
	d.On("AddAttributes", documents.CollaboratorsAccess{}, false, attrs).Return(
		errors.New("failed to add attributes")).Once()
	_, err = s.AddAttributes(ctx, docID, attrs)
	assert.Error(t, err)

	// success
	repo.On("Update", did[:], docID, d).Return(nil).Once()
	d.On("AddAttributes", documents.CollaboratorsAccess{}, false, attrs).Return(nil).Once()
	_, err = s.AddAttributes(ctx, docID, attrs)
	assert.NoError(t, err)
}

func TestService_DeleteAttribute(t *testing.T) {
	s := service{}
	ctx := context.Background()
	docID := utils.RandomSlice(32)
	key, err := documents.AttrKeyFromLabel("test")
	assert.NoError(t, err)

	// missing did from context
	_, err = s.DeleteAttribute(ctx, docID, key)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrIdentityMissingFromContext, err))

	// missing doc
	ctx = testingconfig.CreateAccountContext(t, cfg)
	repo := new(mockRepo)
	repo.On("Get", did[:], docID).Return(nil, errors.New("failed")).Once()
	s.pendingRepo = repo
	_, err = s.DeleteAttribute(ctx, docID, key)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// add attributes fails
	d := new(documents.MockModel)
	repo.On("Get", did[:], docID).Return(d, nil).Times(2)
	d.On("DeleteAttribute", key, false).Return(
		errors.New("failed to delete attribute")).Once()
	_, err = s.DeleteAttribute(ctx, docID, key)
	assert.Error(t, err)

	// success
	repo.On("Update", did[:], docID, d).Return(nil).Once()
	d.On("DeleteAttribute", key, false).Return(nil).Once()
	_, err = s.DeleteAttribute(ctx, docID, key)
	assert.NoError(t, err)
}
