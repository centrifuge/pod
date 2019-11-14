// +build unit integration

package centchain

import (
	"strings"

	"github.com/centrifuge/go-substrate-rpc-client/client"
	"github.com/centrifuge/go-substrate-rpc-client/signature"
	"github.com/centrifuge/go-substrate-rpc-client/types"
	"github.com/stretchr/testify/mock"
)

type MockAPI struct {
	mock.Mock
	API
}

func (m *MockAPI) GetMetadataLatest() (*types.Metadata, error) {
	args := m.Called()
	md, _ := args.Get(0).(*types.Metadata)
	return md, args.Error(1)
}

func (m *MockAPI) SubmitExtrinsic(meta *types.Metadata, c types.Call, krp signature.KeyringPair) (txHash types.Hash, bn types.BlockNumber, sig types.Signature, err error) {
	args := m.Called(meta, c, krp)
	txHash, _ = args.Get(0).(types.Hash)
	bn, _ = args.Get(1).(types.BlockNumber)
	sig, _ = args.Get(2).(types.Signature)
	return txHash, bn, sig, args.Error(3)
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
						Name: "AccountNonce",
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
