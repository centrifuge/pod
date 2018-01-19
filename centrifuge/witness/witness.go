package witness

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"log"
	"strings"

	"github.com/xsleonard/go-merkle"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/sha3"
)

/*
SignedDocument is a struct that handles the four most important aspects of a transaction:
1. JSON payload
2. NextVersionID Nonce
3. MerkleRoot
4. Signatures
*/
type SignedDocument struct {
	Payload          string `json:"payload"`
	PreviousRoot     string
	CurrentVersionID string
	NextVersionID    string
	MerkleRoot       string
	Signatures       [][2]string
}

func createRandomByte32  () (out []byte) {
	r := make([]byte, 32)
	_, err := rand.Read(r)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		panic(err)
	}
	return r
}

// SetNextDocumentID sets a nonce that is used to publish an update of the document
func (doc *SignedDocument) SetNextDocumentID() {
	doc.NextVersionID = base64.StdEncoding.EncodeToString(createRandomByte32())
}

// GenerateMerkleRoot creates a merkle root for payload & nonce
func (doc *SignedDocument) GenerateMerkleRoot() (root string) {
	// Set merkle root: replace this with actual merkle root
	merkleItems := doc.FlattenJSON()

	// If there is a previous merkle root, this needs to be included in the merkle tree as the first item:
	if doc.PreviousRoot != "" {
		merkleItems = append([][]byte{[]byte(doc.PreviousRoot)}, merkleItems...)
	}
	tree := merkle.NewTree()
	tree.Generate(merkleItems, sha3.New256())
	root = base64.URLEncoding.EncodeToString(tree.Root().Hash)
	return
}

// PrepareDocument create a new document
func PrepareDocument(payload string) *SignedDocument {
	// Fills the payload with a random string
	doc := new(SignedDocument)
	doc.Payload = payload
	doc.CurrentVersionID = base64.StdEncoding.EncodeToString(createRandomByte32())
	doc.MerkleRoot = doc.GenerateMerkleRoot()
	return doc
}

// UpdateDocument takes an existing document as a template for a new version.
func UpdateDocument(previousDoc *SignedDocument) *SignedDocument {
	doc := new(SignedDocument)
	doc.Payload = previousDoc.Payload
	doc.CurrentVersionID = previousDoc.NextVersionID
	doc.PreviousRoot = previousDoc.MerkleRoot
	doc.SetNextDocumentID()
	doc.MerkleRoot = doc.GenerateMerkleRoot()
	return doc
}

func (doc *SignedDocument) createSignatureData() []byte {
	signatureElements := [][]byte{[]byte(doc.MerkleRoot), []byte(","), []byte(doc.NextVersionID)}
	signatureString := bytes.Join(signatureElements, []byte(""))
	sigArray := []byte(signatureString)
	return sigArray
}

// Sign a document with a provided public key
func (doc *SignedDocument) Sign(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) {
	sigArray := doc.createSignatureData()
	doc.Signatures = append(doc.Signatures, [2]string{base64.URLEncoding.EncodeToString(publicKey), base64.URLEncoding.EncodeToString(ed25519.Sign(privateKey, sigArray))})
}


// WitnessDocument pushes the calculated merkle root to ethereum using the "Witness" contract.
func (doc *SignedDocument) WitnessDocument() {
	version, _ := base64.StdEncoding.DecodeString(doc.CurrentVersionID)
	root, _ := base64.StdEncoding.DecodeString(doc.MerkleRoot)

	var versionArr, rootArr [32]byte
	copy(versionArr[:], version[:32])
	copy(rootArr[:], root[:32])

	contract := GetWitnessContract()
	opts := GetGethKey()
	tx, err := contract.WitnessDocument(opts, versionArr, rootArr)
	if err != nil {
		log.Fatalf("Transaction error")
	}
	fmt.Printf("Transfer pending: 0x%x\n", tx.Hash())
}



// Verify constists of two checks: verify merkleroot & signature
func (doc *SignedDocument) Verify(publicKey ed25519.PublicKey) (verified bool) {
	if !doc.VerifyMerkleRoot() {
		return false
	}
	if !doc.VerifySignature(publicKey) {
		return false
	}
	return true
}

// VerifyMerkleRoot checks if the merkle root matches the calculation
func (doc *SignedDocument) VerifyMerkleRoot() (verified bool) {
	return doc.MerkleRoot == doc.GenerateMerkleRoot()
}

// VerifySignature by checking if it exists on the document and then validating
// it against the provided public key
func (doc *SignedDocument) VerifySignature(publicKey ed25519.PublicKey) (verified bool) {
	// Find signature first
	var signature string
	for i := range doc.Signatures {
		if doc.Signatures[i][0] == base64.URLEncoding.EncodeToString(publicKey) {
			signature = doc.Signatures[i][1]
			// Found!
			break
		}
	}
	if signature == "" {
		return false
	}

	sigArray := doc.createSignatureData()
	sig, _ := base64.URLEncoding.DecodeString(signature)
	verified = ed25519.Verify(publicKey, sigArray, sig[:])
	return verified
}

// keyValueArray is a structure used to serialize JSON Strings. The array is ordered
// by the by the first item of the element (e.g. k in [[k, v], [k, v]])
type keyValue struct {
	Key   string
	Value string
}

type keyValueArray []keyValue

func (s keyValueArray) Len() int {
	return len(s)
}
func (s keyValueArray) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s keyValueArray) Less(i, j int) bool {
	return strings.Compare(s[i].Key, s[j].Key) == -1
}

// FlattenJSON converts a json map string into an array of a certain structure so a Merkle root
// can be calculated. It only deals with strings and doesn't support more than one level.
func (doc *SignedDocument) FlattenJSON() (flattenedArray [][]byte) {
	var mm map[string]interface{}
	json.Unmarshal([]byte(doc.Payload), &mm)

	var flat []keyValue
	for k, v := range mm {
		item := keyValue{k, v.(string)}
		flat = append(flat, item)
	}
	sort.Sort(keyValueArray(flat))
	for _, element := range flat {
		flattenedArray = append(flattenedArray, []byte(element.Key))
		flattenedArray = append(flattenedArray, []byte(element.Value))
	}
	return
}

func (doc *SignedDocument) SerializeDocument() (jsonString string, err error) {
	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		fmt.Println(err)
		return
	}
	jsonString = string(jsonBytes)
	return
}

func (doc *SignedDocument) DeserializeDocument(jsonString string) {
	err := json.Unmarshal([]byte(jsonString), &doc)
	if err != nil {
		fmt.Println(err)
	}
	return
}
