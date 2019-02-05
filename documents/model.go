package documents

import (
	"crypto/sha256"
	"fmt"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"strings"
)

// Model is an interface to abstract away model specificness like invoice or purchaseOrder
// The interface can cast into the type specified by the model if required
// It should only handle protocol-level Document actions
type Model interface {
	storage.Model
	// PackCoreDocument packs the implementing document into a core document
	// should create the identifiers for the core document if not present
	PackCoreDocument() (*coredocumentpb.CoreDocument, error)

	// UnpackCoreDocument must return the document.Model
	// assumes that core document has valid identifiers set
	UnpackCoreDocument(cd *coredocumentpb.CoreDocument) error

	// CreateProofs creates precise-proofs for given fields
	CreateProofs(fields []string) (coreDoc *coredocumentpb.CoreDocument, proofs []*proofspb.Proof, err error)
}

type CoreDocumentModel struct {
	document *coredocumentpb.CoreDocument
}

// New returns a new core document
// Note: collaborators and salts are to be filled by the caller
func newDocModel() *CoreDocumentModel {
	id := utils.RandomSlice(32)
	cd := &coredocumentpb.CoreDocument{
		DocumentIdentifier: id,
		CurrentVersion:     id,
		NextVersion:        utils.RandomSlice(32),
	}
	return &CoreDocumentModel{
		cd,
	}
}

func (m *CoreDocumentModel) GetDocument() *coredocumentpb.CoreDocument {
	return m.document
}

func (m *CoreDocumentModel) SetDocument(updatedDocumentModel *CoreDocumentModel) *CoreDocumentModel{
	ud := updatedDocumentModel.GetDocument()
	m.document = ud
	return m
}

// PrepareNewVersion creates a copy of the passed coreDocument with the version fields updated
// Adds collaborators and fills salts
// Note: new collaborators are added to the list with old collaborators.
func (m *CoreDocumentModel) PrepareNewVersion (collaborators []string) (*CoreDocumentModel, error) {
	ncd := newDocModel()
	ucs, err := fetchUniqueCollaborators(m.document.Collaborators, collaborators)
	if err != nil {
		return nil, errors.New("failed to decode collaborator: %v", err)
	}

	cs := m.document.Collaborators
	for _, c := range ucs {
		c := c
		cs = append(cs, c[:])
	}

	ncd.document.Collaborators = cs

	// copy read rules and roles
	ncd.document.Roles = m.document.Roles
	ncd.document.ReadRules = m.document.ReadRules
	addCollaboratorsToReadSignRules(ncd.document, ucs)

	err = ncd.fillSalts()
	if err != nil {
		return nil, err
	}

	if m.document.DocumentIdentifier == nil {
		return nil, errors.New("coredocument.DocumentIdentifier is nil")
	}
	ncd.document.DocumentIdentifier = m.document.DocumentIdentifier

	if m.document.CurrentVersion == nil {
		return nil, errors.New("coredocument.CurrentVersion is nil")
	}
	ncd.document.PreviousVersion = m.document.CurrentVersion

	if m.document.NextVersion == nil {
		return nil, errors.New("coredocument.NextVersion is nil")
	}
	ncd.document.CurrentVersion = m.document.NextVersion
	ncd.document.NextVersion = utils.RandomSlice(32)
	if m.document.DocumentRoot == nil {
		return nil, errors.New("coredocument.DocumentRoot is nil")
	}
	ncd.document.PreviousRoot = m.document.DocumentRoot
	return ncd, nil
}

// FillSalts creates a new coredocument.Salts and fills it
func (m *CoreDocumentModel) fillSalts() error {
	salts := new(coredocumentpb.CoreDocumentSalts)
	err := proofs.FillSalts(m.document, salts)
	if err != nil {
		return errors.New("failed to fill coredocument salts: %v", err)
	}

	m.document.CoredocumentSalts = salts
	return nil
}

// GetExternalCollaborators returns collaborators of a document without the own centID.
func (m *CoreDocumentModel) GetExternalCollaborators(selfCentID identity.CentID)  ([][]byte, error) {
	var collabs [][]byte

	for _, collab := range m.document.Collaborators {
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

// GetTypeURL returns the type of the embedded document
func (m *CoreDocumentModel) GetTypeURL() (string, error) {

	if m.document == nil {
		return "", errors.New("core document is nil")
	}

	if m.document.EmbeddedData == nil {
		return "", errors.New("core document doesn't have embedded data")
	}

	if m.document.EmbeddedData.TypeUrl == "" {
		return "", errors.New("typeUrl not set properly")
	}

	return m.document.EmbeddedData.TypeUrl, nil
}

func (m *CoreDocumentModel) SetTypeURL(typeURL string) error {
	if m.document == nil {
		return errors.New("core document is nil")
	}

	if m.document.EmbeddedData == nil {
		return errors.New("core document doesn't have embedded data")
	}

	if m.document.EmbeddedData.TypeUrl == "" {
		return errors.New("typeUrl not set properly")
	}

	m.document.EmbeddedData.TypeUrl = typeURL
	return nil
}


// CreateProofs util function that takes document data tree, coreDocument and a list fo fields and generates proofs
func (m *CoreDocumentModel) CreateProofs(dataTree *proofs.DocumentTree, fields []string) (proofs []*proofspb.Proof, err error) {
	dataRootHashes, err := m.getDataProofHashes()
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	signingRootHashes, err := m.getSigningProofHashes()
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	cdtree, err := m.GetDocumentSigningTree()
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
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
					return nil, errors.New("createProofs error %v", err)
				}
				rootHashes = signingRootHashes
			} else {
				return nil, errors.New("createProofs error %v", err)
			}
		}
		proof.SortedHashes = append(proof.SortedHashes, rootHashes...)
		proofs = append(proofs, &proof)
	}

	return proofs, nil
}


// getDataProofHashes returns the hashes needed to create a proof from DataRoot to SigningRoot. This method is used
// to create field proofs
func (m *CoreDocumentModel) getDataProofHashes() (hashes [][]byte, err error) {
	tree, err := m.GetDocumentSigningTree()
	if err != nil {
		return
	}

	signingProof, err := tree.CreateProof("data_root")
	if err != nil {
		return
	}

	rootProofHashes, err := m.getSigningProofHashes()
	if err != nil {
		return
	}

	return append(signingProof.SortedHashes, rootProofHashes...), err
}

// getSigningProofHashes returns the hashes needed to create a proof for fields from SigningRoot to DataRoot. This method is used
// to create field proofs
func (m *CoreDocumentModel) getSigningProofHashes() (hashes [][]byte, err error) {
	tree, err := m.GetDocumentRootTree()
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
func (m *CoreDocumentModel) CalculateSigningRoot() error {
	tree, err := m.GetDocumentSigningTree()
	if err != nil {
		return err
	}

	m.document.SigningRoot = tree.RootHash()
	return nil
}

// CalculateDocumentRoot calculates the document root of the core document
func (m *CoreDocumentModel) CalculateDocumentRoot() error {
	if len(m.document.SigningRoot) != 32 {
		return errors.New("signing root invalid")
	}

	tree, err := m.GetDocumentRootTree()
	if err != nil {
		return err
	}

	m.document.DocumentRoot = tree.RootHash()
	return nil
}

// GetDocumentRootTree returns the merkle tree for the document root
func (m *CoreDocumentModel) GetDocumentRootTree() (tree *proofs.DocumentTree, err error) {
	h := sha256.New()
	t := proofs.NewDocumentTree(proofs.TreeOptions{EnableHashSorting: true, Hash: h})
	tree = &t

	// The first leave added is the signing_root
	err = tree.AddLeaf(proofs.LeafNode{Hash: m.document.SigningRoot, Hashed: true, Property: proofs.NewProperty("signing_root")})
	if err != nil {
		return nil, err
	}
	// For every signature we create a LeafNode
	sigProperty := proofs.NewProperty("signatures")
	sigLeafList := make([]proofs.LeafNode, len(m.document.Signatures)+1)
	sigLengthNode := proofs.LeafNode{
		Property: sigProperty.LengthProp(),
		Salt:     make([]byte, 32),
		Value:    fmt.Sprintf("%d", len(m.document.Signatures)),
	}
	sigLengthNode.HashNode(h, false)
	sigLeafList[0] = sigLengthNode
	for i, sig := range m.document.Signatures {
		payload := sha256.Sum256(append(sig.EntityId, append(sig.PublicKey, sig.Signature...)...))
		leaf := proofs.LeafNode{
			Hash:     payload[:],
			Hashed:   true,
			Property: sigProperty.SliceElemProp(proofs.FieldNumForSliceLength(i)),
		}
		leaf.HashNode(h, false)
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
func (m *CoreDocumentModel) GetDocumentSigningTree() (tree *proofs.DocumentTree, err error) {
	h := sha256.New()
	t := proofs.NewDocumentTree(proofs.TreeOptions{EnableHashSorting: true, Hash: h})
	tree = &t
	err = tree.AddLeavesFromDocument(m.document, m.document.CoredocumentSalts)
	if err != nil {
		return nil, err
	}

	if m.document.EmbeddedData == nil {
		return nil, errors.New("EmbeddedData cannot be nil when generating signing tree")
	}
	// Adding document type as it is an excluded field in the tree
	documentTypeNode := proofs.LeafNode{
		Property: proofs.NewProperty("document_type"),
		Salt:     make([]byte, 32),
		Value:    m.document.EmbeddedData.TypeUrl,
	}
	documentTypeNode.HashNode(h, false)
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
