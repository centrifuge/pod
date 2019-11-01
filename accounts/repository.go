package accounts

import (
	"github.com/centrifuge/go-centrifuge/storage"
)

const accountPrefix string = "account-"

type repository struct {
	db storage.Repository
}

func newRepository(db storage.Repository) repository {
	db.Register(new(account))
	return repository{db: db}
}

func getAccountKey(id []byte) []byte {
	return append([]byte(accountPrefix), id...)
}

// GetAccount returns the account associated with account ID
func (r repository) GetAccount(id []byte) (Account, error) {
	key := getAccountKey(id)
	model, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}
	return model.(Account), nil
}

// Create creates the account if not present in the DB.
// should error out if the account exists.
func (r repository) CreateAccount(id []byte, account Account) error {
	key := getAccountKey(id)
	return r.db.Create(key, account)
}

// Update strictly updates the account.
// Will error out when the account doesn't exist in the DB.
func (r repository) UpdateAccount(id []byte, account Account) error {
	key := getAccountKey(id)
	return r.db.Update(key, account)
}

// Delete deletes account.
// Will not error out when account doesn't exists in DB.
func (r repository) DeleteAccount(id []byte) error {
	key := getAccountKey(id)
	return r.db.Delete(key)
}
