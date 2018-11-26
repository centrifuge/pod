package coredocument

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
)

// getDataProofHashes returns the hashes needed to create a proof from DataRoot to SigningRoot. This method is used
// to create field proofs
func getDataProofHashes(document *coredocumentpb.CoreDocument) (hashes [][]byte, err error) {
	tree, err := GetDocumentSigningTree(document)
	if err != nil {
		return
	}

	signingProof, err := tree.CreateProof("data_root")
	if err != nil {
		return
	}

	rootProofHashes, err := getSigningProofHashes(document)
	if err != nil {
		return
	}

	return append(signingProof.SortedHashes, rootProofHashes...), err
}

// getSigningProofHashes returns the hashes needed to create a proof for fields from SigningRoot to DataRoot. This method is used
// to create field proofs
func getSigningProofHashes(document *coredocumentpb.CoreDocument) (hashes [][]byte, err error) {
	tree, err := GetDocumentRootTree(document)
	if err != nil {
		return
	}
	rootProof, err := tree.CreateProof("signing_root")
	if err != nil {
		return
	}
	return rootProof.SortedHashes, err
}

// CalculateSigningRoot calculates the signing root of the core document
func CalculateSigningRoot(doc *coredocumentpb.CoreDocument) error {
	tree, err := GetDocumentSigningTree(doc)
	if err != nil {
		return err
	}

	doc.SigningRoot = tree.RootHash()
	return nil
}

// CalculateDocumentRoot calculates the document root of the core document
func CalculateDocumentRoot(document *coredocumentpb.CoreDocument) error {
	if len(document.SigningRoot) != 32 {
		return fmt.Errorf("signing root invalid")
	}

	tree, err := GetDocumentRootTree(document)
	if err != nil {
		return err
	}

	document.DocumentRoot = tree.RootHash()
	return nil
}

// GetDocumentRootTree returns the merkle tree for the document root
func GetDocumentRootTree(document *coredocumentpb.CoreDocument) (tree *proofs.DocumentTree, err error) {
	h := sha256.New()
	t := proofs.NewDocumentTree(proofs.TreeOptions{EnableHashSorting: true, Hash: h})
	tree = &t

	// The first leave added is the signing_root
	err = tree.AddLeaf(proofs.LeafNode{Hash: document.SigningRoot, Hashed: true, Property: "signing_root"})
	if err != nil {
		return nil, err
	}
	// For every signature we create a LeafNode
	sigLeafList := make([]proofs.LeafNode, len(document.Signatures)+1)
	sigLengthNode := proofs.LeafNode{
		Property: "signatures.length",
		Salt:     make([]byte, 32),
		Value:    fmt.Sprintf("%d", len(document.Signatures)),
	}
	sigLengthNode.HashNode(h)
	sigLeafList[0] = sigLengthNode
	for i, sig := range document.Signatures {
		payload := sha256.Sum256(append(sig.EntityId, append(sig.PublicKey, sig.Signature...)...))
		leaf := proofs.LeafNode{
			Hash:     payload[:],
			Hashed:   true,
			Property: fmt.Sprintf("signatures[%d]", i),
		}
		leaf.HashNode(h)
		sigLeafList[i+1] = leaf
	}
	err = tree.AddLeaves(sigLeafList)
	if err != nil {
		return nil, err
	}
	err = tree.Generate()
	if err != nil {
		return nil, err
	}
	return tree, nil
}

// GetDocumentSigningTree returns the merkle tree for the signing root
func GetDocumentSigningTree(document *coredocumentpb.CoreDocument) (tree *proofs.DocumentTree, err error) {
	h := sha256.New()
	t := proofs.NewDocumentTree(proofs.TreeOptions{EnableHashSorting: true, Hash: h})
	tree = &t
	err = tree.AddLeavesFromDocument(document, document.CoredocumentSalts)
	if err != nil {
		return nil, err
	}

	if document.EmbeddedData == nil {
		return nil, fmt.Errorf("EmbeddedData cannot be nil when generating signing tree")
	}
	// Adding document type as it is an excluded field in the tree
	documentTypeNode := proofs.LeafNode{
		Property: "document_type",
		Salt:     make([]byte, 32),
		Value:    document.EmbeddedData.TypeUrl,
	}
	documentTypeNode.HashNode(h)
	err = tree.AddLeaf(documentTypeNode)
	if err != nil {
		return nil, err
	}

	err = tree.Generate()
	if err != nil {
		return nil, err
	}
	return tree, nil
}

// PrepareNewVersion creates a copy of the passed coreDocument with the version fields updated
// Adds collaborators and fills salts
// Note: ignores any collaborators in the oldCD
func PrepareNewVersion(oldCD coredocumentpb.CoreDocument, collaborators []string) (*coredocumentpb.CoreDocument, error) {
	newCD, err := NewWithCollaborators(collaborators)
	if err != nil {
		return nil, err
	}

	if oldCD.DocumentIdentifier == nil {
		return nil, fmt.Errorf("coredocument.DocumentIdentifier is nil")
	}
	newCD.DocumentIdentifier = oldCD.DocumentIdentifier

	if oldCD.CurrentVersion == nil {
		return nil, fmt.Errorf("coredocument.CurrentVersion is nil")
	}
	newCD.PreviousVersion = oldCD.CurrentVersion

	if oldCD.NextVersion == nil {
		return nil, fmt.Errorf("coredocument.NextVersion is nil")
	}
	newCD.CurrentVersion = oldCD.NextVersion
	newCD.NextVersion = utils.RandomSlice(32)
	if oldCD.DocumentRoot == nil {
		return nil, fmt.Errorf("coredocument.DocumentRoot is nil")
	}
	newCD.PreviousRoot = oldCD.DocumentRoot
	return newCD, nil
}

// New returns a new core document
// Note: collaborators and salts are to be filled by the caller
func New() *coredocumentpb.CoreDocument {
	id := utils.RandomSlice(32)
	return &coredocumentpb.CoreDocument{
		DocumentIdentifier: id,
		CurrentVersion:     id,
		NextVersion:        utils.RandomSlice(32),
	}
}

// NewWithCollaborators generates new core document, adds collaborators, and fills salts
func NewWithCollaborators(collaborators []string) (*coredocumentpb.CoreDocument, error) {
	cd := New()
	ids, err := identity.CentIDsFromStrings(collaborators)
	if err != nil {
		return nil, fmt.Errorf("failed to decode collaborator: %v", err)
	}

	for i := range ids {
		cd.Collaborators = append(cd.Collaborators, ids[i][:])
	}

	err = FillSalts(cd)
	if err != nil {
		return nil, err
	}

	return cd, nil
}

// GetExternalCollaborators returns collaborators of a document without the own centID.
func GetExternalCollaborators(selfCentID identity.CentID, doc *coredocumentpb.CoreDocument) ([][]byte, error) {
	var collabs [][]byte

	for _, collab := range doc.Collaborators {
		collabID, err := identity.ToCentID(collab)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to CentID: %v", err)
		}
		if !selfCentID.Equal(collabID) {
			collabs = append(collabs, collab)
		}
	}

	return collabs, nil
}

// FillSalts creates a new coredocument.Salts and fills it
func FillSalts(doc *coredocumentpb.CoreDocument) error {
	salts := &coredocumentpb.CoreDocumentSalts{}
	err := proofs.FillSalts(doc, salts)
	if err != nil {
		return fmt.Errorf("failed to fill coredocument salts: %v", err)
	}

	doc.CoredocumentSalts = salts
	return nil
}

// GetTypeURL returns the type of the embedded document
func GetTypeURL(coreDocument *coredocumentpb.CoreDocument) (string, error) {

	if coreDocument == nil {
		return "", fmt.Errorf("core document is nil")
	}

	if coreDocument.EmbeddedData == nil {
		return "", fmt.Errorf("core document doesn't have embedded data")
	}

	if coreDocument.EmbeddedData.TypeUrl == "" {
		return "", fmt.Errorf("typeUrl not set properly")
	}
	return coreDocument.EmbeddedData.TypeUrl, nil
}

// CreateProofs util function that takes document data tree, coreDocument and a list fo fields and generates proofs
func CreateProofs(dataTree *proofs.DocumentTree, coreDoc *coredocumentpb.CoreDocument, fields []string) (proofs []*proofspb.Proof, err error) {
	dataRootHashes, err := getDataProofHashes(coreDoc)
	if err != nil {
		return nil, fmt.Errorf("createProofs error %v", err)
	}

	signingRootHashes, err := getSigningProofHashes(coreDoc)
	if err != nil {
		return nil, fmt.Errorf("createProofs error %v", err)
	}

	cdtree, err := GetDocumentSigningTree(coreDoc)
	if err != nil {
		return nil, fmt.Errorf("createProofs error %v", err)
	}

	// We support fields that belong to different document trees, as we do not prepend a tree prefix to the field, the approach
	// is to try in both trees to find the field and create the proof accordingly
	for _, field := range fields {
		rootHashes := dataRootHashes
		proof, err := dataTree.CreateProof(field)
		if err != nil {
			if strings.Contains(err.Error(), "No such field") {
				proof, err = cdtree.CreateProof(field)
				if err != nil {
					return nil, fmt.Errorf("createProofs error %v", err)
				}
				rootHashes = signingRootHashes
			} else {
				return nil, fmt.Errorf("createProofs error %v", err)
			}
		}
		proof.SortedHashes = append(proof.SortedHashes, rootHashes...)
		proofs = append(proofs, &proof)
	}

	return proofs, nil
}
