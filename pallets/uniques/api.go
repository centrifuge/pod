package uniques

import (
	"context"

	"github.com/centrifuge/go-centrifuge/pallets/proxy"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"

	"github.com/centrifuge/go-centrifuge/config"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/validation"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

const (
	PalletName = "Uniques"

	CreateCollectionCall = PalletName + ".create"
	MintCall             = PalletName + ".mint"
	SetMetadataCall      = PalletName + ".set_metadata"
	SetAttributeCall     = PalletName + ".set_attribute"

	CollectionStorageMethod = "Class"
	ItemStorageMethod       = "Asset"
	ItemMetadataMethod      = "InstanceMetadataOf"
	AttributeMethod         = "Attribute"

	// MetadataLimit as defined in the Centrifuge chain development runtime.
	MetadataLimit = 256
	// KeyLimit as defined in the Centrifuge chain development runtime.
	KeyLimit = 256
	// ValueLimit as defined in the Centrifuge chain development runtime.
	ValueLimit = 256
)

type API interface {
	CreateCollection(ctx context.Context, collectionID types.U64) (*centchain.ExtrinsicInfo, error)

	Mint(ctx context.Context, collectionID types.U64, itemID types.U128, owner *types.AccountID) (*centchain.ExtrinsicInfo, error)

	GetCollectionDetails(ctx context.Context, collectionID types.U64) (*types.CollectionDetails, error)

	GetItemDetails(ctx context.Context, collectionID types.U64, itemID types.U128) (*types.ItemDetails, error)

	SetMetadata(ctx context.Context, collectionID types.U64, itemID types.U128, data []byte, isFrozen bool) (*centchain.ExtrinsicInfo, error)

	GetItemMetadata(ctx context.Context, collectionID types.U64, itemID types.U128) (*types.ItemMetadata, error)

	SetAttribute(ctx context.Context, collectionID types.U64, itemID types.U128, key []byte, value []byte) (*centchain.ExtrinsicInfo, error)

	GetItemAttribute(ctx context.Context, collectionID types.U64, itemID types.U128, key []byte) ([]byte, error)
}

type api struct {
	cfgService config.Service
	centAPI    centchain.API
	proxyAPI   proxy.API
	log        *logging.ZapEventLogger
}

func NewAPI(cfgService config.Service, centApi centchain.API, proxyAPI proxy.API) API {
	return &api{
		cfgService: cfgService,
		centAPI:    centApi,
		proxyAPI:   proxyAPI,
		log:        logging.Logger("uniques_api"),
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

	meta, err := a.centAPI.GetMetadataLatest()

	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, errors.ErrMetadataRetrieval
	}

	adminMultiAddress, err := types.NewMultiAddressFromAccountID(acc.GetIdentity().ToBytes())

	if err != nil {
		a.log.Errorf("Couldn't create admin multi address: %s", err)

		return nil, ErrAdminMultiAddressCreation
	}

	call, err := types.NewCall(
		meta,
		CreateCollectionCall,
		collectionID,
		adminMultiAddress,
	)

	if err != nil {
		a.log.Errorf("Couldn't create call: %s", err)

		return nil, errors.ErrCallCreation
	}

	podOperator, err := a.cfgService.GetPodOperator()

	if err != nil {
		a.log.Errorf("Couldn't retrieve pod operator: %s", err)

		return nil, errors.ErrPodOperatorRetrieval
	}

	extInfo, err := a.proxyAPI.ProxyCall(
		ctx,
		acc.GetIdentity(),
		podOperator.ToKeyringPair(),
		types.NewOption(proxyType.PodOperation),
		call,
	)

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
		collectionID,
		itemID,
		ownerMultiAddress,
	)

	if err != nil {
		a.log.Errorf("Couldn't create call: %s", err)

		return nil, errors.ErrCallCreation
	}

	podOperator, err := a.cfgService.GetPodOperator()

	if err != nil {
		a.log.Errorf("Couldn't retrieve pod operator: %s", err)

		return nil, errors.ErrPodOperatorRetrieval
	}

	extInfo, err := a.proxyAPI.ProxyCall(
		ctx,
		acc.GetIdentity(),
		podOperator.ToKeyringPair(),
		types.NewOption(proxyType.PodOperation),
		call,
	)

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

	encodedCollectionID, err := codec.Encode(collectionID)

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

	encodedCollectionID, err := codec.Encode(collectionID)

	if err != nil {
		a.log.Errorf("Couldn't encode collection ID: %s", err)

		return nil, ErrCollectionIDEncoding
	}

	encodedItemID, err := codec.Encode(itemID)

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

	meta, err := a.centAPI.GetMetadataLatest()

	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, errors.ErrMetadataRetrieval
	}

	call, err := types.NewCall(
		meta,
		SetMetadataCall,
		collectionID,
		itemID,
		data,
		isFrozen,
	)

	if err != nil {
		a.log.Errorf("Couldn't create call: %s", err)

		return nil, errors.ErrCallCreation
	}

	podOperator, err := a.cfgService.GetPodOperator()

	if err != nil {
		a.log.Errorf("Couldn't retrieve pod operator: %s", err)

		return nil, errors.ErrPodOperatorRetrieval
	}

	extInfo, err := a.proxyAPI.ProxyCall(
		ctx,
		acc.GetIdentity(),
		podOperator.ToKeyringPair(),
		types.NewOption(proxyType.PodOperation),
		call,
	)

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

	encodedCollectionID, err := codec.Encode(collectionID)

	if err != nil {
		a.log.Errorf("Couldn't encode collection ID: %s", err)

		return nil, ErrCollectionIDEncoding
	}

	encodedItemID, err := codec.Encode(itemID)

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

func (a *api) SetAttribute(
	ctx context.Context,
	collectionID types.U64,
	itemID types.U128,
	key []byte,
	value []byte,
) (*centchain.ExtrinsicInfo, error) {
	err := validation.Validate(
		validation.NewValidator(collectionID, CollectionIDValidatorFn),
		validation.NewValidator(itemID, ItemIDValidatorFn),
		validation.NewValidator(key, KeyValidatorFn),
		validation.NewValidator(value, valueValidatorFn),
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

	meta, err := a.centAPI.GetMetadataLatest()

	if err != nil {
		a.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, errors.ErrMetadataRetrieval
	}

	call, err := types.NewCall(
		meta,
		SetAttributeCall,
		collectionID,
		types.NewOption(itemID),
		key,
		value,
	)

	if err != nil {
		a.log.Errorf("Couldn't create call: %s", err)

		return nil, errors.ErrCallCreation
	}

	podOperator, err := a.cfgService.GetPodOperator()

	if err != nil {
		a.log.Errorf("Couldn't retrieve pod operator: %s", err)

		return nil, errors.ErrPodOperatorRetrieval
	}

	extInfo, err := a.proxyAPI.ProxyCall(
		ctx,
		acc.GetIdentity(),
		podOperator.ToKeyringPair(),
		types.NewOption(proxyType.PodOperation),
		call,
	)

	if err != nil {
		a.log.Errorf("Couldn't perform proxy call: %s", err)

		return nil, errors.ErrProxyCall
	}

	return extInfo, nil
}

func (a *api) GetItemAttribute(_ context.Context, collectionID types.U64, itemID types.U128, key []byte) ([]byte, error) {
	err := validation.Validate(
		validation.NewValidator(collectionID, CollectionIDValidatorFn),
		validation.NewValidator(itemID, ItemIDValidatorFn),
		validation.NewValidator(key, KeyValidatorFn),
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

	encodedCollectionID, err := codec.Encode(collectionID)

	if err != nil {
		a.log.Errorf("Couldn't encode collection ID: %s", err)

		return nil, ErrCollectionIDEncoding
	}

	encodedItemID, err := codec.Encode(types.NewOption(itemID))

	if err != nil {
		a.log.Errorf("Couldn't encode item ID: %s", err)

		return nil, ErrItemIDEncoding
	}

	encodedKey, err := codec.Encode(key)

	if err != nil {
		a.log.Errorf("Couldn't encode key: %s", err)

		return nil, ErrKeyEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, PalletName, AttributeMethod, encodedCollectionID, encodedItemID, encodedKey)

	if err != nil {
		a.log.Errorf("Couldn't create storage key: %s", err)

		return nil, errors.ErrStorageKeyCreation
	}

	var value []byte

	ok, err := a.centAPI.GetStorageLatest(storageKey, &value)

	if err != nil {
		a.log.Errorf("Couldn't retrieve item metadata from storage: %s", err)

		return nil, ErrItemAttributeRetrieval
	}

	if !ok {
		return nil, ErrItemAttributeNotFound
	}

	return value, nil
}
