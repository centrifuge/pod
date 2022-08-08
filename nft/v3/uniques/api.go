package uniques

import (
	"context"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity/v2/proxy"
	"github.com/centrifuge/go-centrifuge/validation"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

const (
	PalletName = "Uniques"

	CreateCollectionCall = PalletName + ".create"
	MintCall             = PalletName + ".mint"
	SetMetadataCall      = PalletName + ".set_metadata"

	CollectionStorageMethod = "Collection"
	ItemStorageMethod       = "Item"
	ItemMetadataMethod      = "ItemMetadataOf"

	// StringLimit as defined in the Centrifuge chain for the uniques pallet.
	StringLimit = 256
)

type API interface {
	CreateCollection(ctx context.Context, collectionID types.U64) (*centchain.ExtrinsicInfo, error)

	Mint(ctx context.Context, collectionID types.U64, itemID types.U128, owner *types.AccountID) (*centchain.ExtrinsicInfo, error)

	GetCollectionDetails(ctx context.Context, collectionID types.U64) (*types.CollectionDetails, error)

	GetItemDetails(ctx context.Context, collectionID types.U64, itemID types.U128) (*types.ItemDetails, error)

	SetMetadata(ctx context.Context, collectionID types.U64, itemID types.U128, data []byte, isFrozen bool) (*centchain.ExtrinsicInfo, error)

	GetItemMetadata(ctx context.Context, collectionID types.U64, itemID types.U128) (*types.ItemMetadata, error)
}

type api struct {
	centAPI  centchain.API
	proxyAPI proxy.API
	log      *logging.ZapEventLogger
}

func NewAPI(centApi centchain.API, proxyAPI proxy.API) API {
	return &api{
		centAPI:  centApi,
		proxyAPI: proxyAPI,
		log:      logging.Logger("uniques_api"),
	}
}

func (a *api) CreateCollection(ctx context.Context, collectionID types.U64) (*centchain.ExtrinsicInfo, error) {
	if err := validation.Validate(validation.NewValidator(collectionID, CollectionIDValidatorFn)); err != nil {
		a.log.Errorf("Validation error: %s", err)

		return nil, errors.ErrValidation
	}

	acc, err := contextutil.Account(ctx)

	if err != nil {
		a.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, errors.ErrContextAccountRetrieval
	}

	accProxy, err := acc.GetAccountProxies().WithProxyType(types.NFTManagement)

	if err != nil {
		a.log.Errorf("Couldn't get account proxy: %s", err)

		return nil, errors.ErrAccountProxyRetrieval
	}

	meta, err := a.centAPI.GetMetadataLatest()

	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, errors.ErrMetadataRetrieval
	}

	// NOTE - the admin is the current identity.
	adminMultiAddress, err := types.NewMultiAddressFromAccountID(acc.GetIdentity().ToBytes())

	if err != nil {
		a.log.Errorf("Couldn't create admin multi address: %s", err)

		return nil, ErrAdminMultiAddressCreation
	}

	call, err := types.NewCall(
		meta,
		CreateCollectionCall,
		types.NewUCompactFromUInt(uint64(collectionID)),
		adminMultiAddress,
	)

	if err != nil {
		a.log.Errorf("Couldn't create call: %s", err)

		return nil, errors.ErrCallCreation
	}

	extInfo, err := a.proxyAPI.ProxyCall(ctx, acc.GetIdentity(), accProxy, call)

	if err != nil {
		a.log.Errorf("Couldn't perform proxy call: %s", err)

		return nil, errors.ErrProxyCall
	}

	return extInfo, nil
}

func (a *api) Mint(ctx context.Context, collectionID types.U64, itemID types.U128, owner *types.AccountID) (*centchain.ExtrinsicInfo, error) {
	err := validation.Validate(
		validation.NewValidator(collectionID, CollectionIDValidatorFn),
		validation.NewValidator(itemID, ItemIDValidatorFn),
	)

	if err != nil {
		a.log.Errorf("Validation error: %s", err)

		return nil, errors.ErrValidation
	}

	acc, err := contextutil.Account(ctx)

	if err != nil {
		a.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, errors.ErrContextAccountRetrieval
	}

	accProxy, err := acc.GetAccountProxies().WithProxyType(types.NFTMint)

	if err != nil {
		a.log.Errorf("Couldn't get account proxy: %s", err)

		return nil, errors.ErrAccountProxyRetrieval
	}

	meta, err := a.centAPI.GetMetadataLatest()

	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, errors.ErrMetadataRetrieval
	}

	ownerMultiAddress, err := types.NewMultiAddressFromAccountID(owner.ToBytes())

	if err != nil {
		a.log.Errorf("Couldn't create owner multi address: %s", err)

		return nil, ErrOwnerMultiAddressCreation
	}

	call, err := types.NewCall(
		meta,
		MintCall,
		types.NewUCompactFromUInt(uint64(collectionID)),
		types.NewUCompact(itemID.Int),
		ownerMultiAddress,
	)

	if err != nil {
		a.log.Errorf("Couldn't create call: %s", err)

		return nil, errors.ErrCallCreation
	}

	extInfo, err := a.proxyAPI.ProxyCall(ctx, acc.GetIdentity(), accProxy, call)

	if err != nil {
		a.log.Errorf("Couldn't perform proxy call: %s", err)

		return nil, errors.ErrProxyCall
	}

	return extInfo, nil
}

func (a *api) GetCollectionDetails(_ context.Context, collectionID types.U64) (*types.CollectionDetails, error) {
	if err := validation.Validate(validation.NewValidator(collectionID, CollectionIDValidatorFn)); err != nil {
		a.log.Errorf("Validation error: %s", err)

		return nil, errors.ErrValidation
	}

	meta, err := a.centAPI.GetMetadataLatest()

	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, errors.ErrMetadataRetrieval
	}

	encodedCollectionID, err := types.Encode(collectionID)

	if err != nil {
		a.log.Errorf("Couldn't encode collection ID: %s", err)

		return nil, ErrCollectionIDEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, PalletName, CollectionStorageMethod, encodedCollectionID)

	if err != nil {
		a.log.Errorf("Couldn't create storage key: %s", err)

		return nil, errors.ErrStorageKeyCreation
	}

	var collectionDetails types.CollectionDetails

	ok, err := a.centAPI.GetStorageLatest(storageKey, &collectionDetails)

	if err != nil {
		a.log.Errorf("Couldn't retrieve collection details from storage: %s", err)

		return nil, ErrCollectionDetailsRetrieval
	}

	if !ok {
		return nil, ErrCollectionDetailsNotFound
	}

	return &collectionDetails, nil
}

func (a *api) GetItemDetails(_ context.Context, collectionID types.U64, itemID types.U128) (*types.ItemDetails, error) {
	err := validation.Validate(
		validation.NewValidator(collectionID, CollectionIDValidatorFn),
		validation.NewValidator(itemID, ItemIDValidatorFn),
	)

	if err != nil {
		a.log.Errorf("Validation error: %s", err)

		return nil, errors.ErrValidation
	}

	meta, err := a.centAPI.GetMetadataLatest()
	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, errors.ErrMetadataRetrieval
	}

	encodedCollectionID, err := types.Encode(collectionID)

	if err != nil {
		a.log.Errorf("Couldn't encode collection ID: %s", err)

		return nil, ErrCollectionIDEncoding
	}

	encodedItemID, err := types.Encode(itemID)

	if err != nil {
		a.log.Errorf("Couldn't encode item ID: %s", err)

		return nil, ErrItemIDEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, PalletName, ItemStorageMethod, encodedCollectionID, encodedItemID)

	if err != nil {
		a.log.Errorf("Couldn't create storage key: %s", err)

		return nil, errors.ErrStorageKeyCreation
	}

	var itemDetails types.ItemDetails

	ok, err := a.centAPI.GetStorageLatest(storageKey, &itemDetails)

	if err != nil {
		a.log.Errorf("Couldn't retrieve item details from storage: %s", err)

		return nil, ErrItemDetailsRetrieval
	}

	if !ok {
		return nil, ErrItemDetailsNotFound
	}

	return &itemDetails, nil
}

func (a *api) SetMetadata(
	ctx context.Context,
	collectionID types.U64,
	itemID types.U128,
	data []byte,
	isFrozen bool,
) (*centchain.ExtrinsicInfo, error) {
	err := validation.Validate(
		validation.NewValidator(collectionID, CollectionIDValidatorFn),
		validation.NewValidator(itemID, ItemIDValidatorFn),
		validation.NewValidator(data, metadataValidatorFn),
	)

	if err != nil {
		a.log.Errorf("Validation error: %s", err)

		return nil, errors.ErrValidation
	}

	acc, err := contextutil.Account(ctx)

	if err != nil {
		a.log.Errorf("Couldn't retrieve account from context: %s", err)

		return nil, errors.ErrContextAccountRetrieval
	}

	accProxy, err := acc.GetAccountProxies().WithProxyType(types.NFTMint)

	if err != nil {
		a.log.Errorf("Couldn't get account proxy: %s", err)

		return nil, errors.ErrAccountProxyRetrieval
	}

	meta, err := a.centAPI.GetMetadataLatest()

	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, errors.ErrMetadataRetrieval
	}

	call, err := types.NewCall(
		meta,
		SetMetadataCall,
		types.NewUCompactFromUInt(uint64(collectionID)),
		types.NewUCompact(itemID.Int),
		data,
		isFrozen,
	)

	if err != nil {
		a.log.Errorf("Couldn't create call: %s", err)

		return nil, errors.ErrCallCreation
	}

	extInfo, err := a.proxyAPI.ProxyCall(ctx, acc.GetIdentity(), accProxy, call)

	if err != nil {
		a.log.Errorf("Couldn't perform proxy call: %s", err)

		return nil, errors.ErrProxyCall
	}

	return extInfo, nil
}

func (a *api) GetItemMetadata(_ context.Context, collectionID types.U64, itemID types.U128) (*types.ItemMetadata, error) {
	err := validation.Validate(
		validation.NewValidator(collectionID, CollectionIDValidatorFn),
		validation.NewValidator(itemID, ItemIDValidatorFn),
	)

	if err != nil {
		a.log.Errorf("Validation error: %s", err)

		return nil, errors.ErrValidation
	}

	meta, err := a.centAPI.GetMetadataLatest()

	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, errors.ErrMetadataRetrieval
	}

	encodedCollectionID, err := types.Encode(collectionID)

	if err != nil {
		a.log.Errorf("Couldn't encode collection ID: %s", err)

		return nil, ErrCollectionIDEncoding
	}

	encodedItemID, err := types.Encode(itemID)

	if err != nil {
		a.log.Errorf("Couldn't encode item ID: %s", err)

		return nil, ErrItemIDEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, PalletName, ItemMetadataMethod, encodedCollectionID, encodedItemID)

	if err != nil {
		a.log.Errorf("Couldn't create storage key: %s", err)

		return nil, errors.ErrStorageKeyCreation
	}

	var itemMetadata types.ItemMetadata

	ok, err := a.centAPI.GetStorageLatest(storageKey, &itemMetadata)

	if err != nil {
		a.log.Errorf("Couldn't retrieve item metadata from storage: %s", err)

		return nil, ErrItemMetadataRetrieval
	}

	if !ok {
		return nil, ErrItemMetadataNotFound
	}

	return &itemMetadata, nil
}
