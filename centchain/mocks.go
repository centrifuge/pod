//go:build unit || integration
// +build unit integration

package centchain

import (
	"context"
	"strings"

	"github.com/centrifuge/go-substrate-rpc-client/v4/client"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/mock"
)

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (Bootstrapper) TestTearDown() error {
	return nil
}

type MockAPI struct {
	mock.Mock
	API
}

type MockSubstrateAPI struct {
	mock.Mock
	substrateAPI
}

func (ms *MockSubstrateAPI) GetMetadataLatest() (*types.Metadata, error) {
	args := ms.Called()
	md, _ := args.Get(0).(*types.Metadata)
	return md, args.Error(1)
}

func (ms *MockSubstrateAPI) Call(result interface{}, method string, args ...interface{}) error {
	argss := ms.Called()
	return argss.Error(0)
}

func (ms *MockSubstrateAPI) GetBlockHash(blockNumber uint64) (types.Hash, error) {
	args := ms.Called()
	md, _ := args.Get(0).(types.Hash)
	return md, args.Error(1)
}

func (ms *MockSubstrateAPI) GetBlock(blockHash types.Hash) (*types.SignedBlock, error) {
	args := ms.Called()
	md, _ := args.Get(0).(*types.SignedBlock)
	return md, args.Error(1)
}

func (ms *MockSubstrateAPI) GetStorage(key types.StorageKey, target interface{}, blockHash types.Hash) error {
	args := ms.Called()
	return args.Error(0)
}

func (ms *MockSubstrateAPI) GetBlockLatest() (*types.SignedBlock, error) {
	args := ms.Called()
	md, _ := args.Get(0).(*types.SignedBlock)
	return md, args.Error(1)
}

func (ms *MockSubstrateAPI) GetRuntimeVersionLatest() (*types.RuntimeVersion, error) {
	args := ms.Called()
	md, _ := args.Get(0).(*types.RuntimeVersion)
	return md, args.Error(1)
}

func (ms *MockSubstrateAPI) GetClient() client.Client {
	args := ms.Called()
	md, _ := args.Get(0).(client.Client)
	return md
}

func (ms *MockSubstrateAPI) GetStorageLatest(key types.StorageKey, target interface{}) error {
	args := ms.Called(key, target)
	return args.Error(0)
}

func (m *MockAPI) GetMetadataLatest() (*types.Metadata, error) {
	args := m.Called()
	md, _ := args.Get(0).(*types.Metadata)
	return md, args.Error(1)
}

func (m *MockAPI) SubmitExtrinsic(ctx context.Context, meta *types.Metadata, c types.Call, krp signature.KeyringPair) (txHash types.Hash, bn types.BlockNumber, sig types.MultiSignature, err error) {
	args := m.Called(ctx, meta, c, krp)
	txHash, _ = args.Get(0).(types.Hash)
	bn, _ = args.Get(1).(types.BlockNumber)
	sig, _ = args.Get(2).(types.MultiSignature)
	return txHash, bn, sig, args.Error(3)
}

func (m *MockAPI) SubmitAndWatch(ctx context.Context, meta *types.Metadata, c types.Call,
	krp signature.KeyringPair) (ExtrinsicInfo, error) {
	args := m.Called(c, krp)
	info, _ := args.Get(0).(ExtrinsicInfo)
	return info, args.Error(1)
}

func MetaDataWithCall(call string) *types.Metadata {
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

type MockClient struct {
	mock.Mock
	client.Client
}

func (m *MockClient) Call(result interface{}, method string, args ...interface{}) error {
	arg := m.Called(result, method, args)
	res := arg.Get(0).(string)
	eres := result.(*string)
	*eres = res
	return arg.Error(1)
}
