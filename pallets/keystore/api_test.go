//go:build unit

package keystore

import (
	"context"
	"math/big"
	"math/rand"
	"testing"

	keystoreType "github.com/centrifuge/chain-custom-types/pkg/keystore"
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

func TestAPI_AddKeys(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	keys := []*keystoreType.AddKey{
		{
			Key:     types.NewHash(utils.RandomSlice(32)),
			Purpose: keystoreType.KeyPurposeP2PDocumentSigning,
			KeyType: keystoreType.KeyTypeECDSA,
		},
		{
			Key:     types.NewHash(utils.RandomSlice(32)),
			Purpose: keystoreType.KeyPurposeP2PDocumentSigning,
			KeyType: keystoreType.KeyTypeEDDSA,
		},
	}

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity).
		Once()

	ctx = contextutil.WithAccount(ctx, accountMock)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil).Once()

	var krp signature.KeyringPair

	genericUtils.GetMock[*config.PodOperatorMock](mocks).
		On("ToKeyringPair").
		Return(krp).Once()

	call, err := types.NewCall(meta, AddKeysCall, keys)
	assert.NoError(t, err)

	extInfo := &centchain.ExtrinsicInfo{}

	genericUtils.GetMock[*proxy.APIMock](mocks).On(
		"ProxyCall",
		ctx,
		identity,
		krp,
		types.NewOption(proxyType.KeystoreManagement),
		call,
	).Return(extInfo, nil).Once()

	res, err := api.AddKeys(ctx, keys)
	assert.NoError(t, err)
	assert.Equal(t, extInfo, res)
}

func TestAPI_AddKeys_ValidationError(t *testing.T) {
	ctx := context.Background()

	api, _ := getAPIWithMocks(t)

	res, err := api.AddKeys(ctx, nil)
	assert.ErrorIs(t, err, ErrNoKeysProvided)
	assert.Nil(t, res)

	keys := []*keystoreType.AddKey{
		{
			Key:     emptyKeyHash,
			Purpose: keystoreType.KeyPurposeP2PDocumentSigning,
			KeyType: keystoreType.KeyTypeECDSA,
		},
		{
			Key:     emptyKeyHash,
			Purpose: keystoreType.KeyPurposeP2PDocumentSigning,
			KeyType: keystoreType.KeyTypeEDDSA,
		},
	}

	res, err = api.AddKeys(ctx, keys)
	assert.ErrorIs(t, err, ErrInvalidKey)
	assert.Nil(t, res)
}

func TestAPI_AddKeys_IdentityRetrievalError(t *testing.T) {
	ctx := context.Background()

	api, _ := getAPIWithMocks(t)

	keys := []*keystoreType.AddKey{
		{
			Key:     types.NewHash(utils.RandomSlice(32)),
			Purpose: keystoreType.KeyPurposeP2PDocumentSigning,
			KeyType: keystoreType.KeyTypeECDSA,
		},
		{
			Key:     types.NewHash(utils.RandomSlice(32)),
			Purpose: keystoreType.KeyPurposeP2PDocumentSigning,
			KeyType: keystoreType.KeyTypeEDDSA,
		},
	}

	res, err := api.AddKeys(ctx, keys)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestAPI_AddKeys_MetadataRetrievalError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	keys := []*keystoreType.AddKey{
		{
			Key:     types.NewHash(utils.RandomSlice(32)),
			Purpose: keystoreType.KeyPurposeP2PDocumentSigning,
			KeyType: keystoreType.KeyTypeECDSA,
		},
		{
			Key:     types.NewHash(utils.RandomSlice(32)),
			Purpose: keystoreType.KeyPurposeP2PDocumentSigning,
			KeyType: keystoreType.KeyTypeEDDSA,
		},
	}

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity).
		Once()

	ctx = contextutil.WithAccount(ctx, accountMock)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(nil, errors.New("error")).Once()

	res, err := api.AddKeys(ctx, keys)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_AddKeys_CallCreationError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	keys := []*keystoreType.AddKey{
		{
			Key:     types.NewHash(utils.RandomSlice(32)),
			Purpose: keystoreType.KeyPurposeP2PDocumentSigning,
			KeyType: keystoreType.KeyTypeECDSA,
		},
		{
			Key:     types.NewHash(utils.RandomSlice(32)),
			Purpose: keystoreType.KeyPurposeP2PDocumentSigning,
			KeyType: keystoreType.KeyTypeEDDSA,
		},
	}

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity).
		Once()

	ctx = contextutil.WithAccount(ctx, accountMock)

	var meta types.Metadata

	// NOTE - types.MetadataV14Data does not have info on the Keystore pallet,
	// causing types.NewCall to fail.
	err = codec.DecodeFromHex(types.MetadataV14Data, &meta)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(&meta, nil)

	res, err := api.AddKeys(ctx, keys)
	assert.ErrorIs(t, err, errors.ErrCallCreation)
	assert.Nil(t, res)
}

func TestAPI_AddKeys_ProxyCallError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	keys := []*keystoreType.AddKey{
		{
			Key:     types.NewHash(utils.RandomSlice(32)),
			Purpose: keystoreType.KeyPurposeP2PDocumentSigning,
			KeyType: keystoreType.KeyTypeECDSA,
		},
		{
			Key:     types.NewHash(utils.RandomSlice(32)),
			Purpose: keystoreType.KeyPurposeP2PDocumentSigning,
			KeyType: keystoreType.KeyTypeEDDSA,
		},
	}

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity).
		Once()

	ctx = contextutil.WithAccount(ctx, accountMock)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil).Once()

	var krp signature.KeyringPair

	genericUtils.GetMock[*config.PodOperatorMock](mocks).
		On("ToKeyringPair").
		Return(krp).Once()

	call, err := types.NewCall(meta, AddKeysCall, keys)
	assert.NoError(t, err)

	genericUtils.GetMock[*proxy.APIMock](mocks).On(
		"ProxyCall",
		ctx,
		identity,
		krp,
		types.NewOption(proxyType.KeystoreManagement),
		call,
	).Return(nil, errors.New("error")).Once()

	res, err := api.AddKeys(ctx, keys)
	assert.ErrorIs(t, err, errors.ErrProxyCall)
	assert.Nil(t, res)
}

func TestAPI_RevokeKeys(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	keyPurpose := keystoreType.KeyPurposeP2PDiscovery

	key1 := types.NewHash(utils.RandomSlice(32))
	key2 := types.NewHash(utils.RandomSlice(32))

	keys := []*types.Hash{
		&key1,
		&key2,
	}

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity).
		Once()

	ctx = contextutil.WithAccount(ctx, accountMock)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil).Once()

	var krp signature.KeyringPair

	genericUtils.GetMock[*config.PodOperatorMock](mocks).
		On("ToKeyringPair").
		Return(krp).Once()

	call, err := types.NewCall(meta, RevokeKeysCall, keys, keyPurpose)
	assert.NoError(t, err)

	extInfo := &centchain.ExtrinsicInfo{}

	genericUtils.GetMock[*proxy.APIMock](mocks).On(
		"ProxyCall",
		ctx,
		identity,
		krp,
		types.NewOption(proxyType.KeystoreManagement),
		call,
	).Return(extInfo, nil).Once()

	res, err := api.RevokeKeys(ctx, keys, keyPurpose)
	assert.NoError(t, err)
	assert.Equal(t, extInfo, res)
}

func TestAPI_RevokeKeys_ValidationError(t *testing.T) {
	ctx := context.Background()

	api, _ := getAPIWithMocks(t)

	keyPurpose := keystoreType.KeyPurposeP2PDiscovery

	res, err := api.RevokeKeys(ctx, nil, keyPurpose)
	assert.ErrorIs(t, err, ErrNoKeyHashesProvided)
	assert.Nil(t, res)

	keys := []*types.Hash{
		&emptyKeyHash,
		&emptyKeyHash,
	}

	res, err = api.RevokeKeys(ctx, keys, keyPurpose)
	assert.ErrorIs(t, err, ErrInvalidKey)
	assert.Nil(t, res)
}

func TestAPI_RevokeKeys_IdentityRetrievalError(t *testing.T) {
	ctx := context.Background()

	api, _ := getAPIWithMocks(t)

	keyPurpose := keystoreType.KeyPurposeP2PDiscovery

	key1 := types.NewHash(utils.RandomSlice(32))
	key2 := types.NewHash(utils.RandomSlice(32))

	keys := []*types.Hash{
		&key1,
		&key2,
	}

	res, err := api.RevokeKeys(ctx, keys, keyPurpose)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestAPI_RevokeKeys_MetadataRetrievalError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	keyPurpose := keystoreType.KeyPurposeP2PDiscovery

	key1 := types.NewHash(utils.RandomSlice(32))
	key2 := types.NewHash(utils.RandomSlice(32))

	keys := []*types.Hash{
		&key1,
		&key2,
	}

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity).
		Once()

	ctx = contextutil.WithAccount(ctx, accountMock)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(nil, errors.New("error")).Once()

	res, err := api.RevokeKeys(ctx, keys, keyPurpose)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_RevokeKeys_CallCreationError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	keyPurpose := keystoreType.KeyPurposeP2PDiscovery

	key1 := types.NewHash(utils.RandomSlice(32))
	key2 := types.NewHash(utils.RandomSlice(32))

	keys := []*types.Hash{
		&key1,
		&key2,
	}

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity).
		Once()

	ctx = contextutil.WithAccount(ctx, accountMock)

	var meta types.Metadata

	// NOTE - types.MetadataV14Data does not have info on the Keystore pallet,
	// causing types.NewCall to fail.
	err = codec.DecodeFromHex(types.MetadataV14Data, &meta)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(&meta, nil)

	res, err := api.RevokeKeys(ctx, keys, keyPurpose)
	assert.ErrorIs(t, err, errors.ErrCallCreation)
	assert.Nil(t, res)
}

func TestAPI_RevokeKeys_ProxyCallError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	keyPurpose := keystoreType.KeyPurposeP2PDiscovery

	key1 := types.NewHash(utils.RandomSlice(32))
	key2 := types.NewHash(utils.RandomSlice(32))

	keys := []*types.Hash{
		&key1,
		&key2,
	}

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity).
		Once()

	ctx = contextutil.WithAccount(ctx, accountMock)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil).Once()

	var krp signature.KeyringPair

	genericUtils.GetMock[*config.PodOperatorMock](mocks).
		On("ToKeyringPair").
		Return(krp).Once()

	call, err := types.NewCall(meta, RevokeKeysCall, keys, keyPurpose)
	assert.NoError(t, err)

	genericUtils.GetMock[*proxy.APIMock](mocks).On(
		"ProxyCall",
		ctx,
		identity,
		krp,
		types.NewOption(proxyType.KeystoreManagement),
		call,
	).Return(nil, errors.New("error")).Once()

	res, err := api.RevokeKeys(ctx, keys, keyPurpose)
	assert.ErrorIs(t, err, errors.ErrProxyCall)
	assert.Nil(t, res)
}

func TestAPI_GetKey(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(utils.RandomSlice(32)),
		KeyPurpose: keystoreType.KeyPurposeP2PDiscovery,
	}

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil).Once()

	encodedKeyID, err := codec.Encode(keyID)
	assert.NoError(t, err)

	encodedAccountID, err := codec.Encode(accountID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, KeysStorageName, encodedAccountID, encodedKeyID)
	assert.NoError(t, err)

	keyType := keystoreType.KeyTypeECDSA
	keyPurpose := keystoreType.KeyPurposeP2PDiscovery
	deposit := types.NewU128(*big.NewInt(rand.Int63()))
	revokedAt := types.NewOption[types.U32](types.U32(rand.Int()))

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&keystoreType.Key{})).
		Run(func(args mock.Arguments) {
			key := args.Get(1).(*keystoreType.Key)

			key.KeyType = keyType
			key.KeyPurpose = keyPurpose
			key.Deposit = deposit
			key.RevokedAt = revokedAt
		}).Return(true, nil).Once()

	res, err := api.GetKey(accountID, keyID)
	assert.NoError(t, err)
	assert.Equal(t, keyType, res.KeyType)
	assert.Equal(t, keyPurpose, res.KeyPurpose)
	assert.Equal(t, deposit, res.Deposit)
	assert.Equal(t, revokedAt, res.RevokedAt)
}

func TestAPI_GetKey_ValidationError(t *testing.T) {
	api, _ := getAPIWithMocks(t)

	res, err := api.GetKey(nil, nil)
	assert.ErrorIs(t, err, validation.ErrInvalidAccountID)
	assert.Nil(t, res)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	res, err = api.GetKey(accountID, nil)
	assert.ErrorIs(t, err, ErrInvalidKeyID)
	assert.Nil(t, res)

	keyID := &keystoreType.KeyID{
		Hash:       emptyKeyHash,
		KeyPurpose: keystoreType.KeyPurposeP2PDiscovery,
	}

	res, err = api.GetKey(accountID, keyID)
	assert.ErrorIs(t, err, ErrInvalidKeyIDHash)
	assert.Nil(t, res)
}

func TestAPI_GetKey_MetadataRetrievalError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(utils.RandomSlice(32)),
		KeyPurpose: keystoreType.KeyPurposeP2PDiscovery,
	}

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(nil, errors.New("error")).Once()

	res, err := api.GetKey(accountID, keyID)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_GetKey_StorageKeyCreationError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(utils.RandomSlice(32)),
		KeyPurpose: keystoreType.KeyPurposeP2PDiscovery,
	}

	var meta types.Metadata

	// NOTE - types.MetadataV14Data does not have info on the Keystore pallet,
	// causing types.CreateStorageKey to fail.
	err = codec.DecodeFromHex(types.MetadataV14Data, &meta)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(&meta, nil)

	res, err := api.GetKey(accountID, keyID)
	assert.ErrorIs(t, err, errors.ErrStorageKeyCreation)
	assert.Nil(t, res)
}

func TestAPI_GetKey_StorageRetrievalError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(utils.RandomSlice(32)),
		KeyPurpose: keystoreType.KeyPurposeP2PDiscovery,
	}

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil).Once()

	encodedKeyID, err := codec.Encode(keyID)
	assert.NoError(t, err)

	encodedAccountID, err := codec.Encode(accountID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, KeysStorageName, encodedAccountID, encodedKeyID)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&keystoreType.Key{})).
		Return(false, errors.New("error")).Once()

	res, err := api.GetKey(accountID, keyID)
	assert.ErrorIs(t, err, ErrKeyStorageRetrieval)
	assert.Nil(t, res)
}

func TestAPI_GetKey_KeyNotFoundError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	keyID := &keystoreType.KeyID{
		Hash:       types.NewHash(utils.RandomSlice(32)),
		KeyPurpose: keystoreType.KeyPurposeP2PDiscovery,
	}

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil).Once()

	encodedKeyID, err := codec.Encode(keyID)
	assert.NoError(t, err)

	encodedAccountID, err := codec.Encode(accountID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(meta, PalletName, KeysStorageName, encodedAccountID, encodedKeyID)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&keystoreType.Key{})).
		Return(false, nil).Once()

	res, err := api.GetKey(accountID, keyID)
	assert.ErrorIs(t, err, ErrKeyNotFound)
	assert.Nil(t, res)
}

func TestAPI_GetLastKeyByPurpose(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	keyPurpose := keystoreType.KeyPurposeP2PDiscovery

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil).Once()

	encodedKeyPurpose, err := codec.Encode(keyPurpose)
	assert.NoError(t, err)

	encodedAccountID, err := codec.Encode(accountID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(
		meta,
		PalletName,
		LastKeyByPurposeStorageName,
		encodedAccountID,
		encodedKeyPurpose,
	)
	assert.NoError(t, err)

	hash := types.NewHash(utils.RandomSlice(32))

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&types.Hash{})).
		Run(func(args mock.Arguments) {
			keyHash := args.Get(1).(*types.Hash)

			*keyHash = hash
		}).Return(true, nil).Once()

	res, err := api.GetLastKeyByPurpose(accountID, keyPurpose)
	assert.NoError(t, err)
	assert.Equal(t, hash, *res)
}

func TestAPI_GetLastKeyByPurpose_ValidationError(t *testing.T) {
	api, _ := getAPIWithMocks(t)

	keyPurpose := keystoreType.KeyPurposeP2PDiscovery

	res, err := api.GetLastKeyByPurpose(nil, keyPurpose)
	assert.ErrorIs(t, err, validation.ErrInvalidAccountID)
	assert.Nil(t, res)
}

func TestAPI_GetLastKeyByPurpose_MetadataRetrievalError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	keyPurpose := keystoreType.KeyPurposeP2PDiscovery

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(nil, errors.New("error")).Once()

	res, err := api.GetLastKeyByPurpose(accountID, keyPurpose)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_GetLastKeyByPurpose_StorageKeyCreationError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	keyPurpose := keystoreType.KeyPurposeP2PDiscovery

	var meta types.Metadata

	// NOTE - types.MetadataV14Data does not have info on the Keystore pallet,
	// causing types.CreateStorageKey to fail.
	err = codec.DecodeFromHex(types.MetadataV14Data, &meta)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(&meta, nil)

	res, err := api.GetLastKeyByPurpose(accountID, keyPurpose)
	assert.ErrorIs(t, err, errors.ErrStorageKeyCreation)
	assert.Nil(t, res)
}

func TestAPI_GetLastKeyByPurpose_StorageRetrievalError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	keyPurpose := keystoreType.KeyPurposeP2PDiscovery

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil).Once()

	encodedKeyPurpose, err := codec.Encode(keyPurpose)
	assert.NoError(t, err)

	encodedAccountID, err := codec.Encode(accountID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(
		meta,
		PalletName,
		LastKeyByPurposeStorageName,
		encodedAccountID,
		encodedKeyPurpose,
	)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&types.Hash{})).
		Return(false, errors.New("error")).Once()

	res, err := api.GetLastKeyByPurpose(accountID, keyPurpose)
	assert.ErrorIs(t, err, ErrKeyStorageRetrieval)
	assert.Nil(t, res)
}

func TestAPI_GetLastKeyByPurpose_KeyNotFoundError(t *testing.T) {
	api, mocks := getAPIWithMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	keyPurpose := keystoreType.KeyPurposeP2PDiscovery

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil).Once()

	encodedKeyPurpose, err := codec.Encode(keyPurpose)
	assert.NoError(t, err)

	encodedAccountID, err := codec.Encode(accountID)
	assert.NoError(t, err)

	storageKey, err := types.CreateStorageKey(
		meta,
		PalletName,
		LastKeyByPurposeStorageName,
		encodedAccountID,
		encodedKeyPurpose,
	)
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetStorageLatest", storageKey, mock.IsType(&types.Hash{})).
		Return(false, nil).Once()

	res, err := api.GetLastKeyByPurpose(accountID, keyPurpose)
	assert.ErrorIs(t, err, ErrLastKeyByPurposeNotFound)
	assert.Nil(t, res)
}

func getAPIWithMocks(t *testing.T) (*api, []any) {
	centAPIMock := centchain.NewAPIMock(t)
	proxyAPIMock := proxy.NewAPIMock(t)
	podOperatorMock := config.NewPodOperatorMock(t)

	API := NewAPI(centAPIMock, proxyAPIMock, podOperatorMock)

	return API.(*api), []any{
		centAPIMock,
		proxyAPIMock,
		podOperatorMock,
	}
}
