// +build unit

package configstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_GetConfig_NoConfig(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo)
	cfg, err := svc.GetConfig()
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
}

func TestService_GetConfig(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo)
	nodeCfg := NewNodeConfig(cfg)
	err = repo.CreateConfig(nodeCfg)
	assert.Nil(t, err)
	cfg, err := svc.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
}

func TestService_GetTenant_NoTenant(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo)
	cfg, err := svc.GetTenant([]byte("0x123456789"))
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
}

func TestService_GetTenant(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo)
	tenantCfg, err := NewTenantConfig("main", cfg)
	assert.Nil(t, err)
	tid, _ := tenantCfg.GetIdentityID()
	err = repo.CreateTenant(tid, tenantCfg)
	assert.Nil(t, err)
	cfg, err := svc.GetTenant(tid)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
}

func TestService_CreateConfig(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo)
	nodeCfg := NewNodeConfig(cfg)
	cfgpb, err := svc.CreateConfig(nodeCfg)
	assert.Nil(t, err)
	assert.Equal(t, nodeCfg.GetStoragePath(), cfgpb.GetStoragePath())

	//Config already exists
	_, err = svc.CreateConfig(nodeCfg)
	assert.NotNil(t, err)
}

func TestService_CreateTenant(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo)
	tenantCfg, err := NewTenantConfig("main", cfg)
	assert.Nil(t, err)
	newCfg, err := svc.CreateTenant(tenantCfg)
	assert.Nil(t, err)
	i, err := newCfg.GetIdentityID()
	assert.Nil(t, err)
	tid, err := tenantCfg.GetIdentityID()
	assert.Nil(t, err)
	assert.Equal(t, tid, i)

	//Tenant already exists
	_, err = svc.CreateTenant(tenantCfg)
	assert.NotNil(t, err)
}

func TestService_UpdateConfig(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo)
	nodeCfg := NewNodeConfig(cfg)

	//Config doesn't exists
	_, err = svc.UpdateConfig(nodeCfg)
	assert.NotNil(t, err)

	newCfg, err := svc.CreateConfig(nodeCfg)
	assert.Nil(t, err)
	assert.Equal(t, nodeCfg.GetStoragePath(), newCfg.GetStoragePath())

	n := nodeCfg.(*NodeConfig)
	n.NetworkString = "something"
	newCfg, err = svc.UpdateConfig(n)
	assert.Nil(t, err)
	assert.Equal(t, n.GetNetworkString(), newCfg.GetNetworkString())
}

func TestService_UpdateTenant(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo)
	tenantCfg, err := NewTenantConfig("main", cfg)

	// Tenant doesn't exist
	newCfg, err := svc.UpdateTenant(tenantCfg)
	assert.NotNil(t, err)

	newCfg, err = svc.CreateTenant(tenantCfg)
	assert.Nil(t, err)
	i, err := newCfg.GetIdentityID()
	assert.Nil(t, err)
	tid, err := tenantCfg.GetIdentityID()
	assert.Nil(t, err)
	assert.Equal(t, tid, i)

	tc := tenantCfg.(*TenantConfig)
	tc.EthereumDefaultAccountName = "other"
	newCfg, err = svc.UpdateTenant(tenantCfg)
	assert.Nil(t, err)
	assert.Equal(t, tc.EthereumDefaultAccountName, newCfg.GetEthereumDefaultAccountName())
}

func TestService_DeleteConfig(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo)

	//No config, no error
	err = svc.DeleteConfig()
	assert.Nil(t, err)

	nodeCfg := NewNodeConfig(cfg)
	_, err = svc.CreateConfig(nodeCfg)
	assert.Nil(t, err)

	err = svc.DeleteConfig()
	assert.Nil(t, err)

	_, err = svc.GetConfig()
	assert.NotNil(t, err)
}

func TestService_DeleteTenant(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo)
	tenantCfg, err := NewTenantConfig("main", cfg)
	assert.Nil(t, err)
	tid, err := tenantCfg.GetIdentityID()
	assert.Nil(t, err)

	//No config, no error
	err = svc.DeleteTenant(tid)
	assert.Nil(t, err)

	_, err = svc.CreateTenant(tenantCfg)
	assert.Nil(t, err)

	err = svc.DeleteTenant(tid)
	assert.Nil(t, err)

	_, err = svc.GetTenant(tid)
	assert.NotNil(t, err)
}
