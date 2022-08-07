package v3

import (
	"context"

	"github.com/centrifuge/go-centrifuge/validation"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

const (
	UniquesPalletName = "Uniques"

	CreateCollectionCall = UniquesPalletName + ".create"
	MintCall             = UniquesPalletName + ".mint"
	SetMetadataCall      = UniquesPalletName + ".set_metadata"

	ClassStorageMethod     = "Class"
	AssetStorageMethod     = "Asset"
	InstanceMetadataMethod = "InstanceMetadataOf"

	// StringLimit as defined in the Centrifuge chain for the uniques pallet.
	StringLimit = 256
)

type UniquesAPI interface {
	CreateCollection(ctx context.Context, collectionID types.U64) (*centchain.ExtrinsicInfo, error)

	Mint(ctx context.Context, collectionID types.U64, itemID types.U128, owner *types.AccountID) (*centchain.ExtrinsicInfo, error)

	GetCollectionDetails(ctx context.Context, collectionID types.U64) (*types.CollectionDetails, error)

	GetItemDetails(ctx context.Context, collectionID types.U64, itemID types.U128) (*types.ItemDetails, error)

	SetMetadata(ctx context.Context, classID types.U64, instanceID types.U128, data []byte, isFrozen bool) (*centchain.ExtrinsicInfo, error)

	GetInstanceMetadata(ctx context.Context, classID types.U64, instanceID types.U128) (*types.ItemDetails, error)
}

type uniquesAPI struct {
	api centchain.API
	log *logging.ZapEventLogger
}

func newUniquesAPI(centApi centchain.API) UniquesAPI {
	return &uniquesAPI{
		api: centApi,
		log: logging.Logger("uniques_api"),
	}
}

func (u *uniquesAPI) CreateCollection(ctx context.Context, collectionID types.U64) (*centchain.ExtrinsicInfo, error) {
	if err := validation.Validate(validation.NewValidator(collectionID, collectionIDValidatorFn)); err != nil {
		u.log.Errorf("Validation error: %s", err)

		return nil, ErrValidation
	}

	acc, err := contextutil.Account(ctx)

	if err != nil {
		u.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, ErrAccountFromContextRetrieval
	}

	krp, err := acc.GetCentChainAccount().KeyRingPair()

	if err != nil {
		u.log.Errorf("Couldn't retrieve key ring pair from account: %s", err)

		return nil, ErrKeyRingPairRetrieval
	}

	meta, err := u.api.GetMetadataLatest()

	if err != nil {
		u.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	c, err := types.NewCall(
		meta,
		CreateCollectionCall,
		types.NewUCompactFromUInt(uint64(collectionID)),
		// TODO(cdamian): This should eventually be the identity of the p2p node(s).
		types.NewMultiAddressFromAccountID(krp.PublicKey), // NOTE - the admin is the current account.
	)

	if err != nil {
		u.log.Errorf("Couldn't create call: %s", err)

		return nil, ErrCallCreation
	}

	extInfo, err := u.api.SubmitAndWatch(ctx, meta, c, krp)

	if err != nil {
		u.log.Errorf("Couldn't submit and watch extrinsic: %s", err)

		return nil, ErrSubmitAndWatchExtrinsic
	}

	return &extInfo, nil
}

func (u *uniquesAPI) Mint(ctx context.Context, collectionID types.U64, itemID types.U128, owner *types.AccountID) (*centchain.ExtrinsicInfo, error) {
	err := validation.Validate(
		validation.NewValidator(collectionID, collectionIDValidatorFn),
		validation.NewValidator(itemID, itemIDValidatorFn),
	)

	if err != nil {
		u.log.Errorf("Validation error: %s", err)

		return nil, ErrValidation
	}

	acc, err := contextutil.Account(ctx)

	if err != nil {
		u.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, ErrAccountFromContextRetrieval
	}

	krp, err := acc.GetCentChainAccount().KeyRingPair()

	if err != nil {
		u.log.Errorf("Couldn't retrieve key ring pair from account: %s", err)

		return nil, ErrKeyRingPairRetrieval
	}

	meta, err := u.api.GetMetadataLatest()

	if err != nil {
		u.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	c, err := types.NewCall(
		meta,
		MintCall,
		types.NewUCompactFromUInt(uint64(collectionID)),
		types.NewUCompact(itemID.Int),
		types.NewMultiAddressFromAccountID(owner[:]),
	)

	if err != nil {
		u.log.Errorf("Couldn't create call: %s", err)

		return nil, ErrCallCreation
	}

	extInfo, err := u.api.SubmitAndWatch(ctx, meta, c, krp)

	if err != nil {
		u.log.Errorf("Couldn't submit and watch extrinsic: %s", err)

		return nil, ErrSubmitAndWatchExtrinsic
	}

	return &extInfo, nil
}

func (u *uniquesAPI) GetCollectionDetails(_ context.Context, collectionID types.U64) (*types.CollectionDetails, error) {
	if err := validation.Validate(validation.NewValidator(collectionID, collectionIDValidatorFn)); err != nil {
		u.log.Errorf("Validation error: %s", err)

		return nil, ErrValidation
	}

	meta, err := u.api.GetMetadataLatest()

	if err != nil {
		u.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	encodedCollectionID, err := types.Encode(collectionID)

	if err != nil {
		u.log.Errorf("Couldn't encode collection ID: %s", err)

		return nil, ErrCollectionIDEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, UniquesPalletName, ClassStorageMethod, encodedCollectionID)

	if err != nil {
		u.log.Errorf("Couldn't create storage key: %s", err)

		return nil, ErrStorageKeyCreation
	}

	var collectionDetails types.CollectionDetails

	ok, err := u.api.GetStorageLatest(storageKey, &collectionDetails)

	if err != nil {
		u.log.Errorf("Couldn't retrieve collection details from storage: %s", err)

		return nil, ErrCollectionDetailsRetrieval
	}

	if !ok {
		return nil, ErrCollectionDetailsNotFound
	}

	return &collectionDetails, nil
}

func (u *uniquesAPI) GetItemDetails(_ context.Context, collectionID types.U64, itemID types.U128) (*types.ItemDetails, error) {
	err := validation.Validate(
		validation.NewValidator(collectionID, collectionIDValidatorFn),
		validation.NewValidator(itemID, itemIDValidatorFn),
	)

	if err != nil {
		u.log.Errorf("Validation error: %s", err)

		return nil, ErrValidation
	}

	meta, err := u.api.GetMetadataLatest()
	if err != nil {
		u.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	encodedCollectionID, err := types.Encode(collectionID)

	if err != nil {
		u.log.Errorf("Couldn't encode collection ID: %s", err)

		return nil, ErrCollectionIDEncoding
	}

	encodedItemID, err := types.Encode(itemID)

	if err != nil {
		u.log.Errorf("Couldn't encode item ID: %s", err)

		return nil, ErrItemIDEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, UniquesPalletName, AssetStorageMethod, encodedCollectionID, encodedItemID)

	if err != nil {
		u.log.Errorf("Couldn't create storage key: %s", err)

		return nil, ErrStorageKeyCreation
	}

	var itemDetails types.ItemDetails

	ok, err := u.api.GetStorageLatest(storageKey, &itemDetails)

	if err != nil {
		u.log.Errorf("Couldn't retrieve item details from storage: %s", err)

		return nil, ErrItemDetailsRetrieval
	}

	if !ok {
		return nil, ErrItemDetailsNotFound
	}

	return &itemDetails, nil
}

func (u *uniquesAPI) SetMetadata(
	ctx context.Context,
	collectionID types.U64,
	itemID types.U128,
	data []byte,
	isFrozen bool,
) (*centchain.ExtrinsicInfo, error) {
	err := validation.Validate(
		validation.NewValidator(collectionID, collectionIDValidatorFn),
		validation.NewValidator(itemID, itemIDValidatorFn),
		validation.NewValidator(data, metadataValidatorFn),
	)

	if err != nil {
		u.log.Errorf("Validation error: %s", err)

		return nil, ErrValidation
	}

	acc, err := contextutil.Account(ctx)

	if err != nil {
		u.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, ErrAccountFromContextRetrieval
	}

	krp, err := acc.GetCentChainAccount().KeyRingPair()

	if err != nil {
		u.log.Errorf("Couldn't retrieve key ring pair from account: %s", err)

		return nil, ErrKeyRingPairRetrieval
	}

	meta, err := u.api.GetMetadataLatest()

	if err != nil {
		u.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	c, err := types.NewCall(
		meta,
		SetMetadataCall,
		types.NewUCompactFromUInt(uint64(collectionID)),
		types.NewUCompact(itemID.Int),
		data,
		isFrozen,
	)

	if err != nil {
		u.log.Errorf("Couldn't create call: %s", err)

		return nil, ErrCallCreation
	}

	extInfo, err := u.api.SubmitAndWatch(ctx, meta, c, krp)

	if err != nil {
		u.log.Errorf("Couldn't submit and watch extrinsic: %s", err)

		return nil, ErrSubmitAndWatchExtrinsic
	}

	return &extInfo, nil
}

func (u *uniquesAPI) GetInstanceMetadata(_ context.Context, collectionID types.U64, itemID types.U128) (*types.ItemMetadata, error) {
	err := validation.Validate(
		validation.NewValidator(collectionID, collectionIDValidatorFn),
		validation.NewValidator(itemID, itemIDValidatorFn),
	)

	if err != nil {
		u.log.Errorf("Validation error: %s", err)

		return nil, ErrValidation
	}

	meta, err := u.api.GetMetadataLatest()

	if err != nil {
		u.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	encodedCollectionID, err := types.Encode(collectionID)

	if err != nil {
		u.log.Errorf("Couldn't encode collection ID: %s", err)

		return nil, ErrCollectionIDEncoding
	}

	encodedItemID, err := types.Encode(itemID)

	if err != nil {
		u.log.Errorf("Couldn't encode item ID: %s", err)

		return nil, ErrItemIDEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, UniquesPalletName, InstanceMetadataMethod, encodedCollectionID, encodedItemID)

	if err != nil {
		u.log.Errorf("Couldn't create storage key: %s", err)

		return nil, ErrStorageKeyCreation
	}

	var itemMetadata types.ItemMetadata

	ok, err := u.api.GetStorageLatest(storageKey, &itemMetadata)

	if err != nil {
		u.log.Errorf("Couldn't retrieve item metadata from storage: %s", err)

		return nil, ErrItemMetadataRetrieval
	}

	if !ok {
		return nil, ErrItemMetadataNotFound
	}

	return &itemMetadata, nil
}
