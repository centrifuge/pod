// +build unit

package config

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/config"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestGrpcHandler_GetConfigNoConfig(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo)
	h := GRPCHandler(svc)
	readCfg, err := h.GetConfig(context.Background(), nil)
	assert.NotNil(t, err)
	assert.Nil(t, readCfg)
}

func TestGrpcHandler_GetConfig(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo)
	h := GRPCHandler(svc)
	nodeCfg := NewNodeConfig(cfg)
	_, err = h.CreateConfig(context.Background(), nodeCfg.createProtobuf())
	assert.Nil(t, err)
	readCfg, err := h.GetConfig(context.Background(), nil)
	assert.Nil(t, err)
	assert.NotNil(t, readCfg)
}

func TestGrpcHandler_GetTenantNoConfig(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo)
	h := GRPCHandler(svc)
	readCfg, err := h.GetTenant(context.Background(), &configpb.GetTenantRequest{Identifier: "0x123456789"})
	assert.NotNil(t, err)
	assert.Nil(t, readCfg)
}

func TestGrpcHandler_GetTenant(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo)
	h := GRPCHandler(svc)
	nodeCfg, err := NewTenantConfig("main", cfg)
	assert.Nil(t, err)
	_, err = h.CreateTenant(context.Background(), nodeCfg.createProtobuf())
	assert.Nil(t, err)
	readCfg, err := h.GetTenant(context.Background(), &configpb.GetTenantRequest{Identifier: hexutil.Encode(nodeCfg.IdentityID)})
	assert.Nil(t, err)
	assert.NotNil(t, readCfg)
}

func TestGrpcHandler_GetAllTenants(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo)
	h := GRPCHandler(svc)
	nodeCfg1, err := NewTenantConfig("main", cfg)
	nodeCfg2, err := NewTenantConfig("main", cfg)
	nodeCfg2.IdentityID = []byte("0x123456789")
	_, err = h.CreateTenant(context.Background(), nodeCfg1.createProtobuf())
	assert.Nil(t, err)
	_, err = h.CreateTenant(context.Background(), nodeCfg2.createProtobuf())
	assert.Nil(t, err)

	resp, err := h.GetAllTenants(context.Background(), nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp.Data))
}

func TestGrpcHandler_CreateConfig(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo)
	h := GRPCHandler(svc)
	nodeCfg := NewNodeConfig(cfg)
	_, err = h.CreateConfig(context.Background(), nodeCfg.createProtobuf())
	assert.Nil(t, err)

	// Already exists
	_, err = h.CreateConfig(context.Background(), nodeCfg.createProtobuf())
	assert.NotNil(t, err)
}

func TestGrpcHandler_CreateTenant(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo)
	h := GRPCHandler(svc)
	nodeCfg, err := NewTenantConfig("main", cfg)
	assert.Nil(t, err)
	_, err = h.CreateTenant(context.Background(), nodeCfg.createProtobuf())
	assert.Nil(t, err)

	// Already exists
	_, err = h.CreateTenant(context.Background(), nodeCfg.createProtobuf())
	assert.NotNil(t, err)
}

func TestGrpcHandler_UpdateConfig(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo)
	h := GRPCHandler(svc)
	nodeCfg := NewNodeConfig(cfg)

	// Config doesn't exist
	_, err = h.UpdateConfig(context.Background(), nodeCfg.createProtobuf())
	assert.NotNil(t, err)

	_, err = h.CreateConfig(context.Background(), nodeCfg.createProtobuf())
	assert.Nil(t, err)
	nodeCfg.NetworkString = "other"
	_, err = h.UpdateConfig(context.Background(), nodeCfg.createProtobuf())
	assert.Nil(t, err)

	readCfg, err := h.GetConfig(context.Background(), nil)
	assert.Nil(t, err)
	assert.Equal(t, nodeCfg.NetworkString, readCfg.Network)
}

func TestGrpcHandler_UpdateTenant(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo)
	h := GRPCHandler(svc)
	nodeCfg, err := NewTenantConfig("main", cfg)
	assert.Nil(t, err)

	// Config doesn't exist
	_, err = h.UpdateTenant(context.Background(), &configpb.UpdateTenantRequest{Identifier: hexutil.Encode(nodeCfg.IdentityID), Data: nodeCfg.createProtobuf()})
	assert.NotNil(t, err)

	_, err = h.CreateTenant(context.Background(), nodeCfg.createProtobuf())
	assert.Nil(t, err)
	nodeCfg.EthereumDefaultAccountName = "other"
	_, err = h.UpdateTenant(context.Background(), &configpb.UpdateTenantRequest{Identifier: hexutil.Encode(nodeCfg.IdentityID), Data: nodeCfg.createProtobuf()})
	assert.Nil(t, err)

	readCfg, err := h.GetTenant(context.Background(), &configpb.GetTenantRequest{Identifier: hexutil.Encode(nodeCfg.IdentityID)})
	assert.Nil(t, err)
	assert.Equal(t, nodeCfg.EthereumDefaultAccountName, readCfg.EthDefaultAccountName)
}

func TestGrpcHandler_DeleteConfig(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo)
	h := GRPCHandler(svc)

	//No error when no config
	_, err = h.DeleteConfig(context.Background(), nil)
	assert.Nil(t, err)

	nodeCfg := NewNodeConfig(cfg)
	_, err = h.CreateConfig(context.Background(), nodeCfg.createProtobuf())
	assert.Nil(t, err)
	_, err = h.DeleteConfig(context.Background(), nil)
	assert.Nil(t, err)

	readCfg, err := h.GetConfig(context.Background(), nil)
	assert.NotNil(t, err)
	assert.Nil(t, readCfg)
}

func TestGrpcHandler_DeleteTenant(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo)
	h := GRPCHandler(svc)

	//No error when no config
	_, err = h.DeleteTenant(context.Background(), &configpb.GetTenantRequest{Identifier: "0x12345678"})
	assert.Nil(t, err)

	nodeCfg, err := NewTenantConfig("main", cfg)
	assert.Nil(t, err)
	_, err = h.CreateTenant(context.Background(), nodeCfg.createProtobuf())
	assert.Nil(t, err)
	_, err = h.DeleteTenant(context.Background(), &configpb.GetTenantRequest{Identifier: hexutil.Encode(nodeCfg.IdentityID)})
	assert.Nil(t, err)

	readCfg, err := h.GetTenant(context.Background(), &configpb.GetTenantRequest{Identifier: hexutil.Encode(nodeCfg.IdentityID)})
	assert.NotNil(t, err)
	assert.Nil(t, readCfg)
}
