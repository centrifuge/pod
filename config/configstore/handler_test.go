// +build unit

package configstore

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/account"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
)

func TestGrpcHandler_deriveAllAccountResponseFailure(t *testing.T) {
	idService := &testingcommons.MockIdentityService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)
	h := GRPCAccountHandler(svc)
	accountCfg1, err := NewAccount("main", cfg)
	assert.NoError(t, err)
	accountCfg2, err := NewAccount("main", cfg)
	assert.NoError(t, err)
	tco := accountCfg1.(*Account)
	tco.EthereumAccount = nil
	tcs := []config.Account{tco, accountCfg2}
	hc := h.(*grpcHandler)
	resp, err := hc.deriveAllAccountResponse(tcs)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(resp.Data))
}

func TestGrpcHandler_GetAllAccounts(t *testing.T) {
	idService := &testingcommons.MockIdentityService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)
	h := GRPCAccountHandler(svc)
	accountCfg1, err := NewAccount("main", cfg)
	assert.NoError(t, err)
	accountCfg2, err := NewAccount("main", cfg)
	assert.NoError(t, err)
	acc := accountCfg2.(*Account)
	acc.IdentityID = []byte("0x123456789")
	tc1pb, err := accountCfg1.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.CreateAccount(context.Background(), tc1pb)
	assert.Nil(t, err)
	accpb, err := acc.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.CreateAccount(context.Background(), accpb)
	assert.Nil(t, err)

	resp, err := h.GetAllAccounts(context.Background(), nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp.Data))
}

func TestGrpcHandler_CreateAccount(t *testing.T) {
	idService := &testingcommons.MockIdentityService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)
	h := GRPCAccountHandler(svc)
	tc, err := NewAccount("main", cfg)
	assert.Nil(t, err)
	accpb, err := tc.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.CreateAccount(context.Background(), accpb)
	assert.Nil(t, err)

	// Already exists
	accpb, err = tc.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.CreateAccount(context.Background(), accpb)
	assert.NotNil(t, err)
}

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
	_, err = h.CreateAccount(context.Background(), accpb)
	assert.Nil(t, err)
	acc.EthereumDefaultAccountName = "other"
	tccpb, err := acc.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.UpdateAccount(context.Background(), &accountpb.UpdateAccountRequest{AccountId: hexutil.Encode(accID), Data: tccpb})
	assert.Nil(t, err)

	cfgs, err := h.GetAllAccounts(context.Background(), new(empty.Empty))
	assert.Nil(t, err)
	readCfg := cfgs.Data[0]
	assert.Equal(t, acc.EthereumDefaultAccountName, readCfg.EthDefaultAccountName)
}
