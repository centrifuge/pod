package signatures

import (
	"fmt"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/utils"
	"golang.org/x/crypto/ed25519"
)

// VerifySignature verifies the signature using ed25519
func VerifySignature(pubKey, message, signature []byte) error {
	valid := ed25519.Verify(pubKey, message, signature)
	if !valid {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// Sign the document with the private key and return the signature along with the public key for the verification
// assumes that signing root for the document is generated
func Sign(centIDBytes []byte, privateKey []byte, pubKey []byte, payload []byte) *coredocumentpb.Signature {
	return &coredocumentpb.Signature{
		EntityId:  centIDBytes,
		PublicKey: pubKey,
		Signature: ed25519.Sign(privateKey, payload),
		Timestamp: utils.ToTimestamp(time.Now().UTC()),
	}
}
