package did

import (
	"fmt"
	"math/big"

	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
)

// Key defines a single ERC725 identity key
type Key interface {
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
func NewKey(pk [32]byte, purpose *big.Int, keyType *big.Int) Key {
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
