//go:build unit

package proxy

import (
	"context"
	"math/big"
	"math/rand"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/stretchr/testify/mock"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/validation"

	"github.com/centrifuge/go-centrifuge/utils"

	proxyTypes "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/testingutils"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func TestAPI_AddProxy(t *testing.T) {
	ctx := context.Background()
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	delegate, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	proxyType := proxyTypes.Any
	delay := types.U32(rand.Uint32())
	krp := keyrings.AliceKeyRingPair

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil).Once()

	call, err := types.NewCall(
		meta,
		ProxyAdd,
		delegate,
		proxyType,
		delay,
	)
	assert.NoError(t, err)

	txHash := types.NewHash(utils.RandomSlice(32))
	blockNumber := types.BlockNumber(rand.Int())
	multiSig := types.MultiSignature{}

	centAPIMock.On(
		"SubmitExtrinsic",
		ctx,
		meta,
		call,
		krp,
	).Return(txHash, blockNumber, multiSig, nil).Once()

	err = api.AddProxy(ctx, delegate, proxyType, delay, krp)
	assert.NoError(t, err)
}

func TestAPI_AddProxy_ValidationError(t *testing.T) {
	ctx := context.Background()
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	proxyType := proxyTypes.Any
	delay := types.U32(rand.Uint32())
	krp := keyrings.AliceKeyRingPair

	err := api.AddProxy(ctx, nil, proxyType, delay, krp)
	assert.ErrorIs(t, err, validation.ErrInvalidAccountID)
}

func TestAPI_AddProxy_MetadataRetrievalError(t *testing.T) {
	ctx := context.Background()
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	delegate, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	proxyType := proxyTypes.Any
	delay := types.U32(rand.Uint32())
	krp := keyrings.AliceKeyRingPair

	centAPIMock.On("GetMetadataLatest").
		Return(nil, errors.New("error")).Once()

	err = api.AddProxy(ctx, delegate, proxyType, delay, krp)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
}

func TestAPI_AddProxy_CallCreationError(t *testing.T) {
	ctx := context.Background()
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	delegate, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	proxyType := proxyTypes.Any
	delay := types.U32(rand.Uint32())
	krp := keyrings.AliceKeyRingPair

	// NOTE - types.MetadataV13 does not have info on the Proxy pallet,
	// causing types.NewCall to fail.
	meta := types.NewMetadataV13()

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil)

	err = api.AddProxy(ctx, delegate, proxyType, delay, krp)
	assert.ErrorIs(t, err, errors.ErrCallCreation)
}

func TestAPI_AddProxy_SubmitExtrinsicError(t *testing.T) {
	ctx := context.Background()
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	delegate, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	proxyType := proxyTypes.Any
	delay := types.U32(rand.Uint32())
	krp := keyrings.AliceKeyRingPair

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil).Once()

	call, err := types.NewCall(
		meta,
		ProxyAdd,
		delegate,
		proxyType,
		delay,
	)
	assert.NoError(t, err)

	txHash := types.NewHash(utils.RandomSlice(32))
	blockNumber := types.BlockNumber(rand.Int())
	multiSig := types.MultiSignature{}

	centAPIMock.On(
		"SubmitExtrinsic",
		ctx,
		meta,
		call,
		krp,
	).Return(txHash, blockNumber, multiSig, errors.New("error")).Once()

	err = api.AddProxy(ctx, delegate, proxyType, delay, krp)
	assert.ErrorIs(t, err, errors.ErrExtrinsicSubmission)
}

func TestAPI_ProxyCall(t *testing.T) {
	ctx := context.Background()
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	delegator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	proxyKeyringPair := keyrings.AliceKeyRingPair

	forcedProxyType := types.NewOption[proxyTypes.CentrifugeProxyType](proxyTypes.Any)

	proxiedCall := types.Call{}

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil).Once()

	call, err := types.NewCall(
		meta,
		ProxyCall,
		delegator,
		forcedProxyType,
		proxiedCall,
	)

	assert.NoError(t, err)

	extInfo := centchain.ExtrinsicInfo{}

	centAPIMock.On(
		"SubmitAndWatch",
		ctx,
		meta,
		call,
		proxyKeyringPair,
	).Return(extInfo, nil).Once()

	res, err := api.ProxyCall(ctx, delegator, proxyKeyringPair, forcedProxyType, proxiedCall)
	assert.NoError(t, err)
	assert.Equal(t, extInfo, *res)
}

func TestAPI_ProxyCall_ValidationError(t *testing.T) {
	ctx := context.Background()
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	proxyKeyringPair := keyrings.AliceKeyRingPair

	forcedProxyType := types.NewOption[proxyTypes.CentrifugeProxyType](proxyTypes.Any)

	proxiedCall := types.Call{}

	res, err := api.ProxyCall(ctx, nil, proxyKeyringPair, forcedProxyType, proxiedCall)
	assert.ErrorIs(t, err, validation.ErrInvalidAccountID)
	assert.Nil(t, res)
}

func TestAPI_ProxyCall_MetadataRetrievalError(t *testing.T) {
	ctx := context.Background()
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	delegator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	proxyKeyringPair := keyrings.AliceKeyRingPair

	forcedProxyType := types.NewOption[proxyTypes.CentrifugeProxyType](proxyTypes.Any)

	proxiedCall := types.Call{}

	centAPIMock.On("GetMetadataLatest").
		Return(nil, errors.New("error")).Once()

	res, err := api.ProxyCall(ctx, delegator, proxyKeyringPair, forcedProxyType, proxiedCall)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_ProxyCall_CallCreationError(t *testing.T) {
	ctx := context.Background()
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	delegator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	proxyKeyringPair := keyrings.AliceKeyRingPair

	forcedProxyType := types.NewOption[proxyTypes.CentrifugeProxyType](proxyTypes.Any)

	proxiedCall := types.Call{}

	// NOTE - types.MetadataV13 does not have info on the Proxy pallet,
	// causing types.NewCall to fail.
	meta := types.NewMetadataV13()

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil)

	res, err := api.ProxyCall(ctx, delegator, proxyKeyringPair, forcedProxyType, proxiedCall)
	assert.ErrorIs(t, err, errors.ErrCallCreation)
	assert.Nil(t, res)
}

func TestAPI_ProxyCall_SubmitAndWatchError(t *testing.T) {
	ctx := context.Background()
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	delegator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	proxyKeyringPair := keyrings.AliceKeyRingPair

	forcedProxyType := types.NewOption[proxyTypes.CentrifugeProxyType](proxyTypes.Any)

	proxiedCall := types.Call{}

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil).Once()

	call, err := types.NewCall(
		meta,
		ProxyCall,
		delegator,
		forcedProxyType,
		proxiedCall,
	)

	assert.NoError(t, err)

	extInfo := centchain.ExtrinsicInfo{}

	centAPIMock.On(
		"SubmitAndWatch",
		ctx,
		meta,
		call,
		proxyKeyringPair,
	).Return(extInfo, errors.New("error")).Once()

	res, err := api.ProxyCall(ctx, delegator, proxyKeyringPair, forcedProxyType, proxiedCall)
	assert.ErrorIs(t, err, errors.ErrExtrinsicSubmitAndWatch)
	assert.Nil(t, res)
}

func TestAPI_GetProxies(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	encodedAccountID, err := codec.Encode(accountID)
	assert.NoError(t, err)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil).Once()

	storageKey, err := types.CreateStorageKey(meta, PalletName, ProxiesStorageName, encodedAccountID)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	balance := types.NewU128(*big.NewInt(rand.Int63()))
	proxyDefinitions := []types.ProxyDefinition{
		{
			Delegate:  *accountID1,
			ProxyType: types.U8(proxyTypes.Any),
			Delay:     types.U32(rand.Int()),
		},
	}

	centAPIMock.On("GetStorageLatest", storageKey, mock.IsType(&types.ProxyStorageEntry{})).
		Run(func(args mock.Arguments) {
			storageEntry := args.Get(1).(*types.ProxyStorageEntry)

			storageEntry.Balance = balance
			storageEntry.ProxyDefinitions = proxyDefinitions
		}).Return(true, nil).Once()

	res, err := api.GetProxies(accountID)
	assert.NoError(t, err)
	assert.Equal(t, balance, res.Balance)
	assert.Equal(t, proxyDefinitions, res.ProxyDefinitions)
}

func TestAPI_GetProxies_ValidationError(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	res, err := api.GetProxies(nil)
	assert.ErrorIs(t, err, validation.ErrInvalidAccountID)
	assert.Nil(t, res)
}

func TestAPI_GetProxies_MetadataRetrievalError(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(nil, errors.New("error")).Once()

	res, err := api.GetProxies(accountID)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_GetProxies_StorageKeyCreationError(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	// NOTE - types.MetadataV13 does not have info on the Proxy pallet,
	// causing types.CreateStorageKey to fail.
	meta := types.NewMetadataV13()

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil).Once()

	res, err := api.GetProxies(accountID)
	assert.ErrorIs(t, err, errors.ErrStorageKeyCreation)
	assert.Nil(t, res)
}

func TestAPI_GetProxies_StorageRetrievalError(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	encodedAccountID, err := codec.Encode(accountID)
	assert.NoError(t, err)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil).Once()

	storageKey, err := types.CreateStorageKey(meta, PalletName, ProxiesStorageName, encodedAccountID)

	centAPIMock.On("GetStorageLatest", storageKey, mock.IsType(&types.ProxyStorageEntry{})).
		Return(false, errors.New("error")).Once()

	res, err := api.GetProxies(accountID)
	assert.ErrorIs(t, err, ErrProxyStorageEntryRetrieval)
	assert.Nil(t, res)
}

func TestAPI_GetProxies_ProxiesNotFound(t *testing.T) {
	centAPIMock := centchain.NewAPIMock(t)

	api := NewAPI(centAPIMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	encodedAccountID, err := codec.Encode(accountID)
	assert.NoError(t, err)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	centAPIMock.On("GetMetadataLatest").
		Return(meta, nil).Once()

	storageKey, err := types.CreateStorageKey(meta, PalletName, ProxiesStorageName, encodedAccountID)

	centAPIMock.On("GetStorageLatest", storageKey, mock.IsType(&types.ProxyStorageEntry{})).
		Return(false, nil).Once()

	res, err := api.GetProxies(accountID)
	assert.ErrorIs(t, err, ErrProxiesNotFound)
	assert.Nil(t, res)
}
