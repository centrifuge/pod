package witness

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/xsleonard/go-merkle"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/sha3"
	"github.com/CentrifugeInc/centrifuge-ethereum-contracts/centrifuge/witness"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
)

// SignatureKeyPair is the signature of the merkle root & the associated public key
type SignatureKeyPair struct {
	Key       [32]byte
	Signature [64]byte
}

// SignatureKeyPairArray contains all signatures of the documents merkle root & public keys. The sorting
// is by public key
type SignatureKeyPairArray []SignatureKeyPair

func (s SignatureKeyPairArray) Len() int {
	return len(s)
}
func (s SignatureKeyPairArray) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SignatureKeyPairArray) Less(i, j int) bool {
	return bytes.Compare(s[i].Key[:], s[j].Key[:]) == -1
}

type WitnessExternal interface {
	VerifyWitness() (bool, string)
	WitnessDocument()
}
type WitnessExternalDoc struct {
	doc *SignedDocument
}

/*
SignedDocument is a struct that handles the four most important aspects of a transaction:
1. JSON payload
2. Identifier A random and unique identifier for this document. Each update has a uses a new value
3. NextIdentifier
4. MerkleRoot
5. Signatures
6. WitnessRoot The merkle root of `MerkleRoot` & Signatures
*/
type SignedDocument struct {
	Payload        string `json:"payload"`
	PreviousRoot   [32]byte
	Identifier     [32]byte
	NextIdentifier [32]byte
	MerkleRoot     [32]byte
	Signatures     SignatureKeyPairArray
	WitnessRoot    [32]byte
	WitnessExternal
}

func createRandomByte32() (out [32]byte) {
	r := make([]byte, 32)
	_, err := rand.Read(r)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		panic(err)
	}
	copy(out[:], r[:32])
	return
}

// SetNextIdentifier sets a nonce that is used to publish an update of the document
func (doc *SignedDocument) SetNextIdentifier() {
	doc.NextIdentifier = createRandomByte32()
}

// GenerateMerkleRoot creates a merkle root for payload & nonce
func (doc *SignedDocument) GenerateMerkleRoot() (root [32]byte) {
	merkleItems := doc.FlattenJSON()

	// If there is a previous merkle root, this needs to be included in the merkle tree as the first item:
	if len(doc.PreviousRoot) == 0 {
		previousRoot := make([]byte, 32)
		copy(previousRoot[:], doc.PreviousRoot[:32])
		merkleItems = append([][]byte{previousRoot}, merkleItems...)
	}
	tree := merkle.NewTree()
	tree.Generate(merkleItems, sha3.New256())
	copy(root[:], tree.Root().Hash[:32])
	return
}

// PrepareDocument create a new document
func PrepareDocument(payload string) *SignedDocument {
	// Fills the payload with a random string
	doc := new(SignedDocument)
	doc.Payload = payload
	doc.Identifier = createRandomByte32()
	doc.MerkleRoot = doc.GenerateMerkleRoot()
	doc.WitnessExternal = &WitnessExternalDoc{doc}
	return doc
}

// UpdateDocument takes an existing document as a template to update any data in it
func UpdateDocument(previousDoc *SignedDocument) *SignedDocument {
	doc := new(SignedDocument)
	doc.Payload = previousDoc.Payload
	doc.Identifier = previousDoc.NextIdentifier
	doc.PreviousRoot = previousDoc.MerkleRoot
	doc.SetNextIdentifier()
	doc.MerkleRoot = doc.GenerateMerkleRoot()
	return doc
}

func (doc *SignedDocument) createSignatureData() (signatureData []byte) {
	signatureData = make([]byte, 64)
	copy(signatureData[:32], doc.MerkleRoot[:32])
	copy(signatureData[32:64], doc.NextIdentifier[:32])
	return
}

// Sign a document with a provided public key
func (doc *SignedDocument) Sign(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) {
	sigArray := doc.createSignatureData()
	var key [32]byte
	var signature [64]byte
	copy(key[:], publicKey[:32])
	copy(signature[:], ed25519.Sign(privateKey, sigArray)[:64])
	doc.Signatures = append(doc.Signatures, SignatureKeyPair{key, signature})
}

func (doc *SignedDocument) getSignatureListString() (list []byte) {
	sort.Sort(doc.Signatures)
	for _, keyPair := range doc.Signatures {
		key := keyPair.Key
		signature := keyPair.Signature
		list = append(list, key[:]...)
		list = append(list, signature[:]...)
	}
	return
}

func GetWitnessContract() (witnessContract *witness.EthereumWitness) {
	// Instantiate the contract and display its name
	client := ethereum.GetConnection()
	witnessContract, err := witness.NewEthereumWitness(common.HexToAddress(viper.GetString("witness.ethereum.contractAddress")), client.GetClient())
	if err != nil {
		log.Fatalf("Failed to instantiate the witness contract contract: %v", err)
	}
	return
}

// WitnessDocument pushes the calculated merkle root to ethereum using the "Witness" contract.
func (wes *WitnessExternalDoc) WitnessDocument() {
	var merkleRoot []byte
	copy(merkleRoot[:], wes.doc.MerkleRoot[:32])
	merkleItems := [][]byte{merkleRoot, wes.doc.getSignatureListString()}
	tree := merkle.NewTree()
	tree.Generate(merkleItems, sha3.New256())
	copy(wes.doc.WitnessRoot[:], tree.Root().Hash[:32])

	contract := GetWitnessContract()
	opts := ethereum.GetGethTxOpts()
	tx, err := contract.WitnessDocument(opts, wes.doc.Identifier, wes.doc.WitnessRoot)
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
	verified, err := doc.WitnessExternal.VerifyWitness()
	if err != "" {
		log.Fatal("Error in witness verification")
	}
	if !verified {
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
	var signature [64]byte
	for i := range doc.Signatures {
		if bytes.Equal(doc.Signatures[i].Key[:32], publicKey[:32]) {
			signature = doc.Signatures[i].Signature
			// Found!
			break
		}
	}
	if len(signature) == 0 {
		return false
	}

	signatureData := doc.createSignatureData()
	verified = ed25519.Verify(publicKey, signatureData, signature[:])
	return verified
}

// VerifyWitness checks if the root is present on ethereum and if a root for the next identifier exists.
func (wes *WitnessExternalDoc) VerifyWitness() (verified bool, err string) {
	contract := GetWitnessContract()
	opts := ethereum.GetGethCallOpts()
	data, callErr := contract.GetWitness(opts, wes.doc.Identifier)
	if callErr != nil {
		log.Fatal(callErr)
	}
	if data[0] != wes.doc.WitnessRoot {
		return false, "WitnessRoot doesn't match"
	}
	data, callErr = contract.GetWitness(opts, wes.doc.NextIdentifier)
	if callErr != nil {
		log.Fatal(callErr)
	}
	if data[0] != [32]byte{} {
		return false, "Witnessed Document is not the last version"
	}
	return true, ""
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
