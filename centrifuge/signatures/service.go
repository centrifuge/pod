package signatures

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"golang.org/x/crypto/ed25519"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"math/big"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
)

var signingService SigningService
var once sync.Once

func GetSigningService() *SigningService {
	return &signingService
}

func NewSigningService(srv SigningService) {
	once.Do(func() {
		signingService = srv
	})
	return
}

type KeyInfo struct {
	Key []byte
	Purposes  []int
	RevokedAt *big.Int
}

type SigningService struct {
	IdentityService identity.IdentityService
}

// ValidateSignaturesOnDocument validates all signatures on the current document
func (srv *SigningService) ValidateSignaturesOnDocument(doc *coredocumentpb.CoreDocument) (valid bool, err error) {
	// TODO: Signature Validation not yet implemented
	return false, nil
}

func (srv *SigningService) ValidateSignature(signature *coredocumentpb.Signature, message []byte) (valid bool, err error) {
	valid, err = srv.ValidateKey(signature.EntityId, signature.PublicKey)
	if err != nil {
		return
	}

	valid = ed25519.Verify(signature.PublicKey, message, signature.Signature)
	if !valid {
		return false, errors.New("invalid signature")
	}

	return
}

func (srv *SigningService) GetIDFromKey(key ed25519.PublicKey) (id [32]byte) {
	copy(id[:], key[:32])
	return
}

func (srv *SigningService) GetIdentityKey(identity []byte, key ed25519.PublicKey) (keyInfo identity.IdentityKey, err error) {
	identityInstance, err := srv.IdentityService.LookupIdentityForId(identity)
	if err != nil {
		return keyInfo, err
	}

	keyInstance, err := identityInstance.FetchKey()
	if err != nil {
		return keyInfo, err
	}

	if len(keyInstance.GetKey()) == 0 {
		return keyInfo, errors.New(fmt.Sprintf("key not found for identity: %x", identity))
	}

	return keyInstance, nil
}

// ValidateKey checks if a given key is valid for the given centrifugeId.
func (srv *SigningService) ValidateKey(centrifugeId []byte, key ed25519.PublicKey) (valid bool, err error) {
	identityKey, err := srv.GetIdentityKey(centrifugeId, key)
	if err != nil {
		return false, err
	}

	if !bytes.Equal(key, tools.Byte32ToSlice(identityKey.GetKey())) {
		return false, errors.New(fmt.Sprintf("[Key: %x] Key doesn't match", identityKey.GetKey()))
	}

	if !tools.ContainsBigIntInSlice(big.NewInt(identity.KeyPurposeSigning), identityKey.GetPurposes()) {
		return false, errors.New(fmt.Sprintf("[Key: %x] Key doesn't have purpose [%d]", identityKey.GetKey(), identity.KeyPurposeSigning))
	}

	// TODO Check if revokation block happened before the timeframe of the document signing, for historical validations
	if identityKey.GetRevokedAt().Cmp(big.NewInt(0)) != 0 {
		return false, errors.New(fmt.Sprintf("[Key: %x] Key is currently revoked since block [%d]", identityKey.GetKey(), identityKey.GetRevokedAt()))
	}

	return true, nil
}

func (srv *SigningService) MakeSignature(doc *coredocumentpb.CoreDocument, identity []byte, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) (sig *coredocumentpb.Signature) {
	signature := ed25519.Sign(privateKey, doc.SigningRoot)
	return &coredocumentpb.Signature{EntityId: identity, PublicKey: publicKey, Signature: signature}
}

// Sign a document with a provided public key
func (srv *SigningService) Sign(doc *coredocumentpb.CoreDocument) (err error) {
	identityConfig := identity.NewIdentityConfig()
	sig := srv.MakeSignature(doc, identityConfig.IdentityId, identityConfig.PrivateKey, identityConfig.PublicKey)
	doc.Signatures = append(doc.Signatures, sig)
	return nil
}
