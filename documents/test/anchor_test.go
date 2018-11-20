// +build unit

package documents_test

import (
	"context"
	"fmt"
	"testing"

	"os"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/header"
	"github.com/centrifuge/go-centrifuge/testingutils/coredocument"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
)

var ctx = map[string]interface{}{}
var cfg *config.Configuration

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&config.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[config.BootstrappedConfig].(*config.Configuration)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestAnchorDocument(t *testing.T) {
	ctx := context.Background()
	ctxh, err := header.NewContextHeader(ctx, cfg)
	assert.Nil(t, err)
	updater := func(id []byte, model documents.Model) error {
		return nil
	}

	// pack fails
	m := &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(nil, fmt.Errorf("pack failed")).Once()
	model, err := documents.AnchorDocument(ctxh, m, nil, updater)
	m.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pack failed")

	// prepare fails
	m = &testingdocuments.MockModel{}
	cd := coredocument.New()
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc := &testingcoredocument.MockCoreDocumentProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(fmt.Errorf("error")).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare document for signatures")

	// request signatures failed
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &testingcoredocument.MockCoreDocumentProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctxh, m).Return(fmt.Errorf("error")).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to collect signatures")

	// prepare for anchoring fails
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &testingcoredocument.MockCoreDocumentProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctxh, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(fmt.Errorf("error")).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare for anchoring")

	// anchor fails
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &testingcoredocument.MockCoreDocumentProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctxh, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(nil).Once()
	proc.On("AnchorDocument", m).Return(fmt.Errorf("error")).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to anchor document")

	// send failed
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &testingcoredocument.MockCoreDocumentProcessor{}
	proc.On("PrepareForSignatureRequests", m).Return(nil).Once()
	proc.On("RequestSignatures", ctxh, m).Return(nil).Once()
	proc.On("PrepareForAnchoring", m).Return(nil).Once()
	proc.On("AnchorDocument", m).Return(nil).Once()
	proc.On("SendDocument", ctxh, m).Return(fmt.Errorf("error")).Once()
	model, err = documents.AnchorDocument(ctxh, m, proc, updater)
	m.AssertExpectations(t)
	proc.AssertExpectations(t)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send anchored document")

	// success
	m = &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(cd, nil).Once()
	proc = &testingcoredocument.MockCoreDocumentProcessor{}
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
