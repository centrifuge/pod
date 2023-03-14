//go:build unit

package uniques

import (
	"context"
	"math/big"
	"math/rand"
	"testing"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/centrifuge/pod/centchain"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/pallets/proxy"
	"github.com/centrifuge/pod/testingutils"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	genericUtils "github.com/centrifuge/pod/testingutils/generic"
	"github.com/centrifuge/pod/utils"
	"github.com/centrifuge/pod/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAPI_CreateCollection(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	collectionID := types.U64(rand.Int())

	adminMultiAddress, err := types.NewMultiAddressFromAccountID(identity.ToBytes())
	assert.NoError(t, err)

	call, err := types.NewCall(
		meta,
		CreateCollectionCall,
		collectionID,
		adminMultiAddress,
	)
	assert.NoError(t, err)

	var krp signature.KeyringPair

	genericUtils.GetMock[*config.PodOperatorMock](mocks).
		On("ToKeyringPair").
		Return(krp).Once()

	extInfo := &centchain.ExtrinsicInfo{}

	genericUtils.GetMock[*proxy.APIMock](mocks).
		On(
			"ProxyCall",
			ctx,
			identity,
			krp,
			types.NewOption(proxyType.PodOperation),
			call,
		).
		Return(extInfo, nil).Once()

	res, err := api.CreateCollection(ctx, collectionID)
	assert.NoError(t, err)
	assert.Equal(t, extInfo, res)
}

func TestAPI_CreateCollection_InvalidCollectionID(t *testing.T) {
	ctx := context.Background()

	api, _ := getAPIWithMocks(t)

	res, err := api.CreateCollection(ctx, types.U64(0))
	assert.ErrorIs(t, err, ErrInvalidCollectionID)
	assert.Nil(t, res)
}

func TestAPI_CreateCollection_IdentityRetrievalError(t *testing.T) {
	ctx := context.Background()

	api, _ := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())

	res, err := api.CreateCollection(ctx, collectionID)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestAPI_CreateCollection_MetadataRetrievalError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(nil, errors.New("error"))

	collectionID := types.U64(rand.Int())

	res, err := api.CreateCollection(ctx, collectionID)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_CreateCollection_CallCreationError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	// NOTE - MetadataV13 does not have info on the Uniques pallet,
	// causing types.NewCall to fail.
	meta := types.NewMetadataV13()

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	collectionID := types.U64(rand.Int())

	res, err := api.CreateCollection(ctx, collectionID)
	assert.ErrorIs(t, err, errors.ErrCallCreation)
	assert.Nil(t, res)
}

func TestAPI_CreateCollection_ProxyCallError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	collectionID := types.U64(rand.Int())

	adminMultiAddress, err := types.NewMultiAddressFromAccountID(identity.ToBytes())
	assert.NoError(t, err)

	call, err := types.NewCall(
		meta,
		CreateCollectionCall,
		collectionID,
		adminMultiAddress,
	)
	assert.NoError(t, err)

	var krp signature.KeyringPair

	genericUtils.GetMock[*config.PodOperatorMock](mocks).
		On("ToKeyringPair").
		Return(krp).Once()

	genericUtils.GetMock[*proxy.APIMock](mocks).
		On(
			"ProxyCall",
			ctx,
			identity,
			krp,
			types.NewOption(proxyType.PodOperation),
			call,
		).
		Return(nil, errors.New("error")).Once()

	res, err := api.CreateCollection(ctx, collectionID)
	assert.ErrorIs(t, err, errors.ErrProxyCall)
	assert.Nil(t, res)
}

func TestAPI_Mint(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	owner, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	ownerMultiAddress, err := types.NewMultiAddressFromAccountID(owner.ToBytes())
	assert.NoError(t, err)

	call, err := types.NewCall(
		meta,
		MintCall,
		collectionID,
		itemID,
		ownerMultiAddress,
	)
	assert.NoError(t, err, "unable to create new call")

	var krp signature.KeyringPair

	genericUtils.GetMock[*config.PodOperatorMock](mocks).
		On("ToKeyringPair").
		Return(krp).Once()

	extInfo := &centchain.ExtrinsicInfo{}

	genericUtils.GetMock[*proxy.APIMock](mocks).
		On(
			"ProxyCall",
			ctx,
			identity,
			krp,
			types.NewOption(proxyType.PodOperation),
			call,
		).
		Return(extInfo, nil).Once()

	res, err := api.Mint(ctx, collectionID, itemID, owner)
	assert.NoError(t, err)
	assert.Equal(t, extInfo, res)
}

func TestAPI_Mint_ValidationError(t *testing.T) {
	ctx := context.Background()

	api, _ := getAPIWithMocks(t)

	res, err := api.Mint(ctx, types.U64(0), types.NewU128(*big.NewInt(0)), nil)
	assert.ErrorIs(t, err, ErrInvalidCollectionID)
	assert.Nil(t, res)

	collectionID := types.U64(rand.Int())

	res, err = api.Mint(ctx, collectionID, types.NewU128(*big.NewInt(0)), nil)
	assert.ErrorIs(t, err, ErrInvalidItemID)
	assert.Nil(t, res)

	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	res, err = api.Mint(ctx, collectionID, itemID, nil)
	assert.ErrorIs(t, err, validation.ErrInvalidAccountID)
	assert.Nil(t, res)
}

func TestAPI_Mint_IdentityRetrievalError(t *testing.T) {
	ctx := context.Background()

	api, _ := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	owner, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	res, err := api.Mint(ctx, collectionID, itemID, owner)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestAPI_Mint_MetadataRetrievalError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	owner, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(nil, errors.New("error"))

	res, err := api.Mint(ctx, collectionID, itemID, owner)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_Mint_CallCreationError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	owner, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	// NOTE - MetadataV13 does not have info on the Uniques pallet,
	// causing types.NewCall to fail.
	meta := types.NewMetadataV13()

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	res, err := api.Mint(ctx, collectionID, itemID, owner)
	assert.ErrorIs(t, err, errors.ErrCallCreation)
	assert.Nil(t, res)
}

func TestAPI_Mint_ProxyCallError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	owner, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	ownerMultiAddress, err := types.NewMultiAddressFromAccountID(owner.ToBytes())
	assert.NoError(t, err)

	call, err := types.NewCall(
		meta,
		MintCall,
		collectionID,
		itemID,
		ownerMultiAddress,
	)
	assert.NoError(t, err, "unable to create new call")

	var krp signature.KeyringPair

	genericUtils.GetMock[*config.PodOperatorMock](mocks).
		On("ToKeyringPair").
		Return(krp).Once()

	genericUtils.GetMock[*proxy.APIMock](mocks).
		On(
			"ProxyCall",
			ctx,
			identity,
			krp,
			types.NewOption(proxyType.PodOperation),
			call,
		).
		Return(nil, errors.New("error")).Once()

	res, err := api.Mint(ctx, collectionID, itemID, owner)
	assert.ErrorIs(t, err, errors.ErrProxyCall)
	assert.Nil(t, res)
}

func TestAPI_GetCollectionDetails(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, CollectionStorageMethod, encodedCollectionID)
	assert.NoError(t, err)

	testCollectionDetails := types.CollectionDetails{
		FreeHolding:       true,
		Instances:         1,
		InstanceMetadatas: 2,
		Attributes:        3,
		IsFrozen:          true,
	}

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&types.CollectionDetails{})).
		Run(func(args mock.Arguments) {
			collectionDetails := args.Get(1).(*types.CollectionDetails)

			*collectionDetails = testCollectionDetails
		}).Return(true, nil).Once()

	res, err := api.GetCollectionDetails(collectionID)
	assert.NoError(t, err)
	assert.Equal(t, testCollectionDetails, *res)
}

func TestAPI_GetCollectionDetails_ValidationError(t *testing.T) {
	api, _ := getAPIWithMocks(t)

	res, err := api.GetCollectionDetails(types.U64(0))
	assert.ErrorIs(t, err, ErrInvalidCollectionID)
	assert.Nil(t, res)
}

func TestAPI_GetCollectionDetails_MetadataRetrievalError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(nil, errors.New("error"))

	res, err := api.GetCollectionDetails(collectionID)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_GetCollectionDetails_StorageKeyCreationError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())

	// NOTE - types.MetadataV13 does not have info on the Keystore pallet,
	// causing types.CreateStorageKey to fail.
	meta := types.NewMetadataV13()

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	res, err := api.GetCollectionDetails(collectionID)
	assert.ErrorIs(t, err, errors.ErrStorageKeyCreation)
	assert.Nil(t, res)
}

func TestAPI_GetCollectionDetails_StorageRetrievalError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, CollectionStorageMethod, encodedCollectionID)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&types.CollectionDetails{})).
		Return(false, errors.New("error")).Once()

	res, err := api.GetCollectionDetails(collectionID)
	assert.ErrorIs(t, err, ErrCollectionDetailsRetrieval)
	assert.Nil(t, res)
}

func TestAPI_GetCollectionDetails_NotFoundError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, CollectionStorageMethod, encodedCollectionID)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&types.CollectionDetails{})).
		Return(false, nil).Once()

	res, err := api.GetCollectionDetails(collectionID)
	assert.ErrorIs(t, err, ErrCollectionDetailsNotFound)
	assert.Nil(t, res)
}

func TestAPI_GetItemDetails(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	encodedItemID, err := codec.Encode(itemID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, ItemStorageMethod, encodedCollectionID, encodedItemID)
	assert.NoError(t, err)

	testItemDetails := types.ItemDetails{
		IsFrozen: true,
		Deposit:  types.NewU128(*big.NewInt(rand.Int63())),
	}

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&types.ItemDetails{})).
		Run(func(args mock.Arguments) {
			itemDetails := args.Get(1).(*types.ItemDetails)

			*itemDetails = testItemDetails
		}).Return(true, nil).Once()

	res, err := api.GetItemDetails(collectionID, itemID)
	assert.NoError(t, err)
	assert.Equal(t, testItemDetails, *res)
}

func TestAPI_GetItemDetails_ValidationError(t *testing.T) {
	api, _ := getAPIWithMocks(t)

	res, err := api.GetItemDetails(types.U64(0), types.NewU128(*big.NewInt(0)))
	assert.ErrorIs(t, err, ErrInvalidCollectionID)
	assert.Nil(t, res)

	collectionID := types.U64(rand.Int())

	res, err = api.GetItemDetails(collectionID, types.NewU128(*big.NewInt(0)))
	assert.ErrorIs(t, err, ErrInvalidItemID)
	assert.Nil(t, res)
}

func TestAPI_GetItemDetails_MetadataRetrievalError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(nil, errors.New("error"))

	res, err := api.GetItemDetails(collectionID, itemID)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_GetItemDetails_StorageKeyCreationError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	// NOTE - types.MetadataV13 does not have info on the Keystore pallet,
	// causing types.CreateStorageKey to fail.
	meta := types.NewMetadataV13()

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	res, err := api.GetItemDetails(collectionID, itemID)
	assert.ErrorIs(t, err, errors.ErrStorageKeyCreation)
	assert.Nil(t, res)
}

func TestAPI_GetItemDetails_StorageRetrievalError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	encodedItemID, err := codec.Encode(itemID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, ItemStorageMethod, encodedCollectionID, encodedItemID)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&types.ItemDetails{})).
		Return(false, errors.New("error")).Once()

	res, err := api.GetItemDetails(collectionID, itemID)
	assert.ErrorIs(t, err, ErrItemDetailsRetrieval)
	assert.Nil(t, res)
}

func TestAPI_GetItemDetails_NotFoundError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	encodedItemID, err := codec.Encode(itemID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, ItemStorageMethod, encodedCollectionID, encodedItemID)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&types.ItemDetails{})).
		Return(false, nil).Once()

	res, err := api.GetItemDetails(collectionID, itemID)
	assert.ErrorIs(t, err, ErrItemDetailsNotFound)
	assert.Nil(t, res)
}

func TestAPI_SetMetadata(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	data := utils.RandomSlice(32)
	isFrozen := true

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	call, err := types.NewCall(
		meta,
		SetMetadataCall,
		collectionID,
		itemID,
		data,
		isFrozen,
	)
	assert.NoError(t, err)

	var krp signature.KeyringPair

	genericUtils.GetMock[*config.PodOperatorMock](mocks).
		On("ToKeyringPair").
		Return(krp).Once()

	extInfo := &centchain.ExtrinsicInfo{}

	genericUtils.GetMock[*proxy.APIMock](mocks).
		On(
			"ProxyCall",
			ctx,
			identity,
			krp,
			types.NewOption(proxyType.PodOperation),
			call,
		).Return(extInfo, nil).Once()

	res, err := api.SetMetadata(ctx, collectionID, itemID, data, isFrozen)
	assert.NoError(t, err)
	assert.Equal(t, extInfo, res)
}

func TestAPI_SetMetadata_ValidationError(t *testing.T) {
	ctx := context.Background()

	api, _ := getAPIWithMocks(t)

	isFrozen := true

	res, err := api.SetMetadata(ctx, types.U64(0), types.NewU128(*big.NewInt(0)), nil, isFrozen)
	assert.ErrorIs(t, err, ErrInvalidCollectionID)
	assert.Nil(t, res)

	collectionID := types.U64(rand.Int())

	res, err = api.SetMetadata(ctx, collectionID, types.NewU128(*big.NewInt(0)), nil, isFrozen)
	assert.ErrorIs(t, err, ErrInvalidItemID)
	assert.Nil(t, res)

	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	res, err = api.SetMetadata(ctx, collectionID, itemID, nil, isFrozen)
	assert.ErrorIs(t, err, ErrMissingMetadata)
	assert.Nil(t, res)

	data := utils.RandomSlice(MetadataLimit + 1)

	res, err = api.SetMetadata(ctx, collectionID, itemID, data, isFrozen)
	assert.ErrorIs(t, err, ErrMetadataTooBig)
	assert.Nil(t, res)
}

func TestAPI_SetMetadata_IdentityRetrievalError(t *testing.T) {
	ctx := context.Background()

	api, _ := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	data := utils.RandomSlice(32)
	isFrozen := true

	res, err := api.SetMetadata(ctx, collectionID, itemID, data, isFrozen)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestAPI_SetMetadata_MetadataRetrievalError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	data := utils.RandomSlice(32)
	isFrozen := true

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(nil, errors.New("error"))

	res, err := api.SetMetadata(ctx, collectionID, itemID, data, isFrozen)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_SetMetadata_CallCreationError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	data := utils.RandomSlice(32)
	isFrozen := true

	// NOTE - MetadataV13 does not have info on the Uniques pallet,
	// causing types.NewCall to fail.
	meta := types.NewMetadataV13()

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	res, err := api.SetMetadata(ctx, collectionID, itemID, data, isFrozen)
	assert.ErrorIs(t, err, errors.ErrCallCreation)
	assert.Nil(t, res)
}

func TestAPI_SetMetadata_ProxyCallError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	data := utils.RandomSlice(32)
	isFrozen := true

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	call, err := types.NewCall(
		meta,
		SetMetadataCall,
		collectionID,
		itemID,
		data,
		isFrozen,
	)
	assert.NoError(t, err)

	var krp signature.KeyringPair

	genericUtils.GetMock[*config.PodOperatorMock](mocks).
		On("ToKeyringPair").
		Return(krp).Once()

	genericUtils.GetMock[*proxy.APIMock](mocks).
		On(
			"ProxyCall",
			ctx,
			identity,
			krp,
			types.NewOption(proxyType.PodOperation),
			call,
		).Return(nil, errors.New("error")).Once()

	res, err := api.SetMetadata(ctx, collectionID, itemID, data, isFrozen)
	assert.ErrorIs(t, err, errors.ErrProxyCall)
	assert.Nil(t, res)
}

func TestAPI_GetItemMetadata(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	encodedItemID, err := codec.Encode(itemID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, ItemMetadataMethod, encodedCollectionID, encodedItemID)
	assert.NoError(t, err)

	testItemMetadata := types.ItemMetadata{
		Deposit:  types.NewU128(*big.NewInt(rand.Int63())),
		Data:     utils.RandomSlice(32),
		IsFrozen: false,
	}

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&types.ItemMetadata{})).
		Run(func(args mock.Arguments) {
			itemMetadata := args.Get(1).(*types.ItemMetadata)

			*itemMetadata = testItemMetadata
		}).Return(true, nil).Once()

	res, err := api.GetItemMetadata(collectionID, itemID)
	assert.NoError(t, err)
	assert.Equal(t, testItemMetadata, *res)
}

func TestAPI_GetItemMetadata_ValidationError(t *testing.T) {
	api, _ := getAPIWithMocks(t)

	res, err := api.GetItemMetadata(types.U64(0), types.NewU128(*big.NewInt(0)))
	assert.ErrorIs(t, err, ErrInvalidCollectionID)
	assert.Nil(t, res)

	collectionID := types.U64(rand.Int())

	res, err = api.GetItemMetadata(collectionID, types.NewU128(*big.NewInt(0)))
	assert.ErrorIs(t, err, ErrInvalidItemID)
	assert.Nil(t, res)
}

func TestAPI_GetItemMetadata_MetadataRetrievalError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(nil, errors.New("error"))

	res, err := api.GetItemMetadata(collectionID, itemID)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_GetItemMetadata_StorageKeyCreationError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	// NOTE - types.MetadataV13 does not have info on the Keystore pallet,
	// causing types.CreateStorageKey to fail.
	meta := types.NewMetadataV13()

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	res, err := api.GetItemMetadata(collectionID, itemID)
	assert.ErrorIs(t, err, errors.ErrStorageKeyCreation)
	assert.Nil(t, res)
}

func TestAPI_GetItemMetadata_StorageRetrievalError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	encodedItemID, err := codec.Encode(itemID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, ItemMetadataMethod, encodedCollectionID, encodedItemID)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&types.ItemMetadata{})).
		Return(false, errors.New("error")).Once()

	res, err := api.GetItemMetadata(collectionID, itemID)
	assert.ErrorIs(t, err, ErrItemMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_GetItemMetadata_NotFoundError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	encodedItemID, err := codec.Encode(itemID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, ItemMetadataMethod, encodedCollectionID, encodedItemID)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&types.ItemMetadata{})).
		Return(false, nil).Once()

	res, err := api.GetItemMetadata(collectionID, itemID)
	assert.ErrorIs(t, err, ErrItemMetadataNotFound)
	assert.Nil(t, res)
}

func TestAPI_SetAttribute(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	key := utils.RandomSlice(32)
	value := utils.RandomSlice(32)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	call, err := types.NewCall(
		meta,
		SetAttributeCall,
		collectionID,
		types.NewOption(itemID),
		key,
		value,
	)
	assert.NoError(t, err)

	var krp signature.KeyringPair

	genericUtils.GetMock[*config.PodOperatorMock](mocks).
		On("ToKeyringPair").
		Return(krp).Once()

	extInfo := &centchain.ExtrinsicInfo{}

	genericUtils.GetMock[*proxy.APIMock](mocks).
		On(
			"ProxyCall",
			ctx,
			identity,
			krp,
			types.NewOption(proxyType.PodOperation),
			call,
		).Return(extInfo, nil).Once()

	res, err := api.SetAttribute(ctx, collectionID, itemID, key, value)
	assert.NoError(t, err)
	assert.Equal(t, extInfo, res)
}

func TestAPI_SetAttribute_ValidationError(t *testing.T) {
	ctx := context.Background()

	api, _ := getAPIWithMocks(t)

	res, err := api.SetAttribute(ctx, types.U64(0), types.NewU128(*big.NewInt(0)), nil, nil)
	assert.ErrorIs(t, err, ErrInvalidCollectionID)
	assert.Nil(t, res)

	collectionID := types.U64(rand.Int())

	res, err = api.SetAttribute(ctx, collectionID, types.NewU128(*big.NewInt(0)), nil, nil)
	assert.ErrorIs(t, err, ErrInvalidItemID)
	assert.Nil(t, res)

	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	res, err = api.SetAttribute(ctx, collectionID, itemID, nil, nil)
	assert.ErrorIs(t, err, ErrMissingKey)
	assert.Nil(t, res)

	key := utils.RandomSlice(KeyLimit + 1)

	res, err = api.SetAttribute(ctx, collectionID, itemID, key, nil)
	assert.ErrorIs(t, err, ErrKeyTooBig)
	assert.Nil(t, res)

	key = utils.RandomSlice(KeyLimit)

	res, err = api.SetAttribute(ctx, collectionID, itemID, key, nil)
	assert.ErrorIs(t, err, ErrMissingValue)
	assert.Nil(t, res)

	value := utils.RandomSlice(ValueLimit + 1)

	res, err = api.SetAttribute(ctx, collectionID, itemID, key, value)
	assert.ErrorIs(t, err, ErrValueTooBig)
	assert.Nil(t, res)
}

func TestAPI_SetAttribute_IdentityRetrievalError(t *testing.T) {
	ctx := context.Background()

	api, _ := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	key := utils.RandomSlice(32)
	value := utils.RandomSlice(32)

	res, err := api.SetAttribute(ctx, collectionID, itemID, key, value)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestAPI_SetAttribute_MetadataRetrievalError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	key := utils.RandomSlice(32)
	value := utils.RandomSlice(32)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(nil, errors.New("error"))

	res, err := api.SetAttribute(ctx, collectionID, itemID, key, value)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_SetAttribute_CallCreationError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	key := utils.RandomSlice(32)
	value := utils.RandomSlice(32)

	// NOTE - MetadataV13 does not have info on the Uniques pallet,
	// causing types.NewCall to fail.
	meta := types.NewMetadataV13()

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	res, err := api.SetAttribute(ctx, collectionID, itemID, key, value)
	assert.ErrorIs(t, err, errors.ErrCallCreation)
	assert.Nil(t, res)
}

func TestAPI_SetAttribute_ProxyCallError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	key := utils.RandomSlice(32)
	value := utils.RandomSlice(32)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	call, err := types.NewCall(
		meta,
		SetAttributeCall,
		collectionID,
		types.NewOption(itemID),
		key,
		value,
	)
	assert.NoError(t, err)

	var krp signature.KeyringPair

	genericUtils.GetMock[*config.PodOperatorMock](mocks).
		On("ToKeyringPair").
		Return(krp).Once()

	genericUtils.GetMock[*proxy.APIMock](mocks).
		On(
			"ProxyCall",
			ctx,
			identity,
			krp,
			types.NewOption(proxyType.PodOperation),
			call,
		).Return(nil, errors.New("error")).Once()

	res, err := api.SetAttribute(ctx, collectionID, itemID, key, value)
	assert.ErrorIs(t, err, errors.ErrProxyCall)
	assert.Nil(t, res)
}

func TestAPI_GetItemAttribute(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	key := utils.RandomSlice(32)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	encodedItemID, err := codec.Encode(types.NewOption(itemID))
	assert.NoError(t, err)

	encodedKey, err := codec.Encode(key)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, AttributeMethod, encodedCollectionID, encodedItemID, encodedKey)
	assert.NoError(t, err)

	testItemAttribute := utils.RandomSlice(32)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&[]byte{})).
		Run(func(args mock.Arguments) {
			itemAttribute := args.Get(1).(*[]byte)

			*itemAttribute = testItemAttribute
		}).Return(true, nil).Once()

	res, err := api.GetItemAttribute(collectionID, itemID, key)
	assert.NoError(t, err)
	assert.Equal(t, testItemAttribute, res)
}

func TestAPI_GetItemAttribute_ValidationError(t *testing.T) {
	api, _ := getAPIWithMocks(t)

	res, err := api.GetItemAttribute(types.U64(0), types.NewU128(*big.NewInt(0)), nil)
	assert.ErrorIs(t, err, ErrInvalidCollectionID)
	assert.Nil(t, res)

	collectionID := types.U64(rand.Int())

	res, err = api.GetItemAttribute(collectionID, types.NewU128(*big.NewInt(0)), nil)
	assert.ErrorIs(t, err, ErrInvalidItemID)
	assert.Nil(t, res)

	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	res, err = api.GetItemAttribute(collectionID, itemID, nil)
	assert.ErrorIs(t, err, ErrMissingKey)
	assert.Nil(t, res)

	key := utils.RandomSlice(KeyLimit + 1)

	res, err = api.GetItemAttribute(collectionID, itemID, key)
	assert.ErrorIs(t, err, ErrKeyTooBig)
	assert.Nil(t, res)
}

func TestAPI_GetItemAttribute_MetadataRetrievalError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	key := utils.RandomSlice(32)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(nil, errors.New("error"))

	res, err := api.GetItemAttribute(collectionID, itemID, key)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_GetItemAttribute_StorageKeyCreationError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	key := utils.RandomSlice(32)

	meta := types.NewMetadataV13()

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	res, err := api.GetItemAttribute(collectionID, itemID, key)
	assert.ErrorIs(t, err, errors.ErrStorageKeyCreation)
	assert.Nil(t, res)
}

func TestAPI_GetItemAttribute_StorageRetrievalError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	key := utils.RandomSlice(32)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	encodedItemID, err := codec.Encode(types.NewOption(itemID))
	assert.NoError(t, err)

	encodedKey, err := codec.Encode(key)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, AttributeMethod, encodedCollectionID, encodedItemID, encodedKey)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&[]byte{})).
		Return(false, errors.New("error")).Once()

	res, err := api.GetItemAttribute(collectionID, itemID, key)
	assert.ErrorIs(t, err, ErrItemAttributeRetrieval)
	assert.Nil(t, res)
}

func TestAPI_GetItemAttribute_NotFoundError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	collectionID := types.U64(rand.Int())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	key := utils.RandomSlice(32)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	encodedItemID, err := codec.Encode(types.NewOption(itemID))
	assert.NoError(t, err)

	encodedKey, err := codec.Encode(key)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, AttributeMethod, encodedCollectionID, encodedItemID, encodedKey)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&[]byte{})).
		Return(false, nil).Once()

	res, err := api.GetItemAttribute(collectionID, itemID, key)
	assert.ErrorIs(t, err, ErrItemAttributeNotFound)
	assert.Nil(t, res)
}

func getAPIWithMocks(t *testing.T) (*api, []any) {
	centAPIMock := centchain.NewAPIMock(t)
	proxyAPIMock := proxy.NewAPIMock(t)
	podOperatorMock := config.NewPodOperatorMock(t)

	uniquesAPI := NewAPI(centAPIMock, proxyAPIMock, podOperatorMock)

	return uniquesAPI.(*api), []any{
		centAPIMock,
		proxyAPIMock,
		podOperatorMock,
	}
}
