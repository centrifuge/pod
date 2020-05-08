package configstore

import (
	"context"
	"fmt"
	"os"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/ipfs/go-log"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/ethereum/go-ethereum/common"
)

const (
	signingPubKeyName  = "signingKey.pub.pem"
	signingPrivKeyName = "signingKey.key.pem"
)

var accLog = log.Logger("accounts")

// ProtocolSetter sets the protocol on host for the centID
type ProtocolSetter interface {
	InitProtocolForDID(DID *identity.DID)
}

type service struct {
	repo                 Repository
	idFactory            identity.Factory
	idService            identity.Service
	protocolSetterFinder func() ProtocolSetter
}

// DefaultService returns an implementation of the config.Service
func DefaultService(repository Repository, idService identity.Service) config.Service {
	return &service{repo: repository, idService: idService}
}

func (s service) GetConfig() (config.Configuration, error) {
	return s.repo.GetConfig()
}

func (s service) GetAccount(identifier []byte) (config.Account, error) {
	return s.repo.GetAccount(identifier)
}

func (s service) GetAccounts() ([]config.Account, error) {
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
	id := data.GetIdentityID()
	return data, s.repo.CreateAccount(id, data)
}

func (s service) GenerateAccount(cacc config.CentChainAccount) (config.Account, error) {
	if cacc.ID == "" || cacc.Secret == "" || cacc.SS58Addr == "" {
		return nil, errors.New("Centrifuge Chain account is required")
	}

	nc, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	// copy the main account for basic settings
	acc, err := NewAccount(nc.GetEthereumDefaultAccountName(), nc)
	if nil != err {
		return nil, err
	}

	acc.(*Account).CentChainAccount = cacc
	ctx, err := contextutil.New(context.Background(), acc)
	if err != nil {
		return nil, err
	}

	acci := acc.(*Account)

	addr, err := s.idFactory.CalculateIdentityAddress(ctx)
	if err != nil {
		return nil, err
	}

	acc, err = generateAccountKeys(nc.GetAccountsKeystore(), acci, addr)
	if err != nil {
		return nil, err
	}

	DID, err := s.idFactory.CreateIdentity(ctx)
	if err != nil {
		return nil, err
	}

	acci.IdentityID = DID[:]

	err = s.repo.CreateAccount(DID[:], acc)
	if err != nil {
		return nil, err
	}

	// initiate network handling
	s.protocolSetterFinder().InitProtocolForDID(DID)
	return acc, nil
}

// generateAccountKeys generates signing keys
func generateAccountKeys(keystore string, acc *Account, addr *common.Address) (*Account, error) {
	sPub, err := createKeyPath(keystore, addr, signingPubKeyName)
	if err != nil {
		return nil, err
	}
	sPriv, err := createKeyPath(keystore, addr, signingPrivKeyName)
	if err != nil {
		return nil, err
	}
	acc.SigningKeyPair = KeyPair{
		Pub: sPub,
		Pvt: sPriv,
	}
	err = crypto.GenerateSigningKeyPair(acc.SigningKeyPair.Pub, acc.SigningKeyPair.Pvt, crypto.CurveSecp256K1)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

func createKeyPath(keyStorepath string, addr *common.Address, keyName string) (string, error) {
	tdir := fmt.Sprintf("%s/%s", keyStorepath, addr.String())
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
	id := data.GetIdentityID()
	return data, s.repo.UpdateAccount(id, data)
}

func (s service) DeleteAccount(identifier []byte) error {
	return s.repo.DeleteAccount(identifier)
}

// Sign signs the payload using the account's secret key and returns a signature.
func (s service) Sign(accountID, payload []byte) (*coredocumentpb.Signature, error) {
	acc, err := s.GetAccount(accountID)
	if err != nil {
		return nil, err
	}

	return acc.SignMsg(payload)
}

// RetrieveConfig retrieves system config giving priority to db stored config
func RetrieveConfig(dbOnly bool, ctx map[string]interface{}) (config.Configuration, error) {
	var cfg config.Configuration
	var err error
	if cfgService, ok := ctx[config.BootstrappedConfigStorage].(config.Service); ok {
		// may be we need a way to detect a corrupted db here
		cfg, err = cfgService.GetConfig()
		if err != nil {
			accLog.Warningf("could not load config from db: %v", err)
		}
		return cfg, nil
	}

	// we have to allow loading from file in case this is coming from create config cmd where we don't add configs to db
	if _, ok := ctx[bootstrap.BootstrappedConfig]; ok && !dbOnly {
		cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	} else {
		return nil, errors.NewTypedError(config.ErrConfigRetrieve, err)
	}
	return cfg, nil
}
