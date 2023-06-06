//go:build unit

package loans

import (
	"testing"

	"github.com/centrifuge/pod/errors"

	"github.com/centrifuge/pod/validation"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/centrifuge/pod/centchain"
	"github.com/centrifuge/pod/testingutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApi_GetCreatedLoan(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	poolID := types.U64(123)
	loanID := types.U64(0)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.
		On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	encodedPoolID, err := codec.Encode(poolID)
	assert.NoError(t, err)

	encodedLoanID, err := codec.Encode(loanID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, CreatedLoanStorageName, encodedPoolID, encodedLoanID)
	assert.NoError(t, err)

	testStorageEntry := CreatedLoanStorageEntry{}

	centAPIMock.
		On("GetStorageLatest", storageKey, mock.Anything).
		Run(func(args mock.Arguments) {
			storageEntry, ok := args.Get(1).(*CreatedLoanStorageEntry)
			assert.True(t, ok)

			*storageEntry = testStorageEntry
		}).
		Return(true, nil).
		Once()

	res, err := api.GetCreatedLoan(poolID, loanID)
	assert.NoError(t, err)
	assert.Equal(t, &testStorageEntry, res)
}

func TestApi_GetCreatedLoan_InvalidPoolID(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	poolID := types.U64(0)
	loanID := types.U64(0)

	res, err := api.GetCreatedLoan(poolID, loanID)
	assert.ErrorIs(t, err, validation.ErrInvalidU64)
	assert.Nil(t, res)
}

func TestApi_GetCreatedLoan_MetadataRetrievalError(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	poolID := types.U64(123)
	loanID := types.U64(0)

	centAPIMock.
		On("GetMetadataLatest").
		Return(nil, errors.New("error")).
		Once()

	res, err := api.GetCreatedLoan(poolID, loanID)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestApi_GetCreatedLoan_StorageKeyCreationError(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	poolID := types.U64(123)
	loanID := types.U64(0)

	var meta types.Metadata

	// NOTE - types.MetadataV14Data does not have info on the Loans pallet,
	// causing types.CreateStorageKey to fail.
	err := codec.DecodeFromHex(types.MetadataV14Data, &meta)
	assert.NoError(t, err)

	centAPIMock.
		On("GetMetadataLatest").
		Return(&meta, nil).
		Once()

	res, err := api.GetCreatedLoan(poolID, loanID)
	assert.ErrorIs(t, err, errors.ErrStorageKeyCreation)
	assert.Nil(t, res)
}

func TestApi_GetCreatedLoan_StorageRetrievalError(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	poolID := types.U64(123)
	loanID := types.U64(0)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.
		On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	encodedPoolID, err := codec.Encode(poolID)
	assert.NoError(t, err)

	encodedLoanID, err := codec.Encode(loanID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, CreatedLoanStorageName, encodedPoolID, encodedLoanID)
	assert.NoError(t, err)

	centAPIMock.
		On("GetStorageLatest", storageKey, mock.Anything).
		Run(func(args mock.Arguments) {
			_, ok := args.Get(1).(*CreatedLoanStorageEntry)
			assert.True(t, ok)
		}).
		Return(false, errors.New("error")).
		Once()

	res, err := api.GetCreatedLoan(poolID, loanID)
	assert.ErrorIs(t, err, ErrCreatedLoanRetrieval)
	assert.Nil(t, res)
}

func TestApi_GetCreatedLoan_StorageEntryNotFound(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	poolID := types.U64(123)
	loanID := types.U64(0)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.
		On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	encodedPoolID, err := codec.Encode(poolID)
	assert.NoError(t, err)

	encodedLoanID, err := codec.Encode(loanID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, CreatedLoanStorageName, encodedPoolID, encodedLoanID)
	assert.NoError(t, err)

	centAPIMock.
		On("GetStorageLatest", storageKey, mock.Anything).
		Run(func(args mock.Arguments) {
			_, ok := args.Get(1).(*CreatedLoanStorageEntry)
			assert.True(t, ok)
		}).
		Return(false, nil).
		Once()

	res, err := api.GetCreatedLoan(poolID, loanID)
	assert.ErrorIs(t, err, ErrCreatedLoanNotFound)
	assert.Nil(t, res)
}
