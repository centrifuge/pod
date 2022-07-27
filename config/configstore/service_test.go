//go:build unit

package configstore

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/identity"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestService_GetConfig_NoConfig(t *testing.T) {
	idService := &testingcommons.MockIdentityService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterConfig(&NodeConfig{})
	svc := DefaultService(repo, idService)
	cfg, err := svc.GetConfig()
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
}

func TestService_GetConfig(t *testing.T) {
	idService := &testingcommons.MockIdentityService{}
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

func TestService_GetAccount_NoAccount(t *testing.T) {
	idService := &testingcommons.MockIdentityService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)
	cfg, err := svc.GetAccount([]byte("0x123456789"))
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
}

func TestService_GetAccount(t *testing.T) {
	idService := &testingcommons.MockIdentityService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)
	accountCfg, err := NewAccount("main", cfg)
	assert.Nil(t, err)
	accID := accountCfg.GetIdentityID()
	err = repo.CreateAccount(accID, accountCfg)
	assert.Nil(t, err)
	cfg, err := svc.GetAccount(accID)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
}

func TestService_CreateConfig(t *testing.T) {
	idService := &testingcommons.MockIdentityService{}
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

func TestService_Createaccount(t *testing.T) {
	idService := &testingcommons.MockIdentityService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)
	accountCfg, err := NewAccount("main", cfg)
	assert.Nil(t, err)
	newCfg, err := svc.CreateAccount(accountCfg)
	assert.Nil(t, err)
	i := newCfg.GetIdentityID()
	accID := accountCfg.GetIdentityID()
	assert.Equal(t, accID, i)

	//account already exists
	_, err = svc.CreateAccount(accountCfg)
	assert.NotNil(t, err)
}

func TestService_Updateaccount(t *testing.T) {
	idService := &testingcommons.MockIdentityService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)
	accountCfg, err := NewAccount("main", cfg)
	assert.NoError(t, err)

	// account doesn't exist
	newCfg, err := svc.UpdateAccount(accountCfg)
	assert.NotNil(t, err)

	newCfg, err = svc.CreateAccount(accountCfg)
	assert.Nil(t, err)
	i := newCfg.GetIdentityID()
	accID := accountCfg.GetIdentityID()
	assert.Equal(t, accID, i)

	acc := accountCfg.(*Account)
	acc.EthereumDefaultAccountName = "other"
	newCfg, err = svc.UpdateAccount(accountCfg)
	assert.Nil(t, err)
	assert.Equal(t, acc.EthereumDefaultAccountName, newCfg.GetEthereumDefaultAccountName())
}

func TestService_Deleteaccount(t *testing.T) {
	idService := &testingcommons.MockIdentityService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)
	accountCfg, err := NewAccount("main", cfg)
	assert.Nil(t, err)
	accID := accountCfg.GetIdentityID()

	//No config, no error
	err = svc.DeleteAccount(accID)
	assert.Nil(t, err)

	_, err = svc.CreateAccount(accountCfg)
	assert.Nil(t, err)

	err = svc.DeleteAccount(accID)
	assert.Nil(t, err)

	_, err = svc.GetAccount(accID)
	assert.NotNil(t, err)
}

func TestGenerateaccountKeys(t *testing.T) {
	DID, err := identity.NewDIDFromString("0xDcF1695B8a0df44c60825eCD0A8A833dA3875F13")
	assert.NoError(t, err)
	tc, err := generateAccountKeys("/tmp/accounts/", &Account{}, DID)
	assert.Nil(t, err)
	assert.NotNil(t, tc.SigningKeyPair)
	_, err = os.Stat(tc.SigningKeyPair.Pub)
	assert.False(t, os.IsNotExist(err))
	_, err = os.Stat(tc.SigningKeyPair.Pvt)
	assert.False(t, os.IsNotExist(err))
}

func TestService_Sign(t *testing.T) {
	idService := &testingcommons.MockIdentityService{}
	repo, _, err := getRandomStorage()
	assert.Nil(t, err)
	repo.RegisterAccount(&Account{})
	svc := DefaultService(repo, idService)

	payload := utils.RandomSlice(32)
	accountID := utils.RandomSlice(20)

	// missing account
	_, err = svc.Sign(accountID, payload)
	assert.Error(t, err)

	// success
	accountCfg, err := NewAccount("main", cfg)
	assert.Nil(t, err)
	acc, err := svc.CreateAccount(accountCfg)
	assert.NoError(t, err)
	accountID = acc.GetIdentityID()
	sig, err := svc.Sign(accountID, payload)
	assert.NoError(t, err)
	assert.Equal(t, sig.SignerId, accountID)
	assert.True(t, crypto.VerifyMessage(sig.PublicKey, payload, sig.Signature, crypto.CurveSecp256K1))
}
