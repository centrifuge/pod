// +build unit

package coreapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_CreateDocument(t *testing.T) {
	data := map[string]interface{}{
		"scheme": "invoice",
		"data":   invoiceData(),
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "invalid",
				"value": "hello, world",
			},
		},
	}

	d, err := json.Marshal(data)
	assert.NoError(t, err)
	r := httptest.NewRequest("POST", "/documents", bytes.NewReader(d))
	w := httptest.NewRecorder()

	h := handler{}
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "not a valid attribute")

	data = map[string]interface{}{
		"scheme": "invoice",
		"data":   invoiceData(),
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "string",
				"value": "hello, world",
			},
		},
	}
	d, err = json.Marshal(data)
	assert.NoError(t, err)
	docSrv := new(testingdocuments.MockService)
	srv := Service{docService: docSrv}
	h = handler{srv: srv}
	docSrv.On("CreateModel", mock.Anything, mock.Anything).Return(nil, jobs.NilJobID(), errors.New("failed to create model"))
	r = httptest.NewRequest("POST", "/documents", bytes.NewReader(d))
	w = httptest.NewRecorder()
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to create model")
	docSrv.AssertExpectations(t)

	m := new(testingdocuments.MockModel)
	m.On("GetData").Return(data)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, errors.New("failed to get collaborators"))
	docSrv = new(testingdocuments.MockService)
	srv = Service{docService: docSrv}
	h = handler{srv: srv}
	docSrv.On("CreateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	r = httptest.NewRequest("POST", "/documents", bytes.NewReader(d))
	w = httptest.NewRecorder()
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get collaborators")
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)

	m = new(testingdocuments.MockModel)
	m.On("GetCollaborators", mock.Anything).Return(documents.CollaboratorsAccess{}, nil).Once()
	m.On("GetData").Return(data)
	m.On("ID").Return(utils.RandomSlice(32)).Once()
	m.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	m.On("Author").Return(nil, errors.New("somerror"))
	m.On("Timestamp").Return(nil, errors.New("somerror"))
	m.On("NFTs").Return(nil)
	docSrv = new(testingdocuments.MockService)
	srv = Service{docService: docSrv}
	h = handler{srv: srv}
	docSrv.On("CreateModel", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	r = httptest.NewRequest("POST", "/documents", bytes.NewReader(d))
	w = httptest.NewRecorder()
	h.CreateDocument(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}

func TestRegister(t *testing.T) {
	r := chi.NewRouter()
	Register(r, nil, nil)
	assert.Len(t, r.Routes(), 1)
	assert.Equal(t, r.Routes()[0].Pattern, "/documents")
}
