package nft

import (
	"fmt"

	"math/big"

	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TODO remove this when we have a proper dependancy injection mechanism
var poService *paymentObligationService

func setPaymentObligationService(s *paymentObligationService) {
	poService = s
}

func getPaymentObligationService() *paymentObligationService {
	return poService
}

// PaymentObligation is an interface to abstract away payment obligation smart contract
type PaymentObligation interface {
	Mint(opts *bind.TransactOpts, _to common.Address, _tokenId *big.Int, _tokenURI string, _anchorId *big.Int, _merkleRoot [32]byte, _values [3]string, _salts [3][32]byte, _proofs [3][][32]byte) (*types.Transaction, error)
}

// Config is an interface to configurations required by nft package
type Config interface {
	GetIdentityId() ([]byte, error)
	GetEthereumDefaultAccountName() string
}

// paymentObligationService handles all interactions related to minting of NFTs for payment obligations
type paymentObligationService struct {
	paymentObligation PaymentObligation
	identityService   identity.Service
	config            Config
}

// NewPaymentObligationService creates paymentObligationService given the parameters
func NewPaymentObligationService(paymentObligation PaymentObligation, identityService identity.Service, config Config) *paymentObligationService {
	return &paymentObligationService{paymentObligation: paymentObligation, identityService: identityService, config: config}
}

func (s *paymentObligationService) mintNFT(documentID []byte, docType, registryAddress, depositAddress string, proofFields []string) (string, error) {
	documentService, err := getDocumentService(docType)
	if err != nil {
		return "", err
	}

	model, err := documentService.GetLastVersion([]byte(documentID))
	if err != nil {
		return "", err
	}

	corDoc, err := model.PackCoreDocument()
	if err != nil {
		return "", err
	}

	proofs, err := documentService.CreateProofs(documentID, proofFields)
	if err != nil {
		return "", err
	}

	toAddress, err := s.getIdentityAddress()
	if err != nil {
		return "", nil
	}

	anchorID, err := anchors.NewAnchorID(corDoc.CurrentVersion)
	if err != nil {
		return "", nil
	}

	rootHash, err := anchors.NewDocRoot(corDoc.DocumentRoot)
	if err != nil {
		return "", nil
	}

	requestData, err := NewMintRequest(toAddress, anchorID, proofs.FieldProofs, rootHash)
	if err != nil {
		return "", err
	}

	opts, err := ethereum.GetConnection().GetTxOpts(s.config.GetEthereumDefaultAccountName())
	if err != nil {
		return "", err
	}

	_, err = s.paymentObligation.Mint(opts, requestData.To, requestData.TokenId, requestData.TokenURI, requestData.AnchorId,
		requestData.MerkleRoot, requestData.Values, requestData.Salts, requestData.Proofs)
	if err != nil {
		return "", err
	}

	return requestData.TokenId.String(), nil
}

func (s *paymentObligationService) getIdentityAddress() (common.Address, error) {
	centIDBytes, err := s.config.GetIdentityId()
	if err != nil {
		return common.Address{}, err
	}

	centID, err := identity.ToCentID(centIDBytes)
	if err != nil {
		return common.Address{}, err
	}

	address, err := s.identityService.GetIdentityAddress(centID)
	if err != nil {
		return common.Address{}, err
	}
	return address, nil
}

// MintRequest holds the data needed to mint and NFT from a Centrifuge document
type MintRequest struct {

	// To is the address of the recipient of the minted token
	To common.Address

	// TokenId is the ID for the minted token
	TokenId *big.Int

	// TokenURI is the metadata uri
	TokenURI string

	// AnchorId is the ID of the document as identified by the set up anchorRegistry.
	AnchorId *big.Int

	// MerkleRoot is the root hash of the merkle proof/doc
	MerkleRoot [32]byte

	// Values are the values of the leafs that is being proved Will be converted to string and concatenated for proof verification as outlined in precise-proofs library.
	Values [3]string

	// salts are the salts for the field that is being proved Will be concatenated for proof verification as outlined in precise-proofs library.
	Salts [3][32]byte

	// Proofs are the documents proofs that are needed
	Proofs [3][][32]byte
}

// NewMintRequest converts the parameters and returns a struct with needed parameter for minting
func NewMintRequest(to common.Address, anchorID anchors.AnchorID, proofs []*proofspb.Proof, rootHash anchors.DocRoot) (*MintRequest, error) {
	tokenId := tools.ByteSliceToBigInt(tools.RandomSlice(256))
	// TODO move this to config
	tokenURI := "http:=//www.centrifuge.io/DUMMY_URI_SERVICE"
	proofData, err := fillProofs(proofs)
	if err != nil {
		return nil, err
	}

	return &MintRequest{
		To:         to,
		TokenId:    tokenId,
		TokenURI:   tokenURI,
		AnchorId:   anchorID.BigInt(),
		MerkleRoot: rootHash,
		Values:     proofData.Values,
		Salts:      proofData.Salts,
		Proofs:     proofData.Proofs}, nil
}

type proofData struct {
	Values [3]string
	Salts  [3][32]byte
	Proofs [3][][32]byte
}

func fillProofs(proofspb []*proofspb.Proof) (*proofData, error) {
	var values [3]string
	var salts [3][32]byte
	var proofs [3][][32]byte

	for i, p := range proofspb {
		values[i] = p.Value
		salt32, err := tools.SliceToByte32(p.Salt)
		if err != nil {
			return nil, err
		}

		salts[i] = salt32
		property, err := convertProofProperty(p.SortedHashes)
		if err != nil {
			return nil, err
		}
		proofs[i] = property
	}

	return &proofData{Values: values, Salts: salts, Proofs: proofs}, nil
}

func convertProofProperty(sortedHashes [][]byte) ([][32]byte, error) {
	var property [][32]byte
	for _, hash := range sortedHashes {
		hash32, err := tools.SliceToByte32(hash)
		if err != nil {
			return nil, err
		}
		property = append(property, hash32)
	}

	return property, nil
}

func getDocumentService(documentType string) (documents.Service, error) {
	docService, err := documents.GetRegistryInstance().LocateService(documentType)
	if err != nil {
		return nil, err
	}

	service, ok := docService.(documents.Service)
	if !ok {
		return nil, fmt.Errorf("couldn't find service for needed document type")

	}
	return service, nil
}
