//go:build unit
// +build unit

package centchain

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration

func TestMain(m *testing.M) {
	ibootstrappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstrappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstrappers)
	os.Exit(result)
}

func TestApi_GetMetadataLatest(t *testing.T) {
	mockSAPI := NewSubstrateApiMock(t)
	mockSAPI.On("GetMetadataLatest").Return(types.NewMetadataV8(), nil).Once()
	api := NewAPI(mockSAPI, nil, nil)
	meta, err := api.GetMetadataLatest()
	assert.NoError(t, err)
	assert.Equal(t, types.NewMetadataV8(), meta)
}

func TestApi_Call(t *testing.T) {
	mockSAPI := NewSubstrateApiMock(t)
	mockSAPI.On("Call", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	api := NewAPI(mockSAPI, nil, nil)
	err := api.Call(nil, "", nil)
	assert.NoError(t, err)
}

func TestApi_SubmitWithRetries(t *testing.T) {
	meta := MetaDataWithCall("Anchor.commit")
	c, err := types.NewCall(
		meta,
		"Anchor.commit",
		types.NewHash(utils.RandomSlice(32)),
		types.NewHash(utils.RandomSlice(32)),
		types.NewHash(utils.RandomSlice(32)),
		types.NewMoment(time.Now()))
	assert.NoError(t, err)
	cacc, err := cfg.GetCentChainAccount()
	assert.NoError(t, err)
	krp, err := cacc.KeyRingPair()
	assert.NoError(t, err)

	mockRetries := testingutils.MockConfigOption(cfg, "centChain.maxRetries", 3)
	defer mockRetries()

	mockSAPI := NewSubstrateApiMock(t)
	iapi := NewAPI(mockSAPI, cfg, nil)
	tapi := iapi.(*api)

	// Failed to get nonce from chain
	ctx := context.Background()
	mockSAPI.On("GetStorageLatest", mock.Anything, mock.Anything).Return(false, errors.New("failed to get nonce from storage")).Once()
	_, _, _, err = tapi.SubmitExtrinsic(ctx, meta, c, krp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get nonce from storage")

	// Irrecoverable failure to submit extrinsic
	mockSAPI.On("GetStorageLatest", mock.Anything, mock.Anything).Return(false, nil).Once()
	mockSAPI.On("GetBlockHash", mock.Anything).Return(types.Hash{}, errors.New("failed to get block hash")).Once()
	_, _, _, err = tapi.SubmitExtrinsic(ctx, meta, c, krp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get block hash")

	// Recoverable failure to submit extrinsic, max retrials reached
	mockSAPI.On("GetStorageLatest", mock.Anything, mock.Anything).Return(false, nil).Times(3)
	mockSAPI.On("GetBlockHash", mock.Anything).Return(types.Hash{}, ErrNonceTooLow).Times(3)
	_, _, _, err = tapi.SubmitExtrinsic(ctx, meta, c, krp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max concurrent transaction tries reached")

	// Success
	mockSAPI.On("GetBlockHash", mock.Anything).Return(types.Hash(utils.RandomByte32()), nil).Once()
	mockSAPI.On("GetRuntimeVersionLatest").Return(types.NewRuntimeVersion(), nil)
	mockClient := new(MockClient)
	mockSAPI.On("GetClient").Return(mockClient)
	mockSAPI.On("GetBlockLatest", mock.Anything).Return(new(types.SignedBlock), nil)
	mockClient.On("Call", mock.Anything, mock.Anything, mock.Anything).Return(hexutil.Encode(utils.RandomSlice(32)), nil)
	_, _, _, err = tapi.SubmitExtrinsic(ctx, meta, c, krp)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), tapi.accounts[cacc.ID]) // Incremented nonce
	mockSAPI.AssertExpectations(t)
}
