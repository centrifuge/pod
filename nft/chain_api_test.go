// +build unit

package nft

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/errors"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestApi_ValidateNFT(t *testing.T) {
	centAPI := new(centchain.MockAPI)
	api := api{
		api: centAPI,
	}

	anchorID := utils.RandomByte32()
	var to [20]byte
	copy(to[:], utils.RandomSlice(20))
	var staticProofs [3][32]byte

	// missing account
	err := api.ValidateNFT(context.Background(), anchorID, to, nil, staticProofs)
	assert.Error(t, err)

	// failed to get metadata
	ctx := testingconfig.CreateAccountContext(t, cfg)
	centAPI.On("GetMetadataLatest").Return(nil, errors.New("failed to get metadata")).Once()
	err = api.ValidateNFT(ctx, anchorID, to, nil, staticProofs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get metadata")

	// failed to create call
	centAPI.On("GetMetadataLatest").Return(&types.Metadata{}, nil).Once()
	err = api.ValidateNFT(ctx, anchorID, to, nil, staticProofs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported metadata version")

	// failed to execute job
	meta := centchain.MetaDataWithCall(ValidateMint)
	centAPI.On("GetMetadataLatest").Return(meta, nil).Twice()
	centAPI.On("SubmitAndWatch", mock.Anything, mock.Anything).Return(errors.New("failed to submit")).Once()
	err = api.ValidateNFT(ctx, anchorID, to, []SubstrateProof{}, staticProofs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to submit")

	// success
	centAPI.On("SubmitAndWatch", mock.Anything, mock.Anything).Return(nil).Once()
	err = api.ValidateNFT(ctx, anchorID, to, []SubstrateProof{}, staticProofs)
	assert.NoError(t, err)
	centAPI.AssertExpectations(t)
}
