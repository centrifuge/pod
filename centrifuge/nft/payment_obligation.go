package nft

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type WatchMint struct {
	MintRequestData *MintRequestData
	Error      error
}

/*
   * @param To address The recipient of the minted token
   * @param TokenId uint256 The ID for the minted token
   * @param TokenURI string The metadata uri
   * @param AnchorId bytes32 The ID of the document as identified
   * by the set up anchorRegistry.
   * @param MerkleRoot bytes32 The root hash of the merkle proof/doc
   * @param Values bytes32[3] The values of the leafs that is being proved
   * Will be converted to string and concatenated for proof verification as outlined in
   * precise-proofs library.
   * @param Salts bytes32[3] The salts for the field that is being proved
   * Will be concatenated for proof verification as outlined in
   * precise-proofs library.
   * @param Proofs bytes32[][3] Documents proofs that are needed
*/
type MintRequestData struct {
	To common.Address
	TokenId *big.Int
	TokenURI string
	AnchorId *big.Int
	MerkleRoot [32]byte
	Values [3]string
	Salts [3][32]byte
	Proofs [3][][32]byte

}

func getIdentityAddress() (*common.Address, error) {
	centIDBytes, err := config.Config.GetIdentityId()
	if err != nil {
		return nil, err
	}

	centID, err := identity.ToCentID(centIDBytes)

	if err != nil {
		return nil, err
	}

	ethereumIdentity, err := identity.IDService.LookupIdentityForID(centID)

	if err != nil {
		return nil, err
	}

	return ethereumIdentity.GetIdentityAddress()

}


func NewMintRequestData(anchorId []byte,proofs []*proofspb.Proof) (*MintRequestData ,error) {

	/*
	to, err :=  getIdentityAddress()

	if err != nil {
		return nil, err
	}

	tokenId := tools.ByteSliceToBigInt(tools.RandomSlice(256))
	tokenURI := "http:=//www.centrifuge.io/DUMMY_URI_SERVICE"

	//anchorID, err := anchors.NewAnchorID(document.CurrentIdentifier)

*/

return nil, nil
}


type PaymentObligation interface {
	Mint(to common.Address, tokenId *big.Int, tokenURI string, anchorId *big.Int, merkleRoot [32]byte, values [3]string, salts [3][32]byte, proofs [3][][32]byte) (<-chan *WatchMint, error)

}

func getConfiguredPaymentObligation() PaymentObligation {
	//todo not implemented yet, should return Ethereum PaymentObligation
	return nil
}
