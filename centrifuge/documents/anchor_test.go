// +build unit

package documents_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
)

func TestAnchorDocument(t *testing.T) {
	ctx := context.Background()

	// pack fails
	m := &testingdocuments.MockModel{}
	m.On("PackCoreDocument").Return(nil, fmt.Errorf("pack failed")).Once()
	model, err := documents.AnchorDocument(ctx, m, nil, nil)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pack failed")

	proc := coredocumentprocessor.DefaultProcessor(nil, nil, nil)

	// prepare fails
	m = &testingdocuments.MockModel{}
	cd := coredocument.New()
	m.On("PackCoreDocument").Return(cd, nil).Twice()
	model, err = documents.AnchorDocument(ctx, m, proc, nil)
	assert.Nil(t, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unpack failed")

	//poSrv := purchaseorder.DefaultService(nil, nil)
	//po, err := poSrv.DeriveFromCreatePayload(testingdocuments.CreatePOPayload())
	//assert.Nil(t, err)

}
