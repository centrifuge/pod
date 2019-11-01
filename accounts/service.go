package accounts

// Service defines the accounts service.
type Service interface {
	GetAccount(id []byte) (Account, error)
	CreateAccount(Account) (Account, error)
	UpdateAccount(Account) (Account, error)
	DeleteAccount(id []byte) error
}

type service struct {
	repo repository
}

func newService(repo repository) Service {
	return service{repo: repo}
}

// GetAccount returns the account associated with id.
func (s service) GetAccount(id []byte) (Account, error) {
	return s.repo.GetAccount(id)
}

// CreateAccount creates a new Account.
func (s service) CreateAccount(acc Account) (Account, error) {
	err := s.repo.CreateAccount(acc.AccountID(), acc)
	return acc, err
}

// UpdateAccount updates a specific account.
// returns error if the account is missing in DB.
func (s service) UpdateAccount(acc Account) (Account, error) {
	err := s.repo.UpdateAccount(acc.AccountID(), acc)
	return acc, err
}

// DeleteAccount deletes the account from the DB.
func (s service) DeleteAccount(id []byte) error {
	return s.repo.DeleteAccount(id)
}
