// +build unit

package signatures

import (
	"testing"
	"time"
	"golang.org/x/crypto/ed25519"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	"os"
	"bytes"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
)


var (
	signingService SigningService
	testKeys []KeyInfo
	key1Pub, key2Pub, key3Pub, key4Pub ed25519.PublicKey
	key1, key2, key3, key4 ed25519.PrivateKey
	id1 = []byte("1")
	id2 = []byte("2")
	id3 = []byte("3")
)

func TestMain(m *testing.M) {
	signingService = SigningService{}
	signingService.KnownKeys =  map[[32]byte]KeyInfo{}

	// Generated with: key1Pub, key1, _ := ed25519.GenerateKey(rand.Reader)
	key1Pub = []byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1 = []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key2Pub = []byte{113, 94, 241, 90, 37, 21, 191, 41, 217, 134, 68, 34, 107, 207, 87, 203, 0, 204, 113, 213, 126, 0, 98, 57, 220, 45, 155, 242, 151, 76, 39, 231}
	key2 = []byte{250, 254, 250, 250, 96, 33, 238, 120, 185, 198, 83, 69, 160, 207, 138, 239, 51, 172, 90, 252, 23, 83, 19, 166, 136, 131, 142, 18, 119, 36, 176, 0, 113, 94, 241, 90, 37, 21, 191, 41, 217, 134, 68, 34, 107, 207, 87, 203, 0, 204, 113, 213, 126, 0, 98, 57, 220, 45, 155, 242, 151, 76, 39, 231}
	key3Pub = []byte{153, 38, 210, 185, 3, 120, 208, 57, 159, 178, 175, 184, 253, 124, 188, 72, 228, 245, 66, 98, 7, 1, 124, 244, 49, 92, 75, 59, 147, 186, 37, 145}
	key3 = []byte{169, 105, 156, 14, 78, 30, 37, 102, 84, 55, 7, 105, 163, 167, 111, 93, 125, 15, 81, 210, 84, 27, 148, 50, 240, 36, 50, 3, 105, 106, 251, 135, 153, 38, 210, 185, 3, 120, 208, 57, 159, 178, 175, 184, 253, 124, 188, 72, 228, 245, 66, 98, 7, 1, 124, 244, 49, 92, 75, 59, 147, 186, 37, 145}
	key4 = []byte{122, 14, 67, 51, 130, 66, 68, 193, 100, 229, 106, 56, 134, 179, 142, 245, 213, 199, 100, 108, 18, 156, 27, 10, 143, 156, 250, 85, 253, 123, 12, 255}
	key4Pub = []byte{142, 110, 173, 1, 3, 27, 149, 34, 232, 120, 203, 249, 23, 246, 46, 66, 9, 136, 96, 204, 31, 123, 166, 133, 36, 102, 58, 27, 26, 252, 239, 116, 122, 14, 67, 51, 130, 66, 68, 193, 100, 229, 106, 56, 134, 179, 142, 245, 213, 199, 100, 108, 18, 156, 27, 10, 143, 156, 250, 85, 253, 123, 12, 255}

		// Valid key (for one hour)
	testKeys = []KeyInfo{
		{
			ed25519.PublicKey(key1Pub),
			time.Now(),
			time.Now().Add(1 * time.Hour),
			id1,
		},
		// Expired key
		KeyInfo{
			ed25519.PublicKey(key2Pub),
			time.Now().Add(-1 * time.Hour),
			time.Now().Add(-1 * time.Hour),
			id2,
		},
		// Valid key (unlimited)
		KeyInfo{
			ed25519.PublicKey(key3Pub),
			time.Now(),
			time.Time{},
			id3,
		}}

	for _, keyInfo := range testKeys {
		k := signingService.GetIDFromKey(keyInfo.PublicKey)
		signingService.KnownKeys[k] = keyInfo
	}

	signingService.PublicKey = ed25519.PublicKey(key1Pub)
	signingService.PrivateKey = ed25519.PrivateKey(key1)
	signingService.IdentityId = id1

	os.Exit(m.Run())
}

func TestSignatureValidation(t *testing.T) {
	valid, err := signingService.ValidateKey(id1, key1Pub, time.Now())
	if !valid || err != nil {
		t.Fatal("Key should be valid")
	}

	// Signature timestamp is too early
	valid, err = signingService.ValidateKey(id1, key1Pub, time.Now().Add(-10*time.Hour))
	if valid || err == nil {
		t.Fatal("Key should be invalid", err)
	}
	// Signature timestamp is too late
	valid, err = signingService.ValidateKey(id1, key1Pub, time.Now().Add(5*time.Hour))
	if valid || err == nil {
		t.Fatal("Key should be invalid", err)
	}

	// Signature is using an incorrect key
	valid, err = signingService.ValidateKey(id1, key4Pub, time.Now())
	if valid || err == nil {
		t.Fatal("Key should be invalid", err)
	}

}

func TestDocumentSignatures(t *testing.T) {
	// Any 32byte value for these identifiers is ok (such as a ed25519 public key)
	dataMerkleRoot := key1Pub
	documentIdentifier := key1Pub
	nextIdentifier := key1Pub

	doc := &coredocumentpb.CoreDocument{
		DataMerkleRoot: dataMerkleRoot,
		DocumentIdentifier: documentIdentifier,
		NextIdentifier: nextIdentifier,
	}

	message := signingService.createSignatureData(doc)

	sig := signingService.MakeSignature(doc, id1, key1, key1Pub)
	if !bytes.Equal(sig.Signature, []byte{5, 169, 99, 150, 8, 150, 149, 31, 190, 248, 184, 102, 154, 71, 40, 148, 5, 1,
	76, 171, 65, 244, 188, 133, 230, 13, 7, 186, 60, 183, 181, 124, 87, 18, 183, 23, 12, 33, 181, 78, 43, 32, 221, 18,
	239, 237, 221, 147, 85, 241, 205, 29, 233, 5, 82, 118, 130, 149, 199, 98, 57, 234, 219, 15}) {
		t.Fatal("Signature does not match")
	}
	valid, err := signingService.ValidateSignature(sig, message)
	if !valid || err != nil {
		t.Fatal("Signature validation failed")
	}

	sig = signingService.MakeSignature(doc, id3, key3, key3Pub)
	if !bytes.Equal(sig.Signature, []byte{139, 128, 127, 135, 12, 92, 236, 22, 141, 63, 147, 137, 73, 70, 76, 194, 178,
	75, 252, 100, 7, 160, 170, 231, 238, 18, 120, 230, 35, 10, 53, 69, 76, 179, 38, 45, 183, 237, 29, 147, 213, 189, 110,
	43, 128, 36, 6, 178, 201, 12, 181, 163, 144, 190, 204, 87, 62, 153, 140, 201, 28, 226, 177, 10}) {
		t.Fatal("Signature does not match")
	}
	valid, err = signingService.ValidateSignature(sig, message)
	if !valid || err != nil {
		t.Fatal("Signature validation failed")
	}

	sig = signingService.MakeSignature(doc, id2, key2, key2Pub)
	if !bytes.Equal(sig.Signature, []byte{170, 120, 97, 240, 230, 21, 119, 206, 164, 52, 120, 202, 207, 224, 72, 225,
	236, 45, 195, 239, 34, 152, 75, 172, 207, 136, 199, 119, 140, 71, 229, 243, 19, 93, 202, 6, 210, 110, 252, 83, 86,
	64, 207, 149, 213, 160, 158, 98, 2, 67, 246, 225, 67, 16, 217, 99, 147, 234, 134, 192, 200, 65, 210, 13}) {
		t.Fatal("Signature does not match")
	}
	valid, err = signingService.ValidateSignature(sig, message)
	if valid || err == nil {
		t.Fatal("Signature validation succeeded")
	}
}

func TestDocumentSigning(t *testing.T) {
	dataMerkleRoot := testingutils.Rand32Bytes()
	documentIdentifier := testingutils.Rand32Bytes()
	nextIdentifier := testingutils.Rand32Bytes()

	doc := &coredocumentpb.CoreDocument{
		DataMerkleRoot: dataMerkleRoot,
		DocumentIdentifier: documentIdentifier,
		NextIdentifier: nextIdentifier,
	}

	signingService.Sign(doc)
	valid, err := signingService.ValidateSignaturesOnDocument(doc)
	if !valid || err != nil {
		t.Fatal("Signature validation failed")
	}
}