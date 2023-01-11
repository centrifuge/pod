//go:build unit

package centchain

import (
	"bytes"
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/testingutils"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApi_Call(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

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
	dispatcherMock := jobs.NewDispatcherMock(t)

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

func TestApi_SubmitExtrinsic(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

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
		Return(false, errors.New("failed to get nonce from storage")).
		Once()

	_, _, _, err = api.SubmitExtrinsic(ctx, meta, c, krp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get nonce from storage")

	// Irrecoverable failure to submit extrinsic
	substrateAPIMock.On("GetStorageLatest", storageKey, mock.IsType(&types.AccountInfo{})).
		Return(true, nil).
		Once()

	substrateAPIMock.On("GetBlockHash", uint64(0)).
		Return(types.Hash{}, errors.New("failed to get block hash")).
		Once()

	_, _, _, err = api.SubmitExtrinsic(ctx, meta, c, krp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get block hash")

	// Recoverable failure to submit extrinsic, max retries reached
	substrateAPIMock.On("GetStorageLatest", storageKey, mock.IsType(&types.AccountInfo{})).
		Return(true, nil).
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

	substrateAPIMock.On("GetBlockLatest", mock.Anything).
		Return(new(types.SignedBlock), nil)

	ext := types.NewExtrinsic(c)

	extrinsicHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("SubmitExtrinsic", mock.IsType(ext)).
		Run(func(args mock.Arguments) {
			callExt, ok := args.Get(0).(types.Extrinsic)
			assert.True(t, ok, "expected first arg to be types.Extrinsic")

			extVersion := ext.Version | types.ExtrinsicBitSigned

			assert.Equal(t, ext.Method, callExt.Method)
			assert.Equal(t, extVersion, callExt.Version)
		}).
		Return(extrinsicHash, nil)

	_, _, _, err = api.SubmitExtrinsic(ctx, meta, c, krp)
	assert.NoError(t, err)
}

func TestApi_SubmitAndWatch(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

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

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	krp := keyrings.AliceKeyRingPair

	accountInfoKey, err := types.CreateStorageKey(meta, "System", "Account", krp.PublicKey)
	assert.NoError(t, err)

	accountNonce := uint64(11)

	substrateAPIMock.On("GetStorageLatest", accountInfoKey, mock.IsType(&types.AccountInfo{})).
		Run(func(args mock.Arguments) {
			ai := args.Get(1).(*types.AccountInfo)
			ai.Nonce = types.U32(accountNonce)
		}).
		Return(true, nil).
		Once()

	genesisHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("GetBlockHash", uint64(0)).
		Return(genesisHash, nil).
		Once()

	runtimeVersion := types.NewRuntimeVersion()

	substrateAPIMock.On("GetRuntimeVersionLatest").
		Return(runtimeVersion, nil)

	substrateAPIMock.On("GetBlockLatest").
		Return(new(types.SignedBlock), nil)

	ext := types.NewExtrinsic(c)

	extrinsicHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("SubmitExtrinsic", mock.IsType(ext)).
		Run(func(args mock.Arguments) {
			callExt, ok := args.Get(0).(types.Extrinsic)
			assert.True(t, ok, "expected first arg to be types.Extrinsic")

			extVersion := ext.Version | types.ExtrinsicBitSigned

			assert.Equal(t, ext.Method, callExt.Method)
			assert.Equal(t, extVersion, callExt.Version)
		}).
		Return(extrinsicHash, nil)

	dispatcherTaskName := getTaskName(extrinsicHash)

	dispatcherMock.On("RegisterRunnerFunc", dispatcherTaskName, mock.Anything).
		Return(true)

	dispatcherResult := jobs.NewResultMock(t)

	dispatcherMock.On("Dispatch", accountID, mock.IsType(new(gocelery.Job))).
		Return(dispatcherResult, nil)

	extInfo := new(ExtrinsicInfo)

	dispatcherResult.On("Await", context.Background()).
		Return(*extInfo, nil)

	res, err := api.SubmitAndWatch(ctx, meta, c, krp)
	assert.NoError(t, err)
	assert.Equal(t, *extInfo, res)
}

func TestApi_SubmitAndWatch_IdentityRetrievalError(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

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

	ctx := context.Background()

	krp := keyrings.AliceKeyRingPair

	var extInfo ExtrinsicInfo

	res, err := api.SubmitAndWatch(ctx, meta, c, krp)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Equal(t, extInfo, res)
}

func TestApi_SubmitAndWatch_SubmitExtrinsicError(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

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

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	krp := keyrings.AliceKeyRingPair

	accountInfoKey, err := types.CreateStorageKey(meta, "System", "Account", krp.PublicKey)
	assert.NoError(t, err)

	substrateAPIMock.On("GetStorageLatest", accountInfoKey, mock.IsType(&types.AccountInfo{})).
		Return(false, errors.New("storage error")).
		Once()

	var extInfo ExtrinsicInfo

	res, err := api.SubmitAndWatch(ctx, meta, c, krp)
	assert.ErrorIs(t, err, ErrExtrinsicSubmission)
	assert.Equal(t, extInfo, res)
}

func TestApi_SubmitAndWatch_DispatcherError(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

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

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	krp := keyrings.AliceKeyRingPair

	accountInfoKey, err := types.CreateStorageKey(meta, "System", "Account", krp.PublicKey)
	assert.NoError(t, err)

	accountNonce := uint64(11)

	substrateAPIMock.On("GetStorageLatest", accountInfoKey, mock.IsType(&types.AccountInfo{})).
		Run(func(args mock.Arguments) {
			ai := args.Get(1).(*types.AccountInfo)
			ai.Nonce = types.U32(accountNonce)
		}).
		Return(true, nil).
		Once()

	genesisHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("GetBlockHash", uint64(0)).
		Return(genesisHash, nil).
		Once()

	runtimeVersion := types.NewRuntimeVersion()

	substrateAPIMock.On("GetRuntimeVersionLatest").
		Return(runtimeVersion, nil)

	substrateAPIMock.On("GetBlockLatest").
		Return(new(types.SignedBlock), nil)

	ext := types.NewExtrinsic(c)

	extrinsicHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("SubmitExtrinsic", mock.IsType(ext)).
		Run(func(args mock.Arguments) {
			callExt, ok := args.Get(0).(types.Extrinsic)
			assert.True(t, ok, "expected first arg to be types.Extrinsic")

			extVersion := ext.Version | types.ExtrinsicBitSigned

			assert.Equal(t, ext.Method, callExt.Method)
			assert.Equal(t, extVersion, callExt.Version)
		}).
		Return(extrinsicHash, nil)

	dispatcherTaskName := getTaskName(extrinsicHash)

	dispatcherMock.On("RegisterRunnerFunc", dispatcherTaskName, mock.Anything).
		Return(true)

	dispatcherMock.On("Dispatch", accountID, mock.IsType(new(gocelery.Job))).
		Return(nil, errors.New("dispatcher error"))

	var extInfo ExtrinsicInfo

	res, err := api.SubmitAndWatch(ctx, meta, c, krp)
	assert.Error(t, err)
	assert.Equal(t, extInfo, res)
}

func TestApi_SubmitAndWatch_DispatcherResultError(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

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

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	krp := keyrings.AliceKeyRingPair

	accountInfoKey, err := types.CreateStorageKey(meta, "System", "Account", krp.PublicKey)
	assert.NoError(t, err)

	accountNonce := uint64(11)

	substrateAPIMock.On("GetStorageLatest", accountInfoKey, mock.IsType(&types.AccountInfo{})).
		Run(func(args mock.Arguments) {
			ai := args.Get(1).(*types.AccountInfo)
			ai.Nonce = types.U32(accountNonce)
		}).
		Return(true, nil).
		Once()

	genesisHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("GetBlockHash", uint64(0)).
		Return(genesisHash, nil).
		Once()

	runtimeVersion := types.NewRuntimeVersion()

	substrateAPIMock.On("GetRuntimeVersionLatest").
		Return(runtimeVersion, nil)

	substrateAPIMock.On("GetBlockLatest").
		Return(new(types.SignedBlock), nil)

	ext := types.NewExtrinsic(c)

	extrinsicHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("SubmitExtrinsic", mock.IsType(ext)).
		Run(func(args mock.Arguments) {
			callExt, ok := args.Get(0).(types.Extrinsic)
			assert.True(t, ok, "expected first arg to be types.Extrinsic")

			extVersion := ext.Version | types.ExtrinsicBitSigned

			assert.Equal(t, ext.Method, callExt.Method)
			assert.Equal(t, extVersion, callExt.Version)
		}).
		Return(extrinsicHash, nil)

	dispatcherTaskName := getTaskName(extrinsicHash)

	dispatcherMock.On("RegisterRunnerFunc", dispatcherTaskName, mock.Anything).
		Return(true)

	dispatcherResult := jobs.NewResultMock(t)

	dispatcherMock.On("Dispatch", accountID, mock.IsType(new(gocelery.Job))).
		Return(dispatcherResult, nil)

	var extInfo ExtrinsicInfo

	dispatcherResult.On("Await", context.Background()).
		Return(extInfo, errors.New("dispatcher result error"))

	res, err := api.SubmitAndWatch(ctx, meta, c, krp)
	assert.Error(t, err)
	assert.Equal(t, extInfo, res)
}

func TestApi_GetStorageLatest(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

	api := NewAPI(substrateAPIMock, dispatcherMock, 1, 5*time.Second)

	meta := types.NewMetadataV14()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountStorageKey, err := types.CreateStorageKey(meta, "System", "Account", accountID.ToBytes())

	var accountInfo types.AccountInfo

	substrateAPIMock.On("GetStorageLatest", accountStorageKey, accountInfo).
		Return(true, nil).
		Once()

	ok, err := api.GetStorageLatest(accountStorageKey, accountInfo)
	assert.NoError(t, err)
	assert.True(t, ok)

	apiErr := errors.New("api error")

	substrateAPIMock.On("GetStorageLatest", accountStorageKey, accountInfo).
		Return(false, apiErr).
		Once()

	ok, err = api.GetStorageLatest(accountStorageKey, accountInfo)
	assert.ErrorIs(t, err, apiErr)
	assert.False(t, ok)
}

func TestApi_GetBlockLatest(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

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

func TestApi_GetPendingExtrinsics(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

	api := NewAPI(substrateAPIMock, dispatcherMock, 1, 5*time.Second)

	pendingExtrinsics := []types.Extrinsic{
		{
			Version:   0,
			Signature: types.ExtrinsicSignatureV4{},
			Method:    types.Call{},
		},
		{
			Version:   1,
			Signature: types.ExtrinsicSignatureV4{},
			Method:    types.Call{},
		},
	}

	substrateAPIMock.On("GetPendingExtrinsics").
		Return(pendingExtrinsics, nil).Once()

	res, err := api.GetPendingExtrinsics()
	assert.NoError(t, err)
	assert.Equal(t, pendingExtrinsics, res)

	pendingExtrinsicsError := errors.New("error")

	substrateAPIMock.On("GetPendingExtrinsics").
		Return(nil, pendingExtrinsicsError).Once()

	res, err = api.GetPendingExtrinsics()
	assert.ErrorIs(t, err, pendingExtrinsicsError)
	assert.Nil(t, res)
}

func TestApi_dispatcherRunnerFunc(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

	centApi := NewAPI(substrateAPIMock, dispatcherMock, 3, 1*time.Second)
	a := centApi.(*api)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	blockNumber := types.BlockNumber(11)
	txHash := types.NewHash(utils.RandomSlice(32))
	sig := types.NewSignature(utils.RandomSlice(64))

	fn := a.getDispatcherRunnerFunc(&blockNumber, txHash, sig, meta)

	blockHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("GetBlockHash", uint64(blockNumber)).
		Return(blockHash, nil)

	block := &types.SignedBlock{
		Block: types.Block{
			Extrinsics: []types.Extrinsic{
				{
					Signature: types.ExtrinsicSignatureV4{
						Signature: types.MultiSignature{
							IsSr25519: true,
							AsSr25519: sig,
						},
					},
				},
			},
		},
	}

	substrateAPIMock.On("GetBlock", blockHash).
		Return(block, nil)

	eventsStorageKey, err := types.CreateStorageKey(meta, "System", "Events")
	assert.NoError(t, err)

	extInfo := ExtrinsicInfo{
		Hash:      txHash,
		BlockHash: blockHash,
		Index:     0, // Index of the above signature.
	}

	substrateAPIMock.On("GetStorage", eventsStorageKey, mock.IsType(new(types.EventRecordsRaw)), blockHash).
		Run(func(args mock.Arguments) {
			rawEvents := args.Get(1).(*types.EventRecordsRaw)

			var b []byte
			buf := bytes.NewBuffer(b)

			enc := scale.NewEncoder(buf)

			// Push number of events
			err = enc.EncodeUintCompact(*big.NewInt(1))
			assert.NoError(t, err)

			err = enc.Encode(types.Phase{
				IsApplyExtrinsic: true,
				AsApplyExtrinsic: 0, // Index of the above signature.
			})
			assert.NoError(t, err)

			// Extrinsic success event ID

			err = enc.Encode(types.EventID{0, 0})
			assert.NoError(t, err)

			err = enc.Encode(types.DispatchInfo{
				Weight: types.NewWeight(types.NewUCompactFromUInt(123), types.NewUCompactFromUInt(345)),
				Class: types.DispatchClass{
					IsNormal: true,
				},
				PaysFee: types.Pays{
					IsYes: true,
				},
			})
			assert.NoError(t, err)

			err = enc.Encode([]types.Hash{
				types.NewHash(utils.RandomSlice(32)),
			})
			assert.NoError(t, err)

			encodedEvents := buf.Bytes()

			*rawEvents = encodedEvents
			extInfo.EventsRaw = encodedEvents
		}).
		Return(nil)

	res, err := fn(nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, extInfo, res)
}

func TestApi_dispatcherRunnerFunc_ed25519Signature(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

	centApi := NewAPI(substrateAPIMock, dispatcherMock, 3, 1*time.Second)
	a := centApi.(*api)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	blockNumber := types.BlockNumber(11)
	txHash := types.NewHash(utils.RandomSlice(32))
	sig := types.NewSignature(utils.RandomSlice(64))

	fn := a.getDispatcherRunnerFunc(&blockNumber, txHash, sig, meta)

	blockHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("GetBlockHash", uint64(blockNumber)).
		Return(blockHash, nil)

	block := &types.SignedBlock{
		Block: types.Block{
			Extrinsics: []types.Extrinsic{
				{
					Signature: types.ExtrinsicSignatureV4{
						Signature: types.MultiSignature{
							IsEd25519: true,
							AsEd25519: sig,
						},
					},
				},
			},
		},
	}

	substrateAPIMock.On("GetBlock", blockHash).
		Return(block, nil)

	eventsStorageKey, err := types.CreateStorageKey(meta, "System", "Events")
	assert.NoError(t, err)

	extInfo := ExtrinsicInfo{
		Hash:      txHash,
		BlockHash: blockHash,
		Index:     0, // Index of the above signature.
	}

	substrateAPIMock.On("GetStorage", eventsStorageKey, mock.IsType(new(types.EventRecordsRaw)), blockHash).
		Run(func(args mock.Arguments) {
			rawEvents := args.Get(1).(*types.EventRecordsRaw)

			var b []byte
			buf := bytes.NewBuffer(b)

			enc := scale.NewEncoder(buf)

			// Push number of events
			err = enc.EncodeUintCompact(*big.NewInt(1))
			assert.NoError(t, err)

			err = enc.Encode(types.Phase{
				IsApplyExtrinsic: true,
				AsApplyExtrinsic: 0, // Index of the above signature.
			})
			assert.NoError(t, err)

			// Extrinsic success event ID

			err = enc.Encode(types.EventID{0, 0})
			assert.NoError(t, err)

			err = enc.Encode(types.DispatchInfo{
				Weight: types.NewWeight(types.NewUCompactFromUInt(123), types.NewUCompactFromUInt(345)),
				Class: types.DispatchClass{
					IsNormal: true,
				},
				PaysFee: types.Pays{
					IsYes: true,
				},
			})
			assert.NoError(t, err)

			err = enc.Encode([]types.Hash{
				types.NewHash(utils.RandomSlice(32)),
			})
			assert.NoError(t, err)

			encodedEvents := buf.Bytes()

			*rawEvents = encodedEvents
			extInfo.EventsRaw = encodedEvents
		}).
		Return(nil)

	res, err := fn(nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, extInfo, res)
}

func TestApi_dispatcherRunnerFunc_BlockHashError(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

	centApi := NewAPI(substrateAPIMock, dispatcherMock, 3, 1*time.Second)
	a := centApi.(*api)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	blockNumber := types.BlockNumber(11)
	txHash := types.NewHash(utils.RandomSlice(32))
	sig := types.NewSignature(utils.RandomSlice(64))

	fn := a.getDispatcherRunnerFunc(&blockNumber, txHash, sig, meta)

	substrateAPIMock.On("GetBlockHash", uint64(blockNumber)).
		Return(nil, errors.New("error"))

	res, err := fn(nil, nil)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestApi_dispatcherRunnerFunc_BlockError(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

	centApi := NewAPI(substrateAPIMock, dispatcherMock, 3, 1*time.Second)
	a := centApi.(*api)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	blockNumber := types.BlockNumber(11)
	txHash := types.NewHash(utils.RandomSlice(32))
	sig := types.NewSignature(utils.RandomSlice(64))

	fn := a.getDispatcherRunnerFunc(&blockNumber, txHash, sig, meta)

	blockHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("GetBlockHash", uint64(blockNumber)).
		Return(blockHash, nil)

	substrateAPIMock.On("GetBlock", blockHash).
		Return(nil, errors.New("error"))

	res, err := fn(nil, nil)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestApi_dispatcherRunnerFunc_NoSignature(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

	centApi := NewAPI(substrateAPIMock, dispatcherMock, 3, 1*time.Second)
	a := centApi.(*api)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	blockNumber := types.BlockNumber(11)
	txHash := types.NewHash(utils.RandomSlice(32))
	sig := types.NewSignature(utils.RandomSlice(64))
	invalidSig := types.NewSignature(utils.RandomSlice(64))

	fn := a.getDispatcherRunnerFunc(&blockNumber, txHash, sig, meta)

	blockHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("GetBlockHash", uint64(blockNumber)).
		Return(blockHash, nil)

	block := &types.SignedBlock{
		Block: types.Block{
			Extrinsics: []types.Extrinsic{
				{
					Signature: types.ExtrinsicSignatureV4{
						Signature: types.MultiSignature{
							IsSr25519: true,
							AsSr25519: invalidSig,
						},
					},
				},
			},
		},
	}

	substrateAPIMock.On("GetBlock", blockHash).
		Return(block, nil)

	res, err := fn(nil, nil)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestApi_dispatcherRunnerFunc_EventStorageError(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

	centApi := NewAPI(substrateAPIMock, dispatcherMock, 3, 1*time.Second)
	a := centApi.(*api)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	blockNumber := types.BlockNumber(11)
	txHash := types.NewHash(utils.RandomSlice(32))
	sig := types.NewSignature(utils.RandomSlice(64))

	fn := a.getDispatcherRunnerFunc(&blockNumber, txHash, sig, meta)

	blockHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("GetBlockHash", uint64(blockNumber)).
		Return(blockHash, nil)

	block := &types.SignedBlock{
		Block: types.Block{
			Extrinsics: []types.Extrinsic{
				{
					Signature: types.ExtrinsicSignatureV4{
						Signature: types.MultiSignature{
							IsSr25519: true,
							AsSr25519: sig,
						},
					},
				},
			},
		},
	}

	substrateAPIMock.On("GetBlock", blockHash).
		Return(block, nil)

	eventsStorageKey, err := types.CreateStorageKey(meta, "System", "Events")
	assert.NoError(t, err)

	substrateAPIMock.On("GetStorage", eventsStorageKey, mock.IsType(new(types.EventRecordsRaw)), blockHash).
		Return(errors.New("error"))

	res, err := fn(nil, nil)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestApi_dispatcherRunnerFunc_EventDecodeError(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

	centApi := NewAPI(substrateAPIMock, dispatcherMock, 3, 1*time.Second)
	a := centApi.(*api)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	blockNumber := types.BlockNumber(11)
	txHash := types.NewHash(utils.RandomSlice(32))
	sig := types.NewSignature(utils.RandomSlice(64))

	fn := a.getDispatcherRunnerFunc(&blockNumber, txHash, sig, meta)

	blockHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("GetBlockHash", uint64(blockNumber)).
		Return(blockHash, nil)

	block := &types.SignedBlock{
		Block: types.Block{
			Extrinsics: []types.Extrinsic{
				{
					Signature: types.ExtrinsicSignatureV4{
						Signature: types.MultiSignature{
							IsSr25519: true,
							AsSr25519: sig,
						},
					},
				},
			},
		},
	}

	substrateAPIMock.On("GetBlock", blockHash).
		Return(block, nil)

	eventsStorageKey, err := types.CreateStorageKey(meta, "System", "Events")
	assert.NoError(t, err)

	substrateAPIMock.On("GetStorage", eventsStorageKey, mock.IsType(new(types.EventRecordsRaw)), blockHash).
		Run(func(args mock.Arguments) {
			rawEvents := args.Get(1).(*types.EventRecordsRaw)

			var b []byte
			buf := bytes.NewBuffer(b)

			enc := scale.NewEncoder(buf)

			// Don't push number of events which will cause a decoding error

			//err = enc.EncodeUintCompact(*big.NewInt(1))
			//assert.NoError(t, err)

			err = enc.Encode(types.Phase{
				IsApplyExtrinsic: true,
				AsApplyExtrinsic: 0, // Index of the above signature.
			})
			assert.NoError(t, err)

			// Extrinsic success event ID

			err = enc.Encode(types.EventID{0, 0})
			assert.NoError(t, err)

			err = enc.Encode(types.DispatchInfo{
				Weight: types.NewWeight(types.NewUCompactFromUInt(123), types.NewUCompactFromUInt(345)),
				Class: types.DispatchClass{
					IsNormal: true,
				},
				PaysFee: types.Pays{
					IsYes: true,
				},
			})
			assert.NoError(t, err)

			err = enc.Encode([]types.Hash{
				types.NewHash(utils.RandomSlice(32)),
			})
			assert.NoError(t, err)

			encodedEvents := buf.Bytes()

			*rawEvents = encodedEvents
		}).
		Return(nil)

	res, err := fn(nil, nil)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestApi_dispatcherRunnerFunc_FailedExtrinsic(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

	centApi := NewAPI(substrateAPIMock, dispatcherMock, 3, 1*time.Second)
	a := centApi.(*api)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	blockNumber := types.BlockNumber(11)
	txHash := types.NewHash(utils.RandomSlice(32))
	sig := types.NewSignature(utils.RandomSlice(64))

	fn := a.getDispatcherRunnerFunc(&blockNumber, txHash, sig, meta)

	blockHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("GetBlockHash", uint64(blockNumber)).
		Return(blockHash, nil)

	block := &types.SignedBlock{
		Block: types.Block{
			Extrinsics: []types.Extrinsic{
				{
					Signature: types.ExtrinsicSignatureV4{
						Signature: types.MultiSignature{
							IsSr25519: true,
							AsSr25519: sig,
						},
					},
				},
			},
		},
	}

	substrateAPIMock.On("GetBlock", blockHash).
		Return(block, nil)

	eventsStorageKey, err := types.CreateStorageKey(meta, "System", "Events")
	assert.NoError(t, err)

	substrateAPIMock.On("GetStorage", eventsStorageKey, mock.IsType(new(types.EventRecordsRaw)), blockHash).
		Run(func(args mock.Arguments) {
			rawEvents := args.Get(1).(*types.EventRecordsRaw)

			var b []byte
			buf := bytes.NewBuffer(b)

			enc := scale.NewEncoder(buf)

			// Push number of events
			err = enc.EncodeUintCompact(*big.NewInt(1))
			assert.NoError(t, err)

			err = enc.Encode(types.Phase{
				IsApplyExtrinsic: true,
				AsApplyExtrinsic: 0, // Index of the above signature.
			})
			assert.NoError(t, err)

			// Extrinsic failed event ID

			err = enc.Encode(types.EventID{0, 1})
			assert.NoError(t, err)

			err = enc.Encode(types.DispatchError{
				IsBadOrigin: true,
			})
			assert.NoError(t, err)

			err = enc.Encode(types.DispatchInfo{
				Weight: types.NewWeight(types.NewUCompactFromUInt(123), types.NewUCompactFromUInt(345)),
				Class: types.DispatchClass{
					IsNormal: true,
				},
				PaysFee: types.Pays{
					IsYes: true,
				},
			})
			assert.NoError(t, err)

			err = enc.Encode([]types.Hash{
				types.NewHash(utils.RandomSlice(32)),
			})
			assert.NoError(t, err)

			encodedEvents := buf.Bytes()

			*rawEvents = encodedEvents
		}).
		Return(nil)

	res, err := fn(nil, nil)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestApi_dispatcherRunnerFunc_NoEvents(t *testing.T) {
	substrateAPIMock := NewSubstrateAPIMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)

	centApi := NewAPI(substrateAPIMock, dispatcherMock, 3, 1*time.Second)
	a := centApi.(*api)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	blockNumber := types.BlockNumber(11)
	txHash := types.NewHash(utils.RandomSlice(32))
	sig := types.NewSignature(utils.RandomSlice(64))

	fn := a.getDispatcherRunnerFunc(&blockNumber, txHash, sig, meta)

	blockHash := types.NewHash(utils.RandomSlice(32))

	substrateAPIMock.On("GetBlockHash", uint64(blockNumber)).
		Return(blockHash, nil)

	block := &types.SignedBlock{
		Block: types.Block{
			Extrinsics: []types.Extrinsic{
				{
					Signature: types.ExtrinsicSignatureV4{
						Signature: types.MultiSignature{
							IsSr25519: true,
							AsSr25519: sig,
						},
					},
				},
			},
		},
	}

	substrateAPIMock.On("GetBlock", blockHash).
		Return(block, nil)

	eventsStorageKey, err := types.CreateStorageKey(meta, "System", "Events")
	assert.NoError(t, err)

	substrateAPIMock.On("GetStorage", eventsStorageKey, mock.IsType(new(types.EventRecordsRaw)), blockHash).
		Run(func(args mock.Arguments) {
			rawEvents := args.Get(1).(*types.EventRecordsRaw)

			var b []byte
			buf := bytes.NewBuffer(b)

			enc := scale.NewEncoder(buf)

			// Push number of events
			err = enc.EncodeUintCompact(*big.NewInt(0))
			assert.NoError(t, err)

			encodedEvents := buf.Bytes()

			*rawEvents = encodedEvents
		}).
		Return(nil)

	res, err := fn(nil, nil)
	assert.Error(t, err)
	assert.Nil(t, res)
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
