package identity

import (
	"encoding/base64"
	"fmt"
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
	GetCentrifugeId() string
	GetLastB58KeyForType(keyType int) (ret string, err error)
	AddKeyToIdentity(keyType int, confirmations chan<- *Identity) (err error)
	CheckIdentityExists() (exists bool, err error)
}

// IdentityService is used to fetch identities
type IdentityService interface {
	LookupIdentityForId(centrifugeId []byte) (id Identity, err error)
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
