//go:build unit

package utility

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/testingutils"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
)

func TestAPI_BatchAll(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	callCreationFn1 := centchain.CallProviderFn(func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(meta, "System.remark", []byte{1, 2, 3})

		if err != nil {
			return nil, err
		}

		return &call, nil
	})

	callCreationFn2 := centchain.CallProviderFn(func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(meta, "System.remark", []byte{1, 2, 3})

		if err != nil {
			return nil, err
		}

		return &call, nil
	})

	batchCall, err := BatchCalls(callCreationFn1, callCreationFn2)(meta)
	assert.NoError(t, err)

	var krp signature.KeyringPair

	genericUtils.GetMock[*config.PodOperatorMock](mocks).
		On("ToKeyringPair").
		Return(krp).Once()

	extInfo := &centchain.ExtrinsicInfo{}

	genericUtils.GetMock[*proxy.APIMock](mocks).
		On(
			"ProxyCall",
			ctx,
			identity,
			krp,
			types.NewOption(proxyType.PodOperation),
			*batchCall,
		).
		Return(extInfo, nil).Once()

	res, err := api.BatchAll(ctx, callCreationFn1, callCreationFn2)
	assert.NoError(t, err)
	assert.Equal(t, extInfo, res)
}

func TestAPI_BatchAll_IdentityRetrievalError(t *testing.T) {
	ctx := context.Background()

	api, _ := getAPIWithMocks(t)

	callCreationFn1 := centchain.CallProviderFn(func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(meta, "System.remark", []byte{1, 2, 3})

		if err != nil {
			return nil, err
		}

		return &call, nil
	})

	callCreationFn2 := centchain.CallProviderFn(func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(meta, "System.remark", []byte{1, 2, 3})

		if err != nil {
			return nil, err
		}

		return &call, nil
	})

	res, err := api.BatchAll(ctx, callCreationFn1, callCreationFn2)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestAPI_BatchAll_MetadataRetrievalError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(nil, errors.New("error"))

	callCreationFn1 := centchain.CallProviderFn(func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(meta, "System.remark", []byte{1, 2, 3})

		if err != nil {
			return nil, err
		}

		return &call, nil
	})

	callCreationFn2 := centchain.CallProviderFn(func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(meta, "System.remark", []byte{1, 2, 3})

		if err != nil {
			return nil, err
		}

		return &call, nil
	})

	res, err := api.BatchAll(ctx, callCreationFn1, callCreationFn2)
	assert.ErrorIs(t, err, errors.ErrMetadataRetrieval)
	assert.Nil(t, res)
}

func TestAPI_BatchAll_BatchCallCreationError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	callCreationFn1 := centchain.CallProviderFn(func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(meta, "System.remark", []byte{1, 2, 3})

		if err != nil {
			return nil, err
		}

		return &call, nil
	})

	callCreationFn2 := centchain.CallProviderFn(func(meta *types.Metadata) (*types.Call, error) {
		return nil, errors.New("error")
	})

	res, err := api.BatchAll(ctx, callCreationFn1, callCreationFn2)
	assert.ErrorIs(t, err, ErrBatchCallCreation)
	assert.Nil(t, res)
}

func TestAPI_BatchAll_ProxyCallError(t *testing.T) {
	ctx := context.Background()

	api, mocks := getAPIWithMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	meta, err := testingutils.GetTestMetadata()
	assert.NoError(t, err)

	genericUtils.GetMock[*centchain.APIMock](mocks).
		On("GetMetadataLatest").
		Return(meta, nil)

	callCreationFn1 := centchain.CallProviderFn(func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(meta, "System.remark", []byte{1, 2, 3})

		if err != nil {
			return nil, err
		}

		return &call, nil
	})

	callCreationFn2 := centchain.CallProviderFn(func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(meta, "System.remark", []byte{1, 2, 3})

		if err != nil {
			return nil, err
		}

		return &call, nil
	})

	batchCall, err := BatchCalls(callCreationFn1, callCreationFn2)(meta)
	assert.NoError(t, err)

	var krp signature.KeyringPair

	genericUtils.GetMock[*config.PodOperatorMock](mocks).
		On("ToKeyringPair").
		Return(krp).Once()

	genericUtils.GetMock[*proxy.APIMock](mocks).
		On(
			"ProxyCall",
			ctx,
			identity,
			krp,
			types.NewOption(proxyType.PodOperation),
			*batchCall,
		).
		Return(nil, errors.New("error")).Once()

	res, err := api.BatchAll(ctx, callCreationFn1, callCreationFn2)
	assert.ErrorIs(t, err, errors.ErrProxyCall)
	assert.Nil(t, res)
}

func getAPIWithMocks(t *testing.T) (*api, []any) {
	centAPIMock := centchain.NewAPIMock(t)
	proxyAPIMock := proxy.NewAPIMock(t)
	podOperatorMock := config.NewPodOperatorMock(t)

	utilityAPI := NewAPI(centAPIMock, proxyAPIMock, podOperatorMock)

	return utilityAPI.(*api), []any{
		centAPIMock,
		proxyAPIMock,
		podOperatorMock,
	}
}
