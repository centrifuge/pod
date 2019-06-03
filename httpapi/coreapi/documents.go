package coreapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("core_api")

type handler struct {
	srv           Service
	tokenRegistry documents.TokenRegistry
}

// CreateDocumentRequest defines the payload for creating documents.
type CreateDocumentRequest struct {
	Scheme      string               `json:"scheme" enums:"invoice"`
	ReadAccess  []identity.DID       `json:"read_access"`
	WriteAccess []identity.DID       `json:"write_access"`
	Data        json.RawMessage      `json:"data" enums:"invoice.Data"`
	Attributes  map[string]Attribute `json:"attributes"`
}

// Attribute defines a single attribute.
type Attribute struct {
	Type  string `json:"type" enums:"integer,decimal,string,bytes,timestamp"`
	Value string `json:"value"`
}

type NFT struct {
	Registry   string `json:"registry"`
	Owner      string `json:"owner"`
	TokenID    string `json:"token_id"`
	TokenIndex string `json:"token_index"`
}

type ResponseHeader struct {
	DocumentID  string         `json:"document_id"`
	Version     string         `json:"version"`
	Author      string         `json:"author"`
	CreatedAt   string         `json:"created_at"`
	ReadAccess  []identity.DID `json:"read_access" swaggertype:"string"`
	WriteAccess []identity.DID `json:"write_access" swaggertype:"string"`
	JobID       string         `json:"job_id"`
	NFTs        []NFT          `json:"nfts"`
}

type DocumentResponse struct {
	Header ResponseHeader `json:"header"`
	Data   interface{}    `json:"data" enums:"invoice.Data"`
}

type HTTPError struct {
	Message string `json:"message"`
}

func convertAttributes(cattrs map[string]Attribute) (map[documents.AttrKey]documents.Attribute, error) {
	attrs := make(map[documents.AttrKey]documents.Attribute)
	for k, v := range cattrs {
		attr, err := documents.NewAttribute(k, documents.AttributeType(v.Type), v.Value)
		if err != nil {
			return nil, err
		}

		attrs[attr.Key] = attr
	}

	return attrs, nil
}

func toDocumentsCreatePayload(request CreateDocumentRequest) (documents.CreatePayload, error) {
	payload := documents.CreatePayload{
		Scheme: request.Scheme,
		Collaborators: documents.CollaboratorsAccess{
			ReadCollaborators:      request.ReadAccess,
			ReadWriteCollaborators: request.WriteAccess,
		},
		Data: request.Data,
	}

	attrs, err := convertAttributes(request.Attributes)
	if err != nil {
		return payload, err
	}

	payload.Attributes = attrs
	return payload, nil
}

func convertNFTs(tokenRegistry documents.TokenRegistry, nfts []*coredocumentpb.NFT) (nnfts []NFT, err error) {
	for _, n := range nfts {
		regAddress := common.BytesToAddress(n.RegistryId[:common.AddressLength])
		i, errn := tokenRegistry.CurrentIndexOfToken(regAddress, n.TokenId)
		if errn != nil || i == nil {
			err = errors.AppendError(err, errors.New("token index received is nil or other error: %v", errn))
			continue
		}

		o, errn := tokenRegistry.OwnerOf(regAddress, n.TokenId)
		if errn != nil {
			err = errors.AppendError(err, errn)
			continue
		}

		nnfts = append(nnfts, NFT{
			Registry:   regAddress.Hex(),
			Owner:      o.Hex(),
			TokenID:    hexutil.Encode(n.TokenId),
			TokenIndex: hexutil.Encode(i.Bytes()),
		})
	}
	return nnfts, err
}

func deriveResponseHeader(tokenRegistry documents.TokenRegistry, model documents.Model, id jobs.JobID) (response ResponseHeader, err error) {
	cs, err := model.GetCollaborators()
	if err != nil {
		return response, err
	}

	// we ignore error here because it can happen when a model is first created but its not anchored yet
	author, _ := model.Author()

	// we ignore error here because it can happen when a model is first created but its not anchored yet
	var ts string
	t, err := model.Timestamp()
	if err == nil {
		ts = t.UTC().Format(time.RFC3339)
	}

	nfts := model.NFTs()
	cnfts, err := convertNFTs(tokenRegistry, nfts)
	if err != nil {
		// this could be a temporary failure, so we ignore but warn about the error
		log.Warningf("errors encountered when trying to set nfts to the response: %v", err)
	}

	return ResponseHeader{
		DocumentID:  hexutil.Encode(model.ID()),
		Version:     hexutil.Encode(model.CurrentVersion()),
		Author:      author.String(),
		CreatedAt:   ts,
		ReadAccess:  cs.ReadCollaborators,
		WriteAccess: cs.ReadWriteCollaborators,
		NFTs:        cnfts,
		JobID:       id.String(),
	}, nil
}

// CreateDocument creates a document.
// @summary Creates a new document and anchors it.
// @description Creates a new document and anchors it.
// @id create_document
// @tags Documents
// @accept json
// @produce json
// @success 200 {object} health.Pong
// @router /documents [post]
// TODO(ved) fine tune this
// 1. finish the create
// 2. Add AUth
// 3. tests
// 4. swagger
func (h handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	var err error
	var code int
	defer func() {
		if err == nil {
			return
		}

		render.Status(r, code)
		render.JSON(w, r, HTTPError{Message: err.Error()})
	}()

	ctx := r.Context()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		code = http.StatusInternalServerError
		return
	}

	var request CreateDocumentRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		code = http.StatusBadRequest
		return
	}

	payload, err := toDocumentsCreatePayload(request)
	if err != nil {
		code = http.StatusBadRequest
		return
	}

	model, jobID, err := h.srv.CreateDocument(ctx, payload)
	if err != nil {
		code = http.StatusBadRequest
		return
	}

	docData := model.GetData()
	header, err := deriveResponseHeader(h.tokenRegistry, model, jobID)
	if err != nil {
		code = http.StatusInternalServerError
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, DocumentResponse{Header: header, Data: docData})
}

// Register registers the core apis to the router.
func Register(r *chi.Mux, registry documents.TokenRegistry, docSrv documents.Service) {
	h := handler{srv: Service{docService: docSrv}, tokenRegistry: registry}
	r.Post("/documents", h.CreateDocument)
}
