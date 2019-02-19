package identity

import (
	"context"
	"fmt"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"math/big"

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

	// BootstrappedIDService is used as a key to map the configured ID Service through context.
	BootstrappedIDService string = "BootstrappedIDService"

	// CentIDLength is the length in bytes of the CentrifugeID
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

// ToAddress returns the DID as common.Address
func (d DID) ToAddress() common.Address {
	return common.Address(d)
}

// String returns the DID as HEX String
func (d DID) String() string {
	return d.ToAddress().String()
}

// BigInt returns DID in bigInt
func (c DID) BigInt() *big.Int {
	return utils.ByteSliceToBigInt(c[:])
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

// NewDIDFromBytes returns a DID based on a bytes input
func NewDIDFromBytes(bAddr []byte) DID {
	return DID(common.BytesToAddress(bAddr))
}

// Factory is the interface for factory related interactions
type Factory interface {
	CreateIdentity(ctx context.Context) (id *DID, err error)
	CalculateIdentityAddress(ctx context.Context) (*common.Address, error)
}

// Service interface contains the methods to interact with the identity contract
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

// Key defines a single ERC725 identity key
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
