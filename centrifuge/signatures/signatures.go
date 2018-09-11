package signatures

import (
	"fmt"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"golang.org/x/crypto/ed25519"
)

// ValidateSignaturesOnDocument validates all signatures on the current document
func ValidateSignaturesOnDocument(doc *coredocumentpb.CoreDocument) (valid bool, err error) {
	for _, sig := range doc.Signatures {
		valid, err := ValidateSignature(sig, doc.SigningRoot)
		if err != nil || !valid {
			return false, err
		}
	}
	return true, nil
}

// ValidateSignature verifies the signature on the document
func ValidateSignature(signature *coredocumentpb.Signature, message []byte) (valid bool, err error) {
	centid, err := identity.NewCentID(signature.EntityId)
	if err != nil {
		return valid, err
	}
	valid, err = identity.ValidateKey(centid, signature.PublicKey)
	if err != nil {
		return valid, err
	}

	return verifySignature(signature.PublicKey, message, signature.Signature)
}

// verifySignature verifies the signature using ed25519
func verifySignature(pubKey, message, signature []byte) (bool, error) {
	valid := ed25519.Verify(pubKey, message, signature)
	if !valid {
		return false, fmt.Errorf("invalid signature")
	}

	return true, nil
}

// Sign the document with the private key and return the signature along with the public key for the verification
// assumes that signing root for the document is generated
func Sign(idConfig *config.IdentityConfig, payload []byte) *coredocumentpb.Signature {
	signature := ed25519.Sign(idConfig.PrivateKey, payload)
	return &coredocumentpb.Signature{EntityId: idConfig.ID, PublicKey: idConfig.PublicKey, Signature: signature}
}
