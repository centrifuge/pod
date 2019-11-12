// +build unit integration

package centchain

import (
	"strings"

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
	meta := types.NewMetadataV4()
	meta.AsMetadataV4.Modules = []types.ModuleMetadataV4{
		types.ModuleMetadataV4{
			Name:       types.Text(data[0]),
			Prefix:     "",
			HasStorage: false,
			Storage:    nil,
			HasCalls:   true,
			Calls: []types.FunctionMetadataV4{{
				Name: types.Text(data[1]),
			}},
			HasEvents: false,
			Events:    nil,
		},
	}
	return meta
}
