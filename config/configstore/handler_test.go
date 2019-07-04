// +build unit

package configstore

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/account"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestGrpcHandler_UpdateAccount(t *testing.T) {
	idService := &testingcommons.MockIdentityService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)
	h := GRPCAccountHandler(svc)
	tcfg, err := NewAccount("main", cfg)
	assert.Nil(t, err)

	accID := tcfg.GetIdentityID()
	acc := tcfg.(*Account)

	// Config doesn't exist
	accpb, err := tcfg.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.UpdateAccount(context.Background(), &accountpb.UpdateAccountRequest{AccountId: hexutil.Encode(accID), Data: accpb})
	assert.NotNil(t, err)

	accpb, err = tcfg.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.(*grpcHandler).CreateAccount(context.Background(), accpb)
	assert.Nil(t, err)
	acc.EthereumDefaultAccountName = "other"
	tccpb, err := acc.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.UpdateAccount(context.Background(), &accountpb.UpdateAccountRequest{AccountId: hexutil.Encode(accID), Data: tccpb})
	assert.Nil(t, err)

	cfgs, err := svc.GetAccounts()
	assert.Nil(t, err)
	readCfg := cfgs[0]
	assert.Equal(t, acc.EthereumDefaultAccountName, readCfg.GetEthereumDefaultAccountName())
}
