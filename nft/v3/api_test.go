//go:build unit
// +build unit

package v3

import (
	"context"
	"math/big"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUniquesAPI_CreateClass(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	testAcc, err := contextutil.Account(ctx)
	assert.NoError(t, err, "unable to retrieve account from context")

	testKRP, err := testAcc.GetCentChainAccount().KeyRingPair()
	assert.NoError(t, err, "unable to retrieve key ring pair")

	var meta types.Metadata

	err = types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)

	testCall, err := types.NewCall(
		&meta,
		CreateClassCall,
		types.NewUCompactFromUInt(uint64(classID)),
		types.NewMultiAddressFromAccountID(testKRP.PublicKey),
	)
	assert.NoError(t, err, "unable to create new call")

	extInfo := centchain.ExtrinsicInfo{
		Hash:      types.NewHash([]byte("some_bytes")),
		BlockHash: types.NewHash([]byte("some_more_bytes")),
	}

	centAPIMock.On("SubmitAndWatch", ctx, &meta, testCall, testKRP).
		Return(extInfo, nil)

	res, err := uniquesAPI.CreateClass(ctx, classID)
	assert.NoError(t, err, "unable to create class")
	assert.Equal(t, &extInfo, res, "extrinsic infos should be equal")
}

func TestUniquesAPI_CreateClass_InvalidClassID(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	res, err := uniquesAPI.CreateClass(context.Background(), types.U64(0))
	assert.ErrorIs(t, err, ErrValidation, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestUniquesAPI_CreateClass_CtxAccountError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	classID := types.U64(1234)

	res, err := uniquesAPI.CreateClass(context.Background(), classID)
	assert.ErrorIs(t, err, ErrAccountFromContextRetrieval, "errors should match")
	assert.Nil(t, res, "expected nil extrinsic info")
}

func TestUniquesAPI_CreateClass_KeyRingPairError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	classID := types.U64(1234)

	mockAccount := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	ccAcc := config.CentChainAccount{
		ID: "non-hex-string",
	}

	mockAccount.On("GetCentChainAccount").
		Return(ccAcc)

	res, err := uniquesAPI.CreateClass(ctx, classID)
	assert.ErrorIs(t, err, ErrKeyRingPairRetrieval, "errors should match")
	assert.Nil(t, res, "expected nil extrinsic info")
}

func TestUniquesAPI_CreateClass_MetadataError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	centAPIMock.On("GetMetadataLatest").
		Return(nil, errors.New("metadata error"))

	classID := types.U64(1234)

	res, err := uniquesAPI.CreateClass(ctx, classID)
	assert.ErrorIs(t, err, ErrMetadataRetrieval, "errors should match")
	assert.Nil(t, res, "extrinsic should be nil")
}

func TestUniquesAPI_CreateClass_NewCallError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	invalidMeta := &types.Metadata{}

	centAPIMock.On("GetMetadataLatest").
		Return(invalidMeta, nil)

	classID := types.U64(1234)

	res, err := uniquesAPI.CreateClass(ctx, classID)
	assert.ErrorIs(t, err, ErrCallCreation, "errors should match")
	assert.Nil(t, res, "extrinsic info should be nil")
}

func TestUniquesAPI_CreateClass_SubmitAndWatchError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	testAcc, err := contextutil.Account(ctx)
	assert.NoError(t, err, "unable to retrieve account from context")

	testKRP, err := testAcc.GetCentChainAccount().KeyRingPair()
	assert.NoError(t, err, "unable to retrieve key ring pair")

	var meta types.Metadata

	err = types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)

	testCall, err := types.NewCall(
		&meta,
		CreateClassCall,
		types.NewUCompactFromUInt(uint64(classID)),
		types.NewMultiAddressFromAccountID(testKRP.PublicKey),
	)
	assert.NoError(t, err, "unable to create new call")

	extInfo := centchain.ExtrinsicInfo{}

	centAPIMock.On("SubmitAndWatch", ctx, &meta, testCall, testKRP).
		Return(extInfo, errors.New("submit and watch error"))

	res, err := uniquesAPI.CreateClass(ctx, classID)
	assert.ErrorIs(t, err, ErrSubmitAndWatchExtrinsic, "errors should match")
	assert.Nil(t, res, "extrinsic info should be nil")
}

func TestUniquesAPI_MintInstance(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	testAcc, err := contextutil.Account(ctx)
	assert.NoError(t, err, "unable to retrieve account from context")

	testKRP, err := testAcc.GetCentChainAccount().KeyRingPair()
	assert.NoError(t, err, "unable to retrieve key ring pair")

	var meta types.Metadata

	err = types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))
	accountID := types.NewAccountID([]byte("account-id"))

	testCall, err := types.NewCall(
		&meta,
		MintInstanceCall,
		types.NewUCompactFromUInt(uint64(classID)),
		types.NewUCompact(instanceID.Int),
		types.NewMultiAddressFromAccountID(accountID[:]),
	)
	assert.NoError(t, err, "unable to create new call")

	extInfo := centchain.ExtrinsicInfo{
		Hash:      types.NewHash([]byte("some_bytes")),
		BlockHash: types.NewHash([]byte("some_more_bytes")),
	}

	centAPIMock.On("SubmitAndWatch", ctx, &meta, testCall, testKRP).
		Return(extInfo, nil)

	res, err := uniquesAPI.MintInstance(ctx, classID, instanceID, accountID)
	assert.NoError(t, err, "unable to mint instance")
	assert.Equal(t, &extInfo, res, "extrinsic infos should be equal")
}

func TestUniquesAPI_MintInstance_InvalidData(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	res, err := uniquesAPI.MintInstance(
		context.Background(),
		types.U64(0),
		types.NewU128(*big.NewInt(5678)),
		types.NewAccountID([]byte("acc_id")),
	)
	assert.ErrorIs(t, err, ErrValidation, "errors should match")
	assert.Nil(t, res, "expected no response")

	res, err = uniquesAPI.MintInstance(
		context.Background(),
		types.U64(1234),
		types.NewU128(*big.NewInt(0)),
		types.NewAccountID([]byte("acc_id")),
	)
	assert.ErrorIs(t, err, ErrValidation, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestUniquesAPI_MintInstance_CtxAccountError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))
	accountID := types.NewAccountID([]byte("account-id"))

	res, err := uniquesAPI.MintInstance(context.Background(), classID, instanceID, accountID)
	assert.ErrorIs(t, err, ErrAccountFromContextRetrieval, "errors should match")
	assert.Nil(t, res, "extrinsic info should be nil")
}

func TestUniquesAPI_MintInstance_KeyRingPairError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))
	accountID := types.NewAccountID([]byte("account-id"))

	mockAccount := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	ccAcc := config.CentChainAccount{
		ID: "non-hex-string",
	}

	mockAccount.On("GetCentChainAccount").
		Return(ccAcc)

	res, err := uniquesAPI.MintInstance(ctx, classID, instanceID, accountID)
	assert.ErrorIs(t, err, ErrKeyRingPairRetrieval, "errors should match")
	assert.Nil(t, res, "extrinsic info should be nil")
}

func TestUniquesAPI_MintInstance_MetadataError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	centAPIMock.On("GetMetadataLatest").
		Return(nil, errors.New("metadata error"))

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))
	accountID := types.NewAccountID([]byte("account-id"))

	res, err := uniquesAPI.MintInstance(ctx, classID, instanceID, accountID)
	assert.ErrorIs(t, err, ErrMetadataRetrieval, "errors should match")
	assert.Nil(t, res, "extrinsic info should be nil")
}

func TestUniquesAPI_MintInstance_NewCallError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	invalidMeta := &types.Metadata{}

	centAPIMock.On("GetMetadataLatest").
		Return(invalidMeta, nil)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))
	accountID := types.NewAccountID([]byte("account-id"))

	res, err := uniquesAPI.MintInstance(ctx, classID, instanceID, accountID)
	assert.ErrorIs(t, err, ErrCallCreation, "errors should match")
	assert.Nil(t, res, "extrinsic info should be nil")
}

func TestUniquesAPI_MintInstance_SubmitAndWatchError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	testAcc, err := contextutil.Account(ctx)
	assert.NoError(t, err, "unable to retrieve account from context")

	testKRP, err := testAcc.GetCentChainAccount().KeyRingPair()
	assert.NoError(t, err, "unable to retrieve key ring pair")

	var meta types.Metadata

	err = types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))
	accountID := types.NewAccountID([]byte("account-id"))

	testCall, err := types.NewCall(
		&meta,
		MintInstanceCall,
		types.NewUCompactFromUInt(uint64(classID)),
		types.NewUCompact(instanceID.Int),
		types.NewMultiAddressFromAccountID(accountID[:]),
	)
	assert.NoError(t, err, "unable to create new call")

	extInfo := centchain.ExtrinsicInfo{}
	centAPIMock.On("SubmitAndWatch", ctx, &meta, testCall, testKRP).
		Return(extInfo, errors.New("submit and watch error"))

	res, err := uniquesAPI.MintInstance(ctx, classID, instanceID, accountID)
	assert.ErrorIs(t, err, ErrSubmitAndWatchExtrinsic, "errors should match")
	assert.Nil(t, res, "extrinsic info should be nil")
}

func TestUniquesAPI_GetClassDetails(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	var meta types.Metadata

	err := types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)

	encodedClassID, err := types.EncodeToBytes(classID)
	assert.Nil(t, err, "unable to encode class ID")

	storageKey, err := types.CreateStorageKey(&meta, UniquesPalletName, ClassStorageMethod, encodedClassID)
	assert.Nil(t, err, "unable to create storage key")

	centAPIMock.On("GetStorageLatest", storageKey, mock.Anything).
		Return(true, nil)

	res, err := uniquesAPI.GetClassDetails(context.Background(), classID)
	assert.Nil(t, err, "unable to retrieve class details")
	assert.IsType(t, res, &types.ClassDetails{}, "type should match")
}

func TestUniquesAPI_GetClassDetails_InvalidClassID(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	res, err := uniquesAPI.GetClassDetails(context.Background(), types.U64(0))
	assert.ErrorIs(t, err, ErrValidation, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestUniquesAPI_GetClassDetails_MetadataError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	centAPIMock.On("GetMetadataLatest").
		Return(nil, errors.New("metadata error"))

	classID := types.U64(1234)

	res, err := uniquesAPI.GetClassDetails(context.Background(), classID)
	assert.ErrorIs(t, err, ErrMetadataRetrieval, "errors should match")
	assert.Nil(t, res, "expected nil class details")
}

func TestUniquesAPI_GetClassDetails_StorageKeyError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	invalidMeta := types.Metadata{}

	centAPIMock.On("GetMetadataLatest").
		Return(&invalidMeta, nil)

	classID := types.U64(1234)

	res, err := uniquesAPI.GetClassDetails(context.Background(), classID)
	assert.ErrorIs(t, err, ErrStorageKeyCreation, "errors should match")
	assert.Nil(t, res, "expected nil class details")
}

func TestUniquesAPI_GetClassDetails_StorageError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	var meta types.Metadata

	err := types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)

	encodedClassID, err := types.EncodeToBytes(classID)
	assert.Nil(t, err, "unable to encode class ID")

	storageKey, err := types.CreateStorageKey(&meta, UniquesPalletName, ClassStorageMethod, encodedClassID)
	assert.Nil(t, err, "unable to create storage key")

	centAPIMock.On("GetStorageLatest", storageKey, mock.Anything).
		Return(false, errors.New("storage error"))

	res, err := uniquesAPI.GetClassDetails(context.Background(), classID)
	assert.ErrorIs(t, err, ErrClassDetailsRetrieval, "errors should match")
	assert.Nil(t, res, "expected nil class details")
}

func TestUniquesAPI_GetClassDetails_EmptyStorage(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	var meta types.Metadata

	err := types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)

	encodedClassID, err := types.EncodeToBytes(classID)
	assert.Nil(t, err, "unable to encode class ID")

	storageKey, err := types.CreateStorageKey(&meta, UniquesPalletName, ClassStorageMethod, encodedClassID)
	assert.Nil(t, err, "unable to create storage key")

	centAPIMock.On("GetStorageLatest", storageKey, mock.Anything).
		Return(false, nil)

	res, err := uniquesAPI.GetClassDetails(context.Background(), classID)
	assert.ErrorIs(t, err, ErrClassDetailsNotFound, "errors should match")
	assert.Nil(t, res, "expected nil class details")
}

func TestUniquesAPI_GetInstanceDetails(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	var meta types.Metadata

	err := types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	encodedClassID, err := types.EncodeToBytes(classID)
	assert.Nil(t, err, "unable to encode class ID")

	encodedInstanceID, err := types.EncodeToBytes(instanceID)
	assert.Nil(t, err, "unable to encode instance ID")

	storageKey, err := types.CreateStorageKey(&meta, UniquesPalletName, AssetStorageMethod, encodedClassID, encodedInstanceID)
	assert.Nil(t, err, "unable to create storage key")

	centAPIMock.On("GetStorageLatest", storageKey, mock.Anything).
		Return(true, nil)

	res, err := uniquesAPI.GetInstanceDetails(context.Background(), classID, instanceID)
	assert.Nil(t, err, "unable to retrieve instance details")
	assert.IsType(t, res, &types.InstanceDetails{}, "type should match")
}

func TestUniquesAPI_GetInstanceDetails_InvalidData(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	res, err := uniquesAPI.GetInstanceDetails(context.Background(), types.U64(0), types.NewU128(*big.NewInt(5678)))
	assert.ErrorIs(t, err, ErrValidation, "errors should match")
	assert.Nil(t, res, "expected no response")

	res, err = uniquesAPI.GetInstanceDetails(context.Background(), types.U64(1234), types.NewU128(*big.NewInt(0)))
	assert.ErrorIs(t, err, ErrValidation, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestUniquesAPI_GetInstanceDetails_MetadataError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	centAPIMock.On("GetMetadataLatest").
		Return(nil, errors.New("metadata error"))

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	res, err := uniquesAPI.GetInstanceDetails(context.Background(), classID, instanceID)
	assert.ErrorIs(t, err, ErrMetadataRetrieval, "errors should match")
	assert.Nil(t, res, "expected nil instance details")
}

func TestUniquesAPI_GetInstanceDetails_StorageKeyError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	invalidMeta := types.Metadata{}

	centAPIMock.On("GetMetadataLatest").
		Return(&invalidMeta, nil)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	res, err := uniquesAPI.GetInstanceDetails(context.Background(), classID, instanceID)
	assert.ErrorIs(t, err, ErrStorageKeyCreation, "errors should match")
	assert.Nil(t, res, "expected nil instance details")
}

func TestUniquesAPI_GetInstanceDetails_StorageError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	var meta types.Metadata

	err := types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	encodedClassID, err := types.EncodeToBytes(classID)
	assert.Nil(t, err, "unable to encode class ID")

	encodedInstanceID, err := types.EncodeToBytes(instanceID)
	assert.Nil(t, err, "unable to encode instance ID")

	storageKey, err := types.CreateStorageKey(&meta, UniquesPalletName, AssetStorageMethod, encodedClassID, encodedInstanceID)
	assert.Nil(t, err, "unable to create storage key")

	centAPIMock.On("GetStorageLatest", storageKey, mock.Anything).
		Return(false, errors.New("storage error"))

	res, err := uniquesAPI.GetInstanceDetails(context.Background(), classID, instanceID)
	assert.ErrorIs(t, err, ErrInstanceDetailsRetrieval, "errors should match")
	assert.Nil(t, res, "expected nil instance details")
}

func TestUniquesAPI_GetInstanceDetails_EmptyStorage(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	var meta types.Metadata

	err := types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	encodedClassID, err := types.EncodeToBytes(classID)
	assert.Nil(t, err, "unable to encode class ID")

	encodedInstanceID, err := types.EncodeToBytes(instanceID)
	assert.Nil(t, err, "unable to encode instance ID")

	storageKey, err := types.CreateStorageKey(&meta, UniquesPalletName, AssetStorageMethod, encodedClassID, encodedInstanceID)
	assert.Nil(t, err, "unable to create storage key")

	centAPIMock.On("GetStorageLatest", storageKey, mock.Anything).
		Return(false, nil)

	res, err := uniquesAPI.GetInstanceDetails(context.Background(), classID, instanceID)
	assert.ErrorIs(t, err, ErrInstanceDetailsNotFound, "errors should match")
	assert.Nil(t, res, "expected nil instance details")
}

func TestUniquesAPI_SetMetadata(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	testAcc, err := contextutil.Account(ctx)
	assert.NoError(t, err, "unable to retrieve account from context")

	testKRP, err := testAcc.GetCentChainAccount().KeyRingPair()
	assert.NoError(t, err, "unable to retrieve key ring pair")

	var meta types.Metadata

	err = types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))
	data := utils.RandomSlice(StringLimit)

	testCall, err := types.NewCall(
		&meta,
		SetMetadataCall,
		types.NewUCompactFromUInt(uint64(classID)),
		types.NewUCompact(instanceID.Int),
		data,
		false,
	)
	assert.NoError(t, err, "unable to create new call")

	extInfo := centchain.ExtrinsicInfo{
		Hash:      types.NewHash([]byte("some_bytes")),
		BlockHash: types.NewHash([]byte("some_more_bytes")),
	}

	centAPIMock.On("SubmitAndWatch", ctx, &meta, testCall, testKRP).
		Return(extInfo, nil)

	res, err := uniquesAPI.SetMetadata(ctx, classID, instanceID, data, false)
	assert.NoError(t, err, "unable to set metadata")
	assert.Equal(t, &extInfo, res, "extrinsic infos should be equal")
}

func TestUniquesAPI_SetMetadata_InvalidData(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	res, err := uniquesAPI.SetMetadata(ctx, types.U64(0), types.NewU128(*big.NewInt(5678)), utils.RandomSlice(StringLimit-1), false)
	assert.ErrorIs(t, err, ErrValidation, "errors should match")
	assert.Nil(t, res, "expected no response")

	res, err = uniquesAPI.SetMetadata(ctx, types.U64(1234), types.NewU128(*big.NewInt(0)), utils.RandomSlice(StringLimit-1), false)
	assert.ErrorIs(t, err, ErrValidation, "errors should match")
	assert.Nil(t, res, "expected no response")

	res, err = uniquesAPI.SetMetadata(ctx, types.U64(1234), types.NewU128(*big.NewInt(5678)), utils.RandomSlice(StringLimit+1), false)
	assert.ErrorIs(t, err, ErrValidation, "errors should match")
	assert.Nil(t, res, "expected no response")

	res, err = uniquesAPI.SetMetadata(ctx, types.U64(1234), types.NewU128(*big.NewInt(5678)), nil, false)
	assert.ErrorIs(t, err, ErrValidation, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestUniquesAPI_SetMetadata_CtxAccountError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))
	data := utils.RandomSlice(StringLimit)

	res, err := uniquesAPI.SetMetadata(context.Background(), classID, instanceID, data, false)
	assert.ErrorIs(t, err, ErrAccountFromContextRetrieval, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestUniquesAPI_SetMetadata_KeyRingPairError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	mockAccount := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	ccAcc := config.CentChainAccount{
		ID: "non-hex-string",
	}

	mockAccount.On("GetCentChainAccount").
		Return(ccAcc)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))
	data := utils.RandomSlice(StringLimit)

	res, err := uniquesAPI.SetMetadata(ctx, classID, instanceID, data, false)
	assert.ErrorIs(t, err, ErrKeyRingPairRetrieval, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestUniquesAPI_SetMetadata_MetadataError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	centAPIMock.On("GetMetadataLatest").
		Return(nil, errors.New("metadata error"))

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))
	data := utils.RandomSlice(StringLimit)

	res, err := uniquesAPI.SetMetadata(ctx, classID, instanceID, data, false)
	assert.ErrorIs(t, err, ErrMetadataRetrieval, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestUniquesAPI_SetMetadata_SubmitAndWatchError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	testAcc, err := contextutil.Account(ctx)
	assert.NoError(t, err, "unable to retrieve account from context")

	testKRP, err := testAcc.GetCentChainAccount().KeyRingPair()
	assert.NoError(t, err, "unable to retrieve key ring pair")

	var meta types.Metadata

	err = types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))
	data := utils.RandomSlice(StringLimit)

	testCall, err := types.NewCall(
		&meta,
		SetMetadataCall,
		types.NewUCompactFromUInt(uint64(classID)),
		types.NewUCompact(instanceID.Int),
		data,
		false,
	)
	assert.NoError(t, err, "unable to create new call")

	centAPIMock.On("SubmitAndWatch", ctx, &meta, testCall, testKRP).
		Return(centchain.ExtrinsicInfo{}, errors.New("submit and watch error"))

	res, err := uniquesAPI.SetMetadata(ctx, classID, instanceID, data, false)
	assert.ErrorIs(t, err, ErrSubmitAndWatchExtrinsic, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestUniquesAPI_GetInstanceMetadata(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	var meta types.Metadata

	err := types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	encodedClassID, err := types.EncodeToBytes(classID)
	assert.Nil(t, err, "unable to encode class ID")

	encodedInstanceID, err := types.EncodeToBytes(instanceID)
	assert.Nil(t, err, "unable to encode instance ID")

	storageKey, err := types.CreateStorageKey(&meta, UniquesPalletName, InstanceMetadataMethod, encodedClassID, encodedInstanceID)
	assert.Nil(t, err, "unable to create storage key")

	centAPIMock.On("GetStorageLatest", storageKey, mock.Anything).
		Return(true, nil)

	res, err := uniquesAPI.GetInstanceMetadata(context.Background(), classID, instanceID)
	assert.Nil(t, err, "unable to retrieve instance metadata")
	assert.IsType(t, &types.InstanceMetadata{}, res, "type should match")
}

func TestUniquesAPI_GetInstanceMetadata_InvalidData(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	res, err := uniquesAPI.GetInstanceMetadata(context.Background(), types.U64(0), types.NewU128(*big.NewInt(5678)))
	assert.ErrorIs(t, err, ErrValidation, "errors should match")
	assert.Nil(t, res, "expected no response")

	res, err = uniquesAPI.GetInstanceMetadata(context.Background(), types.U64(1234), types.NewU128(*big.NewInt(0)))
	assert.ErrorIs(t, err, ErrValidation, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestUniquesAPI_GetInstanceMetadata_MetadataError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	centAPIMock.On("GetMetadataLatest").
		Return(nil, errors.New("metadata error"))

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	res, err := uniquesAPI.GetInstanceMetadata(context.Background(), classID, instanceID)
	assert.ErrorIs(t, err, ErrMetadataRetrieval, "errors should match")
	assert.Nil(t, res, "expected nil instance metadata")
}

func TestUniquesAPI_GetInstanceMetadata_StorageKeyError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	invalidMeta := types.Metadata{}

	centAPIMock.On("GetMetadataLatest").
		Return(&invalidMeta, nil)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	res, err := uniquesAPI.GetInstanceMetadata(context.Background(), classID, instanceID)
	assert.ErrorIs(t, err, ErrStorageKeyCreation, "errors should match")
	assert.Nil(t, res, "expected nil instance metadata")
}

func TestUniquesAPI_GetInstanceMetadata_StorageError(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	var meta types.Metadata

	err := types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	encodedClassID, err := types.EncodeToBytes(classID)
	assert.Nil(t, err, "unable to encode class ID")

	encodedInstanceID, err := types.EncodeToBytes(instanceID)
	assert.Nil(t, err, "unable to encode instance ID")

	storageKey, err := types.CreateStorageKey(&meta, UniquesPalletName, InstanceMetadataMethod, encodedClassID, encodedInstanceID)
	assert.Nil(t, err, "unable to create storage key")

	centAPIMock.On("GetStorageLatest", storageKey, mock.Anything).
		Return(false, errors.New("storage error"))

	res, err := uniquesAPI.GetInstanceMetadata(context.Background(), classID, instanceID)
	assert.ErrorIs(t, err, ErrInstanceMetadataRetrieval, "errors should match")
	assert.Nil(t, res, "expected nil instance metadata")
}

func TestUniquesAPI_GetInstanceMetadata_EmptyStorage(t *testing.T) {
	centAPIMock := centchain.NewApiMock(t)

	uniquesAPI := newUniquesAPI(centAPIMock)

	var meta types.Metadata

	err := types.DecodeFromHexString(types.MetadataV14Data, &meta)
	assert.NoError(t, err, "unable to decode metadata V14")

	centAPIMock.On("GetMetadataLatest").
		Return(&meta, nil)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	encodedClassID, err := types.EncodeToBytes(classID)
	assert.Nil(t, err, "unable to encode class ID")

	encodedInstanceID, err := types.EncodeToBytes(instanceID)
	assert.Nil(t, err, "unable to encode instance ID")

	storageKey, err := types.CreateStorageKey(&meta, UniquesPalletName, InstanceMetadataMethod, encodedClassID, encodedInstanceID)
	assert.Nil(t, err, "unable to create storage key")

	centAPIMock.On("GetStorageLatest", storageKey, mock.Anything).
		Return(false, nil)

	res, err := uniquesAPI.GetInstanceMetadata(context.Background(), classID, instanceID)
	assert.ErrorIs(t, err, ErrInstanceMetadataNotFound, "errors should match")
	assert.Nil(t, res, "expected nil instance metadata")
}
