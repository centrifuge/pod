package configstore

import (
	logging "github.com/ipfs/go-log"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
)

// ProtocolSetter sets the protocol on host for the centID
type ProtocolSetter interface {
	InitProtocolForDID(identity.DID)
}

type service struct {
	repo                 Repository
	configStore          config.Service
	dispatcher           jobs.Dispatcher
	log                  *logging.ZapEventLogger
	protocolSetterFinder func() ProtocolSetter
}

// NewService returns an implementation of the config.Service
func NewService(repository Repository) config.Service {
	return &service{repo: repository}
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

func (s service) CreateConfig(config config.Configuration) (config.Configuration, error) {
	_, err := s.repo.GetConfig()
	if err != nil {
		return config, s.repo.CreateConfig(config)
	}

	return config, s.repo.UpdateConfig(config)
}

func (s service) CreateAccount(data config.Account) (config.Account, error) {
	id := data.GetIdentity()
	return data, s.repo.CreateAccount(id[:], data)
}

// generateAccountKeys generates signing keys
//func generateAccountKeys(keystore string, acc *Account, did identity.DID) (*Account, error) {
//	acc.IdentityID = did[:]
//	sPub, err := createKeyPath(keystore, did, signingPubKeyName)
//	if err != nil {
//		return nil, err
//	}
//	sPriv, err := createKeyPath(keystore, did, signingPrivKeyName)
//	if err != nil {
//		return nil, err
//	}
//	acc.SigningKeyPair = KeyPair{
//		Pub: sPub,
//		Pvt: sPriv,
//	}
//	err = crypto.GenerateSigningKeyPair(acc.SigningKeyPair.Pub, acc.SigningKeyPair.Pvt, crypto.CurveEd25519)
//	if err != nil {
//		return nil, err
//	}
//
//	return acc, nil
//}
//
//func createKeyPath(keyStorepath string, did identity.DID, keyName string) (string, error) {
//	tdir := fmt.Sprintf("%s/%s", keyStorepath, did.String())
//	// create account specific key dir
//	if _, err := os.Stat(tdir); os.IsNotExist(err) {
//		err := os.MkdirAll(tdir, os.ModePerm)
//		if err != nil {
//			return "", err
//		}
//	}
//	return fmt.Sprintf("%s/%s", tdir, keyName), nil
//}

func (s service) UpdateAccount(data config.Account) (config.Account, error) {
	id := data.GetIdentity()
	return data, s.repo.UpdateAccount(id[:], data)
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
