package signatures

import (
	"testing"
	"crypto/rand"
	"time"
	"golang.org/x/crypto/ed25519"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
)

func rand32Bytes () []byte {
	randbytes := make([]byte, 32)
	rand.Read(randbytes)
	return randbytes
}


func TestSignatureValidation(t *testing.T) {
	var signingService = SigningService{}
	var testKeys []KeyInfo
	key1Pub, key1, _ := ed25519.GenerateKey(rand.Reader)
	key2Pub, key2, _ := ed25519.GenerateKey(rand.Reader)
	key3Pub, key3, _ := ed25519.GenerateKey(rand.Reader)
	key4Pub, _, _ := ed25519.GenerateKey(rand.Reader)
	id1 := []byte("1")
	id2 := []byte("2")
	id3 := []byte("3")


	// Valid key (for one hour)
	testKeys = append(testKeys, KeyInfo{
		ed25519.PublicKey(key1Pub),
		time.Now(),
		time.Now().Add(1*time.Hour),
		id1,
	})
	// Expired key
	testKeys =  append(testKeys, KeyInfo{
		ed25519.PublicKey(key2Pub),
		time.Now().Add(-1*time.Hour),
		time.Now().Add(-1*time.Hour),
		id2,
	})

	// Valid key (unlimited)
	testKeys =  append(testKeys, KeyInfo{
		ed25519.PublicKey(key3Pub),
		time.Now(),
		time.Time{},
		id3,
	})

	signingService.KnownKeys =  map[[32]byte]KeyInfo{}

	for _, keyInfo := range testKeys {
		k := signingService.GetIDFromKey(keyInfo.PublicKey)
		signingService.KnownKeys[k] = keyInfo
	}

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

	dataMerkleRoot := rand32Bytes()
	documentIdentifier := rand32Bytes()
	nextIdentifier := rand32Bytes()

	doc := &coredocument.CoreDocument{
		DataMerkleRoot: dataMerkleRoot,
		DocumentIdentifier: documentIdentifier,
		NextIdentifier: nextIdentifier,
	}

	signingService.Sign(doc, id1, key1, key1Pub)

	valid, err = signingService.ValidateSignaturesOnDocument(doc)
	if len(doc.Signatures) != 1 {
		t.Fatal("Signature is missing")
	}
	if !valid || err != nil {
		t.Fatal("Signature validation failed")
	}

	signingService.Sign(doc, id3, key3, key3Pub)
	valid, err = signingService.ValidateSignaturesOnDocument(doc)
	if len(doc.Signatures) != 2 {
		t.Fatal("Signature is missing")
	}
	if !valid || err != nil {
		t.Fatal("Signature validation failed")
	}

	// Adding an invalid signature
	// TODO: Probably creating the signature shouldn't even work because it's invalid
	signingService.Sign(doc, id2, key2, key2Pub)
	valid, err = signingService.ValidateSignaturesOnDocument(doc)
	if len(doc.Signatures) != 3 {
		t.Fatal("Signature is missing")
	}
	if valid || err == nil {
		t.Fatal("Signature validation succeeded")
	}



}