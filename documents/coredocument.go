package documents

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/any"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
)

const (
	// CDRootField represents the coredocument root property of a tree
	CDRootField = "cd_root"

	// DataRootField represents the data root property of a tree
	DataRootField = "data_root"

	// DocumentTypeField represents the doc type property of a tree
	DocumentTypeField = "document_type"

	// SignaturesField represents the signatures property of a tree
	SignaturesField = "signatures"

	// SigningRootField represents the signature root property of a tree
	SigningRootField = "signing_root"

	// idSize represents the size of identifiers, roots etc..
	idSize = 32

	// nftByteCount is the length of combined bytes of registry and tokenID
	nftByteCount = 52

	// CDTreePrefix is the human readable prefix for core doc tree props
	CDTreePrefix = "cd_tree"

	// SigningTreePrefix is the human readable prefix for signing tree props
	SigningTreePrefix = "signing_tree"
)

func compactProperties(key string) []byte {
	m := map[string][]byte{
		CDRootField:       {0, 0, 0, 7},
		DataRootField:     {0, 0, 0, 5},
		DocumentTypeField: {0, 0, 0, 100},
		SignaturesField:   {0, 0, 0, 6},
		SigningRootField:  {0, 0, 0, 10},

		// tree prefixes use the first byte of a 4 byte slice by convention
		CDTreePrefix:      {1, 0, 0, 0},
		SigningTreePrefix: {2, 0, 0, 0},
	}
	return m[key]
}

// CoreDocument is a wrapper for CoreDocument Protobuf.
type CoreDocument struct {
	Document coredocumentpb.CoreDocument
}

// newCoreDocument returns a new CoreDocument.
func newCoreDocument() *CoreDocument {
	id := utils.RandomSlice(idSize)
	cd := coredocumentpb.CoreDocument{
		DocumentIdentifier: id,
		CurrentVersion:     id,
		NextVersion:        utils.RandomSlice(idSize),
	}

	return &CoreDocument{cd}
}

// NewCoreDocumentFromProtobuf returns CoreDocument from the CoreDocument Protobuf.
func NewCoreDocumentFromProtobuf(cd coredocumentpb.CoreDocument) *CoreDocument {
	cd.EmbeddedDataSalts = nil
	cd.EmbeddedData = nil
	return &CoreDocument{Document: cd}
}

// NewCoreDocumentWithCollaborators generates new core Document, adds collaborators, adds read rules and fills salts
func NewCoreDocumentWithCollaborators(collaborators []string) (*CoreDocument, error) {
	cd := newCoreDocument()
	ids, err := identity.NewDIDsFromStrings(collaborators)
	if err != nil {
		return nil, errors.New("failed to decode collaborators: %v", err)
	}

	cd.initReadRules(ids)
	if err := cd.setSalts(); err != nil {
		return nil, err
	}

	return cd, nil
}

// ID returns the Document identifier
func (cd *CoreDocument) ID() []byte {
	return cd.Document.DocumentIdentifier
}

// CurrentVersion returns the current version of the Document
func (cd *CoreDocument) CurrentVersion() []byte {
	return cd.Document.CurrentVersion
}

// PreviousVersion returns the previous version of the Document.
func (cd *CoreDocument) PreviousVersion() []byte {
	return cd.Document.PreviousVersion
}

// NextVersion returns the next version of the Document.
func (cd *CoreDocument) NextVersion() []byte {
	return cd.Document.NextVersion
}

// PreviousDocumentRoot returns the Document root of the previous version.
func (cd *CoreDocument) PreviousDocumentRoot() []byte {
	return cd.Document.PreviousRoot
}

// AppendSignatures appends signatures to core Document.
func (cd *CoreDocument) AppendSignatures(signs ...*coredocumentpb.Signature) {
	cd.Document.Signatures = append(cd.Document.Signatures, signs...)
}

// setSalts generate salts for core Document.
// This is no-op if the salts are already generated.
func (cd *CoreDocument) setSalts() error {
	if cd.Document.CoredocumentSalts != nil {
		return nil
	}

	pSalts, err := GenerateNewSalts(&cd.Document, CDTreePrefix, compactProperties(CDTreePrefix))
	if err != nil {
		return err
	}

	cd.Document.CoredocumentSalts = ConvertToProtoSalts(pSalts)
	return nil
}

// PrepareNewVersion prepares the next version of the CoreDocument
// if initSalts is true, salts will be generated for new version.
func (cd *CoreDocument) PrepareNewVersion(collaborators []string, initSalts bool) (*CoreDocument, error) {
	if len(cd.Document.DocumentRoot) != idSize {
		return nil, errors.New("Document root is invalid")
	}

	cs, err := identity.NewDIDsFromStrings(collaborators)
	if err != nil {
		return nil, err
	}

	// get all the old collaborators
	oldCs, err := cd.GetCollaborators()
	if err != nil {
		return nil, err
	}

	ucs := filterCollaborators(cs, oldCs...)
	cdp := coredocumentpb.CoreDocument{
		DocumentIdentifier: cd.Document.DocumentIdentifier,
		PreviousVersion:    cd.Document.CurrentVersion,
		CurrentVersion:     cd.Document.NextVersion,
		NextVersion:        utils.RandomSlice(32),
		PreviousRoot:       cd.Document.DocumentRoot,
		Roles:              cd.Document.Roles,
		ReadRules:          cd.Document.ReadRules,
		TransitionRules:	cd.Document.TransitionRules,
		Nfts:               cd.Document.Nfts,
	}

	ncd := &CoreDocument{Document: cdp}
	ncd.addCollaboratorsToReadSignRules(ucs)

	if !initSalts {
		return ncd, nil
	}

	err = ncd.setSalts()
	if err != nil {
		return nil, errors.New("failed to init salts: %v", err)
	}

	return ncd, nil
}

// addCollaboratorsToReadSignRules adds the given collaborators to a new read rule with READ_SIGN capability.
// The operation is no-op if no collaborators is provided.
// The operation is not idempotent. So calling twice with same accounts will lead to read rules duplication.
func (cd *CoreDocument) addCollaboratorsToReadSignRules(collaborators []identity.DID) {
	if len(collaborators) == 0 {
		return
	}

	// create a role for given collaborators
	role := new(coredocumentpb.Role)
	role.RoleKey = utils.RandomSlice(idSize)
	for _, c := range collaborators {
		c := c
		role.Collaborators = append(role.Collaborators, c[:])
	}

	cd.addNewRule(role, coredocumentpb.Action_ACTION_READ_SIGN)
}

// addNewRule creates a new rule as per the role and action.
func (cd *CoreDocument) addNewRule(role *coredocumentpb.Role, action coredocumentpb.Action) {
	cd.Document.Roles = append(cd.Document.Roles, role)
	rule := new(coredocumentpb.ReadRule)
	rule.Roles = append(rule.Roles, role.RoleKey)
	rule.Action = action
	cd.Document.ReadRules = append(cd.Document.ReadRules, rule)
}

// CreateProofs takes Document data tree and list to fields and generates proofs.
// we will try generating proofs from the dataTree. If failed, we will generate proofs from CoreDocument.
// errors out when the proof generation is failed on core Document tree.
func (cd *CoreDocument) CreateProofs(docType string, dataTree *proofs.DocumentTree, fields []string) (proofs []*proofspb.Proof, err error) {
	srpHashes, err := cd.getSigningRootProofHashes()
	if err != nil {
		return nil, errors.New("failed to generate signing root proofs: %v", err)
	}

	cdTree, err := cd.documentTree(docType)
	if err != nil {
		return nil, errors.New("failed to generate core Document tree: %v", err)
	}

	dataRoot := dataTree.RootHash()
	cdRoot := cdTree.RootHash()

	// try generating proofs from data root first
	proofs, missedPfs := generateProofs(dataTree, fields, append([][]byte{cdRoot}, srpHashes...))
	if len(missedPfs) == 0 {
		return proofs, nil
	}

	// generate proofs from cdTree. fail if any proofs are missed after this
	pfs, missedPfs := generateProofs(cdTree, missedPfs, append([][]byte{dataRoot}, srpHashes...))
	if len(missedPfs) > 0 {
		return nil, errors.New("failed to generate proofs for %v", missedPfs)
	}

	proofs = append(proofs, pfs...)
	return proofs, nil
}

func generateProofs(tree *proofs.DocumentTree, fields []string, appendHashes [][]byte) (proofs []*proofspb.Proof, missedProofs []string) {
	for _, f := range fields {
		proof, err := tree.CreateProof(f)
		if err != nil {
			// add the missed proof to the map
			missedProofs = append(missedProofs, f)
			continue
		}

		proof.SortedHashes = append(proof.SortedHashes, appendHashes...)
		proofs = append(proofs, &proof)
	}

	return proofs, missedProofs
}

// getSigningRootProofHashes returns the hashes needed to create a proof for fields from SigningRoot to DocumentRoot.
// The returned proofs are appended to the proofs generated from the data tree and core Document tree for a successful verification.
func (cd *CoreDocument) getSigningRootProofHashes() (hashes [][]byte, err error) {
	tree, err := cd.DocumentRootTree()
	if err != nil {
		return
	}

	rootProof, err := tree.CreateProof("signing_root")
	if err != nil {
		return
	}

	return rootProof.SortedHashes, err
}

// DocumentRootTree returns the merkle tree for the Document root.
func (cd *CoreDocument) DocumentRootTree() (tree *proofs.DocumentTree, err error) {
	if len(cd.Document.SigningRoot) != idSize {
		return nil, errors.New("signing root is invalid")
	}

	tree = NewDefaultTree(ConvertToProofSalts(cd.Document.CoredocumentSalts))

	// The first leave added is the signing_root
	err = tree.AddLeaf(proofs.LeafNode{
		Hash:     cd.Document.SigningRoot,
		Hashed:   true,
		Property: NewLeafProperty(SigningRootField, compactProperties(SigningRootField))})
	if err != nil {
		return nil, err
	}

	// For every signature we create a LeafNode
	sigProperty := NewLeafProperty(SignaturesField, compactProperties(SignaturesField))
	sigLeafList := make([]proofs.LeafNode, len(cd.Document.Signatures)+1)
	sigLengthNode := proofs.LeafNode{
		Property: sigProperty.LengthProp(proofs.DefaultSaltsLengthSuffix),
		Salt:     make([]byte, idSize),
		Value:    []byte(fmt.Sprintf("%d", len(cd.Document.Signatures))),
	}
	h := sha256.New()
	err = sigLengthNode.HashNode(h, true)
	if err != nil {
		return nil, err
	}

	sigLeafList[0] = sigLengthNode
	for i, sig := range cd.Document.Signatures {
		payload := sha256.Sum256(append(sig.EntityId, append(sig.PublicKey, sig.Signature...)...))
		leaf := proofs.LeafNode{
			Hash:     payload[:],
			Hashed:   true,
			Property: sigProperty.SliceElemProp(proofs.FieldNumForSliceLength(i)),
		}

		err = leaf.HashNode(h, true)
		if err != nil {
			return nil, err
		}

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

// signingRootTree returns the merkle tree for the signing root.
func (cd *CoreDocument) signingRootTree(docType string) (tree *proofs.DocumentTree, err error) {
	if len(cd.Document.DataRoot) != idSize {
		return nil, errors.New("data root is invalid")
	}

	cdTree, err := cd.documentTree(docType)
	if err != nil {
		return nil, err
	}

	// create the signing tree with data root and coredoc root as siblings
	tree = NewDefaultTreeWithPrefix(ConvertToProofSalts(cd.Document.CoredocumentSalts), SigningTreePrefix, compactProperties(SigningTreePrefix))
	prefixProp := NewLeafProperty(SigningTreePrefix, compactProperties(SigningTreePrefix))

	err = tree.AddLeaves([]proofs.LeafNode{
		{
			Property: prefixProp.FieldProp(DataRootField, binary.LittleEndian.Uint32(compactProperties(DataRootField))),
			Hash:     cd.Document.DataRoot,
			Hashed:   true,
		},
		{
			Property: prefixProp.FieldProp(CDRootField, binary.LittleEndian.Uint32(compactProperties(CDRootField))),
			Hash:     cdTree.RootHash(),
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

// documentTree returns the merkle tree of the core Document.
func (cd *CoreDocument) documentTree(docType string) (tree *proofs.DocumentTree, err error) {
	tree = NewDefaultTreeWithPrefix(ConvertToProofSalts(cd.Document.CoredocumentSalts), CDTreePrefix, compactProperties(CDTreePrefix))
	err = tree.AddLeavesFromDocument(&cd.Document)
	if err != nil {
		return nil, err
	}

	prefixProp := NewLeafProperty(CDTreePrefix, compactProperties(CDTreePrefix))
	// Adding document type as it is an excluded field in the tree
	documentTypeNode := proofs.LeafNode{
		Property: prefixProp.FieldProp(DocumentTypeField, binary.LittleEndian.Uint32(compactProperties(DocumentTypeField))),
		Salt:     make([]byte, 32),
		Value:    []byte(docType),
	}

	err = documentTypeNode.HashNode(sha256.New(), true)
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

// GetSignerCollaborators returns the collaborators excluding the filteredIDs
// returns collaborators with Read_Sign permissions.
func (cd *CoreDocument) GetSignerCollaborators(filterIDs ...identity.DID) ([]identity.DID, error) {
	cs, err := cd.getCollaborators(coredocumentpb.Action_ACTION_READ_SIGN)
	if err != nil {
		return nil, err
	}

	return filterCollaborators(cs, filterIDs...), nil
}

// GetCollaborators returns the collaborators excluding the filteredIDs
// returns collaborators with Read and Read_Sign permissions.
func (cd *CoreDocument) GetCollaborators(filterIDs ...identity.DID) ([]identity.DID, error) {
	cs, err := cd.getCollaborators(coredocumentpb.Action_ACTION_READ_SIGN, coredocumentpb.Action_ACTION_READ)
	if err != nil {
		return nil, err
	}

	return filterCollaborators(cs, filterIDs...), nil
}

// getCollaborators returns all the collaborators who belongs to the actions passed.
func (cd *CoreDocument) getCollaborators(actions ...coredocumentpb.Action) (ids []identity.DID, err error) {
	findRole(cd.Document, func(_, _ int, role *coredocumentpb.Role) bool {
		if len(role.Collaborators) < 1 {
			return false
		}

		for _, c := range role.Collaborators {
			// TODO(ved): we should ideally check the address length of 20
			// we will still keep the error return to the function so that once check is in, we don't have to refactor this function
			ids = append(ids, identity.NewDIDFromBytes(c))
		}

		return false
	}, actions...)

	if err != nil {
		return nil, err
	}

	return ids, nil
}

// filterCollaborators removes the filterIDs if any from cs and returns the result
func filterCollaborators(cs []identity.DID, filterIDs ...identity.DID) (filteredIDs []identity.DID) {
	filter := make(map[string]struct{})
	for _, c := range filterIDs {
		cs := strings.ToLower(c.String())
		filter[cs] = struct{}{}
	}

	for _, id := range cs {
		if _, ok := filter[strings.ToLower(id.String())]; ok {
			continue
		}

		filteredIDs = append(filteredIDs, id)
	}

	return filteredIDs
}

// CalculateDocumentRoot calculates the Document root of the core Document.
func (cd *CoreDocument) CalculateDocumentRoot() ([]byte, error) {
	tree, err := cd.DocumentRootTree()
	if err != nil {
		return nil, err
	}

	cd.Document.DocumentRoot = tree.RootHash()
	return cd.Document.DocumentRoot, nil
}

// SetDataRoot sets the document data root to core document.
func (cd *CoreDocument) SetDataRoot(dr []byte) {
	cd.Document.DataRoot = dr
}

// CalculateSigningRoot calculates the signing root of the core Document.
func (cd *CoreDocument) CalculateSigningRoot(docType string) ([]byte, error) {
	tree, err := cd.signingRootTree(docType)
	if err != nil {
		return nil, err
	}

	cd.Document.SigningRoot = tree.RootHash()
	return cd.Document.SigningRoot, nil
}

// PackCoreDocument prepares the document into a core document.
func (cd *CoreDocument) PackCoreDocument(data *any.Any, salts []*coredocumentpb.DocumentSalt) coredocumentpb.CoreDocument {
	// lets copy the value so that mutations on the returned doc wont be reflected on Document we are holding
	cdp := cd.Document
	cdp.EmbeddedData = data
	cdp.EmbeddedDataSalts = salts
	return cdp
}

// Signatures returns the copy of the signatures on the Document.
func (cd *CoreDocument) Signatures() (signatures []coredocumentpb.Signature) {
	for _, s := range cd.Document.Signatures {
		signatures = append(signatures, *s)
	}

	return signatures
}
