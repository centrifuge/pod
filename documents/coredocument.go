package documents

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

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

	// SignaturesField represents the signatures property of a tree
	SignaturesField = "signatures"

	// SigningRootField represents the signature root property of a tree
	SigningRootField = "signing_root"

	// byteSize represents the size of identifiers, roots etc..
	byteSize = 32

	// ErrZeroCollaborators error when no collaborators are passed
	ErrZeroCollaborators = errors.Error("require at least one collaborator")

	// ErrNFTRoleMissing errors when role to generate proof doesn't exist
	ErrNFTRoleMissing = errors.Error("NFT Role doesn't exist")

	// nftByteCount is the length of combined bytes of registry and tokenID
	nftByteCount = 52
)

var compactProperties = map[string][]byte{
	CDRootField:       {0, 0, 0, 7},
	DataRootField:     {0, 0, 0, 5},
	DocumentTypeField: {0, 0, 0, 100},
	SignaturesField:   {0, 0, 0, 6},
	SigningRootField:  {0, 0, 0, 10},
}

// CoreDocument holds the protobuf coredocument.
type CoreDocument struct {
	document coredocumentpb.CoreDocument
}

// jsonCD is an intermediate document type used for marshalling and un-marshaling CoreDocument to/from json.
type jsonCD struct {
	CoreDocument coredocumentpb.CoreDocument `json:"core_document"`
}

// MarshalJSON returns a JSON formatted representation of the CoreDocument
func (cd *CoreDocument) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonCD{CoreDocument: cd.document})
}

// UnmarshalJSON loads the json formatted CoreDocument.
func (cd *CoreDocument) UnmarshalJSON(data []byte) error {
	jcd := new(jsonCD)
	err := json.Unmarshal(data, jcd)
	if err != nil {
		return err
	}

	cd.document = jcd.CoreDocument
	return nil
}

// newCoreDocument returns a new CoreDocument.
func newCoreDocument() *CoreDocument {
	id := utils.RandomSlice(byteSize)
	cd := coredocumentpb.CoreDocument{
		DocumentIdentifier: id,
		CurrentVersion:     id,
		NextVersion:        utils.RandomSlice(byteSize),
	}

	return &CoreDocument{cd}
}

// NewCoreDocumentWithCollaborators generates new core document, adds collaborators, adds read rules and fills salts
func NewCoreDocumentWithCollaborators(collaborators []string) (*CoreDocument, error) {
	cd := newCoreDocument()
	ids, err := identity.CentIDsFromStrings(collaborators)
	if err != nil {
		return nil, errors.New("failed to decode collaborators: %v", err)
	}

	for i := range ids {
		cd.document.Collaborators = append(cd.document.Collaborators, ids[i][:])
	}

	cd.initReadRules(ids)
	if err := cd.setSalts(); err != nil {
		return nil, err
	}

	return cd, nil
}

// setSalts generate salts for core document.
// This is no-op if the salts are already generated.
func (cd *CoreDocument) setSalts() error {
	if cd.document.CoredocumentSalts != nil {
		return nil
	}

	pSalts, err := GenerateNewSalts(&cd.document, "", nil)
	if err != nil {
		return err
	}

	cd.document.CoredocumentSalts = ConvertToProtoSalts(pSalts)
	return nil
}

// PrepareNewVersion prepares the next version of the CoreDocument
// Note: salts needs to be filled by the caller
func (cd *CoreDocument) PrepareNewVersion(collaborators []string, initSalts bool) (*CoreDocument, error) {
	if len(cd.document.DocumentRoot) != byteSize {
		return nil, errors.New("Document root is invalid")
	}

	ucs, err := fetchUniqueAccounts(cd.document.Collaborators, collaborators)
	if err != nil {
		return nil, errors.New("failed to fetch new collaborators: %v", err)
	}

	cs := cd.document.Collaborators
	for _, c := range ucs {
		c := c
		cs = append(cs, c[:])
	}

	cdp := coredocumentpb.CoreDocument{
		DocumentIdentifier: cd.document.DocumentIdentifier,
		PreviousVersion:    cd.document.CurrentVersion,
		CurrentVersion:     cd.document.NextVersion,
		NextVersion:        utils.RandomSlice(32),
		PreviousRoot:       cd.document.DocumentRoot,
		Collaborators:      cs,
		Roles:              cd.document.Roles,
		ReadRules:          cd.document.ReadRules,
		Nfts:               cd.document.Nfts,
	}

	ncd := &CoreDocument{document: cdp}
	cd.addCollaboratorsToReadSignRules(ucs)

	if !initSalts {
		return ncd, nil
	}

	err = cd.setSalts()
	if err != nil {
		return nil, errors.New("failed to init salts: %v", err)
	}

	return ncd, nil
}

// addCollaboratorsToReadSignRules adds the given collaborators to a new read rule with READ_SIGN capability.
// The operation is no-op if no collaborators is provided.
// The operation is not idempotent. So calling twice with same accounts will lead to read rules duplication.
func (cd *CoreDocument) addCollaboratorsToReadSignRules(collaborators []identity.CentID) {
	if len(collaborators) == 0 {
		return
	}

	// create a role for given collaborators
	role := new(coredocumentpb.Role)
	role.RoleKey = utils.RandomSlice(byteSize)
	for _, c := range collaborators {
		c := c
		role.Collaborators = append(role.Collaborators, c[:])
	}

	cd.addNewRule(role, coredocumentpb.Action_ACTION_READ_SIGN)
}

// addNewRule creates a new rule as per the role and action.
func (cd *CoreDocument) addNewRule(role *coredocumentpb.Role, action coredocumentpb.Action) {
	cd.document.Roles = append(cd.document.Roles, role)
	rule := new(coredocumentpb.ReadRule)
	rule.Roles = append(rule.Roles, role.RoleKey)
	rule.Action = action
	cd.document.ReadRules = append(cd.document.ReadRules, rule)
}

// fetchUniqueAccounts fetches the unique accounts that are not present in oldAccounts.
func fetchUniqueAccounts(oldAccounts [][]byte, newAccounts []string) (uniqueAccounts []identity.CentID, err error) {
	ocsm := make(map[string]struct{})
	for _, c := range oldAccounts {
		ocsm[hexutil.Encode(c)] = struct{}{}
	}

	var uc []string
	for _, c := range newAccounts {
		if _, ok := ocsm[c]; ok {
			continue
		}

		uc = append(uc, c)
	}

	for _, c := range uc {
		account, err := identity.CentIDFromString(c)
		if err != nil {
			return nil, err
		}

		uniqueAccounts = append(uniqueAccounts, account)
	}

	return uniqueAccounts, nil
}

// GenerateProofs takes document data tree and list to fields and generates proofs.
// we will try generating proofs from the dataTree. If failed, we will generate proofs from CoreDocument.
// errors out when the proof generation is failed on core Document tree.
func (cd *CoreDocument) GenerateProofs(dataTree *proofs.DocumentTree, fields []string) (proofs []*proofspb.Proof, err error) {
	srpHashes, err := cd.getSigningRootProofHashes()
	if err != nil {
		return nil, errors.New("failed to generate signing root proofs: %v", err)
	}

	cdTree, err := cd.documentTree()
	if err != nil {
		return nil, errors.New("failed to generate core document tree: %v", err)
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
	if len(missedPfs) != 0 {
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
// The returned proofs are appended to the proofs generated from the data tree and core document tree for a successful verification.
func (cd *CoreDocument) getSigningRootProofHashes() (hashes [][]byte, err error) {
	tree, err := cd.documentRootTree()
	if err != nil {
		return
	}

	rootProof, err := tree.CreateProof("signing_root")
	if err != nil {
		return
	}

	return rootProof.SortedHashes, err
}

// documentRootTree returns the merkle tree for the document root.
func (cd *CoreDocument) documentRootTree() (tree *proofs.DocumentTree, err error) {
	if len(cd.document.SigningRoot) != byteSize {
		return nil, errors.New("signing root is invalid")
	}

	tree = NewDefaultTree(ConvertToProofSalts(cd.document.CoredocumentSalts))
	// The first leave added is the signing_root
	err = tree.AddLeaf(proofs.LeafNode{
		Hash:     cd.document.SigningRoot,
		Hashed:   true,
		Property: NewLeafProperty(SigningRootField, compactProperties[SigningRootField]),
	})
	if err != nil {
		return nil, err
	}

	// For every signature we create a LeafNode
	sigProperty := NewLeafProperty(SignaturesField, compactProperties[SignaturesField])
	sigLeafList := make([]proofs.LeafNode, len(cd.document.Signatures)+1)
	sigLengthNode := proofs.LeafNode{
		Property: sigProperty.LengthProp(proofs.DefaultSaltsLengthSuffix),
		Salt:     make([]byte, byteSize),
		Value:    []byte(fmt.Sprintf("%d", len(cd.document.Signatures))),
	}

	h := sha256.New()
	err = sigLengthNode.HashNode(h, true)
	if err != nil {
		return nil, err
	}

	sigLeafList[0] = sigLengthNode
	for i, sig := range cd.document.Signatures {
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
func (cd *CoreDocument) signingRootTree() (tree *proofs.DocumentTree, err error) {
	if len(cd.document.DataRoot) != byteSize {
		return nil, errors.New("data root is invalid")
	}

	cdTree, err := cd.documentTree()
	if err != nil {
		return nil, err
	}

	// create the signing tree with data root and coredoc root as siblings
	tree = NewDefaultTree(ConvertToProofSalts(cd.document.CoredocumentSalts))
	err = tree.AddLeaves([]proofs.LeafNode{
		{
			Property: NewLeafProperty(DataRootField, compactProperties[DataRootField]),
			Hash:     cd.document.DataRoot,
			Hashed:   true,
		},
		{
			Property: NewLeafProperty(CDRootField, compactProperties[CDRootField]),
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

// documentTree returns the merkle tree of the core document.
func (cd *CoreDocument) documentTree() (tree *proofs.DocumentTree, err error) {
	if cd.document.EmbeddedData == nil {
		return nil, errors.New("EmbeddedData cannot be nil when generating document tree")
	}

	tree = NewDefaultTree(ConvertToProofSalts(cd.document.CoredocumentSalts))
	err = tree.AddLeavesFromDocument(&cd.document)
	if err != nil {
		return nil, err
	}

	// Adding document type as it is an excluded field in the tree
	documentTypeNode := proofs.LeafNode{
		Property: NewLeafProperty(DocumentTypeField, compactProperties[DocumentTypeField]),
		Salt:     make([]byte, 32),
		Value:    []byte(cd.document.EmbeddedData.TypeUrl),
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

// calculateDocumentRoot calculates the document root of the core document.
func (cd *CoreDocument) calculateDocumentRoot() error {
	tree, err := cd.documentRootTree()
	if err != nil {
		return err
	}

	cd.document.DocumentRoot = tree.RootHash()
	return nil
}

// calculateSigningRoot calculates the signing root of the core document
func (cd *CoreDocument) calculateSigningRoot() error {
	tree, err := cd.signingRootTree()
	if err != nil {
		return err
	}

	cd.document.SigningRoot = tree.RootHash()
	return nil
}
