//go:build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	proofspb "github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestConvertProofs(t *testing.T) {
	// nil input
	pfs := ConvertProofs(nil)
	assert.Empty(t, pfs)

	var input []*proofspb.Proof
	p0 := &proofspb.Proof{
		Property: &proofspb.Proof_CompactName{CompactName: utils.RandomSlice(32)},
		Value:    utils.RandomSlice(32),
		Salt:     utils.RandomSlice(32),
		Hash:     []byte{},
		SortedHashes: [][]byte{
			utils.RandomSlice(32),
			utils.RandomSlice(32),
		},
	}
	p1 := &proofspb.Proof{
		Property: &proofspb.Proof_CompactName{CompactName: utils.RandomSlice(32)},
		Value:    utils.RandomSlice(32),
		Salt:     utils.RandomSlice(32),
		Hash:     []byte{},
		SortedHashes: [][]byte{
			utils.RandomSlice(32),
			utils.RandomSlice(32),
		},
	}
	input = append(input, p0)
	input = append(input, p1)

	pfs = ConvertProofs(input)
	assert.Len(t, pfs, 2)
	assert.Equal(t, hexutil.Encode(p0.GetCompactName()), pfs[0].Property.String())
	assert.Equal(t, hexutil.Encode(p1.GetCompactName()), pfs[1].Property.String())
	assert.Len(t, pfs[0].SortedHashes, 2)
	assert.Equal(t, hexutil.Encode(p0.SortedHashes[0]), pfs[0].SortedHashes[0].String())
}
