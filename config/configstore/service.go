package configstore

import (
	"context"
	"fmt"
	"os"

	"github.com/centrifuge/go-centrifuge/crypto"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
)

const (
	signingPubKeyName  = "signingKey.pub.pem"
	signingPrivKeyName = "signingKey.key.pem"
	ethAuthPubKeyName  = "ethauth.pub.pem"
	ethAuthPrivKeyName = "ethauth.key.pem"
)

// ProtocolSetter sets the protocol on host for the centID
type ProtocolSetter interface {
	InitProtocolForCID(CID identity.CentID)
}

type service struct {
	repo                 repository
	idService            identity.Service
	protocolSetterFinder func() ProtocolSetter
}

// DefaultService returns an implementation of the config.Service
func DefaultService(repository repository, idService identity.Service) config.Service {
	return &service{repo: repository, idService: idService}
}

func (s service) GetConfig() (config.Configuration, error) {
	return s.repo.GetConfig()
}

func (s service) GetAccount(identifier []byte) (config.Account, error) {
	return s.repo.GetAccount(identifier)
}

func (s service) GetAllAccounts() ([]config.Account, error) {
	return s.repo.GetAllAccounts()
}

func (s service) CreateConfig(data config.Configuration) (config.Configuration, error) {
	_, err := s.repo.GetConfig()
	if err != nil {
		return data, s.repo.CreateConfig(data)
	}
	return data, s.repo.UpdateConfig(data)
}

func (s service) CreateAccount(data config.Account) (config.Account, error) {
	id, err := data.GetIdentityID()
	if err != nil {
		return nil, err
	}
	return data, s.repo.CreateAccount(id, data)
}

func (s service) GenerateAccount() (config.Account, error) {
	nc, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	// copy the main account for basic settings
	mtc, err := NewAccount(nc.GetEthereumDefaultAccountName(), nc)
	if nil != err {
		return nil, err
	}
	ctx, err := contextutil.NewCentrifugeContext(context.Background(), mtc)
	if err != nil {
		return nil, err
	}

	id, confirmations, err := s.idService.CreateIdentity(ctx, identity.RandomCentID())
	if err != nil {
		return nil, err
	}
	<-confirmations

	// copy the main account again to create the new account
	tc, err := NewAccount(nc.GetEthereumDefaultAccountName(), nc)
	if err != nil {
		return nil, err
	}

	CID := id.CentID()
	tc, err = generateAccountKeys(nc.GetAccountsKeystore(), tc.(*Account), CID)
	if err != nil {
		return nil, err
	}

	// minor hack to set same p2p keys as node to account: Set the new account ID to copy of main account and create p2p keys
	mtcc := mtc.(*Account)
	mtcc.IdentityID = CID[:]
	err = s.idService.AddKeyFromConfig(mtcc, identity.KeyPurposeP2P)
	if err != nil {
		return nil, err
	}

	err = s.idService.AddKeyFromConfig(tc, identity.KeyPurposeSigning)
	if err != nil {
		return nil, err
	}

	err = s.idService.AddKeyFromConfig(tc, identity.KeyPurposeEthMsgAuth)
	if err != nil {
		return nil, err
	}

	err = s.repo.CreateAccount(CID[:], tc)
	if err != nil {
		return nil, err
	}

	// initiate network handling
	s.protocolSetterFinder().InitProtocolForCID(CID)
	return tc, nil
}

func generateAccountKeys(keystore string, tc *Account, CID identity.CentID) (*Account, error) {
	tc.IdentityID = CID[:]
	Pub, err := createKeyPath(keystore, CID, signingPubKeyName)
	if err != nil {
		return nil, err
	}
	Priv, err := createKeyPath(keystore, CID, signingPrivKeyName)
	if err != nil {
		return nil, err
	}
	tc.SigningKeyPair = KeyPair{
		Pub:  Pub,
		Priv: Priv,
	}
	ePub, err := createKeyPath(keystore, CID, ethAuthPubKeyName)
	if err != nil {
		return nil, err
	}
	ePriv, err := createKeyPath(keystore, CID, ethAuthPrivKeyName)
	if err != nil {
		return nil, err
	}
	tc.EthAuthKeyPair = KeyPair{
		Pub:  ePub,
		Priv: ePriv,
	}
	err = crypto.GenerateSigningKeyPair(tc.SigningKeyPair.Pub, tc.SigningKeyPair.Priv, "ed25519")
	if err != nil {
		return nil, err
	}
	err = crypto.GenerateSigningKeyPair(tc.EthAuthKeyPair.Pub, tc.EthAuthKeyPair.Priv, "secp256k1")
	if err != nil {
		return nil, err
	}
	return tc, nil
}

func createKeyPath(keyStorepath string, CID identity.CentID, keyName string) (string, error) {
	tdir := fmt.Sprintf("%s/%s", keyStorepath, CID.String())
	// create account specific key dir
	if _, err := os.Stat(tdir); os.IsNotExist(err) {
		err := os.MkdirAll(tdir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%s/%s", tdir, keyName), nil
}

func (s service) UpdateAccount(data config.Account) (config.Account, error) {
	id, err := data.GetIdentityID()
	if err != nil {
		return nil, err
	}
	return data, s.repo.UpdateAccount(id, data)
}

func (s service) DeleteAccount(identifier []byte) error {
	return s.repo.DeleteAccount(identifier)
}

// RetrieveConfig retrieves system config giving priority to db stored config
func RetrieveConfig(dbOnly bool, ctx map[string]interface{}) (config.Configuration, error) {
	var cfg config.Configuration
	var err error
	if cfgService, ok := ctx[config.BootstrappedConfigStorage].(config.Service); ok {
		// may be we need a way to detect a corrupted db here
		cfg, err = cfgService.GetConfig()
		if err != nil {
			apiLog.Warningf("could not load config from db: %v", err)
		}
		return cfg, nil
	}

	// we have to allow loading from file in case this is coming from create config cmd where we don't add configs to db
	if _, ok := ctx[bootstrap.BootstrappedConfig]; ok && cfg == nil && !dbOnly {
		cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	} else {
		return nil, errors.NewTypedError(config.ErrConfigRetrieve, err)
	}
	return cfg, nil
}
