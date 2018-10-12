package nft

import (
	"math/big"

	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
)

type WatchMint struct {
	MintRequestData *MintRequest
	Error           error
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

// NewMintRequest converts the parameters and returns a struct with needed parameter for minting
func NewMintRequest(to *common.Address, anchorId []byte, proofs []*proofspb.Proof, rootHash []byte) (*MintRequest, error) {
	tokenId := tools.ByteSliceToBigInt(tools.RandomSlice(256))
	tokenURI := "http:=//www.centrifuge.io/DUMMY_URI_SERVICE"
	anchorID := tools.ByteSliceToBigInt(anchorId)
	merkleRoot, err := tools.SliceToByte32(rootHash)
	if err != nil {
		return nil, err
	}

	proofData, err := fillProofs(proofs)
	if err != nil {
		return nil, err
	}

	return &MintRequest{
		To:         *to,
		TokenId:    tokenId,
		TokenURI:   tokenURI,
		AnchorId:   anchorID,
		MerkleRoot: merkleRoot,
		Values:     proofData.Values,
		Salts:      proofData.Salts,
		Proofs:     proofData.Proofs}, nil
}

type PaymentObligation interface {
	Mint(to common.Address, tokenId *big.Int, tokenURI string, anchorId *big.Int, merkleRoot [32]byte, values [3]string, salts [3][32]byte, proofs [3][][32]byte) (<-chan *WatchMint, error)
}

func getConfiguredPaymentObligation() PaymentObligation {
	//todo not implemented yet, should return Ethereum PaymentObligation
	return nil
}
