package configstore

import (
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/storage"
)

const (
	configPrefix  string = "config"
	accountPrefix string = "account-"
)

// Repository defines the required methods for the config repository.
type Repository interface {
	// RegisterAccount registers account in DB
	RegisterAccount(acc config.Account)

	// RegisterConfig registers node config in DB
	RegisterConfig(cfg config.Configuration)

	// GetAccount returns the Account associated with account ID
	GetAccount(id []byte) (config.Account, error)

	// GetConfig returns the node config model
	GetConfig() (config.Configuration, error)

	// GetAllAccounts returns a list of all account models in the config DB
	GetAllAccounts() ([]config.Account, error)

	// CreateAccount creates the account model if not present in the DB.
	// should error out if the config exists.
	CreateAccount(id []byte, acc config.Account) error

	// CreateConfig creates the node config model if not present in the DB.
	// should error out if the config exists.
	CreateConfig(cfg config.Configuration) error

	// UpdateAccount strictly updates the account model.
	// Will error out when the account model doesn't exist in the DB.
	UpdateAccount(id []byte, acc config.Account) error

	// UpdateConfig strictly updates the node config model.
	// Will error out when the config model doesn't exist in the DB.
	UpdateConfig(cfg config.Configuration) error

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

func getAccountKey(id []byte) []byte {
	return append([]byte(accountPrefix), id...)
}

func getConfigKey() []byte {
	return []byte(configPrefix)
}

// NewDBRepository creates instance of Config repository
func NewDBRepository(db storage.Repository) Repository {
	return &repo{db: db}
}

// RegisterAccount registers account in DB
func (r *repo) RegisterAccount(config config.Account) {
	r.db.Register(config)
}

// RegisterConfig registers node config in DB
func (r *repo) RegisterConfig(config config.Configuration) {
	r.db.Register(config)
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

// CreateAccount creates the account model if not present in the DB.
// should error out if the config exists.
func (r *repo) CreateAccount(id []byte, account config.Account) error {
	key := getAccountKey(id)
	return r.db.Create(key, account)
}

// CreateConfig creates the node config model if not present in the DB.
// should error out if the config exists.
func (r *repo) CreateConfig(config config.Configuration) error {
	key := getConfigKey()
	return r.db.Create(key, config)
}

// UpdateAccount strictly updates the account model.
// Will error out when the config model doesn't exist in the DB.
func (r *repo) UpdateAccount(id []byte, account config.Account) error {
	key := getAccountKey(id)
	return r.db.Update(key, account)
}

// UpdateConfig strictly updates the node config model.
// Will error out when the config model doesn't exist in the DB.
func (r *repo) UpdateConfig(config config.Configuration) error {
	key := getConfigKey()
	return r.db.Update(key, config)
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
