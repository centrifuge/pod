package documents

import (
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	proofspb "github.com/centrifuge/precise-proofs/proofs/proto"
)

// Proof represents a single proof
type Proof struct {
	Property     byteutils.HexBytes   `json:"property" swaggertype:"primitive,string"`
	Value        byteutils.HexBytes   `json:"value" swaggertype:"primitive,string"`
	Salt         byteutils.HexBytes   `json:"salt" swaggertype:"primitive,string"`
	Hash         byteutils.HexBytes   `json:"hash" swaggertype:"primitive,string"`
	SortedHashes []byteutils.HexBytes `json:"sorted_hashes" swaggertype:"array,string"`
}

// ConvertProofs converts proto proofs to JSON struct
func ConvertProofs(fieldProofs []*proofspb.Proof) []Proof {
	var proofs []Proof
	for _, pf := range fieldProofs {
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

	return proofs
}