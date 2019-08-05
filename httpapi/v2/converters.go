package v2

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/jobs"
)

func toDocumentsPayload(req CreateDocumentRequest) (payload documents.UpdatePayload, err error) {
	cp, err := coreapi.ToDocumentsCreatePayload(req.DocumentRequest)
	if err != nil {
		return payload, err
	}

	return documents.UpdatePayload{CreatePayload: cp, DocumentID: req.DocumentID.Bytes()}, nil
}

func toDocumentResponse(doc documents.Model, tokenRegistry documents.TokenRegistry, jobID jobs.JobID) (coreapi.DocumentResponse, error) {
	resp, err := coreapi.GetDocumentResponse(doc, tokenRegistry, jobID)
	if err != nil {
		return resp, err
	}

	resp.Header.Status = string(doc.GetStatus())
	return resp, err
}

// unmarshalBody unmarshals req.Body to val.
// val should always be a pointer to the struct.
func unmarshalBody(r *http.Request, val interface{}) error {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, val)
}
