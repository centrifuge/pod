package configstore

import (
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/storage"
)

const (
	configPrefix      string = "config"
	accountPrefix     string = "account-"
	nodeAdminPrefix   string = "node-admin"
	podOperatorPrefix string = "pod-operator"
)

//go:generate mockery --name Repository --structname RepositoryMock --filename repository_mock.go --inpackage

// Repository defines the required methods for the config repository.
type Repository interface {
	// RegisterNodeAdmin registers node admin in DB
	RegisterNodeAdmin(nodeAdmin config.NodeAdmin)

	// RegisterAccount registers account in DB
	RegisterAccount(acc config.Account)

	// RegisterConfig registers node config in DB
	RegisterConfig(cfg config.Configuration)

	// RegisterPodOperator registers pod operator in DB
	RegisterPodOperator(podOperator config.PodOperator)

	// GetNodeAdmin returns the node admin
	GetNodeAdmin() (config.NodeAdmin, error)

	// GetAccount returns the Account associated with account ID
	GetAccount(id []byte) (config.Account, error)

	// GetConfig returns the node config model
	GetConfig() (config.Configuration, error)

	// GetPodOperator returns the pod operator model
	GetPodOperator() (config.PodOperator, error)

	// GetAllAccounts returns a list of all account models in the config DB
	GetAllAccounts() ([]config.Account, error)

	// CreateNodeAdmin stores the node admin in the DB.
	// Should error out if the node admin exists.
	CreateNodeAdmin(nodeAdmin config.NodeAdmin) error

	// CreateAccount creates the account model if not present in the DB.
	// Should error out if the account exists.
	CreateAccount(acc config.Account) error

	// CreateConfig creates the node config model if not present in the DB.
	// Should error out if the config exists.
	CreateConfig(cfg config.Configuration) error

	// CreatePodOperator creates the pod operator model if not present in the DB.
	// Should error out if the pod operator exists.
	CreatePodOperator(podOperator config.PodOperator) error

	// UpdateNodeAdmin strictly updates the node admin model.
	// Will error out when the node admin model doesn't exist in the DB.
	UpdateNodeAdmin(nodeAdmin config.NodeAdmin) error

	// UpdateAccount strictly updates the account model.
	// Will error out when the account model doesn't exist in the DB.
	UpdateAccount(acc config.Account) error

	// UpdateConfig strictly updates the node config model.
	// Will error out when the config model doesn't exist in the DB.
	UpdateConfig(cfg config.Configuration) error

	// UpdatePodOperator strictly updates the pod operator model.
	// Will error out when the pod operator model doesn't exist in the DB.
	UpdatePodOperator(podOperator config.PodOperator) error

	// DeleteAccount deletes account config
	// Will not error out when account model doesn't exist in DB
	DeleteAccount(id []byte) error

	// DeleteConfig deletes node config
	// Will not error out when config model doesn't exist in DB
	DeleteConfig() error
}

type repo struct {
	db storage.Repository
}

func getNodeAdminKey() []byte {
	return []byte(nodeAdminPrefix)
}

func getAccountKey(id []byte) []byte {
	return append([]byte(accountPrefix), id...)
}

func getConfigKey() []byte {
	return []byte(configPrefix)
}

func getPodOperatorKey() []byte {
	return []byte(podOperatorPrefix)
}

// NewDBRepository creates instance of Config repository
func NewDBRepository(db storage.Repository) Repository {
	return &repo{db: db}
}

// RegisterNodeAdmin registers a node admin in DB
func (r *repo) RegisterNodeAdmin(nodeAdmin config.NodeAdmin) {
	r.db.Register(nodeAdmin)
}

// RegisterAccount registers account in DB
func (r *repo) RegisterAccount(account config.Account) {
	r.db.Register(account)
}

// RegisterConfig registers node config in DB
func (r *repo) RegisterConfig(config config.Configuration) {
	r.db.Register(config)
}

// RegisterPodOperator registers pod operator in DB
func (r *repo) RegisterPodOperator(podOperator config.PodOperator) {
	r.db.Register(podOperator)
}

func (r *repo) GetNodeAdmin() (config.NodeAdmin, error) {
	key := getNodeAdminKey()

	model, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}

	return model.(config.NodeAdmin), nil
}

// GetAccount returns the account Document associated with account ID
func (r *repo) GetAccount(id []byte) (config.Account, error) {
	key := getAccountKey(id)
	model, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}
	return model.(config.Account), nil
}

// GetConfig returns the node config model
func (r *repo) GetConfig() (config.Configuration, error) {
	key := getConfigKey()
	model, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}
	return model.(config.Configuration), nil
}

// GetPodOperator returns the pod operator model
func (r *repo) GetPodOperator() (config.PodOperator, error) {
	key := getPodOperatorKey()

	model, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}

	return model.(config.PodOperator), nil
}

// GetAllAccounts iterates over all account entries in DB and returns a list of Models
// If an error occur reading a account, throws a warning and continue
func (r *repo) GetAllAccounts() (accountConfigs []config.Account, err error) {
	models, err := r.db.GetAllByPrefix(accountPrefix)
	if err != nil {
		return nil, err
	}
	for _, acc := range models {
		accountConfigs = append(accountConfigs, acc.(config.Account))
	}
	return accountConfigs, nil
}

// CreateNodeAdmin stores the node admin in the DB.
func (r *repo) CreateNodeAdmin(nodeAdmin config.NodeAdmin) error {
	return r.db.Create(getNodeAdminKey(), nodeAdmin)
}

// CreateAccount creates the account model if not present in the DB.
// should error out if the config exists.
func (r *repo) CreateAccount(account config.Account) error {
	key := getAccountKey(account.GetIdentity().ToBytes())
	return r.db.Create(key, account)
}

// CreateConfig creates the node config model if not present in the DB.
// should error out if the config exists.
func (r *repo) CreateConfig(config config.Configuration) error {
	key := getConfigKey()
	return r.db.Create(key, config)
}

// CreatePodOperator creates the pod operator model if not present in the DB.
// should error out if the config exists.
func (r *repo) CreatePodOperator(podOperator config.PodOperator) error {
	key := getPodOperatorKey()
	return r.db.Create(key, podOperator)
}

// UpdateNodeAdmin strictly updates the node admin model.
// Will error out when the node admin model doesn't exist in the DB.
func (r *repo) UpdateNodeAdmin(nodeAdmin config.NodeAdmin) error {
	return r.db.Update(getNodeAdminKey(), nodeAdmin)
}

// UpdateAccount strictly updates the account model.
// Will error out when the account model doesn't exist in the DB.
func (r *repo) UpdateAccount(account config.Account) error {
	key := getAccountKey(account.GetIdentity().ToBytes())
	return r.db.Update(key, account)
}

// UpdateConfig strictly updates the node config model.
// Will error out when the config model doesn't exist in the DB.
func (r *repo) UpdateConfig(config config.Configuration) error {
	key := getConfigKey()
	return r.db.Update(key, config)
}

// UpdatePodOperator strictly updates the pod operator model.
// Will error out when the config model doesn't exist in the DB.
func (r *repo) UpdatePodOperator(podOperator config.PodOperator) error {
	key := getPodOperatorKey()
	return r.db.Update(key, podOperator)
}

// DeleteAccount deletes account
// Will not error out when config model doesn't  in DB
func (r *repo) DeleteAccount(id []byte) error {
	key := getAccountKey(id)
	return r.db.Delete(key)
}

// DeleteConfig deletes node config
// Will not error out when config model doesn't exist in DB
func (r *repo) DeleteConfig() error {
	key := getConfigKey()
	return r.db.Delete(key)
}
