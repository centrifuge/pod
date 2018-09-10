package identity

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/big"

	"sync"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
)

const (
	CentIdByteLength     = 6
	ActionCreate         = "create"
	ActionAddKey         = "addkey"
	KeyPurposeP2p        = 1
	KeyPurposeSigning    = 2
	KeyPurposeEthMsgAuth = 3
)

// idService is a default implementation of the Service
var idService Service

// once guard the idService from multiple sets
var once sync.Once

// SetIdentityService sets the srv to default identity service
func SetIdentityService(srv Service) {
	once.Do(func() {
		idService = srv
	})
}

// GetIdentityService returns the default identity service
// panics if service is not set
func GetIdentityService() Service {
	if idService == nil {
		log.Fatal("identity service not initialised yet")
	}

	return idService
}

// Identity defines an Identity on chain
type Identity interface {
	String() string
	GetCentrifugeID() []byte
	CentrifugeIDString() string
	// todo convert this to a - type CentrifugeId [CentIdByteLength]byte
	CentrifugeIDBytes() [CentIdByteLength]byte
	CentrifugeIDBigInt() *big.Int
	SetCentrifugeID(b []byte) error
	GetCurrentP2PKey() (ret string, err error)
	GetLastKeyForPurpose(keyPurpose int) (key []byte, err error)
	AddKeyToIdentity(keyPurpose int, key []byte) (confirmations chan *WatchIdentity, err error)
	CheckIdentityExists() (exists bool, err error)
	FetchKey(key []byte) (Key, error)
}

// Key defines a single ERC725 identity key
type Key interface {
	GetKey() [32]byte
	GetPurposes() []*big.Int
	GetRevokedAt() *big.Int
}

type WatchIdentity struct {
	Identity Identity
	Error    error
}

// Service is used to fetch identities
type Service interface {
	LookupIdentityForID(centrifugeId []byte) (id Identity, err error)
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

// GetClientP2PURL returns the p2p url associated with the centID
func GetClientP2PURL(centID []byte) (url string, err error) {
	target, err := idService.LookupIdentityForID(centID)
	if err != nil {
		return url, errors.Wrap(err, "error fetching receiver identity")
	}

	p2pKey, err := target.GetCurrentP2PKey()
	if err != nil {
		return url, errors.Wrap(err, "error fetching p2p key")
	}

	return fmt.Sprintf("/ipfs/%s", p2pKey), nil
}

// GetClientsP2PURLs returns p2p urls associated with each centIDs
// will error out at first failure
func GetClientsP2PURLs(centIDs [][]byte) ([]string, error) {
	var p2pURLs []string
	for _, id := range centIDs {
		url, err := GetClientP2PURL(id)
		if err != nil {
			return nil, err
		}

		p2pURLs = append(p2pURLs, url)
	}

	return p2pURLs, nil
}

// GetIdentityKey returns the key for provided identity
func GetIdentityKey(identity, pubKey []byte) (keyInfo Key, err error) {
	id, err := idService.LookupIdentityForID(identity)
	if err != nil {
		return keyInfo, err
	}

	key, err := id.FetchKey(pubKey)
	if err != nil {
		return keyInfo, err
	}

	if tools.IsEmptyByte32(key.GetKey()) {
		return keyInfo, fmt.Errorf(fmt.Sprintf("key not found for identity: %x", identity))
	}

	return key, nil
}

// ValidateKey checks if a given key is valid for the given centrifugeID.
func ValidateKey(centrifugeId []byte, key []byte) (valid bool, err error) {
	idKey, err := GetIdentityKey(centrifugeId, key)
	if err != nil {
		return false, err
	}

	if !bytes.Equal(key, tools.Byte32ToSlice(idKey.GetKey())) {
		return false, fmt.Errorf(fmt.Sprintf("[Key: %x] Key doesn't match", idKey.GetKey()))
	}

	if !tools.ContainsBigIntInSlice(big.NewInt(KeyPurposeSigning), idKey.GetPurposes()) {
		return false, fmt.Errorf(fmt.Sprintf("[Key: %x] Key doesn't have purpose [%d]", idKey.GetKey(), KeyPurposeSigning))
	}

	// TODO Check if revokation block happened before the timeframe of the document signing, for historical validations
	if idKey.GetRevokedAt().Cmp(big.NewInt(0)) != 0 {
		return false, fmt.Errorf(fmt.Sprintf("[Key: %x] Key is currently revoked since block [%d]", idKey.GetKey(), idKey.GetRevokedAt()))
	}

	return true, nil
}
