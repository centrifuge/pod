package configstore

import (
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

// ProtocolSetter sets the protocol on host for the centID
type ProtocolSetter interface {
	InitProtocolForDID(*types.AccountID)
}

type service struct {
	log                  *logging.ZapEventLogger
	repo                 Repository
	dispatcher           jobs.Dispatcher
	protocolSetterFinder func() ProtocolSetter
}

// NewService returns an implementation of the config.Service
func NewService(
	repo Repository,
	dispatcher jobs.Dispatcher,
	protocolSetterFinder func() ProtocolSetter,
) config.Service {
	log := logging.Logger("configstore_service")

	return &service{
		log,
		repo,
		dispatcher,
		protocolSetterFinder,
	}
}

func (s service) GetConfig() (config.Configuration, error) {
	return s.repo.GetConfig()
}

func (s service) GetNodeAdmin() (config.NodeAdmin, error) {
	return s.repo.GetNodeAdmin()
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

func (s service) CreateNodeAdmin(nodeAdmin config.NodeAdmin) (config.NodeAdmin, error) {
	return nodeAdmin, s.repo.CreateNodeAdmin(nodeAdmin)
}

func (s service) CreateAccount(data config.Account) (config.Account, error) {
	id := data.GetIdentity()
	return data, s.repo.CreateAccount(id.ToBytes(), data)
}

func (s service) UpdateAccount(data config.Account) (config.Account, error) {
	id := data.GetIdentity()
	return data, s.repo.UpdateAccount(id.ToBytes(), data)
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
