//go:build unit

package permissions

import (
	"math/rand"
	"testing"

	"github.com/centrifuge/pod/errors"

	"github.com/centrifuge/pod/validation"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/centrifuge/pod/testingutils"
	"github.com/stretchr/testify/mock"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/centchain"
	"github.com/centrifuge/pod/utils"
	"github.com/stretchr/testify/assert"
)

func TestApi_GetPermissionRoles(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	accountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	poolID := types.U64(rand.Uint32())

	meta, err := testingutils.GetTestMetadata()

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	encodedAccountID, err := codec.Encode(accountID)
	assert.NoError(t, err)

	scope := PermissionScope{
		IsPool: true,
		AsPool: poolID,
	}

	encodedScope, err := codec.Encode(scope)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(
		meta,
		PalletName,
		PermissionStorageName,
		encodedAccountID,
		encodedScope,
	)
	assert.NoError(t, err)

	permissionRoles := PermissionRoles{
		PoolAdmin:     POOL_ADMIN,
		CurrencyAdmin: PERMISSIONED_ASSET_ISSUER,
	}

	centAPIMock.On("GetStorageLatest", storageKey, mock.Anything).
		Run(func(args mock.Arguments) {
			storageEntry, ok := args.Get(1).(*PermissionRoles)
			assert.True(t, ok)

			*storageEntry = permissionRoles
		}).
		Return(true, nil).
		Once()

	res, err := api.GetPermissionRoles(accountID, poolID)
	assert.NoError(t, err)
	assert.Equal(t, &permissionRoles, res)
}

func TestApi_GetPermissionRoles_ValidationErrors(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	accountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	poolID := types.U64(rand.Uint32())

	res, err := api.GetPermissionRoles(nil, poolID)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, validation.ErrInvalidAccountID)

	res, err = api.GetPermissionRoles(accountID, 0)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, validation.ErrInvalidU64)
}

func TestApi_GetPermissionRoles_MetadataRetrievalError(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	accountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	poolID := types.U64(rand.Uint32())

	centAPIMock.On("GetMetadataLatest").
		Return(nil, errors.New("error")).
		Once()

	res, err := api.GetPermissionRoles(accountID, poolID)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
}

func TestApi_GetPermissionRoles_StorageKeyCreationError(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	accountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	poolID := types.U64(rand.Uint32())

	var meta types.Metadata

	// NOTE - types.MetadataV14Data does not have info on the Permissions pallet,
	// causing types.CreateStorageKey to fail.
	err = codec.DecodeFromHex(types.MetadataV14Data, &meta)
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil).
		Once()

	res, err := api.GetPermissionRoles(accountID, poolID)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, errors.ErrStorageKeyCreation)
}

func TestApi_GetPermissionRoles_StorageRetrievalError(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	accountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	poolID := types.U64(rand.Uint32())

	meta, err := testingutils.GetTestMetadata()

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	encodedAccountID, err := codec.Encode(accountID)
	assert.NoError(t, err)

	scope := PermissionScope{
		IsPool: true,
		AsPool: poolID,
	}

	encodedScope, err := codec.Encode(scope)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(
		meta,
		PalletName,
		PermissionStorageName,
		encodedAccountID,
		encodedScope,
	)
	assert.NoError(t, err)

	centAPIMock.On("GetStorageLatest", storageKey, mock.Anything).
		Run(func(args mock.Arguments) {
			_, ok := args.Get(1).(*PermissionRoles)
			assert.True(t, ok)
		}).
		Return(false, errors.New("error")).
		Once()

	res, err := api.GetPermissionRoles(accountID, poolID)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrPermissionRolesRetrieval)
}

func TestApi_GetPermissionRoles_StorageEntryNotFound(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	accountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	poolID := types.U64(rand.Uint32())

	meta, err := testingutils.GetTestMetadata()

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil).
		Once()

	encodedAccountID, err := codec.Encode(accountID)
	assert.NoError(t, err)

	scope := PermissionScope{
		IsPool: true,
		AsPool: poolID,
	}

	encodedScope, err := codec.Encode(scope)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(
		meta,
		PalletName,
		PermissionStorageName,
		encodedAccountID,
		encodedScope,
	)
	assert.NoError(t, err)

	centAPIMock.On("GetStorageLatest", storageKey, mock.Anything).
		Run(func(args mock.Arguments) {
			_, ok := args.Get(1).(*PermissionRoles)
			assert.True(t, ok)
		}).
		Return(false, nil).
		Once()

	res, err := api.GetPermissionRoles(accountID, poolID)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrPermissionRolesNotFound)
}
