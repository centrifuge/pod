// +build unit

package centchain

import (
	"os"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/client"
	"github.com/centrifuge/go-substrate-rpc-client/types"
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
	api := api{getMetadataLatest: func() (*types.Metadata, error) {
		return types.NewMetadataV8(), nil
	}}

	meta, err := api.getMetadataLatest()
	assert.NoError(t, err)
	assert.Equal(t, types.NewMetadataV8(), meta)
}

func TestApi_SubmitExtrinsic(t *testing.T) {
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

	// failed getGenesisHash
	api := api{}
	api.getBlockHash = func(u uint64) (hashes types.Hash, e error) {
		return hashes, errors.New("failed to get block hash")
	}
	_, _, _, err = api.SubmitExtrinsic(meta, c, krp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get block hash")

	// failed to get runtime version latest
	api.getBlockHash = func(u uint64) (hashes types.Hash, e error) { return types.Hash(utils.RandomByte32()), nil }
	api.getRuntimeVersionLatest = func() (version *types.RuntimeVersion, e error) {
		return nil, errors.New("failed to get runtime version")
	}
	_, _, _, err = api.SubmitExtrinsic(meta, c, krp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get runtime version")

	// failed to get storage latest
	api.getRuntimeVersionLatest = func() (version *types.RuntimeVersion, e error) { return types.NewRuntimeVersion(), nil }
	api.getStorageLatest = func(key types.StorageKey, target interface{}) error {
		return errors.New("failed to get storage latest")
	}
	_, _, _, err = api.SubmitExtrinsic(meta, c, krp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get storage latest")

	// failed to get latest block
	api.getStorageLatest = func(key types.StorageKey, target interface{}) error { return nil }
	cl := new(MockClient)
	api.getClient = func() client.Client { return cl }
	api.getBlockLatest = func() (block *types.SignedBlock, e error) {
		return nil, errors.New("failed to get latest block")
	}
	_, _, _, err = api.SubmitExtrinsic(meta, c, krp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get latest block")

	// success
	api.getBlockLatest = func() (block *types.SignedBlock, e error) { return new(types.SignedBlock), nil }
	cl.On("Call", mock.Anything, mock.Anything, mock.Anything).Return(hexutil.Encode(utils.RandomSlice(32)), nil)
	_, _, _, err = api.SubmitExtrinsic(meta, c, krp)
	assert.NoError(t, err)
	cl.AssertExpectations(t)
}
