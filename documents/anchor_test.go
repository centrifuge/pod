// +build unit

package documents_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
)

func TestAnchorDocument(t *testing.T) {
	ctx := context.Background()
	updater := func(id []byte, model documents.Model) error {
		return nil
	}

	// pack fails
	m := &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(nil, fmt.Errorf("pack failed")).Once()
	model, err := documents.AnchorDocument(ctx, m, nil, updater)
	m.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pack failed")

	// prepare fails
	m = &testingdocuments.MockModel{}
	cd := coredocument.New()
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc := &testingutils.MockCoreDocumentProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(fmt.Errorf("error")).Once()
	model, err = documents.AnchorDocument(ctx, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare document for signatures")

	// request signatures failed
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &testingutils.MockCoreDocumentProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctx, m).Return(fmt.Errorf("error")).Once()
	model, err = documents.AnchorDocument(ctx, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to collect signatures")

	// prepare for anchoring fails
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &testingutils.MockCoreDocumentProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctx, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(fmt.Errorf("error")).Once()
	model, err = documents.AnchorDocument(ctx, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare for anchoring")

	// anchor fails
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &testingutils.MockCoreDocumentProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctx, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(nil).Once()
	proc.On("AnchorDocument", m).Return(fmt.Errorf("error")).Once()
	model, err = documents.AnchorDocument(ctx, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to anchor document")

	// send failed
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &testingutils.MockCoreDocumentProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctx, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(nil).Once()
	proc.On("AnchorDocument", m).Return(nil).Once()
	proc.On("SendDocument", ctx, m).Return(fmt.Errorf("error")).Once()
	model, err = documents.AnchorDocument(ctx, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send anchored document")

	// success
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &testingutils.MockCoreDocumentProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctx, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(nil).Once()
	proc.On("AnchorDocument", m).Return(nil).Once()
	proc.On("SendDocument", ctx, m).Return(nil).Once()
	model, err = documents.AnchorDocument(ctx, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, model)
}
