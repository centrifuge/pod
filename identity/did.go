package identity

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/ethereum/go-ethereum/common"
)

const (

	// ErrMalformedAddress standard error for malformed address
	ErrMalformedAddress = errors.Error("malformed address provided")

	// BootstrappedDIDFactory stores the id of the factory
	BootstrappedDIDFactory string = "BootstrappedDIDFactory"

	// BootstrappedDIDService stores the id of the service
	BootstrappedDIDService string = "BootstrappedDIDService"

	// CentIDLength is the length in bytes of the DID
	CentIDLength = 6

	// KeyPurposeP2P represents a key used for p2p txns
	KeyPurposeP2P = 1

	// KeyPurposeSigning represents a key used for signing
	KeyPurposeSigning = 2

	// KeyPurposeEthMsgAuth represents a key used for ethereum txns
	KeyPurposeEthMsgAuth = 3

	// KeyTypeECDSA has the value one in the ERC725 identity contract
	KeyTypeECDSA = 1
)

// DID stores the identity address of the user
type DID common.Address

// DIDLength contains the length of a DID
const DIDLength = common.AddressLength

// ToAddress returns the DID as common.Address
func (d DID) ToAddress() common.Address {
	return common.Address(d)
}

// String returns the DID as HEX String
func (d DID) String() string {
	return d.ToAddress().String()
}

// BigInt returns DID in bigInt
func (d DID) BigInt() *big.Int {
	return utils.ByteSliceToBigInt(d[:])
}

// Equal checks if d == other
func (d DID) Equal(other DID) bool {
	for i := range d {
		if d[i] != other[i] {
			return false
		}
	}
	return true
}

// NewDID returns a DID based on a common.Address
func NewDID(address common.Address) DID {
	return DID(address)
}

// NewDIDFromString returns a DID based on a hex string
func NewDIDFromString(address string) (DID, error) {
	if !common.IsHexAddress(address) {
		return DID{}, ErrMalformedAddress
	}
	return DID(common.HexToAddress(address)), nil
}

// NewDIDsFromStrings converts hex ids to DIDs
func NewDIDsFromStrings(ids []string) ([]DID, error) {
	var cids []DID
	for _, id := range ids {
		cid, err := NewDIDFromString(id)
		if err != nil {
			return nil, err
		}

		cids = append(cids, cid)
	}

	return cids, nil
}

// NewDIDFromBytes returns a DID based on a bytes input
func NewDIDFromBytes(bAddr []byte) DID {
	return DID(common.BytesToAddress(bAddr))
}

//// NewDIDFromContext returns DID from context.Account
//func NewDIDFromContext(ctx context.Context) (DID, error) {
//	tc, err := contextutil.Account(ctx)
//	if err != nil {
//		return DID{}, err
//	}
//
//	addressByte, err := tc.GetIdentityID()
//	if err != nil {
//		return DID{}, err
//	}
//	return NewDID(common.BytesToAddress(addressByte)), nil
//}

// Factory is the interface for factory related interactions
type Factory interface {
	CreateIdentity(ctx context.Context) (id *DID, err error)
	CalculateIdentityAddress(ctx context.Context) (*common.Address, error)
}

// NewDIDFromByte returns a DID based on a byte slice
func NewDIDFromByte(did []byte) DID {
	return DID(common.BytesToAddress(did))
}

// ServiceDID interface contains the methods to interact with the identity contract
type ServiceDID interface {
	// AddKey adds a key to identity contract
	AddKey(ctx context.Context, key KeyDID) error

	// AddKeysForAccount adds key from configuration
	AddKeysForAccount(acc config.Account) error

	// GetKey return a key from the identity contract
	GetKey(did DID, key [32]byte) (*KeyResponse, error)

	// RawExecute calls the execute method on the identity contract
	RawExecute(ctx context.Context, to common.Address, data []byte) error

	// Execute creates the abi encoding an calls the execute method on the identity contract
	Execute(ctx context.Context, to common.Address, contractAbi, methodName string, args ...interface{}) error

	// IsSignedWithPurpose verifies if a message is signed with one of the identities specific purpose keys
	IsSignedWithPurpose(did DID, message [32]byte, signature []byte, purpose *big.Int) (bool, error)

	// AddMultiPurposeKey adds a key with multiple purposes
	AddMultiPurposeKey(context context.Context, key [32]byte, purposes []*big.Int, keyType *big.Int) error

	// RevokeKey revokes an existing key in the smart contract
	RevokeKey(ctx context.Context, key [32]byte) error

	// GetClientP2PURL returns the p2p url associated with the did
	GetClientP2PURL(did DID) (string, error)

	//Exists checks if an identity contract exists
	Exists(ctx context.Context, did DID) error

	// ValidateKey checks if a given key is valid for the given centrifugeID.
	ValidateKey(ctx context.Context, did DID, key []byte, purpose int64) error

	// ValidateSignature checks if signature is valid for given identity
	ValidateSignature(signature *coredocumentpb.Signature, message []byte) error

	// CurrentP2PKey retrieves the last P2P key stored in the identity
	CurrentP2PKey(did DID) (ret string, err error)

	// GetClientsP2PURLs returns p2p urls associated with each centIDs
	// will error out at first failure
	GetClientsP2PURLs(dids []*DID) ([]string, error)

	// GetKeysByPurpose returns keys grouped by purpose from the identity contract.
	GetKeysByPurpose(did DID, purpose *big.Int) ([][32]byte, error)
}

// KeyDID defines a single ERC725 identity key
type KeyDID interface {
	GetKey() [32]byte
	GetPurpose() *big.Int
	GetRevokedAt() *big.Int
	GetType() *big.Int
}

// KeyResponse contains the needed fields of the GetKey response
type KeyResponse struct {
	Key       [32]byte
	Purposes  []*big.Int
	RevokedAt *big.Int
}

// Key holds the identity related details
type key struct {
	Key       [32]byte
	Purpose   *big.Int
	RevokedAt *big.Int
	Type      *big.Int
}

//NewKey returns a new key struct
func NewKey(pk [32]byte, purpose *big.Int, keyType *big.Int) KeyDID {
	return &key{pk, purpose, big.NewInt(0), keyType}
}

// GetKey returns the public key
func (idk *key) GetKey() [32]byte {
	return idk.Key
}

// GetPurposes returns the purposes intended for the key
func (idk *key) GetPurpose() *big.Int {
	return idk.Purpose
}

// GetRevokedAt returns the block at which the identity is revoked
func (idk *key) GetRevokedAt() *big.Int {
	return idk.RevokedAt
}

// GetType returns the type of the key
func (idk *key) GetType() *big.Int {
	return idk.Type
}

// String prints the peerID extracted from the key
func (idk *key) String() string {
	peerID, _ := ed25519.PublicKeyToP2PKey(idk.Key)
	return fmt.Sprintf("%s", peerID.Pretty())
}

// IDKey represents a key pair
type IDKey struct {
	PublicKey  []byte
	PrivateKey []byte
}

// IDKeys holds key of an identity
type IDKeys struct {
	ID   []byte
	Keys map[int]IDKey
}

// Config defines methods required for the package identity.
type Config interface {
	GetEthereumDefaultAccountName() string
	GetIdentityID() ([]byte, error)
	GetP2PKeyPair() (pub, priv string)
	GetSigningKeyPair() (pub, priv string)
	GetEthAuthKeyPair() (pub, priv string)
	GetEthereumContextWaitTimeout() time.Duration
}

// IDConfig holds information about the identity
// Deprecated
type IDConfig struct {
	ID   DID
	Keys map[int]IDKey
}

// GetIdentityConfig returns the identity and keys associated with the node.
func GetIdentityConfig(config Config) (*IDConfig, error) {
	centIDBytes, err := config.GetIdentityID()
	if err != nil {
		return nil, err
	}
	centID := NewDIDFromBytes(centIDBytes)

	//ed25519 keys
	keys := map[int]IDKey{}

	pk, sk, err := ed25519.GetSigningKeyPair(config.GetP2PKeyPair())
	if err != nil {
		return nil, err
	}
	keys[KeyPurposeP2P] = IDKey{PublicKey: pk, PrivateKey: sk}

	pk, sk, err = ed25519.GetSigningKeyPair(config.GetSigningKeyPair())
	if err != nil {
		return nil, err
	}
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

// Sign the document with the private key and return the signature along with the public key for the verification
// assumes that signing root for the document is generated
// Deprecated
func Sign(idConfig *IDConfig, purpose int, payload []byte) *coredocumentpb.Signature {
	return crypto.Sign(idConfig.ID[:], idConfig.Keys[purpose].PrivateKey, idConfig.Keys[purpose].PublicKey, payload)
}
