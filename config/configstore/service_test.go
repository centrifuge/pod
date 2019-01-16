// +build unit

package configstore

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/testingutils/commons"

	"github.com/stretchr/testify/assert"
)

func TestService_GetConfig_NoConfig(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo, idService)
	cfg, err := svc.GetConfig()
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
}

func TestService_GetConfig(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo, idService)
	nodeCfg := NewNodeConfig(cfg)
	err = repo.CreateConfig(nodeCfg)
	assert.Nil(t, err)
	cfg, err := svc.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
}

func TestService_GetTenant_NoTenant(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo, idService)
	cfg, err := svc.GetTenant([]byte("0x123456789"))
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
}

func TestService_GetTenant(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo, idService)
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
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo, idService)
	nodeCfg := NewNodeConfig(cfg)
	cfgpb, err := svc.CreateConfig(nodeCfg)
	assert.Nil(t, err)
	assert.Equal(t, nodeCfg.GetStoragePath(), cfgpb.GetStoragePath())

	//Config already exists
	_, err = svc.CreateConfig(nodeCfg)
	assert.Nil(t, err)
}

func TestService_CreateTenant(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo, idService)
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

func TestService_UpdateTenant(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo, idService)
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

func TestService_DeleteTenant(t *testing.T) {
	idService := &testingcommons.MockIDService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterTenant(&TenantConfig{})
	svc := DefaultService(repo, idService)
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

func TestGenerateTenantKeys(t *testing.T) {
	tc, err := generateTenantKeys("/tmp/tenants/", &TenantConfig{}, identity.RandomCentID())
	assert.Nil(t, err)
	assert.NotNil(t, tc.SigningKeyPair)
	assert.NotNil(t, tc.EthAuthKeyPair)
	_, err = os.Stat(tc.SigningKeyPair.Pub)
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(tc.SigningKeyPair.Priv)
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(tc.EthAuthKeyPair.Pub)
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(tc.EthAuthKeyPair.Priv)
	assert.False(t, os.IsNotExist(err))
}
