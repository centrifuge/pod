// +build unit

package nft

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func Test_toSubstrateProofs(t *testing.T) {
	props := [][]byte{
		common.FromHex("0x392614ecdd98ce9b86b6c82242ae1b85aaf53ebe6f52490ed44539c88215b17a"),
	}

	values := [][]byte{
		common.FromHex("0xd6ad85800460ea404f3289484f9300ed787dc951203cb3f0ef5fa0fa4db283cc"),
	}

	salts := [][32]byte{
		byteSliceToByteArray32(common.FromHex("0x34ea1aa3061dca2e1e23573c3b8866f80032d18fd85934d90339c8bafcab0408")),
	}

	sortedHash := [][32]byte{
		utils.RandomByte32(),
		utils.RandomByte32(),
		utils.RandomByte32(),
	}

	sortedHashes := [][][32]byte{sortedHash}
	proofs := toSubstrateProofs(props, values, salts, sortedHashes)
	assert.Len(t, proofs, 1)
	assert.Equal(t, hexutil.Encode(proofs[0].LeafHash[:]), "0xe07c38c0af7a55b6c3bf4ce68856d5d16d841c728519a7c84145567857c0b989")
	assert.Equal(t, proofs[0].SortedHashes, sortedHash)
}
