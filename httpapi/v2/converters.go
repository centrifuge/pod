package v2

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
)

func toDocumentsPayload(req DocumentRequest, docID []byte) (payload documents.UpdatePayload, err error) {
	cp, err := coreapi.ToDocumentsCreatePayload(req)
	if err != nil {
		return payload, err
	}

	return documents.UpdatePayload{CreatePayload: cp, DocumentID: docID}, nil
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

func toClientRole(r *coredocumentpb.Role) Role {
	var cs []byteutils.HexBytes
	for _, c := range r.Collaborators {
		c := c
		cs = append(cs, c)
	}

	return Role{
		ID:            r.RoleKey,
		Collaborators: cs,
	}
}
