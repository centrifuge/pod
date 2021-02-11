package configstore

import (
	"fmt"
	"os"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
)

const (
	signingPubKeyName  = "signingKey.pub.pem"
	signingPrivKeyName = "signingKey.key.pem"
)

// ProtocolSetter sets the protocol on host for the centID
type ProtocolSetter interface {
	InitProtocolForDID(identity.DID)
}

type service struct {
	repo                 Repository
	idFactory            identity.Factory
	idFactoryV2          identity.Factory
	idService            identity.Service
	dispatcher           jobs.Dispatcher
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

func (s service) GenerateAccountAsync(cacc config.CentChainAccount) (didBytes []byte, jobID []byte, err error) {
	if cacc.ID == "" || cacc.Secret == "" || cacc.SS58Addr == "" {
		return nil, nil, errors.New("Centrifuge Chain account is required")
	}

	nc, err := s.GetConfig()
	if err != nil {
		return nil, nil, err
	}

	// copy the main account for basic settings
	acc, err := NewAccount(nc.GetEthereumDefaultAccountName(), nc)
	if nil != err {
		return nil, nil, err
	}
	acc.(*Account).CentChainAccount = cacc
	did, err := s.idFactoryV2.NextIdentityAddress()
	if err != nil {
		return nil, nil, err
	}

	acc, err = generateAccountKeys(nc.GetAccountsKeystore(), acc.(*Account), did)
	if err != nil {
		return nil, nil, err
	}

	err = s.repo.CreateAccount(did[:], acc)
	if err != nil {
		return nil, nil, err
	}

	valid := nc.GetTaskValidDuration()
	jobID, err = StartGenerateIdentityJob(did, s.dispatcher, time.Now().UTC().Add(valid))
	if err != nil {
		return nil, nil, err
	}

	// initiate network handling
	s.protocolSetterFinder().InitProtocolForDID(did)
	return did[:], jobID, nil
}

// generateAccountKeys generates signing keys
func generateAccountKeys(keystore string, acc *Account, did identity.DID) (*Account, error) {
	acc.IdentityID = did[:]
	sPub, err := createKeyPath(keystore, did, signingPubKeyName)
	if err != nil {
		return nil, err
	}
	sPriv, err := createKeyPath(keystore, did, signingPrivKeyName)
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

func createKeyPath(keyStorepath string, did identity.DID, keyName string) (string, error) {
	tdir := fmt.Sprintf("%s/%s", keyStorepath, did.String())
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
