package configstore

import (
	"github.com/centrifuge/pod/config"
)

type service struct {
	repo Repository
}

// NewService returns an implementation of the config.Service
func NewService(repo Repository) config.Service {
	return &service{
		repo: repo,
	}
}

func (s service) CreateConfig(config config.Configuration) error {
	if _, err := s.repo.GetConfig(); err != nil {
		return s.repo.CreateConfig(config)
	}

	return s.repo.UpdateConfig(config)
}

func (s service) CreatePodAdmin(nodeAdmin config.PodAdmin) error {
	if _, err := s.repo.GetPodAdmin(); err != nil {
		return s.repo.CreatePodAdmin(nodeAdmin)
	}

	return s.repo.UpdatePodAdmin(nodeAdmin)
}

func (s service) CreateAccount(account config.Account) error {
	return s.repo.CreateAccount(account)
}

func (s service) CreatePodOperator(podOperator config.PodOperator) error {
	if _, err := s.repo.GetPodOperator(); err != nil {
		return s.repo.CreatePodOperator(podOperator)
	}

	return s.repo.UpdatePodOperator(podOperator)
}

func (s service) GetConfig() (config.Configuration, error) {
	return s.repo.GetConfig()
}

func (s service) GetPodAdmin() (config.PodAdmin, error) {
	return s.repo.GetPodAdmin()
}

func (s service) GetAccount(identifier []byte) (config.Account, error) {
	return s.repo.GetAccount(identifier)
}

func (s service) GetAccounts() ([]config.Account, error) {
	return s.repo.GetAllAccounts()
}

func (s service) GetPodOperator() (config.PodOperator, error) {
	return s.repo.GetPodOperator()
}

func (s service) UpdateAccount(account config.Account) error {
	return s.repo.UpdateAccount(account)
}

func (s service) DeleteAccount(identifier []byte) error {
	return s.repo.DeleteAccount(identifier)
}
