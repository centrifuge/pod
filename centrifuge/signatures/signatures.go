package signatures

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
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
	valid, err = ValidateKey(idSrv, signature.EntityId, signature.PublicKey)
	if err != nil {
		return valid, err
	}

	valid = ed25519.Verify(signature.PublicKey, message, signature.Signature)
	if !valid {
		return false, errors.New("invalid signature")
	}

	return
}

// ValidateKey checks if a given key is valid for the given centrifugeID.
func ValidateKey(idSrv identity.Service, centrifugeId []byte, key ed25519.PublicKey) (valid bool, err error) {
	idKey, err := identity.GetIdentityKey(idSrv, centrifugeId, key)
	if err != nil {
		return false, err
	}

	if !bytes.Equal(key, tools.Byte32ToSlice(idKey.GetKey())) {
		return false, errors.New(fmt.Sprintf("[Key: %x] Key doesn't match", idKey.GetKey()))
	}

	if !tools.ContainsBigIntInSlice(big.NewInt(identity.KeyPurposeSigning), idKey.GetPurposes()) {
		return false, errors.New(fmt.Sprintf("[Key: %x] Key doesn't have purpose [%d]", idKey.GetKey(), identity.KeyPurposeSigning))
	}

	// TODO Check if revokation block happened before the timeframe of the document signing, for historical validations
	if idKey.GetRevokedAt().Cmp(big.NewInt(0)) != 0 {
		return false, errors.New(fmt.Sprintf("[Key: %x] Key is currently revoked since block [%d]", idKey.GetKey(), idKey.GetRevokedAt()))
	}

	return true, nil
}

// Sign the document with the private key and return the signature along with the public key for the verification
// assumes that signing root for the document is generated
func Sign(idConfig *config.IdentityConfig, doc *coredocumentpb.CoreDocument) *coredocumentpb.Signature {
	signature := ed25519.Sign(idConfig.PrivateKey, doc.SigningRoot)
	return &coredocumentpb.Signature{EntityId: idConfig.IdentityId, PublicKey: idConfig.PublicKey, Signature: signature}
}
