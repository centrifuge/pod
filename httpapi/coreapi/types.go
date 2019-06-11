package coreapi

import (
	"encoding/json"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// AttributeMap defines a map of attributes with attribute key as key
type AttributeMap map[string]Attribute

// CreateDocumentRequest defines the payload for creating documents.
type CreateDocumentRequest struct {
	Scheme      string           `json:"scheme" enums:"invoice,purchaseorder,entity"`
	ReadAccess  []common.Address `json:"read_access" swaggertype:"array,string"`
	WriteAccess []common.Address `json:"write_access" swaggertype:"array,string"`
	Data        interface{}      `json:"data"`
	Attributes  AttributeMap     `json:"attributes"`
}

// UpdateDocumentRequest defines the payload for updating documents.
type UpdateDocumentRequest struct {
	CreateDocumentRequest
	DocumentID byteutils.HexBytes `json:"document_id" swaggertype:"primitive,string"`
}

// Attribute defines a single attribute.
type Attribute struct {
	Type  string `json:"type" enums:"integer,decimal,string,bytes,timestamp"`
	Value string `json:"value"`
}

// NFT defines a single NFT.
type NFT struct {
	Registry   string `json:"registry"`
	Owner      string `json:"owner"`
	TokenID    string `json:"token_id"`
	TokenIndex string `json:"token_index"`
}

// ResponseHeader holds the common response header fields
type ResponseHeader struct {
	DocumentID  string           `json:"document_id"`
	Version     string           `json:"version"`
	Author      string           `json:"author"`
	CreatedAt   string           `json:"created_at"`
	ReadAccess  []common.Address `json:"read_access" swaggertype:"array,string"`
	WriteAccess []common.Address `json:"write_access" swaggertype:"array,string"`
	JobID       string           `json:"job_id,omitempty"`
	NFTs        []NFT            `json:"nfts"`
}

// DocumentResponse is the common response for Document APIs.
type DocumentResponse struct {
	Header     ResponseHeader `json:"header"`
	Scheme     string         `json:"scheme" enums:"invoice,purchaseorder,entity"`
	Data       interface{}    `json:"data"`
	Attributes AttributeMap   `json:"attributes"`
}

func toDocumentAttributes(cattrs map[string]Attribute) (map[documents.AttrKey]documents.Attribute, error) {
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
			ReadCollaborators:      identity.AddressToDIDs(request.ReadAccess...),
			ReadWriteCollaborators: identity.AddressToDIDs(request.WriteAccess...),
		},
	}

	data, err := json.Marshal(request.Data)
	if err != nil {
		return payload, err
	}
	payload.Data = data

	attrs, err := toDocumentAttributes(request.Attributes)
	if err != nil {
		return payload, err
	}
	payload.Attributes = attrs

	return payload, nil
}

// toDocumentsUpdatePayload converts the update request to UpdatePayload.
func toDocumentsUpdatePayload(request UpdateDocumentRequest) (payload documents.UpdatePayload, err error) {
	createPayload, err := toDocumentsCreatePayload(request.CreateDocumentRequest)
	if err != nil {
		return payload, err
	}

	return documents.UpdatePayload{
		CreatePayload: createPayload,
		DocumentID:    request.DocumentID,
	}, nil
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

func convertAttributes(attrs []documents.Attribute) (AttributeMap, error) {
	m := make(AttributeMap)
	for _, v := range attrs {
		val, err := v.Value.String()
		if err != nil {
			return nil, err
		}

		m[v.KeyLabel] = Attribute{
			Type:  v.Value.Type.String(),
			Value: val,
		}
	}

	return m, nil
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
		ReadAccess:  identity.DIDsToAddress(cs.ReadCollaborators...),
		WriteAccess: identity.DIDsToAddress(cs.ReadWriteCollaborators...),
		NFTs:        cnfts,
		JobID:       id.String(),
	}, nil
}

func getDocumentResponse(model documents.Model, tokenRegistry documents.TokenRegistry, jobID jobs.JobID) (resp DocumentResponse, err error) {
	docData := model.GetData()
	scheme := model.Scheme()
	attrMap, err := convertAttributes(model.GetAttributes())
	if err != nil {
		return resp, err
	}

	header, err := deriveResponseHeader(tokenRegistry, model, jobID)
	if err != nil {
		return resp, err
	}

	return DocumentResponse{Header: header, Scheme: scheme, Data: docData, Attributes: attrMap}, nil
}

// ProofsRequest holds the fields for which proofs are generated.
type ProofsRequest struct {
	Fields []string `json:"fields"`
}

// ProofResponseHeader holds the document details.
type ProofResponseHeader struct {
	DocumentID byteutils.HexBytes `json:"document_id" swaggertype:"primitive,string"`
	VersionID  byteutils.HexBytes `json:"version_id" swaggertype:"primitive,string"`
	State      string             `json:"state"`
}

// Proof represents a single proof
type Proof struct {
	Property     byteutils.HexBytes   `json:"property" swaggertype:"primitive,string"`
	Value        byteutils.HexBytes   `json:"value" swaggertype:"primitive,string"`
	Salt         byteutils.HexBytes   `json:"salt" swaggertype:"primitive,string"`
	Hash         byteutils.HexBytes   `json:"hash" swaggertype:"primitive,string"`
	SortedHashes []byteutils.HexBytes `json:"sorted_hashes" swaggertype:"array,string"`
}

// ProofsResponse holds the proofs for the fields given for a document.
type ProofsResponse struct {
	Header      ProofResponseHeader `json:"header"`
	FieldProofs []Proof             `json:"field_proofs"`
}

func convertProofs(proof *documents.DocumentProof) ProofsResponse {
	resp := ProofsResponse{
		Header: ProofResponseHeader{
			DocumentID: proof.DocumentID,
			VersionID:  proof.VersionID,
			State:      proof.State,
		},
	}

	var proofs []Proof
	for _, pf := range proof.FieldProofs {
		pff := Proof{
			Value:    pf.Value,
			Hash:     pf.Hash,
			Salt:     pf.Salt,
			Property: pf.GetCompactName(),
		}

		var hashes []byteutils.HexBytes
		for _, h := range pf.SortedHashes {
			h := h
			hashes = append(hashes, h)
		}

		pff.SortedHashes = hashes
		proofs = append(proofs, pff)
	}

	resp.FieldProofs = proofs
	return resp
}

// MintNFTRequest holds required fields for minting NFT
type MintNFTRequest struct {
	DocumentID               byteutils.HexBytes `json:"document_id" swaggertype:"primitive,string"`
	RegistryAddress          common.Address     `json:"registry_address" swaggertype:"primitive,string"`
	DepositAddress           common.Address     `json:"deposit_address" swaggertype:"primitive,string"`
	ProofFields              []string           `json:"proof_fields"`
	GrantNFTReadAccess       bool               `json:"grant_nft_read_access"`
	SubmitTokenProof         bool               `json:"submit_token_proof"`
	SubmitNFTReadAccessProof bool               `json:"submit_nft_read_access_proof"`
}

// NFTResponseHeader holds the NFT mint job ID.
type NFTResponseHeader struct {
	JobID string `json:"job_id"`
}

// MintNFTResponse holds the details of the minted NFT.
type MintNFTResponse struct {
	Header          NFTResponseHeader  `json:"header"`
	DocumentID      byteutils.HexBytes `json:"document_id" swaggertype:"primitive,string"`
	TokenID         string             `json:"token_id"`
	RegistryAddress common.Address     `json:"registry_address" swaggertype:"primitive,string"`
	DepositAddress  common.Address     `json:"deposit_address" swaggertype:"primitive,string"`
}

func toNFTMintRequest(req MintNFTRequest) nft.MintNFTRequest {
	return nft.MintNFTRequest{
		DocumentID:               req.DocumentID,
		DepositAddress:           req.DepositAddress,
		GrantNFTReadAccess:       req.GrantNFTReadAccess,
		ProofFields:              req.ProofFields,
		RegistryAddress:          req.RegistryAddress,
		SubmitNFTReadAccessProof: req.SubmitNFTReadAccessProof,
		SubmitTokenProof:         req.SubmitTokenProof,
	}
}

// TransferNFTRequest holds Registry Address and To address for NFT transfer
type TransferNFTRequest struct {
	RegistryAddress common.Address `json:"registry_address" swaggertype:"primitive,string"`
	To              common.Address `json:"to" swaggertype:"primitive,string"`
}

// TransferNFTResponse is the response for NFT transfer.
type TransferNFTResponse struct {
	Header          NFTResponseHeader `json:"header"`
	TokenID         string            `json:"token_id"`
	RegistryAddress common.Address    `json:"registry_address" swaggertype:"primitive,string"`
	To              common.Address    `json:"to" swaggertype:"primitive,string"`
}
