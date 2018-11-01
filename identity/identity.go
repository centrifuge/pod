package identity

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/keytools/ed25519"
	"github.com/centrifuge/go-centrifuge/keytools/secp256k1"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	CentIDLength         = 6
	ActionCreate         = "create"
	ActionAddKey         = "addkey"
	KeyPurposeP2p        = 1
	KeyPurposeSigning    = 2
	KeyPurposeEthMsgAuth = 3
)

// IDService is a default implementation of the Service
var IDService Service

type CentID [CentIDLength]byte

// IdentityConfig holds information about the identity
type IdentityConfig struct {
	ID   CentID
	Keys map[int]IdentityKey
}

// IdentityKey represents a key pair
type IdentityKey struct {
	PublicKey  []byte
	PrivateKey []byte
}

// GetIdentityConfig returns the identity and keys associated with the node
func GetIdentityConfig() (*IdentityConfig, error) {
	centIDBytes, err := config.Config.GetIdentityID()
	if err != nil {
		return nil, err
	}
	centID, err := ToCentID(centIDBytes)
	if err != nil {
		return nil, err
	}

	//ed25519 keys
	keys := map[int]IdentityKey{}
	pk, sk, err := ed25519.GetSigningKeyPairFromConfig()
	if err != nil {
		return nil, err
	}
	keys[KeyPurposeP2p] = IdentityKey{PublicKey: pk, PrivateKey: sk}
	keys[KeyPurposeSigning] = IdentityKey{PublicKey: pk, PrivateKey: sk}

	//secp256k1 keys
	pk, sk, err = secp256k1.GetEthAuthKeyFromConfig()
	if err != nil {
		return nil, err
	}
	pubKey, err := hexutil.Decode(secp256k1.GetAddress(pk))
	if err != nil {
		return nil, err
	}
	keys[KeyPurposeEthMsgAuth] = IdentityKey{PublicKey: pubKey, PrivateKey: sk}

	return &IdentityConfig{ID: centID, Keys: keys}, nil
}

// ToCentID takes bytes and return CentID
// errors out if bytes are empty, nil, or len(bytes) > CentIDLength
func ToCentID(bytes []byte) (centID CentID, err error) {
	if utils.IsEmptyByteSlice(bytes) {
		return centID, fmt.Errorf("empty bytes provided")
	}

	if !utils.IsValidByteSliceForLength(bytes, CentIDLength) {
		return centID, errors.New("invalid length byte slice provided for centID")
	}

	copy(centID[:], bytes[:CentIDLength])
	return centID, nil
}

// CentIDFromString takes an hex string and returns a CentID
func CentIDFromString(id string) (centID CentID, err error) {
	decID, err := hexutil.Decode(id)
	if err != nil {
		return centID, centerrors.Wrap(err, "failed to decode id")
	}

	return ToCentID(decID)
}

// CentIDsFromStrings converts hex ids to centIDs
func CentIDsFromStrings(ids []string) ([]CentID, error) {
	var cids []CentID
	for _, id := range ids {
		cid, err := CentIDFromString(id)
		if err != nil {
			return nil, err
		}

		cids = append(cids, cid)
	}

	return cids, nil
}

func NewRandomCentID() CentID {
	ID, _ := ToCentID(utils.RandomSlice(CentIDLength))
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
	return hexutil.Encode(c[:])
}

func (c CentID) MarshalBinary() (data []byte, err error) {
	return c[:], nil
}

func (c CentID) BigInt() *big.Int {
	return utils.ByteSliceToBigInt(c[:])
}

func (c CentID) ByteArray() [CentIDLength]byte {
	var idBytes [CentIDLength]byte
	copy(idBytes[:], c[:CentIDLength])
	return idBytes
}

func ParseCentIDs(centIDByteArray [][]byte) (errs []error, centIDs []CentID) {
	for _, element := range centIDByteArray {
		centID, err := ToCentID(element)
		if err != nil {
			err = centerrors.Wrap(err, "error parsing receiver centId")
			errs = append(errs, err)
			continue
		}
		centIDs = append(centIDs, centID)
	}
	return errs, centIDs
}

// Identity defines an Identity on chain
type Identity interface {
	fmt.Stringer
	GetCentrifugeID() CentID
	CentrifugeID(cenId CentID)
	GetCurrentP2PKey() (ret string, err error)
	GetLastKeyForPurpose(keyPurpose int) (key []byte, err error)
	AddKeyToIdentity(ctx context.Context, keyPurpose int, key []byte) (confirmations chan *WatchIdentity, err error)
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

	// LookupIdentityForID looks up if the identity for given CentID exists on ethereum
	LookupIdentityForID(centrifugeID CentID) (id Identity, err error)

	// CreateIdentity creates an identity representing the id on ethereum
	CreateIdentity(centrifugeID CentID) (id Identity, confirmations chan *WatchIdentity, err error)

	// CheckIdentityExists checks if the identity represented by id actually exists on ethereum
	CheckIdentityExists(centrifugeID CentID) (exists bool, err error)

	// GetIdentityAddress gets the address of the ethereum identity contract for the given CentID
	GetIdentityAddress(centID CentID) (common.Address, error)
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

	if utils.IsEmptyByte32(key.GetKey()) {
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

	if !bytes.Equal(key, utils.Byte32ToSlice(idKey.GetKey())) {
		return fmt.Errorf(fmt.Sprintf("[Key: %x] Key doesn't match", idKey.GetKey()))
	}

	if !utils.ContainsBigIntInSlice(big.NewInt(int64(purpose)), idKey.GetPurposes()) {
		return fmt.Errorf(fmt.Sprintf("[Key: %x] Key doesn't have purpose [%d]", idKey.GetKey(), purpose))
	}

	if idKey.GetRevokedAt().Cmp(big.NewInt(0)) != 0 {
		return fmt.Errorf(fmt.Sprintf("[Key: %x] Key is currently revoked since block [%d]", idKey.GetKey(), idKey.GetRevokedAt()))
	}

	return nil
}

// AddKeyFromConfig adds a key previously generated and indexed in the configuration file to the identity specified in such config file
func AddKeyFromConfig(purpose int) error {
	identityConfig, err := GetIdentityConfig()
	if err != nil {
		return err
	}

	id, err := IDService.LookupIdentityForID(identityConfig.ID)
	if err != nil {
		return err
	}

	ctx, cancel := ethereum.DefaultWaitForTransactionMiningContext()
	defer cancel()
	confirmations, err := id.AddKeyToIdentity(ctx, purpose, identityConfig.Keys[purpose].PublicKey)
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
