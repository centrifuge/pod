// +build unit

package configstore

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/accounts"
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
	repo.RegisterTenant(&TenantConfig{})
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
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo, idService)
	h := GRPCAccountHandler(svc)
	tenantCfg, err := NewTenantConfig("main", cfg)
	assert.Nil(t, err)
	_, err = h.CreateAccount(context.Background(), tenantCfg.CreateProtobuf())
	assert.Nil(t, err)
	tid, err := tenantCfg.GetIdentityID()
	assert.Nil(t, err)
	readCfg, err := h.GetAccount(context.Background(), &accountpb.GetAccountRequest{Identifier: hexutil.Encode(tid)})
	assert.Nil(t, err)
	assert.NotNil(t, readCfg)
}

func TestGrpcHandler_GetAllTenants(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo, idService)
	h := GRPCAccountHandler(svc)
	tenantCfg1, err := NewTenantConfig("main", cfg)
	tenantCfg2, err := NewTenantConfig("main", cfg)
	tc := tenantCfg2.(*TenantConfig)
	tc.IdentityID = []byte("0x123456789")
	_, err = h.CreateAccount(context.Background(), tenantCfg1.CreateProtobuf())
	assert.Nil(t, err)
	_, err = h.CreateAccount(context.Background(), tc.CreateProtobuf())
	assert.Nil(t, err)

	resp, err := h.GetAllAccounts(context.Background(), nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp.Data))
}

func TestGrpcHandler_CreateTenant(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo, idService)
	h := GRPCAccountHandler(svc)
	nodeCfg, err := NewTenantConfig("main", cfg)
	assert.Nil(t, err)
	_, err = h.CreateAccount(context.Background(), nodeCfg.CreateProtobuf())
	assert.Nil(t, err)

	// Already exists
	_, err = h.CreateAccount(context.Background(), nodeCfg.CreateProtobuf())
	assert.NotNil(t, err)
}

func TestGrpcHandler_GenerateTenant(t *testing.T) {
	s := MockService{}
	t1, _ := NewTenantConfig(cfg.GetEthereumDefaultAccountName(), cfg)
	s.On("GenerateTenant").Return(t1, nil)
	h := GRPCAccountHandler(s)
	tc, err := h.GenerateAccount(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, tc)
}

func TestGrpcHandler_UpdateTenant(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo, idService)
	h := GRPCAccountHandler(svc)
	nodeCfg, err := NewTenantConfig("main", cfg)
	assert.Nil(t, err)

	tid, err := nodeCfg.GetIdentityID()
	assert.Nil(t, err)

	tc := nodeCfg.(*TenantConfig)

	// Config doesn't exist
	_, err = h.UpdateAccount(context.Background(), &accountpb.UpdateAccountRequest{Identifier: hexutil.Encode(tid), Data: nodeCfg.CreateProtobuf()})
	assert.NotNil(t, err)

	_, err = h.CreateAccount(context.Background(), nodeCfg.CreateProtobuf())
	assert.Nil(t, err)
	tc.EthereumDefaultAccountName = "other"
	_, err = h.UpdateAccount(context.Background(), &accountpb.UpdateAccountRequest{Identifier: hexutil.Encode(tid), Data: tc.CreateProtobuf()})
	assert.Nil(t, err)

	readCfg, err := h.GetAccount(context.Background(), &accountpb.GetAccountRequest{Identifier: hexutil.Encode(tid)})
	assert.Nil(t, err)
	assert.Equal(t, tc.EthereumDefaultAccountName, readCfg.EthDefaultAccountName)
}
