package signatures

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"golang.org/x/crypto/ed25519"
	"sync"
	"time"
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
	PublicKey  ed25519.PublicKey
	ValidFrom  time.Time
	ValidUntil time.Time
	Identity   []byte
}

type SigningService struct {
	// For simplicity we only support one active identity for now.
	IdentityId []byte
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

func (srv *SigningService) LoadIdentityKeyFromConfig() {
	srv.IdentityId = config.Config.GetIdentityId()
	srv.PublicKey, srv.PrivateKey = keytools.GetSigningKeyPairFromConfig()
}

// ValidateSignaturesOnDocument validates all signatures on the current document
func (srv *SigningService) ValidateSignaturesOnDocument(doc *coredocumentpb.CoreDocument) (valid bool, err error) {
	// TODO: Signature Validation not yet implemented
	return false, nil
}

func (srv *SigningService) ValidateSignature(signature *coredocumentpb.Signature, message []byte) (valid bool, err error) {
	valid, err = srv.ValidateKey(signature.EntityId, signature.PublicKey, time.Now())
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

func (srv *SigningService) GetKeyInfo(key ed25519.PublicKey) (keyInfo KeyInfo, err error) {
	exists := false
	// TODO: Get Key Info not yet implemented
	if !exists {
		return keyInfo, errors.New("key not found")
	}
	return
}

// ValidateKey checks if a given key is valid for the given timestamp.
func (srv *SigningService) ValidateKey(identity []byte, key ed25519.PublicKey, timestamp time.Time) (valid bool, err error) {
	keyInfo, err := srv.GetKeyInfo(key)

	if err != nil {
		return false, err
	}

	if !bytes.Equal(identity, keyInfo.Identity) {
		return false, errors.New(fmt.Sprintf("[Key: %s] Key Identity doesn't match", srv.GetIDFromKey(keyInfo.PublicKey)))
	}

	if !keyInfo.ValidFrom.IsZero() && timestamp.Unix() < keyInfo.ValidFrom.Unix() {
		return false, errors.New(fmt.Sprintf("[Key: %s] Signature timestamp is before key was added", srv.GetIDFromKey(keyInfo.PublicKey)))
	}

	if !keyInfo.ValidUntil.IsZero() && timestamp.Unix() > keyInfo.ValidUntil.Unix() {
		return false, errors.New(fmt.Sprintf("[Key: %s] Signature timestamp is past key revocation", srv.GetIDFromKey(keyInfo.PublicKey)))
	}

	return true, nil
}

func (srv *SigningService) MakeSignature(doc *coredocumentpb.CoreDocument, identity []byte, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) (sig *coredocumentpb.Signature) {
	signature := ed25519.Sign(privateKey, doc.SigningRoot)
	return &coredocumentpb.Signature{EntityId: identity, PublicKey: publicKey, Signature: signature}
}

// Sign a document with a provided public key
func (srv *SigningService) Sign(doc *coredocumentpb.CoreDocument) (err error) {
	sig := srv.MakeSignature(doc, srv.IdentityId, srv.PrivateKey, srv.PublicKey)
	doc.Signatures = append(doc.Signatures, sig)
	return nil
}
