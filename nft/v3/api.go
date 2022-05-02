package v3

import (
	"context"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const (
	UniquesPalletName = "Uniques"

	CreateClassCall  = UniquesPalletName + ".create"
	MintInstanceCall = UniquesPalletName + ".mint"

	ClassStorageMethod = "Class"
	AssetStorageMethod = "Asset"
)

type UniquesAPI interface {
	CreateClass(ctx context.Context, classID types.U64) (*centchain.ExtrinsicInfo, error)

	MintInstance(ctx context.Context, classID types.U64, instanceID types.U128, owner types.AccountID) (*centchain.ExtrinsicInfo, error)

	GetClassDetails(ctx context.Context, classID types.U64) (*types.ClassDetails, error)

	GetInstanceDetails(ctx context.Context, classID types.U64, instanceID types.U128) (*types.InstanceDetails, error)
}

type uniquesAPI struct {
	api centchain.API
}

func newUniquesAPI(centApi centchain.API) UniquesAPI {
	return &uniquesAPI{centApi}
}

func (u *uniquesAPI) CreateClass(ctx context.Context, classID types.U64) (*centchain.ExtrinsicInfo, error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	krp, err := acc.GetCentChainAccount().KeyRingPair()
	if err != nil {
		return nil, err
	}

	meta, err := u.api.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	c, err := types.NewCall(
		meta,
		CreateClassCall,
		types.NewUCompactFromUInt(uint64(classID)),
		// TODO(cdamian): This should eventually be the identity of the p2p node(s).
		types.NewMultiAddressFromAccountID(krp.PublicKey), // NOTE - the admin is the current account.
	)

	if err != nil {
		return nil, err
	}

	extInfo, err := u.api.SubmitAndWatch(ctx, meta, c, krp)

	if err != nil {
		return nil, err
	}

	return &extInfo, nil
}

func (u *uniquesAPI) MintInstance(ctx context.Context, classID types.U64, instanceID types.U128, owner types.AccountID) (*centchain.ExtrinsicInfo, error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	krp, err := acc.GetCentChainAccount().KeyRingPair()
	if err != nil {
		return nil, err
	}

	meta, err := u.api.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	c, err := types.NewCall(
		meta,
		MintInstanceCall,
		types.NewUCompactFromUInt(uint64(classID)),
		types.NewUCompact(instanceID.Int),
		//krp.PublicKey, // NOTE - the owner is the current account
		owner,
	)

	if err != nil {
		return nil, err
	}

	extInfo, err := u.api.SubmitAndWatch(ctx, meta, c, krp)

	if err != nil {
		return nil, err
	}

	return &extInfo, nil
}

func (u *uniquesAPI) GetClassDetails(_ context.Context, classID types.U64) (*types.ClassDetails, error) {
	meta, err := u.api.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	encodedClassID, err := types.EncodeToBytes(classID)

	if err != nil {
		return nil, err
	}

	storageKey, err := types.CreateStorageKey(meta, UniquesPalletName, ClassStorageMethod, encodedClassID)

	if err != nil {
		return nil, err
	}

	var classDetails types.ClassDetails

	if err := u.api.GetStorageLatest(storageKey, &classDetails); err != nil {
		return nil, err
	}

	return &classDetails, nil
}

func (u *uniquesAPI) GetInstanceDetails(_ context.Context, classID types.U64, instanceID types.U128) (*types.InstanceDetails, error) {
	meta, err := u.api.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	encodedClassID, err := types.EncodeToBytes(classID)

	if err != nil {
		return nil, err
	}

	encodedInstanceID, err := types.EncodeToBytes(instanceID)

	if err != nil {
		return nil, err
	}

	storageKey, err := types.CreateStorageKey(meta, UniquesPalletName, AssetStorageMethod, encodedClassID, encodedInstanceID)

	if err != nil {
		return nil, err
	}

	var instanceDetails types.InstanceDetails

	if err := u.api.GetStorageLatest(storageKey, &instanceDetails); err != nil {
		return nil, err
	}

	return &instanceDetails, nil
}
