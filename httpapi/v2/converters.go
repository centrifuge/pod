package v2

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
)

func toDocumentsPayload(req CreateDocumentRequest) (payload documents.UpdatePayload, err error) {
	cp, err := coreapi.ToDocumentsCreatePayload(req.DocumentRequest)
	if err != nil {
		return payload, err
	}

	return documents.UpdatePayload{CreatePayload: cp, DocumentID: req.DocumentID.Bytes()}, nil
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
