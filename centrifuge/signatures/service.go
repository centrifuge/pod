package signatures

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ed25519"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"time"
	"errors"
	"bytes"
	"fmt"
)

type KeyInfo struct {
	PublicKey ed25519.PublicKey
	ValidFrom time.Time
	ValidUntil time.Time
	Identity []byte
}

type SigningService struct {
	// For now we will hard code a few known signing keys. This should later be replaced with ethereum based identities
	// Structure is: Identity ID, Key
	KnownKeys map[[32]byte]KeyInfo
}

// ValidateSignaturesOnDocument validates all signatures on the current document
func (srv *SigningService) ValidateSignaturesOnDocument(doc *coredocument.CoreDocument) (valid bool, err error) {
	message := srv.createSignatureData(doc)
	for _, signature := range doc.Signatures {
		valid, err := srv.ValidateSignature(signature, message)
		if !valid {
			return valid, err
		}
	}
	return true, nil
}

func (srv *SigningService) ValidateSignature(signature *coredocument.Signature, message []byte) (valid bool, err error) {
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

// LoadPublicKeys just loads public keys from the config for now until identity management does this for us.
func (srv *SigningService) LoadPublicKeys () {
	keys := viper.GetStringMapString("keys.knownSigningKeys")
	for k, v := range keys {
		key := keytools.GetPublicSigningKey(v)
		var i [32]byte
		copy(i[:], key[:32])
		srv.KnownKeys[i] = KeyInfo{
			PublicKey: key,
			ValidUntil: time.Time{},
			ValidFrom: time.Now(),
			Identity: []byte(k),
		}

	}
}

func (srv *SigningService) GetIDFromKey(key ed25519.PublicKey) (id [32]byte) {
	copy(id[:], key[:32])
	return
}

func (srv *SigningService) GetKeyInfo(key ed25519.PublicKey) (keyInfo KeyInfo, err error) {
	keyInfo, exists := srv.KnownKeys[srv.GetIDFromKey(key)]
	if !exists {
		return keyInfo, errors.New("key not found")
	}
	return
}

// ValidateKey checks if a given key is valid for the given timestamp.
func (srv *SigningService) ValidateKey(identity []byte, key ed25519.PublicKey, timestamp time.Time) (valid bool, err error) {
	keyInfo, err :=  srv.GetKeyInfo(key)

	if err != nil {
		return false, errors.New("key not found")
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

func (srv *SigningService) createSignatureData (doc *coredocument.CoreDocument) (signatureData []byte) {
	signatureData = make([]byte, 64)
	copy(signatureData[:32], doc.DataMerkleRoot[:32])
	copy(signatureData[32:64], doc.NextIdentifier[:32])
	return
}

// Sign a document with a provided public key
func (srv *SigningService) Sign (doc *coredocument.CoreDocument, identity []byte, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) {
	sigArray := srv.createSignatureData(doc)
	signature := ed25519.Sign(privateKey, sigArray)
	doc.Signatures = append(doc.Signatures, &coredocument.Signature{identity,publicKey, signature})
}

