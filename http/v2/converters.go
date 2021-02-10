package v2

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
)

func toDocumentsPayload(req coreapi.CreateDocumentRequest, docID []byte) (payload documents.UpdatePayload, err error) {
	cp, err := coreapi.ToDocumentsCreatePayload(req)
	if err != nil {
		return payload, err
	}

	return documents.UpdatePayload{CreatePayload: cp, DocumentID: docID}, nil
}

func toDocumentResponse(doc documents.Document, tokenRegistry documents.TokenRegistry, jobID jobs.JobID) (coreapi.DocumentResponse, error) {
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
	return Role{
		ID:            r.RoleKey,
		Collaborators: byteutils.ToHexByteSlice(r.Collaborators),
	}
}

func toClientRule(r *coredocumentpb.TransitionRule) TransitionRule {
	return TransitionRule{
		RuleID:               r.RuleKey,
		Action:               coredocumentpb.TransitionAction_name[int32(r.Action)],
		Roles:                byteutils.ToHexByteSlice(r.Roles),
		Field:                r.Field,
		AttributeLabels:      byteutils.ToHexByteSlice(r.ComputeFields),
		Wasm:                 r.ComputeCode,
		TargetAttributeLabel: string(r.ComputeTargetField),
	}
}

func toClientRules(rules []*coredocumentpb.TransitionRule) (tr TransitionRules) {
	for _, r := range rules {
		tr.Rules = append(tr.Rules, toClientRule(r))
	}

	return tr
}

func toDocumentAttributes(attrs coreapi.AttributeMapRequest) ([]documents.Attribute, error) {
	cattrs, err := coreapi.ToDocumentAttributes(attrs)
	if err != nil {
		return nil, err
	}

	var rattrs []documents.Attribute
	for _, v := range cattrs {
		rattrs = append(rattrs, v)
	}

	return rattrs, nil
}
