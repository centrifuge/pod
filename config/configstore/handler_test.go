// +build unit

package configstore

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/account"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestGrpcHandler_GetConfigNoConfig(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo, idService)
	h := GRPCHandler(svc)
	readCfg, err := h.GetConfig(context.Background(), nil)
	assert.NotNil(t, err)
	assert.Nil(t, readCfg)
}

func TestGrpcHandler_GetConfig(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo, idService)
	h := GRPCHandler(svc)
	nodeCfg := NewNodeConfig(cfg)
	_, err = svc.CreateConfig(nodeCfg)
	assert.Nil(t, err)
	readCfg, err := h.GetConfig(context.Background(), nil)
	assert.Nil(t, err)
	assert.NotNil(t, readCfg)
}

func TestGrpcHandler_GetAccountNotExist(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)
	h := GRPCAccountHandler(svc)
	readCfg, err := h.GetAccount(context.Background(), &accountpb.GetAccountRequest{Identifier: "0x123456789"})
	assert.NotNil(t, err)
	assert.Nil(t, readCfg)
}

func TestGrpcHandler_GetTenant(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)
	h := GRPCAccountHandler(svc)
	tenantCfg, err := NewAccount("main", cfg)
	assert.Nil(t, err)
	tcpb, err := tenantCfg.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.CreateAccount(context.Background(), tcpb)
	assert.Nil(t, err)
	tid, err := tenantCfg.GetIdentityID()
	assert.Nil(t, err)
	readCfg, err := h.GetAccount(context.Background(), &accountpb.GetAccountRequest{Identifier: hexutil.Encode(tid)})
	assert.Nil(t, err)
	assert.NotNil(t, readCfg)
}

func TestGrpcHandler_deriveAllTenantResponseFailure(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo, idService)
	h := GRPCAccountHandler(svc)
	tenantCfg1, err := NewTenantConfig("main", cfg)
	tenantCfg2, err := NewTenantConfig("main", cfg)
	tco := tenantCfg1.(*TenantConfig)
	tco.EthereumAccount = nil
	tcs := []config.TenantConfiguration{tco, tenantCfg2}
	hc := h.(*grpcHandler)
	resp, err := hc.deriveAllTenantResponse(tcs)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(resp.Data))
}

func TestGrpcHandler_GetAllTenants(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)
	h := GRPCAccountHandler(svc)
	tenantCfg1, err := NewAccount("main", cfg)
	tenantCfg2, err := NewAccount("main", cfg)
	tc := tenantCfg2.(*Account)
	tc.IdentityID = []byte("0x123456789")
	tc1pb, err := tenantCfg1.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.CreateAccount(context.Background(), tc1pb)
	assert.Nil(t, err)
	tcpb, err := tc.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.CreateAccount(context.Background(), tcpb)
	assert.Nil(t, err)

	resp, err := h.GetAllAccounts(context.Background(), nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp.Data))
}

func TestGrpcHandler_CreateTenant(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)
	h := GRPCAccountHandler(svc)
	tc, err := NewAccount("main", cfg)
	assert.Nil(t, err)
	tcpb, err := tc.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.CreateAccount(context.Background(), tcpb)
	assert.Nil(t, err)

	// Already exists
	tcpb, err = tc.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.CreateAccount(context.Background(), tcpb)
	assert.NotNil(t, err)
}

func TestGrpcHandler_GenerateTenant(t *testing.T) {
	s := MockService{}
	t1, _ := NewAccount(cfg.GetEthereumDefaultAccountName(), cfg)
	s.On("GenerateAccount").Return(t1, nil)
	h := GRPCAccountHandler(s)
	tc, err := h.GenerateAccount(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, tc)
}

func TestGrpcHandler_UpdateTenant(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)
	h := GRPCAccountHandler(svc)
	tcfg, err := NewAccount("main", cfg)
	assert.Nil(t, err)

	tid, err := tcfg.GetIdentityID()
	assert.Nil(t, err)

	tc := tcfg.(*Account)

	// Config doesn't exist
	tcpb, err := tcfg.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.UpdateAccount(context.Background(), &accountpb.UpdateAccountRequest{Identifier: hexutil.Encode(tid), Data: tcpb})
	assert.NotNil(t, err)

	tcpb, err = tcfg.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.CreateAccount(context.Background(), tcpb)
	assert.Nil(t, err)
	tc.EthereumDefaultAccountName = "other"
	tccpb, err := tc.CreateProtobuf()
	assert.NoError(t, err)
	_, err = h.UpdateAccount(context.Background(), &accountpb.UpdateAccountRequest{Identifier: hexutil.Encode(tid), Data: tccpb})
	assert.Nil(t, err)

	readCfg, err := h.GetAccount(context.Background(), &accountpb.GetAccountRequest{Identifier: hexutil.Encode(tid)})
	assert.Nil(t, err)
	assert.Equal(t, tc.EthereumDefaultAccountName, readCfg.EthDefaultAccountName)
}
