package coreapi

import (
	"encoding/json"
	"math/big"
	"strings"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/ethereum/go-ethereum/common"
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

// AttributeMapRequest defines a map of attributes with attribute key as key
type AttributeMapRequest map[string]AttributeRequest

// CreateDocumentRequest defines the payload for creating documents.
type CreateDocumentRequest struct {
	Scheme      string              `json:"scheme" enums:"generic,entity"`
	ReadAccess  []identity.DID      `json:"read_access" swaggertype:"array,string"`
	WriteAccess []identity.DID      `json:"write_access" swaggertype:"array,string"`
	Data        interface{}         `json:"data"`
	Attributes  AttributeMapRequest `json:"attributes"`
}

// GenerateAccountPayload holds required fields to generate account with defaults.
type GenerateAccountPayload struct {
	CentChainAccount config.CentChainAccount `json:"centrifuge_chain_account"`
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
	Key byteutils.HexBytes `json:"key" swaggertype:"primitive,string"`
}

// AttributeMapResponse maps attribute label to AttributeResponse
type AttributeMapResponse map[string]AttributeResponse

// NFT defines a single NFT.
type NFT struct {
	Registry   string `json:"registry"`
	Owner      string `json:"owner"`
	TokenID    string `json:"token_id"`
	TokenIndex string `json:"token_index"`
}

// ResponseHeader holds the common response header fields
type ResponseHeader struct {
	DocumentID  string         `json:"document_id"`
	VersionID   string         `json:"version_id"`
	Author      string         `json:"author"`
	CreatedAt   string         `json:"created_at"`
	ReadAccess  []identity.DID `json:"read_access" swaggertype:"array,string"`
	WriteAccess []identity.DID `json:"write_access" swaggertype:"array,string"`
	JobID       string         `json:"job_id,omitempty"`
	NFTs        []NFT          `json:"nfts"`
	Status      string         `json:"status,omitempty"`
}

// DocumentResponse is the common response for Document APIs.
type DocumentResponse struct {
	Header     ResponseHeader       `json:"header"`
	Scheme     string               `json:"scheme" enums:"generic,entity"`
	Data       interface{}          `json:"data"`
	Attributes AttributeMapResponse `json:"attributes"`
}

func toDocumentAttributes(cattrs map[string]AttributeRequest) (map[documents.AttrKey]documents.Attribute, error) {
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

	attrs, err := toDocumentAttributes(request.Attributes)
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
			// Optional value to be part of the document response
			log.Debug(errors.New("token index received is nil or other error: %v", errn))
			i = new(big.Int)
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

func toAttributeMapResponse(attrs []documents.Attribute) (AttributeMapResponse, error) {
	m := make(AttributeMapResponse)
	for _, v := range attrs {
		vx := v // convert to value
		var attrReq AttributeRequest
		switch vx.Value.Type {
		case documents.AttrMonetary:
			id := string(vx.Value.Monetary.ID)
			if vx.Value.Monetary.Type == documents.MonetaryToken {
				id = hexutil.Encode(vx.Value.Monetary.ID)
			}
			attrReq = AttributeRequest{
				Type: vx.Value.Type.String(),
				MonetaryValue: &MonetaryValue{
					Value:   vx.Value.Monetary.Value,
					ChainID: vx.Value.Monetary.ChainID,
					ID:      id,
				},
			}
		default:
			val, err := vx.Value.String()
			if err != nil {
				return nil, err
			}
			attrReq = AttributeRequest{
				Type:  vx.Value.Type.String(),
				Value: val,
			}
		}

		m[vx.KeyLabel] = AttributeResponse{
			AttributeRequest: attrReq,
			Key:              vx.Key[:],
		}
	}

	return m, nil
}

// DeriveResponseHeader derives an appropriate response header
func DeriveResponseHeader(tokenRegistry documents.TokenRegistry, model documents.Model, id jobs.JobID) (response ResponseHeader, err error) {
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
		VersionID:   hexutil.Encode(model.CurrentVersion()),
		Author:      author.String(),
		CreatedAt:   ts,
		ReadAccess:  cs.ReadCollaborators,
		WriteAccess: cs.ReadWriteCollaborators,
		NFTs:        cnfts,
		JobID:       id.String(),
	}, nil
}

// GetDocumentResponse converts model to a client api format.
func GetDocumentResponse(model documents.Model, tokenRegistry documents.TokenRegistry, jobID jobs.JobID) (resp DocumentResponse, err error) {
	docData := model.GetData()
	scheme := model.Scheme()
	attrMap, err := toAttributeMapResponse(model.GetAttributes())
	if err != nil {
		return resp, err
	}

	header, err := DeriveResponseHeader(tokenRegistry, model, jobID)
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

func convertProofs(proof *documents.DocumentProof) ProofsResponse {
	return ProofsResponse{
		Header: ProofResponseHeader{
			DocumentID: proof.DocumentID,
			VersionID:  proof.VersionID,
			State:      proof.State,
		},
		FieldProofs: documents.ConvertProofs(proof.FieldProofs),
	}
}

// MintNFTRequest holds required fields for minting NFT
type MintNFTRequest struct {
	DocumentID               byteutils.HexBytes `json:"document_id" swaggertype:"primitive,string"`
	DepositAddress           common.Address     `json:"deposit_address" swaggertype:"primitive,string"`
	ProofFields              []string           `json:"proof_fields"`
	GrantNFTReadAccess       bool               `json:"grant_nft_access"`
	SubmitTokenProof         bool               `json:"submit_token_proof"`
	SubmitNFTReadAccessProof bool               `json:"submit_nft_owner_access_proof"`
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

func toNFTMintRequest(req MintNFTRequest, registryAddress common.Address) nft.MintNFTRequest {
	return nft.MintNFTRequest{
		DocumentID:               req.DocumentID,
		DepositAddress:           req.DepositAddress,
		GrantNFTReadAccess:       req.GrantNFTReadAccess,
		ProofFields:              req.ProofFields,
		RegistryAddress:          registryAddress,
		SubmitNFTReadAccessProof: req.SubmitNFTReadAccessProof,
		SubmitTokenProof:         req.SubmitTokenProof,
	}
}

// TransferNFTRequest holds Registry Address and To address for NFT transfer
type TransferNFTRequest struct {
	To common.Address `json:"to" swaggertype:"primitive,string"`
}

// TransferNFTResponse is the response for NFT transfer.
type TransferNFTResponse struct {
	Header          NFTResponseHeader `json:"header"`
	TokenID         string            `json:"token_id"`
	RegistryAddress common.Address    `json:"registry_address" swaggertype:"primitive,string"`
	To              common.Address    `json:"to" swaggertype:"primitive,string"`
}

// NFTOwnerResponse is the response for NFT owner request.
type NFTOwnerResponse struct {
	TokenID         string         `json:"token_id"`
	RegistryAddress common.Address `json:"registry_address" swaggertype:"primitive,string"`
	Owner           common.Address `json:"owner" swaggertype:"primitive,string"`
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

// KeyPair represents the public and private key.
type KeyPair struct {
	Pub string `json:"pub"`
	Pvt string `json:"pvt"`
}

// EthAccount holds address of the account.
type EthAccount struct {
	Address  string `json:"address"`
	Key      string `json:"key,omitempty"`
	Password string `json:"password,omitempty"`
}

// Account holds the single account details.
type Account struct {
	EthereumAccount                  EthAccount              `json:"eth_account"`
	EthereumDefaultAccountName       string                  `json:"eth_default_account_name"`
	ReceiveEventNotificationEndpoint string                  `json:"receive_event_notification_endpoint"`
	IdentityID                       byteutils.HexBytes      `json:"identity_id" swaggertype:"primitive,string"`
	SigningKeyPair                   KeyPair                 `json:"signing_key_pair"`
	P2PKeyPair                       KeyPair                 `json:"p2p_key_pair"`
	CentChainAccount                 config.CentChainAccount `json:"centrifuge_chain_account"`
}

// Accounts holds a list of accounts
type Accounts struct {
	Data []Account `json:"data"`
}

func toClientAccount(acc config.Account) Account {
	var p2pkp, signingkp KeyPair
	p2pkp.Pub, p2pkp.Pvt = acc.GetP2PKeyPair()
	signingkp.Pub, signingkp.Pvt = acc.GetSigningKeyPair()
	ccacc := acc.GetCentChainAccount()
	ccacc.Secret = ""
	return Account{
		EthereumAccount: EthAccount{
			Address: acc.GetEthereumAccount().Address,
		},
		IdentityID:                       acc.GetIdentityID(),
		ReceiveEventNotificationEndpoint: acc.GetReceiveEventNotificationEndpoint(),
		EthereumDefaultAccountName:       acc.GetEthereumDefaultAccountName(),
		P2PKeyPair:                       p2pkp,
		SigningKeyPair:                   signingkp,
		CentChainAccount:                 ccacc,
	}
}

func toClientAccounts(accs []config.Account) Accounts {
	var caccs Accounts
	for _, acc := range accs {
		caccs.Data = append(caccs.Data, toClientAccount(acc))
	}

	return caccs
}

func isKeyPairEmpty(kp *KeyPair) bool {
	kp.Pvt, kp.Pub = strings.TrimSpace(kp.Pvt), strings.TrimSpace(kp.Pub)
	return kp.Pvt == "" || kp.Pub == ""
}

func fromClientAccount(cacc Account) (config.Account, error) {
	acc := new(configstore.Account)
	if cacc.EthereumAccount.Address == "" || cacc.EthereumAccount.Key == "" {
		return nil, errors.New("ethereum address/key cannot be empty")
	}

	if cacc.CentChainAccount.ID == "" || cacc.CentChainAccount.Secret == "" || cacc.CentChainAccount.SS58Addr == "" {
		return nil, errors.New("centrifuge chain account cannot be empty ")
	}

	acc.CentChainAccount = cacc.CentChainAccount
	ca := config.AccountConfig(cacc.EthereumAccount)
	acc.EthereumAccount = &ca
	acc.EthereumDefaultAccountName = cacc.EthereumDefaultAccountName

	if isKeyPairEmpty(&cacc.P2PKeyPair) {
		return nil, errors.New("p2p key pair is invalid")
	}
	acc.P2PKeyPair = configstore.KeyPair(cacc.P2PKeyPair)

	if isKeyPairEmpty(&cacc.SigningKeyPair) {
		return nil, errors.New("signing key pair is invalid")
	}
	acc.SigningKeyPair = configstore.KeyPair(cacc.SigningKeyPair)

	if len(cacc.IdentityID) < 1 {
		return nil, errors.New("Identity ID cannot be empty")
	}

	acc.IdentityID = cacc.IdentityID
	acc.ReceiveEventNotificationEndpoint = cacc.ReceiveEventNotificationEndpoint
	return acc, nil
}
