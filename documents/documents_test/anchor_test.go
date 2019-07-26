// +build unit

package documents_test

import (
	"context"
	"errors"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAnchorProcessor struct {
	mock.Mock
}

func (m *mockAnchorProcessor) PreAnchorDocument(ctx context.Context, model documents.Model) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}

func (m *mockAnchorProcessor) Send(ctx context.Context, cd coredocumentpb.CoreDocument, recipient identity.DID) (err error) {
	args := m.Called(ctx, cd, recipient)
	return args.Error(0)
}

func (m *mockAnchorProcessor) Anchor(
	ctx context.Context,
	coreDocument *coredocumentpb.CoreDocument,
	saveState func(*coredocumentpb.CoreDocument) error) (err error) {
	args := m.Called(ctx, coreDocument, saveState)
	if saveState != nil {
		err := saveState(coreDocument)
		if err != nil {
			return err
		}
	}
	return args.Error(0)
}

func (m *mockAnchorProcessor) PrepareForSignatureRequests(ctx context.Context, model documents.Model) error {
	args := m.Called(model)
	return args.Error(0)
}

func (m *mockAnchorProcessor) RequestSignatures(ctx context.Context, model documents.Model) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}

func (m *mockAnchorProcessor) PrepareForAnchoring(model documents.Model) error {
	args := m.Called(model)
	return args.Error(0)
}

func (m *mockAnchorProcessor) AnchorDocument(ctx context.Context, model documents.Model) error {
	args := m.Called(model)
	return args.Error(0)
}

func (m *mockAnchorProcessor) SendDocument(ctx context.Context, model documents.Model) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}

func (m *mockAnchorProcessor) RequestDocumentWithAccessToken(ctx context.Context, tokenIdentifier, entityIdentifier, entityRelationIdentifier []byte) (*p2ppb.GetDocumentResponse, error) {
	args := m.Called(ctx, tokenIdentifier, entityIdentifier, entityRelationIdentifier)
	return args.Get(0).(*p2ppb.GetDocumentResponse), args.Error(0)
}

func TestAnchorDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	updater := func(id []byte, model documents.Model) error {
		return nil
	}

	// prepare fails
	id := utils.RandomSlice(32)
	m := &documents.MockModel{}
	m.On("CurrentVersion").Return(id).Once()
	proc := &mockAnchorProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(errors.New("error")).Once()
	model, err := documents.AnchorDocument(ctxh, m, proc, updater, false)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare document for signatures")

	// request signatures failed
	m = &documents.MockModel{}
	m.On("CurrentVersion").Return(id).Once()
	proc = &mockAnchorProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctxh, m).Return(errors.New("error")).Once()
	proc.On("PreAnchorDocument", ctxh, m).Return(nil).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater, true)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to collect signatures")

	// prepare for anchoring fails
	m = &documents.MockModel{}
	m.On("CurrentVersion").Return(id).Once()
	proc = &mockAnchorProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctxh, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(errors.New("error")).Once()
	proc.On("PreAnchorDocument", ctxh, m).Return(nil).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater, true)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare for anchoring")

	// anchor fails
	m = &documents.MockModel{}
	m.On("CurrentVersion").Return(id).Once()
	proc = &mockAnchorProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctxh, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(nil).Once()
	proc.On("PreAnchorDocument", ctxh, m).Return(nil).Once()
	proc.On("AnchorDocument", m).Return(errors.New("error")).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater, true)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to anchor document")

	// send failed
	m = &documents.MockModel{}
	m.On("CurrentVersion").Return(id).Once()
	m.On("SetStatus", documents.Committed).Return(nil)
	proc = &mockAnchorProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctxh, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(nil).Once()
	proc.On("PreAnchorDocument", ctxh, m).Return(nil).Once()
	proc.On("AnchorDocument", m).Return(nil).Once()
	proc.On("SendDocument", ctxh, m).Return(errors.New("error")).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater, true)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send anchored document")

	// success
	m = &documents.MockModel{}
	m.On("CurrentVersion").Return(id).Once()
	m.On("SetStatus", documents.Committed).Return(nil)
	proc = &mockAnchorProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctxh, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(nil).Once()
	proc.On("AnchorDocument", m).Return(nil).Once()
	proc.On("SendDocument", ctxh, m).Return(nil).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater, false)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, model)
}
