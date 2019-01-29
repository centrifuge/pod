// +build unit

package documents_test

import (
	"context"
	"errors"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAnchorProcessor struct {
	mock.Mock
}

func (m *mockAnchorProcessor) Send(ctx context.Context, coreDocument *coredocumentpb.CoreDocument, recipient identity.CentID) (err error) {
	args := m.Called(coreDocument, ctx, recipient)
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

func (m *mockAnchorProcessor) GetDataProofHashes(coreDocument *coredocumentpb.CoreDocument) (hashes [][]byte, err error) {
	args := m.Called(coreDocument)
	return args.Get(0).([][]byte), args.Error(1)
}

func TestAnchorDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	updater := func(id []byte, model documents.Model) error {
		return nil
	}

	// pack fails
	m := &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(nil, errors.New("pack failed")).Once()
	model, err := documents.AnchorDocument(ctxh, m, nil, updater)
	m.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pack failed")

	// prepare fails
	m = &testingdocuments.MockModel{}
	cd := coredocument.New()
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc := &mockAnchorProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(errors.New("error")).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare document for signatures")

	// request signatures failed
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &mockAnchorProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctxh, m).Return(errors.New("error")).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to collect signatures")

	// prepare for anchoring fails
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &mockAnchorProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctxh, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(errors.New("error")).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare for anchoring")

	// anchor fails
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &mockAnchorProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctxh, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(nil).Once()
	proc.On("AnchorDocument", m).Return(errors.New("error")).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to anchor document")

	// send failed
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &mockAnchorProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctxh, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(nil).Once()
	proc.On("AnchorDocument", m).Return(nil).Once()
	proc.On("SendDocument", ctxh, m).Return(errors.New("error")).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send anchored document")

	// success
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &mockAnchorProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctxh, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(nil).Once()
	proc.On("AnchorDocument", m).Return(nil).Once()
	proc.On("SendDocument", ctxh, m).Return(nil).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, model)
}
