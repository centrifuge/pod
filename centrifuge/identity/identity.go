package identity

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"

	"github.com/centrifuge/go-centrifuge/centrifuge/ethereum"

	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/keytools/ed25519"
	"github.com/centrifuge/go-centrifuge/centrifuge/keytools/secp256k1"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
)

const (
	CentIDByteLength     = 6
	ActionCreate         = "create"
	ActionAddKey         = "addkey"
	KeyPurposeP2p        = 1
	KeyPurposeSigning    = 2
	KeyPurposeEthMsgAuth = 3
)

type CentID [CentIDByteLength]byte

func NewCentID(centIDBytes []byte) (CentID, error) {
	var centBytes [CentIDByteLength]byte
	if !tools.IsValidByteSliceForLength(centIDBytes, CentIDByteLength) {
		return centBytes, errors.New("invalid length byte slice provided for centId")
	}
	copy(centBytes[:], centIDBytes[:CentIDByteLength])
	return centBytes, nil
}

func NewRandomCentID() CentID {
	ID, _ := NewCentID(tools.RandomSlice(CentIDByteLength))
	return ID
}

func (c CentID) Equal(other CentID) bool {
	for i := range c {
		if c[i] != other[i] {
			return false
		}
	}

	return true
}

func (c CentID) String() string {
	return base64.StdEncoding.EncodeToString(c[:])
}

func (c CentID) MarshalBinary() (data []byte, err error) {
	return c[:], nil
}

func (c CentID) BigInt() *big.Int {
	return tools.ByteSliceToBigInt(c[:])
}

func (c CentID) ByteArray() [CentIDByteLength]byte {
	var idBytes [CentIDByteLength]byte
	copy(idBytes[:], c[:CentIDByteLength])
	return idBytes
}

func ParseCentIDs(centIDByteArray [][]byte) (errs []error, centIDs []CentID) {
	for _, element := range centIDByteArray {
		centID, err := NewCentID(element)
		if err != nil {
			err = centerrors.Wrap(err, "error parsing receiver centId")
			errs = append(errs, err)
			continue
		}
		centIDs = append(centIDs, centID)
	}
	return errs, centIDs
}

// IDService is a default implementation of the Service
var IDService Service

// Identity defines an Identity on chain
type Identity interface {
	fmt.Stringer
	GetCentrifugeID() CentID
	CentrifugeID(cenId CentID)
	GetCurrentP2PKey() (ret string, err error)
	GetLastKeyForPurpose(keyPurpose int) (key []byte, err error)
	AddKeyToIdentity(ctx context.Context, keyPurpose int, key []byte) (confirmations chan *WatchIdentity, err error)
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
	LookupIdentityForID(centrifugeID CentID) (id Identity, err error)
	CreateIdentity(centrifugeID CentID) (id Identity, confirmations chan *WatchIdentity, err error)
	CheckIdentityExists(centrifugeID CentID) (exists bool, err error)
}

// CentrifugeIdStringToSlice takes a string and decodes it using base64 to convert it into a slice
// of length 32.
func CentrifugeIdStringToSlice(s string) (id CentID, err error) {
	centBytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return [CentIDByteLength]byte{}, err
	}
	centId, err := NewCentID(centBytes)
	if err != nil {
		return [CentIDByteLength]byte{}, err
	}
	return centId, nil
}

// GetClientP2PURL returns the p2p url associated with the centID
func GetClientP2PURL(centID CentID) (url string, err error) {
	target, err := IDService.LookupIdentityForID(centID)
	if err != nil {
		return url, centerrors.Wrap(err, "error fetching receiver identity")
	}

	p2pKey, err := target.GetCurrentP2PKey()
	if err != nil {
		return url, centerrors.Wrap(err, "error fetching p2p key")
	}

	return fmt.Sprintf("/ipfs/%s", p2pKey), nil
}

// GetClientsP2PURLs returns p2p urls associated with each centIDs
// will error out at first failure
func GetClientsP2PURLs(centIDs []CentID) ([]string, error) {
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
func GetIdentityKey(identity CentID, pubKey []byte) (keyInfo Key, err error) {
	id, err := IDService.LookupIdentityForID(identity)
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
func ValidateKey(centrifugeId CentID, key []byte, purpose int) error {
	idKey, err := GetIdentityKey(centrifugeId, key)
	if err != nil {
		return err
	}

	if !bytes.Equal(key, tools.Byte32ToSlice(idKey.GetKey())) {
		return fmt.Errorf(fmt.Sprintf("[Key: %x] Key doesn't match", idKey.GetKey()))
	}

	if !tools.ContainsBigIntInSlice(big.NewInt(int64(purpose)), idKey.GetPurposes()) {
		return fmt.Errorf(fmt.Sprintf("[Key: %x] Key doesn't have purpose [%d]", idKey.GetKey(), purpose))
	}

	// TODO Check if revokation block happened before the timeframe of the document signing, for historical validations
	if idKey.GetRevokedAt().Cmp(big.NewInt(0)) != 0 {
		return fmt.Errorf(fmt.Sprintf("[Key: %x] Key is currently revoked since block [%d]", idKey.GetKey(), idKey.GetRevokedAt()))
	}

	return nil
}

// AddKeyFromConfig adds a key previously generated and indexed in the configuration file to the identity specified in such config file
func AddKeyFromConfig(purpose int) error {
	identityService := EthereumIdentityService{}

	var identityConfig *config.IdentityConfig
	var err error

	switch purpose {
	case KeyPurposeP2p:
		identityConfig, err = ed25519.GetIDConfig()
	case KeyPurposeSigning:
		identityConfig, err = ed25519.GetIDConfig()
	case KeyPurposeEthMsgAuth:
		identityConfig, err = secp256k1.GetIDConfig()
	default:
		err = errors.New("Option not supported")
	}

	if err != nil {
		return err
	}

	centId, err := NewCentID(identityConfig.ID)
	if err != nil {
		return err
	}
	id, err := identityService.LookupIdentityForID(centId)
	if err != nil {
		return err
	}

	ctx, cancel := ethereum.DefaultWaitForTransactionMiningContext()
	defer cancel()
	confirmations, err := id.AddKeyToIdentity(ctx, purpose, identityConfig.PublicKey)
	if err != nil {
		return err
	}
	watchAddedToIdentity := <-confirmations

	lastKey, errLocal := watchAddedToIdentity.Identity.GetLastKeyForPurpose(purpose)
	if errLocal != nil {
		return err
	}

	log.Infof("Key [%v] with type [$s] Added to Identity [%s]", lastKey, purpose, watchAddedToIdentity.Identity)

	return nil
}
