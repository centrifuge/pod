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

	CreateClassCall  = UniquesPalletName + ".create"
	MintInstanceCall = UniquesPalletName + ".mint"
	SetMetadataCall  = UniquesPalletName + ".set_metadata"

	ClassStorageMethod     = "Class"
	AssetStorageMethod     = "Asset"
	InstanceMetadataMethod = "InstanceMetadataOf"

	// StringLimit as defined in the Centrifuge chain for the uniques pallet.
	StringLimit = 256
)

type UniquesAPI interface {
	CreateClass(ctx context.Context, classID types.U64) (*centchain.ExtrinsicInfo, error)

	MintInstance(ctx context.Context, classID types.U64, instanceID types.U128, owner types.AccountID) (*centchain.ExtrinsicInfo, error)

	GetClassDetails(ctx context.Context, classID types.U64) (*types.ClassDetails, error)

	GetInstanceDetails(ctx context.Context, classID types.U64, instanceID types.U128) (*types.InstanceDetails, error)

	SetMetadata(ctx context.Context, classID types.U64, instanceID types.U128, data []byte, isFrozen bool) (*centchain.ExtrinsicInfo, error)

	GetInstanceMetadata(ctx context.Context, classID types.U64, instanceID types.U128) (*types.InstanceMetadata, error)
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

func (u *uniquesAPI) CreateClass(ctx context.Context, classID types.U64) (*centchain.ExtrinsicInfo, error) {
	if err := validation.Validate(validation.NewValidator(classID, classIDValidatorFn)); err != nil {
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
		CreateClassCall,
		types.NewUCompactFromUInt(uint64(classID)),
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

func (u *uniquesAPI) MintInstance(ctx context.Context, classID types.U64, instanceID types.U128, owner types.AccountID) (*centchain.ExtrinsicInfo, error) {
	err := validation.Validate(
		validation.NewValidator(classID, classIDValidatorFn),
		validation.NewValidator(instanceID, instanceIDValidatorFn),
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
		MintInstanceCall,
		types.NewUCompactFromUInt(uint64(classID)),
		types.NewUCompact(instanceID.Int),
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

func (u *uniquesAPI) GetClassDetails(_ context.Context, classID types.U64) (*types.ClassDetails, error) {
	if err := validation.Validate(validation.NewValidator(classID, classIDValidatorFn)); err != nil {
		u.log.Errorf("Validation error: %s", err)

		return nil, ErrValidation
	}

	meta, err := u.api.GetMetadataLatest()

	if err != nil {
		u.log.Errorf("Couldn't retrieve latest metadata: %s", err)

		return nil, ErrMetadataRetrieval
	}

	encodedClassID, err := types.EncodeToBytes(classID)

	if err != nil {
		u.log.Errorf("Couldn't encode class ID: %s", err)

		return nil, ErrClassIDEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, UniquesPalletName, ClassStorageMethod, encodedClassID)

	if err != nil {
		u.log.Errorf("Couldn't create storage key: %s", err)

		return nil, ErrStorageKeyCreation
	}

	var classDetails types.ClassDetails

	ok, err := u.api.GetStorageLatest(storageKey, &classDetails)

	if err != nil {
		u.log.Errorf("Couldn't retrieve class details from storage: %s", err)

		return nil, ErrClassDetailsRetrieval
	}

	if !ok {
		return nil, ErrClassDetailsNotFound
	}

	return &classDetails, nil
}

func (u *uniquesAPI) GetInstanceDetails(_ context.Context, classID types.U64, instanceID types.U128) (*types.InstanceDetails, error) {
	err := validation.Validate(
		validation.NewValidator(classID, classIDValidatorFn),
		validation.NewValidator(instanceID, instanceIDValidatorFn),
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

	encodedClassID, err := types.EncodeToBytes(classID)

	if err != nil {
		u.log.Errorf("Couldn't encode class ID: %s", err)

		return nil, ErrClassIDEncoding
	}

	encodedInstanceID, err := types.EncodeToBytes(instanceID)

	if err != nil {
		u.log.Errorf("Couldn't encode instance ID: %s", err)

		return nil, ErrInstanceIDEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, UniquesPalletName, AssetStorageMethod, encodedClassID, encodedInstanceID)

	if err != nil {
		u.log.Errorf("Couldn't create storage key: %s", err)

		return nil, ErrStorageKeyCreation
	}

	var instanceDetails types.InstanceDetails

	ok, err := u.api.GetStorageLatest(storageKey, &instanceDetails)

	if err != nil {
		u.log.Errorf("Couldn't retrieve instance details from storage: %s", err)

		return nil, ErrInstanceDetailsRetrieval
	}

	if !ok {
		return nil, ErrInstanceDetailsNotFound
	}

	return &instanceDetails, nil
}

func (u *uniquesAPI) SetMetadata(
	ctx context.Context,
	classID types.U64,
	instanceID types.U128,
	data []byte,
	isFrozen bool,
) (*centchain.ExtrinsicInfo, error) {
	err := validation.Validate(
		validation.NewValidator(classID, classIDValidatorFn),
		validation.NewValidator(instanceID, instanceIDValidatorFn),
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
		types.NewUCompactFromUInt(uint64(classID)),
		types.NewUCompact(instanceID.Int),
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

func (u *uniquesAPI) GetInstanceMetadata(_ context.Context, classID types.U64, instanceID types.U128) (*types.InstanceMetadata, error) {
	err := validation.Validate(
		validation.NewValidator(classID, classIDValidatorFn),
		validation.NewValidator(instanceID, instanceIDValidatorFn),
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

	encodedClassID, err := types.EncodeToBytes(classID)

	if err != nil {
		u.log.Errorf("Couldn't encode class ID: %s", err)

		return nil, ErrClassIDEncoding
	}

	encodedInstanceID, err := types.EncodeToBytes(instanceID)

	if err != nil {
		u.log.Errorf("Couldn't encode instance ID: %s", err)

		return nil, ErrInstanceIDEncoding
	}

	storageKey, err := types.CreateStorageKey(meta, UniquesPalletName, InstanceMetadataMethod, encodedClassID, encodedInstanceID)

	if err != nil {
		u.log.Errorf("Couldn't create storage key: %s", err)

		return nil, ErrStorageKeyCreation
	}

	var instanceMetadata types.InstanceMetadata

	ok, err := u.api.GetStorageLatest(storageKey, &instanceMetadata)

	if err != nil {
		u.log.Errorf("Couldn't retrieve instance metadata from storage: %s", err)

		return nil, ErrInstanceMetadataRetrieval
	}

	if !ok {
		return nil, ErrInstanceMetadataNotFound
	}

	return &instanceMetadata, nil
}
