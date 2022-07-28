//go:build unit

package centchain

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"

	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"

	dispatcherMock "github.com/centrifuge/go-centrifuge/jobs"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApi_Call(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := dispatcherMock.NewDispatcherMock(t)

	api := NewAPI(substrateAPIMock, dispatcherMock, 1, 5*time.Second)

	result := types.AccountInfo{}
	method := "some_method"
	args := []interface{}{1, 2, 3}

	substrateAPIMock.On("Call", result, method, args).
		Return(nil).
		Once()

	err := api.Call(result, method, args)
	assert.NoError(t, err)

	apiErr := errors.New("api error")

	substrateAPIMock.On("Call", result, method, args).
		Return(apiErr)

	err = api.Call(result, method, args)
	assert.ErrorIs(t, err, apiErr)
}

func TestApi_GetMetadataLatest(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := dispatcherMock.NewDispatcherMock(t)

	api := NewAPI(substrateAPIMock, dispatcherMock, 1, 5*time.Second)

	substrateAPIMock.On("GetMetadataLatest").
		Return(types.NewMetadataV14(), nil).
		Once()

	meta, err := api.GetMetadataLatest()
	assert.NoError(t, err)
	assert.Equal(t, types.NewMetadataV14(), meta)

	apiErr := errors.New("api error")

	substrateAPIMock.On("GetMetadataLatest").Return(nil, apiErr)

	meta, err = api.GetMetadataLatest()
	assert.Nil(t, meta)
	assert.ErrorIs(t, err, apiErr)
}

func TestApi_GetStorageLatest(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := dispatcherMock.NewDispatcherMock(t)

	api := NewAPI(substrateAPIMock, dispatcherMock, 1, 5*time.Second)

	meta := types.NewMetadataV14()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountStorageKey, err := types.CreateStorageKey(meta, "System", "Account", accountID.ToBytes())

	var accountInfo types.AccountInfo

	substrateAPIMock.On("GetStorageLatest", accountStorageKey, accountInfo).
		Return(nil).
		Once()

	err = api.GetStorageLatest(accountStorageKey, accountInfo)
	assert.NoError(t, err)

	apiErr := errors.New("api error")

	substrateAPIMock.On("GetStorageLatest", accountStorageKey, accountInfo).
		Return(apiErr).
		Once()

	err = api.GetStorageLatest(accountStorageKey, accountInfo)
	assert.ErrorIs(t, err, apiErr)
}

func TestApi_GetBlockLatest(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := dispatcherMock.NewDispatcherMock(t)

	api := NewAPI(substrateAPIMock, dispatcherMock, 1, 5*time.Second)

	testBlock := &types.SignedBlock{}

	substrateAPIMock.On("GetBlockLatest").
		Return(testBlock, nil).
		Once()

	block, err := api.GetBlockLatest()
	assert.NoError(t, err)
	assert.Equal(t, testBlock, block)

	apiErr := errors.New("api error")

	substrateAPIMock.On("GetBlockLatest").
		Return(nil, apiErr).
		Once()

	block, err = api.GetBlockLatest()
	assert.ErrorIs(t, err, apiErr)
	assert.Nil(t, block)
}

func TestApi_SubmitExtrinsic(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := dispatcherMock.NewDispatcherMock(t)

	api := NewAPI(substrateAPIMock, dispatcherMock, 3, 1*time.Second)

	meta := metaDataWithCall("Anchor.commit")
	c, err := types.NewCall(
		meta,
		"Anchor.commit",
		types.NewHash(utils.RandomSlice(32)),
		types.NewHash(utils.RandomSlice(32)),
		types.NewHash(utils.RandomSlice(32)),
		types.NewMoment(time.Now()),
	)

	assert.NoError(t, err)

	krp := keyrings.AliceKeyRingPair

	storageKey, err := types.CreateStorageKey(meta, "System", "Account", krp.PublicKey)
	assert.NoError(t, err)

	// Failed to get nonce from chain
	ctx := context.Background()
	substrateAPIMock.On("GetStorageLatest", storageKey, mock.IsType(&types.AccountInfo{})).
		Return(errors.New("failed to get nonce from storage")).
		Once()

	_, _, _, err = api.SubmitExtrinsic(ctx, meta, c, krp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get nonce from storage")

	// Irrecoverable failure to submit extrinsic
	substrateAPIMock.On("GetStorageLatest", storageKey, mock.IsType(&types.AccountInfo{})).
		Return(nil).
		Once()

	substrateAPIMock.On("GetBlockHash", uint64(0)).
		Return(types.Hash{}, errors.New("failed to get block hash")).
		Once()

	_, _, _, err = api.SubmitExtrinsic(ctx, meta, c, krp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get block hash")

	// Recoverable failure to submit extrinsic, max retries reached
	substrateAPIMock.On("GetStorageLatest", storageKey, mock.IsType(&types.AccountInfo{})).
		Return(nil).
		Times(3)

	substrateAPIMock.On("GetBlockHash", uint64(0)).
		Return(types.Hash{}, ErrNonceTooLow).
		Times(3)

	_, _, _, err = api.SubmitExtrinsic(ctx, meta, c, krp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max concurrent transaction tries reached")

	// Success
	substrateAPIMock.On("GetBlockHash", mock.Anything).
		Return(types.Hash(utils.RandomByte32()), nil).
		Once()

	substrateAPIMock.On("GetRuntimeVersionLatest").
		Return(types.NewRuntimeVersion(), nil)

	clientMock := NewClientMock(t)

	substrateAPIMock.On("GetClient").
		Return(clientMock)

	substrateAPIMock.On("GetBlockLatest", mock.Anything).
		Return(new(types.SignedBlock), nil)

	clientMock.On("Call", mock.Anything, mock.Anything, mock.Anything).
		Return(hexutil.Encode(utils.RandomSlice(32)), nil)

	_, _, _, err = api.SubmitExtrinsic(ctx, meta, c, krp)
	assert.NoError(t, err)
}

func TestApi_SubmitAndWatch(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := dispatcherMock.NewDispatcherMock(t)

	api := NewAPI(substrateAPIMock, dispatcherMock, 3, 1*time.Second)

	meta := metaDataWithCall("Anchor.commit")
	c, err := types.NewCall(
		meta,
		"Anchor.commit",
		types.NewHash(utils.RandomSlice(32)),
		types.NewHash(utils.RandomSlice(32)),
		types.NewHash(utils.RandomSlice(32)),
		types.NewMoment(time.Now()),
	)

	assert.NoError(t, err)

	krp := keyrings.AliceKeyRingPair
}

func metaDataWithCall(call string) *types.Metadata {
	data := strings.Split(call, ".")
	meta := types.NewMetadataV8()
	meta.AsMetadataV8.Modules = []types.ModuleMetadataV8{
		{
			Name:       "System",
			HasStorage: true,
			Storage: types.StorageMetadata{
				Prefix: "System",
				Items: []types.StorageFunctionMetadataV5{
					{
						Name: "Account",
						Type: types.StorageFunctionTypeV5{
							IsMap: true,
							AsMap: types.MapTypeV4{
								Hasher: types.StorageHasher{IsBlake2_256: true},
							},
						},
					},
					{
						Name: "Events",
						Type: types.StorageFunctionTypeV5{
							IsMap: true,
							AsMap: types.MapTypeV4{
								Hasher: types.StorageHasher{IsBlake2_256: true},
							},
						},
					},
				},
			},
			HasEvents: true,
			Events: []types.EventMetadataV4{
				{
					Name: "ExtrinsicSuccess",
				},
				{
					Name: "ExtrinsicFailed",
				},
			},
		},
		{
			Name:       types.Text(data[0]),
			HasStorage: true,
			Storage: types.StorageMetadata{
				Prefix: types.Text(data[0]),
				Items: []types.StorageFunctionMetadataV5{
					{
						Name: "Events",
						Type: types.StorageFunctionTypeV5{
							IsMap: true,
							AsMap: types.MapTypeV4{
								Hasher: types.StorageHasher{IsBlake2_256: true},
							},
						},
					},
				},
			},
			HasCalls: true,
			Calls: []types.FunctionMetadataV4{{
				Name: types.Text(data[1]),
			}},
		},
	}
	return meta
}
