package configstore

import (
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
)

type service struct {
	repo Repository
}

// NewService returns an implementation of the config.Service
func NewService(repo Repository) config.Service {
	return &service{
		repo,
	}
}

func (s service) CreateConfig(config config.Configuration) error {
	_, err := s.repo.GetConfig()
	if err != nil {
		return s.repo.CreateConfig(config)
	}

	return s.repo.UpdateConfig(config)
}

func (s service) CreateNodeAdmin(nodeAdmin config.NodeAdmin) error {
	return s.repo.CreateNodeAdmin(nodeAdmin)
}

func (s service) CreateAccount(account config.Account) error {
	return s.repo.CreateAccount(account)
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

func (s service) UpdateNodeAdmin(nodeAdmin config.NodeAdmin) error {
	return s.repo.UpdateNodeAdmin(nodeAdmin)
}

func (s service) UpdateAccount(account config.Account) error {
	return s.repo.UpdateAccount(account)
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
