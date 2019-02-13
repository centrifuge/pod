package coredocument

import (
	"crypto/sha256"
	"strings"

	"github.com/golang/protobuf/proto"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	// CDRootField represents the coredocument root property of a tree
	CDRootField = "cd_root"
	// DataRootField represents the data root property of a tree
	DataRootField = "data_root"
	// DocumentTypeField represents the doc type property of a tree
	DocumentTypeField = "document_type"
	// SignaturesField represents the signatures_data_root property of a tree
	SignatureDataRootField = "signatures_data_root"
	// SigningRootField represents the signature root property of a tree
	SigningRootField = "signing_root"
)

var compactProperties = map[string][]byte{
	CDRootField:       {0, 0, 0, 7},
	DataRootField:     {0, 0, 0, 5},
	DocumentTypeField: {0, 0, 0, 100},
	SignatureDataRootField:   {0, 0, 0, 6},
	SigningRootField:  {0, 0, 0, 10},
}

// NewDefaultTree returns a DocumentTree with default opts
func NewDefaultTree(salts *proofs.Salts) *proofs.DocumentTree {
	return NewDefaultTreeWithPrefix(salts, "", nil)
}

// NewDefaultTreeWithPrefix returns a DocumentTree with default opts passing a prefix to the tree leaves
func NewDefaultTreeWithPrefix(salts *proofs.Salts, prefix string, compactPrefix []byte) *proofs.DocumentTree {
	var prop proofs.Property
	if prefix != "" {
		prop = NewLeafProperty(prefix, compactPrefix)
	}

	t := proofs.NewDocumentTree(proofs.TreeOptions{CompactProperties: true, EnableHashSorting: true, Hash: sha256.New(), ParentPrefix: prop, Salts: salts})
	return &t
}

// NewLeafProperty returns a proof property with the literal and the compact
func NewLeafProperty(literal string, compact []byte) proofs.Property {
	return proofs.NewProperty(literal, compact...)
}

// getDataProofHashes returns the hashes needed to create a proof from DataRoot to SigningRoot. This method is used
// to create field proofs
func getDataProofHashes(document *coredocumentpb.CoreDocument, dataRoot []byte) (hashes [][]byte, err error) {
	tree, err := GetDocumentSigningTree(document, dataRoot)
	if err != nil {
		return
	}

	signingProof, err := tree.CreateProof("data_root")
	if err != nil {
		return
	}

	rootProofHashes, err := getSigningRootProofHashes(document)
	if err != nil {
		return
	}

	return append(signingProof.SortedHashes, rootProofHashes...), err
}

// getSigningRootProofHashes returns the hashes needed to create a proof for fields from SigningRoot to DocumentRoot. This method is used
// to create field proofs
func getSigningRootProofHashes(document *coredocumentpb.CoreDocument) (hashes [][]byte, err error) {
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
func CalculateSigningRoot(doc *coredocumentpb.CoreDocument, dataRoot []byte) error {
	tree, err := GetDocumentSigningTree(doc, dataRoot)
	if err != nil {
		return err
	}

	doc.SigningRoot = tree.RootHash()
	return nil
}

// CalculateDocumentRoot calculates the document root of the core document
func CalculateDocumentRoot(document *coredocumentpb.CoreDocument) error {
	if len(document.SigningRoot) != 32 {
		return errors.New("signing root invalid")
	}

	tree, err := GetDocumentRootTree(document)
	if err != nil {
		return err
	}

	document.DocumentRoot = tree.RootHash()
	return nil
}

func getSignatureDataSalts(document *coredocumentpb.CoreDocument) ([]*coredocumentpb.DocumentSalt, error) {
	if document.SignatureDataSalts == nil {
		proofSalts, err := GenerateNewSalts(document.SignatureData, "", nil)
		if err != nil {
			return nil, err
		}
		document.SignatureDataSalts = ConvertToProtoSalts(proofSalts)
	}
	return document.SignatureDataSalts, nil
}

func GetSignatureDataTree(document *coredocumentpb.CoreDocument) (*proofs.DocumentTree, error) {
	signatureSalts, err := getSignatureDataSalts(document)
	tree := NewDefaultTree(ConvertToProofSalts(signatureSalts))
	if err != nil {
		return nil, err
	}

	err = tree.AddLeavesFromDocument(document.SignatureData)
	if err != nil {
		return nil, err
	}

	err = tree.Generate()
	if err != nil {
		return nil, err
	}
	return tree, nil
}

// GetDocumentRootTree returns the merkle tree for the document root
func GetDocumentRootTree(document *coredocumentpb.CoreDocument) (tree *proofs.DocumentTree, err error) {
	tree = NewDefaultTree(ConvertToProofSalts(document.CoredocumentSalts))

	// The first leave added is the signing_root
	err = tree.AddLeaf(proofs.LeafNode{Hash: document.SigningRoot, Hashed: true, Property: NewLeafProperty(SigningRootField, compactProperties[SigningRootField])})
	if err != nil {
		return nil, err
	}

	signatureTree, err := GetSignatureDataTree(document)
	if err != nil {
		return nil, err
	}
	// The second leave added is the signatures_data_root
	err = tree.AddLeaf(proofs.LeafNode{Hash: signatureTree.RootHash(), Hashed: true, Property: NewLeafProperty(SignatureDataRootField, compactProperties[SignatureDataRootField])})
	if err != nil {
		return nil, err
	}

	err = tree.Generate()
	if err != nil {
		return nil, err
	}
	return tree, nil
}

// GetCoreDocTree returns the merkle tree for the coredoc root
func GetCoreDocTree(document *coredocumentpb.CoreDocument) (tree *proofs.DocumentTree, err error) {
	h := sha256.New()
	tree = NewDefaultTree(ConvertToProofSalts(document.CoredocumentSalts))
	err = tree.AddLeavesFromDocument(document)
	if err != nil {
		return nil, err
	}

	if document.EmbeddedData == nil {
		return nil, errors.New("EmbeddedData cannot be nil when generating signing tree")
	}
	// Adding document type as it is an excluded field in the tree
	documentTypeNode := proofs.LeafNode{
		Property: NewLeafProperty(DocumentTypeField, compactProperties[DocumentTypeField]),
		Salt:     make([]byte, 32),
		Value:    []byte(document.EmbeddedData.TypeUrl),
	}

	err = documentTypeNode.HashNode(h, true)
	if err != nil {
		return nil, err
	}

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

// GetDocumentSigningTree returns the merkle tree for the signing root
func GetDocumentSigningTree(document *coredocumentpb.CoreDocument, dataRoot []byte) (tree *proofs.DocumentTree, err error) {
	// coredoc tree
	coreDocTree, err := GetCoreDocTree(document)
	if err != nil {
		return nil, err
	}

	// create the signing tree with data root and coredoc root as siblings
	tree = NewDefaultTree(ConvertToProofSalts(document.CoredocumentSalts))
	err = tree.AddLeaves([]proofs.LeafNode{
		{
			Property: NewLeafProperty(DataRootField, compactProperties[DataRootField]),
			Hash:     dataRoot,
			Hashed:   true,
		},
		{
			Property: NewLeafProperty(CDRootField, compactProperties[CDRootField]),
			Hash:     coreDocTree.RootHash(),
			Hashed:   true,
		},
	})

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
// Note: new collaborators are added to the list with old collaborators.
func PrepareNewVersion(oldCD coredocumentpb.CoreDocument, collaborators []string) (*coredocumentpb.CoreDocument, error) {
	ncd := New()
	ucs, err := fetchUniqueCollaborators(oldCD.Collaborators, collaborators)
	if err != nil {
		return nil, errors.New("failed to decode collaborator: %v", err)
	}

	cs := oldCD.Collaborators
	for _, c := range ucs {
		c := c
		cs = append(cs, c[:])
	}

	ncd.Collaborators = cs

	// copy read rules and roles
	ncd.Roles = oldCD.Roles
	ncd.ReadRules = oldCD.ReadRules
	err = addCollaboratorsToReadSignRules(ncd, ucs)
	if err != nil {
		return nil, err
	}

	err = FillSalts(ncd)
	if err != nil {
		return nil, err
	}

	if oldCD.DocumentIdentifier == nil {
		return nil, errors.New("coredocument.DocumentIdentifier is nil")
	}
	ncd.DocumentIdentifier = oldCD.DocumentIdentifier

	if oldCD.CurrentVersion == nil {
		return nil, errors.New("coredocument.CurrentVersion is nil")
	}
	ncd.PreviousVersion = oldCD.CurrentVersion

	if oldCD.NextVersion == nil {
		return nil, errors.New("coredocument.NextVersion is nil")
	}
	ncd.CurrentVersion = oldCD.NextVersion
	ncd.NextVersion = utils.RandomSlice(32)
	if oldCD.DocumentRoot == nil {
		return nil, errors.New("coredocument.DocumentRoot is nil")
	}
	ncd.PreviousRoot = oldCD.DocumentRoot
	return ncd, nil
}

// New returns a new core document
// Note: collaborators and salts are to be filled by the caller
func New() *coredocumentpb.CoreDocument {
	id := utils.RandomSlice(32)
	return &coredocumentpb.CoreDocument{
		DocumentIdentifier: id,
		CurrentVersion:     id,
		NextVersion:        utils.RandomSlice(32),
		SignatureData:      new(coredocumentpb.SignatureData),
	}
}

// NewWithCollaborators generates new core document, adds collaborators, adds read rules and fills salts
func NewWithCollaborators(collaborators []string) (*coredocumentpb.CoreDocument, error) {
	cd := New()
	ids, err := identity.CentIDsFromStrings(collaborators)
	if err != nil {
		return nil, errors.New("failed to decode collaborator: %v", err)
	}

	for i := range ids {
		cd.Collaborators = append(cd.Collaborators, ids[i][:])
	}

	err = initReadRules(cd, ids)
	if err != nil {
		return nil, errors.New("failed to init read rules: %v", err)
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
			return nil, errors.New("failed to convert to CentID: %v", err)
		}
		if !selfCentID.Equal(collabID) {
			collabs = append(collabs, collab)
		}
	}

	return collabs, nil
}

// GenerateNewSalts generates salts for new document
func GenerateNewSalts(document proto.Message, prefix string, compactPrefix []byte) (*proofs.Salts, error) {
	docSalts := &proofs.Salts{}
	t := NewDefaultTreeWithPrefix(docSalts, prefix, compactPrefix)
	err := t.AddLeavesFromDocument(document)
	if err != nil {
		return nil, err
	}
	return docSalts, nil
}

// ConvertToProtoSalts converts proofSalts into protocolSalts
func ConvertToProtoSalts(proofSalts *proofs.Salts) []*coredocumentpb.DocumentSalt {
	protoSalts := make([]*coredocumentpb.DocumentSalt, len(*proofSalts))
	if proofSalts == nil {
		return nil
	}

	for i, pSalt := range *proofSalts {
		protoSalts[i] = &coredocumentpb.DocumentSalt{Value: pSalt.Value, Compact: pSalt.Compact}
	}

	return protoSalts
}

// ConvertToProofSalts converts protocolSalts into proofSalts
func ConvertToProofSalts(protoSalts []*coredocumentpb.DocumentSalt) *proofs.Salts {
	proofSalts := make(proofs.Salts, len(protoSalts))
	if protoSalts == nil {
		return nil
	}

	for _, pSalt := range protoSalts {
		proofSalts = append(proofSalts, proofs.Salt{Value: pSalt.Value, Compact: pSalt.Compact})
	}

	return &proofSalts
}

// FillSalts creates a new coredocument.Salts and fills it
func FillSalts(doc *coredocumentpb.CoreDocument) error {
	salts, err := GenerateNewSalts(doc, "", nil)
	if err != nil {
		return err
	}

	doc.CoredocumentSalts = ConvertToProtoSalts(salts)
	return nil
}

// GetTypeURL returns the type of the embedded document
func GetTypeURL(coreDocument *coredocumentpb.CoreDocument) (string, error) {

	if coreDocument == nil {
		return "", errors.New("core document is nil")
	}

	if coreDocument.EmbeddedData == nil {
		return "", errors.New("core document doesn't have embedded data")
	}

	if coreDocument.EmbeddedData.TypeUrl == "" {
		return "", errors.New("typeUrl not set properly")
	}
	return coreDocument.EmbeddedData.TypeUrl, nil
}

// CreateProofs util function that takes document data tree, coreDocument and a list fo fields and generates proofs
func CreateProofs(dataTree *proofs.DocumentTree, coreDoc *coredocumentpb.CoreDocument, fields []string) (proofs []*proofspb.Proof, err error) {
	signingRootProofHashes, err := getSigningRootProofHashes(coreDoc)
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	cdtree, err := GetCoreDocTree(coreDoc)
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	dataRoot := dataTree.RootHash()
	cdRoot := cdtree.RootHash()

	// We support fields that belong to different document trees, as we do not prepend a tree prefix to the field, the approach
	// is to try in both trees to find the field and create the proof accordingly
	for _, field := range fields {
		proof, err := dataTree.CreateProof(field)
		if err != nil {
			if strings.Contains(err.Error(), "No such field") {
				proof, err = cdtree.CreateProof(field)
				if err != nil {
					return nil, errors.New("createProofs error %v", err)
				}
				proof.SortedHashes = append(proof.SortedHashes, dataRoot)
			} else {
				return nil, errors.New("createProofs error %v", err)
			}
		} else {
			proof.SortedHashes = append(proof.SortedHashes, cdRoot)
		}
		proof.SortedHashes = append(proof.SortedHashes, signingRootProofHashes...)
		proofs = append(proofs, &proof)
	}

	return proofs, nil
}

func fetchUniqueCollaborators(oldCollabs [][]byte, newCollabs []string) (ids []identity.CentID, err error) {
	ocsm := make(map[string]struct{})
	for _, c := range oldCollabs {
		ocsm[hexutil.Encode(c)] = struct{}{}
	}

	var uc []string
	for _, c := range newCollabs {
		if _, ok := ocsm[c]; ok {
			continue
		}

		uc = append(uc, c)
	}

	for _, c := range uc {
		id, err := identity.CentIDFromString(c)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}
