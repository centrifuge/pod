package identity

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	centrifugeED25519 "github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/ed25519"
)

const (
	CentIdByteLength     = 6
	ActionCreate         = "create"
	ActionAddKey         = "addkey"
	KeyPurposeP2p        = 1
	KeyPurposeSigning    = 2
	KeyPurposeEthMsgAuth = 3
)

type Identity interface {
	String() string
	GetCentrifugeId() []byte
	CentrifugeIdString() string
	CentrifugeIdBytes() [CentIdByteLength]byte
	CentrifugeIdBigInt() *big.Int
	SetCentrifugeId(b []byte) error
	GetCurrentP2PKey() (ret string, err error)
	GetLastKeyForPurpose(keyPurpose int) (key []byte, err error)
	AddKeyToIdentity(keyPurpose int, key []byte) (confirmations chan *WatchIdentity, err error)
	CheckIdentityExists() (exists bool, err error)
	FetchKey() (IdentityKey, error)
}

type IdentityKey interface {
	GetKey() [32]byte
	GetPurposes() []*big.Int
	GetRevokedAt() *big.Int
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
	if len(id) != CentIdByteLength {
		return []byte{}, fmt.Errorf("CentrifugeId has invalid length [%d]", len(id))
	}
	return id, nil
}

func NewIdentityConfig() (identityConfig *config.IdentityConfig) {
	pk,pvk := centrifugeED25519.GetSigningKeyPairFromConfig()
	identityConfig = &config.IdentityConfig{
		IdentityId: config.Config.GetIdentityId(),
		PublicKey: pk,
		PrivateKey: pvk,
	}
	return
}
