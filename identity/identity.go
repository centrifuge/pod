package identity

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/keytools/ed25519"
	"github.com/centrifuge/go-centrifuge/keytools/secp256k1"
	"github.com/centrifuge/go-centrifuge/signatures"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	// CentIDLength is the length in bytes of the CentrifugeID
	CentIDLength = 6

	// KeyPurposeP2P represents a key used for p2p txns
	KeyPurposeP2P = 1

	// KeyPurposeSigning represents a key used for signing
	KeyPurposeSigning = 2

	// KeyPurposeEthMsgAuth represents a key used for ethereum txns
	KeyPurposeEthMsgAuth = 3
)

// CentID represents a CentIDLength identity of an entity
type CentID [CentIDLength]byte

// IDConfig holds information about the identity
type IDConfig struct {
	ID   CentID
	Keys map[int]IDKey
}

// IDKey represents a key pair
type IDKey struct {
	PublicKey  []byte
	PrivateKey []byte
}

// Equal checks if c == other
func (c CentID) Equal(other CentID) bool {
	for i := range c {
		if c[i] != other[i] {
			return false
		}
	}

	return true
}

// String returns the hex format of CentID
func (c CentID) String() string {
	return hexutil.Encode(c[:])
}

// BigInt returns CentID in bigInt
func (c CentID) BigInt() *big.Int {
	return utils.ByteSliceToBigInt(c[:])
}

// Config defines methods required for the package identity.
type Config interface {
	GetEthereumDefaultAccountName() string
	GetIdentityID() ([]byte, error)
	GetSigningKeyPair() (pub, priv string)
	GetEthAuthKeyPair() (pub, priv string)
	GetEthereumContextWaitTimeout() time.Duration
	GetContractAddress(address string) common.Address
}

// Identity defines an Identity on chain
type Identity interface {
	fmt.Stringer
	CentID() CentID
	SetCentrifugeID(centID CentID)
	CurrentP2PKey() (ret string, err error)
	LastKeyForPurpose(keyPurpose int) (key []byte, err error)
	AddKeyToIdentity(ctx context.Context, keyPurpose int, key []byte) (confirmations chan *WatchIdentity, err error)
	FetchKey(key []byte) (Key, error)
}

// Key defines a single ERC725 identity key
type Key interface {
	GetKey() [32]byte
	GetPurposes() []*big.Int
	GetRevokedAt() *big.Int
}

// WatchIdentity holds the identity received form chain event
type WatchIdentity struct {
	Identity Identity
	Error    error
}

// Service is used to interact with centrifuge identities
type Service interface {

	// LookupIdentityForID looks up if the identity for given CentID exists on ethereum
	LookupIdentityForID(centrifugeID CentID) (id Identity, err error)

	// CreateIdentity creates an identity representing the id on ethereum
	CreateIdentity(centrifugeID CentID) (id Identity, confirmations chan *WatchIdentity, err error)

	// CheckIdentityExists checks if the identity represented by id actually exists on ethereum
	CheckIdentityExists(centrifugeID CentID) (exists bool, err error)

	// GetIdentityAddress gets the address of the ethereum identity contract for the given CentID
	GetIdentityAddress(centID CentID) (common.Address, error)

	// GetClientP2PURL returns the p2p url associated with the centID
	GetClientP2PURL(centID CentID) (url string, err error)

	// GetClientsP2PURLs returns p2p urls associated with each centIDs
	// will error out at first failure
	GetClientsP2PURLs(centIDs []CentID) ([]string, error)

	// GetIdentityKey returns the key for provided identity
	GetIdentityKey(identity CentID, pubKey []byte) (keyInfo Key, err error)

	// ValidateKey checks if a given key is valid for the given centrifugeID.
	ValidateKey(centID CentID, key []byte, purpose int) error

	// AddKeyFromConfig adds a key previously generated and indexed in the configuration file to the identity specified in such config file
	AddKeyFromConfig(purpose int) error

	// ValidateSignature validates a signature on a message based on identity data
	ValidateSignature(signature *coredocumentpb.Signature, message []byte) error
}

// GetIdentityConfig returns the identity and keys associated with the node.
func GetIdentityConfig(config Config) (*IDConfig, error) {
	centIDBytes, err := config.GetIdentityID()
	if err != nil {
		return nil, err
	}
	centID, err := ToCentID(centIDBytes)
	if err != nil {
		return nil, err
	}

	//ed25519 keys
	keys := map[int]IDKey{}
	pk, sk, err := ed25519.GetSigningKeyPair(config.GetSigningKeyPair())
	if err != nil {
		return nil, err
	}
	keys[KeyPurposeP2P] = IDKey{PublicKey: pk, PrivateKey: sk}
	keys[KeyPurposeSigning] = IDKey{PublicKey: pk, PrivateKey: sk}

	//secp256k1 keys
	pk, sk, err = secp256k1.GetEthAuthKey(config.GetEthAuthKeyPair())
	if err != nil {
		return nil, err
	}
	pubKey, err := hexutil.Decode(secp256k1.GetAddress(pk))
	if err != nil {
		return nil, err
	}
	keys[KeyPurposeEthMsgAuth] = IDKey{PublicKey: pubKey, PrivateKey: sk}

	return &IDConfig{ID: centID, Keys: keys}, nil
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

// RandomCentID returns a randomly generated CentID
func RandomCentID() CentID {
	ID, _ := ToCentID(utils.RandomSlice(CentIDLength))
	return ID
}

// ValidateCentrifugeIDBytes validates a centrifuge ID given as bytes
func ValidateCentrifugeIDBytes(givenCentID []byte, centrifugeID CentID) error {
	centIDSignature, err := ToCentID(givenCentID)
	if err != nil {
		return err
	}

	if !centrifugeID.Equal(centIDSignature) {
		return errors.New("provided bytes doesn't match centID")
	}

	return nil
}

// Sign the document with the private key and return the signature along with the public key for the verification
// assumes that signing root for the document is generated
func Sign(idConfig *IDConfig, purpose int, payload []byte) *coredocumentpb.Signature {
	return signatures.Sign(idConfig.ID[:], idConfig.Keys[purpose].PrivateKey, idConfig.Keys[purpose].PublicKey, payload)
}
