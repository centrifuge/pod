package witness

import (
	"fmt"
	"testing"

	"golang.org/x/crypto/ed25519"
)

const examplePayload = `{"amount": "100", "date": "2016-12-12", "state": "due"}`

// GenerateKeypair is a small helper method to generate a signing key
func generateKeypair() (publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	return
}

func TestSign(t *testing.T) {
	alicePublicKey, alicePrivateKey := generateKeypair()
	bobPublicKey, bobPrivateKey := generateKeypair()

	// Prepare document with some empty data
	doc := PrepareDocument(examplePayload)

	// Sign the document
	doc.Sign(alicePrivateKey, alicePublicKey)

	// Check signature by Alice
	verified := doc.Verify(alicePublicKey)
	if !verified {
		t.Fatal("Can't verify signature")
	}

	// Document is not signed by Bob
	verified = doc.Verify(bobPublicKey)
	if verified {
		t.Fatal("Shouldn't verify with incorrect public key")
	}

	// Modify the MerkleRoot
	oldRoot := doc.MerkleRoot
	doc.MerkleRoot = "incorrect"
	fmt.Println("doc.MerkleRoot", doc.MerkleRoot)
	verified = doc.Verify(alicePublicKey)
	if verified {
		t.Fatal("Shouldn't verify with incorrect public key")
	}
	doc.MerkleRoot = oldRoot

	// Modify the JSON post signing
	oldPayload := doc.Payload[:len(doc.Payload)]
	// TODO: test with empty `{}` object -> should raise validation error
	doc.Payload = `{"state":"paid"}`

	verified = doc.Verify(alicePublicKey)
	if verified {
		t.Fatal("Shouldn't verify with incorrect payload")
	}
	doc.Payload = oldPayload

	// Create a new version of the document
	newDoc := UpdateDocument(doc)

	if newDoc.PreviousRoot != doc.MerkleRoot {
		t.Fatal("PreviousRoot doesn't match previous document's MerkleRoot")
	}

	if newDoc.CurrentVersionID != doc.NextVersionID {
		t.Fatal("CurrentVersionId doesn't match previous document's NextVersionId")
	}

	if newDoc.NextVersionID == doc.NextVersionID {
		t.Fatal("Make sure NextVersionNonce is updated")
	}

	newDoc.Sign(bobPrivateKey, bobPublicKey)
	verified = newDoc.Verify(bobPublicKey)
	if !verified {
		t.Fatal("Can't verify Bob's signature")
	}
	newDoc.Sign(alicePrivateKey, alicePublicKey)
	verified = newDoc.Verify(alicePublicKey)
	if !verified {
		t.Fatal("Can't verify Alice's signature")
	}
	fmt.Println("Signatures", newDoc.Signatures)
}
