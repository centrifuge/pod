package identity

import (
	"encoding/base64"
	"fmt"
	"math/big"
)

const (
	ACTION_CREATE       = "create"
	ACTION_ADDKEY       = "addkey"
	KEY_TYPE_PEERID     = 1
	KEY_TYPE_SIGNATURE  = 2
	KEY_TYPE_ENCRYPTION = 3
)

type Identity interface {
	String() string
	GetCentrifugeId() []byte
	CentrifugeIdString() string
	CentrifugeIdB48() [48]byte
	CentrifugeIdBigInt() *big.Int
	SetCentrifugeId(b []byte) error
	GetCurrentP2PKey() (ret string, err error)
	GetLastKeyForType(keyType int) (key []byte, err error)
	AddKeyToIdentity(keyType int, key []byte) (confirmations chan *WatchIdentity, err error)
	CheckIdentityExists() (exists bool, err error)
}

type WatchIdentity struct {
	Identity Identity
	Error    error
}

// IdentityService is used to fetch identities
type IdentityService interface {
	LookupIdentityForId(centrifugeId []byte) (id Identity, err error)
	CreateIdentity(centrifugeId []byte) (id Identity, confirmations chan *WatchIdentity, err error)
	CheckIdentityExists(centrifugeId []byte) (exists bool, err error)
}

// CentrifugeIdStringToSlice takes a string and decodes it using base64 to convert it into a slice
// of length 32.
func CentrifugeIdStringToSlice(s string) (id []byte, err error) {
	id, err = base64.StdEncoding.DecodeString(s)
	if err != nil {
		return []byte{}, err
	}
	if len(id) != 32 {
		return []byte{}, fmt.Errorf("CentrifugeId has invalid length [%d]", len(id))
	}
	return id, nil
}
