package signatures

import (
	"errors"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"golang.org/x/crypto/ed25519"
)

// ValidateSignaturesOnDocument validates all signatures on the current document
func ValidateSignaturesOnDocument(idSrv identity.Service, doc *coredocumentpb.CoreDocument) (valid bool, err error) {
	for _, sig := range doc.Signatures {
		valid, err := ValidateSignature(idSrv, sig, doc.SigningRoot)
		if err != nil || !valid {
			return false, err
		}
	}
	return true, nil
}

// ValidateSignature verifies the signature on the document
func ValidateSignature(idSrv identity.Service, signature *coredocumentpb.Signature, message []byte) (valid bool, err error) {
	valid, err = identity.ValidateKey(idSrv, signature.EntityId, signature.PublicKey)
	if err != nil {
		return valid, err
	}

	valid = ed25519.Verify(signature.PublicKey, message, signature.Signature)
	if !valid {
		return false, errors.New("invalid signature")
	}

	return
}

// Sign the document with the private key and return the signature along with the public key for the verification
// assumes that signing root for the document is generated
func Sign(idConfig *config.IdentityConfig, doc *coredocumentpb.CoreDocument) *coredocumentpb.Signature {
	signature := ed25519.Sign(idConfig.PrivateKey, doc.SigningRoot)
	return &coredocumentpb.Signature{EntityId: idConfig.IdentityId, PublicKey: idConfig.PublicKey, Signature: signature}
}
