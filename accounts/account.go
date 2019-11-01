package accounts

import (
	"encoding/json"
	"reflect"

	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/ipfs/go-log"
)

var accLog = log.Logger("accounts")

type account struct {
	ID       byteutils.HexBytes `json:"id"`
	Secret   byteutils.HexBytes `json:"secret"`
	SS58Addr string             `json:"ss_58_address"`
}

// Account represents a single account on Centrifuge chain.
type Account interface {
	storage.Model
	AccountID() byteutils.HexBytes
	SS58Address() string
}

// NewAccount returns a new Account.
func NewAccount(id, secret byteutils.HexBytes, ss58Addr string) Account {
	return &account{
		ID:       id,
		Secret:   secret,
		SS58Addr: ss58Addr,
	}
}

// AccountID returns the Public key of the account
func (a *account) AccountID() byteutils.HexBytes {
	return a.ID
}

// SS58Address returns the base58 address of the account.
func (a *account) SS58Address() string {
	return a.SS58Addr
}

// Type Returns the underlying type of the Model
func (a *account) Type() reflect.Type {
	return reflect.TypeOf(a)
}

// JSON return the json representation of the model
func (a *account) JSON() ([]byte, error) {
	return json.Marshal(a)
}

// FromJSON initialize the model with a json
func (a *account) FromJSON(data []byte) error {
	return json.Unmarshal(data, a)
}
