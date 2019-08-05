// +build unit

package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/pending"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func invoiceData() map[string]interface{} {
	return map[string]interface{}{
		"number":       "12345",
		"status":       "unpaid",
		"gross_amount": "12.345",
		"recipient":    "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
		"date_due":     "2019-05-24T14:48:44.308854Z", // rfc3339nano
		"date_paid":    "2019-05-24T14:48:44Z",        // rfc3339
		"currency":     "EUR",
		"attachments": []map[string]interface{}{
			{
				"name":      "test",
				"file_type": "pdf",
				"size":      1000202,
				"data":      "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
				"checksum":  "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF3",
			},
		},
	}
}

func invalidDocIDPayload(t *testing.T) io.Reader {
	p := map[string]interface{}{
		"scheme":      "invoice",
		"data":        invoiceData(),
		"document_id": "invalid",
	}

	d, err := json.Marshal(p)
	assert.NoError(t, err)
	return bytes.NewReader(d)
}

func validPayload(t *testing.T) io.Reader {
	p := map[string]interface{}{
		"scheme":      "invoice",
		"data":        invoiceData(),
		"document_id": hexutil.Encode(utils.RandomSlice(32)),
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "string",
				"value": "hello, world",
			},
		},
	}

	d, err := json.Marshal(p)
	assert.NoError(t, err)
	return bytes.NewReader(d)
}

func invalidAttrPayload(t *testing.T) io.Reader {
	p := map[string]interface{}{
		"scheme": "invoice",
		"data":   invoiceData(),
		//"document_id": hexutil.Encode(utils.RandomSlice(32)),
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "invalid",
				"value": "hello, world",
			},
		},
	}

	d, err := json.Marshal(p)
	assert.NoError(t, err)
	return bytes.NewReader(d)
}

func TestHandler_CreateDocument(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/documents", b).WithContext(ctx)
	}

	// failed unmarshal empty body
	ctx := context.Background()
	w, r := getHTTPReqAndResp(ctx, nil)
	pendingSrv := new(pending.MockService)
	h := handler{srv: Service{pendingDocSrv: pendingSrv}}
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed unmarshal invalid doc_id
	w, r = getHTTPReqAndResp(ctx, invalidDocIDPayload(t))
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "hex string without 0x prefix")

	// failed payloadConversion
	w, r = getHTTPReqAndResp(ctx, invalidAttrPayload(t))
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "not a valid attribute type")

	// failed to create document
	pendingSrv.On("Create", ctx, mock.Anything).Return(nil, errors.New("Failed to create document")).Once()
	w, r = getHTTPReqAndResp(ctx, validPayload(t))
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "Failed to create document")

	// failed document conversion
	doc := new(testingdocuments.MockModel)
	doc.On("GetData").Return(invoice.Data{}).Twice()
	doc.On("Scheme").Return("invoice").Twice()
	doc.On("GetAttributes").Return(nil).Twice()
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators")).Once()
	pendingSrv.On("Create", ctx, mock.Anything).Return(doc, nil)
	w, r = getHTTPReqAndResp(ctx, validPayload(t))
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")

	// success
	doc.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	doc.On("ID").Return(utils.RandomSlice(32)).Once()
	doc.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	doc.On("Author").Return(nil, errors.New("somerror")).Once()
	doc.On("Timestamp").Return(nil, errors.New("somerror")).Once()
	doc.On("NFTs").Return(nil).Once()
	doc.On("GetStatus").Return(documents.Pending).Once()
	w, r = getHTTPReqAndResp(ctx, validPayload(t))
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusCreated)
	assert.Contains(t, w.Body.String(), "\"status\":\"pending\"")
	pendingSrv.AssertExpectations(t)
	doc.AssertExpectations(t)
}
