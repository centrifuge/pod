package coreapi

import (
	"encoding/json"
	"time"

	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"

	v2 "github.com/centrifuge/go-centrifuge/identity/v2"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// MonetaryValue defines user format to represent currency type
// Value string representation of decimal number
// ChainID hex bytes representing the chain where the currency is relevant
// ID string representing the Currency (USD|ETH|0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2(DAI)...)
type MonetaryValue struct {
	Value   *documents.Decimal `json:"value" swaggertype:"primitive,string"`
	ChainID byteutils.HexBytes `json:"chain_id" swaggertype:"primitive,string"`
	ID      string             `json:"id"`
}

// SignedValue contains the Identity of who signed the attribute and value which was signed
type SignedValue struct {
	Identity *types.AccountID   `json:"identity" swaggertype:"primitive,string"`
	Value    byteutils.HexBytes `json:"value" swaggertype:"primitive,string"`
}

// AttributeMapRequest defines a map of attributes with attribute key as key
type AttributeMapRequest map[string]AttributeRequest

// CreateDocumentRequest defines the payload for creating documents.
type CreateDocumentRequest struct {
	Scheme      string              `json:"scheme" enums:"generic,entity"`
	ReadAccess  []*types.AccountID  `json:"read_access" swaggertype:"array,string"`
	WriteAccess []*types.AccountID  `json:"write_access" swaggertype:"array,string"`
	Data        interface{}         `json:"data"`
	Attributes  AttributeMapRequest `json:"attributes"`
}

// GenerateAccountPayload holds required fields to generate account with defaults.
type GenerateAccountPayload struct {
	Account Account `json:"account"`
}

func (g *GenerateAccountPayload) ToCreateIdentityRequest() *v2.CreateIdentityRequest {
	return &v2.CreateIdentityRequest{
		Identity:         g.Account.Identity,
		WebhookURL:       g.Account.WebhookURL,
		PrecommitEnabled: g.Account.PrecommitEnabled,
	}
}

// AttributeRequest defines a single attribute.
// Type type of the attribute
// Value simple value of the attribute
// MonetaryValue value for only monetary attribute
type AttributeRequest struct {
	Type          string         `json:"type" enums:"integer,decimal,string,bytes,timestamp,monetary"`
	Value         string         `json:"value"`
	MonetaryValue *MonetaryValue `json:"monetary_value,omitempty"`
}

// AttributeResponse adds key to the attribute.
type AttributeResponse struct {
	AttributeRequest
	Key         byteutils.HexBytes `json:"key" swaggertype:"primitive,string"`
	SignedValue SignedValue        `json:"signed_value"`
}

// AttributeMapResponse maps attribute label to AttributeResponse
type AttributeMapResponse map[string]AttributeResponse

// NFT defines a single NFT.
type NFT struct {
	CollectionID types.U64 `json:"collection_id"`
	ItemID       string    `json:"item_id"`
}

// ResponseHeader holds the common response header fields
type ResponseHeader struct {
	DocumentID        string             `json:"document_id"`
	PreviousVersionID string             `json:"previous_version_id"`
	VersionID         string             `json:"version_id"`
	NextVersionID     string             `json:"next_version_id"`
	Author            string             `json:"author"`
	CreatedAt         string             `json:"created_at"`
	ReadAccess        []*types.AccountID `json:"read_access" swaggertype:"array,string"`
	WriteAccess       []*types.AccountID `json:"write_access" swaggertype:"array,string"`
	JobID             string             `json:"job_id,omitempty"`
	NFTs              []*NFT             `json:"nfts"`
	Status            string             `json:"status,omitempty"`
	Fingerprint       byteutils.HexBytes `json:"fingerprint,omitempty" swaggertype:"primitive,string"`
}

// DocumentResponse is the common response for Document APIs.
type DocumentResponse struct {
	Header     ResponseHeader       `json:"header"`
	Scheme     string               `json:"scheme" enums:"generic,entity"`
	Data       interface{}          `json:"data"`
	Attributes AttributeMapResponse `json:"attributes"`
}

// ToDocumentAttributes converts AttributeRequestMap to document attributes
func ToDocumentAttributes(cattrs map[string]AttributeRequest) (map[documents.AttrKey]documents.Attribute, error) {
	attrs := make(map[documents.AttrKey]documents.Attribute)
	for k, v := range cattrs {
		var attr documents.Attribute
		var err error
		switch documents.AttributeType(v.Type) {
		case documents.AttrMonetary:
			if v.MonetaryValue == nil {
				return nil, errors.NewTypedError(documents.ErrWrongAttrFormat, errors.New("empty value field"))
			}
			attr, err = documents.NewMonetaryAttribute(k, v.MonetaryValue.Value, v.MonetaryValue.ChainID.Bytes(), v.MonetaryValue.ID)
			if err != nil {
				return nil, err
			}
		default:
			attr, err = documents.NewStringAttribute(k, documents.AttributeType(v.Type), v.Value)
			if err != nil {
				return nil, err
			}
		}

		attrs[attr.Key] = attr
	}

	return attrs, nil
}

// ToDocumentsCreatePayload converts CoreAPI create payload to documents payload.
func ToDocumentsCreatePayload(request CreateDocumentRequest) (documents.CreatePayload, error) {
	payload := documents.CreatePayload{
		Scheme: request.Scheme,
		Collaborators: documents.CollaboratorsAccess{
			ReadCollaborators:      request.ReadAccess,
			ReadWriteCollaborators: request.WriteAccess,
		},
	}

	data, err := json.Marshal(request.Data)
	if err != nil {
		return payload, err
	}
	payload.Data = data

	attrs, err := ToDocumentAttributes(request.Attributes)
	if err != nil {
		return payload, err
	}
	payload.Attributes = attrs

	return payload, nil
}

func convertNFTs(nfts []*coredocumentpb.NFT) ([]*NFT, error) {
	var res []*NFT

	for _, n := range nfts {
		var collectionID types.U64

		if err := types.Decode(n.GetRegistryId(), &collectionID); err != nil {
			return nil, err
		}

		var itemID types.U128

		if err := types.Decode(n.GetTokenId(), &itemID); err != nil {
			return nil, err
		}

		res = append(res, &NFT{
			CollectionID: collectionID,
			ItemID:       itemID.String(),
		})
	}

	return res, nil
}

func toAttributeMapResponse(attrs []documents.Attribute) (AttributeMapResponse, error) {
	m := make(AttributeMapResponse)
	for _, v := range attrs {
		vx := v // convert to value
		attrRes := AttributeResponse{
			Key: vx.Key[:],
		}
		switch vx.Value.Type {
		case documents.AttrMonetary:
			id := string(vx.Value.Monetary.ID)
			if vx.Value.Monetary.Type == documents.MonetaryToken {
				id = hexutil.Encode(vx.Value.Monetary.ID)
			}
			attrRes.AttributeRequest = AttributeRequest{
				Type: vx.Value.Type.String(),
				MonetaryValue: &MonetaryValue{
					Value:   vx.Value.Monetary.Value,
					ChainID: vx.Value.Monetary.ChainID,
					ID:      id,
				},
			}
		case documents.AttrSigned:
			signed := SignedValue{
				Identity: v.Value.Signed.Identity,
				Value:    v.Value.Signed.Value,
			}
			attrRes.SignedValue = signed
			attrRes.Type = v.Value.Type.String()
		default:
			val, err := vx.Value.String()
			if err != nil {
				return nil, err
			}
			attrRes.AttributeRequest = AttributeRequest{
				Type:  vx.Value.Type.String(),
				Value: val,
			}
		}

		m[vx.KeyLabel] = attrRes
	}

	return m, nil
}

// DeriveResponseHeader derives an appropriate response header
func DeriveResponseHeader(model documents.Document, jobID string) (response ResponseHeader, err error) {
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

	p, err := model.CalculateTransitionRulesFingerprint()
	if err != nil {
		return response, err
	}

	nfts := model.NFTs()
	cnfts, err := convertNFTs(nfts)
	if err != nil {
		// this could be a temporary failure, so we ignore but warn about the error
		log.Warnf("errors encountered when trying to set nfts to the response: %v", err)
	}

	return ResponseHeader{
		DocumentID:        hexutil.Encode(model.ID()),
		PreviousVersionID: hexutil.Encode(model.PreviousVersion()),
		VersionID:         hexutil.Encode(model.CurrentVersion()),
		NextVersionID:     hexutil.Encode(model.NextVersion()),
		Author:            author.ToHexString(),
		CreatedAt:         ts,
		ReadAccess:        cs.ReadCollaborators,
		WriteAccess:       cs.ReadWriteCollaborators,
		NFTs:              cnfts,
		JobID:             jobID,
		Fingerprint:       p,
	}, nil
}

// GetDocumentResponse converts model to a client api format.
func GetDocumentResponse(model documents.Document, jobID string) (resp DocumentResponse, err error) {
	docData := model.GetData()
	scheme := model.Scheme()
	attrMap, err := toAttributeMapResponse(model.GetAttributes())
	if err != nil {
		return resp, err
	}

	header, err := DeriveResponseHeader(model, jobID)
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

// ProofsResponse holds the proofs for the fields given for a document.
type ProofsResponse struct {
	Header      ProofResponseHeader `json:"header"`
	FieldProofs []documents.Proof   `json:"field_proofs"`
}

// ConvertProofs converts documents.DocumentProof to ProofsResponse
func ConvertProofs(proof *documents.DocumentProof) ProofsResponse {
	return ProofsResponse{
		Header: ProofResponseHeader{
			DocumentID: proof.DocumentID,
			VersionID:  proof.VersionID,
			State:      proof.State,
		},
		FieldProofs: documents.ConvertProofs(proof.FieldProofs),
	}
}

// SignRequest holds the payload to be signed.
type SignRequest struct {
	Payload byteutils.HexBytes `json:"payload" swaggertype:"primitive,string"`
}

// SignResponse holds the signature, pk and Payload for the Sign request.
type SignResponse struct {
	Payload   byteutils.HexBytes `json:"payload" swaggertype:"primitive,string"`
	Signature byteutils.HexBytes `json:"signature" swaggertype:"primitive,string"`
	PublicKey byteutils.HexBytes `json:"public_key" swaggertype:"primitive,string"`
	SignerID  byteutils.HexBytes `json:"signer_id" swaggertype:"primitive,string"`
}

// Account holds identity and proxy information for a Centrifuge account.
type Account struct {
	Identity *types.AccountID `json:"identity" swaggertype:"string"`

	WebhookURL       string `json:"webhook_url"`
	PrecommitEnabled bool   `json:"precommit_enabled"`

	DocumentSigningPublicKey byteutils.HexBytes `json:"document_signing_public_key"`
	P2PPublicSigningKey      byteutils.HexBytes `json:"p2p_public_signing_key"`
	PodOperatorAccountID     *types.AccountID   `json:"pod_operator_account_id"`
}

// Accounts holds a list of accounts
type Accounts struct {
	Data []Account `json:"data"`
}

// NFTResponseHeader holds the NFT mint job ID.
type NFTResponseHeader struct {
	JobID string `json:"job_id"`
}

// MintNFTV3Request holds required fields for minting NFT on the Centrifuge chain.
type MintNFTV3Request struct {
	DocumentID     byteutils.HexBytes `json:"document_id" swaggertype:"primitive,string"`
	Owner          *types.AccountID   `json:"owner" swaggertype:"primitive,string"`
	FreezeMetadata bool               `json:"freeze_metadata"`
	IPFSMetadata   nftv3.IPFSMetadata `json:"ipfs_metadata"`
}

func ToNFTMintRequestV3(req MintNFTV3Request, collectionID types.U64) *nftv3.MintNFTRequest {
	return &nftv3.MintNFTRequest{
		DocumentID:     req.DocumentID,
		CollectionID:   collectionID,
		Owner:          req.Owner,
		IPFSMetadata:   req.IPFSMetadata,
		FreezeMetadata: req.FreezeMetadata,
	}
}

// MintNFTV3Response holds the details of the minted NFT on the Centrifuge chain.
type MintNFTV3Response struct {
	Header         NFTResponseHeader  `json:"header"`
	DocumentID     byteutils.HexBytes `json:"document_id" swaggertype:"primitive,string"`
	CollectionID   types.U64          `json:"collection_id"`
	ItemID         string             `json:"item_id"`
	Owner          *types.AccountID   `json:"owner" swaggertype:"primitive,string"`
	IPFSMetadata   nftv3.IPFSMetadata `json:"ipfs_metadata"`
	FreezeMetadata bool               `json:"freeze_metadata"`
}

type OwnerOfNFTV3Response struct {
	CollectionID types.U64        `json:"collection_id"`
	ItemID       string           `json:"item_id"`
	Owner        *types.AccountID `json:"owner" swaggertype:"primitive,string"`
}

type ItemMetadataOfNFTV3Response struct {
	Deposit string `json:"deposit"`
	// Data contains the IPFS CID of the NFT metadata.
	Data     string `json:"data"`
	IsFrozen bool   `json:"is_frozen"`
}

type ItemAttributeOfNFTV3Response struct {
	Value byteutils.HexBytes `json:"value"`
}

// CreateNFTCollectionV3Request is the request object used for creating an NFT class on Centrifuge chain.
type CreateNFTCollectionV3Request struct {
	CollectionID types.U64 `json:"collection_id"`
}

// CreateNFTCollectionV3Response is the response object for a CreateNFTCollectionV3Request.
type CreateNFTCollectionV3Response struct {
	Header       NFTResponseHeader `json:"header"`
	CollectionID types.U64         `json:"collection_id"`
}

// PodOperatorResponse is the response object for a v3/pod/operator GET request.
type PodOperatorResponse struct {
	PodOperatorAccountID *types.AccountID `json:"pod_operator_account_id" swaggertype:"primitive,string"`
}
