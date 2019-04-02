package documents

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/golang/protobuf/ptypes/any"
)

const (
	// CDRootField represents the coredocument root property of a tree
	CDRootField = "cd_root"

	// DataRootField represents the data root property of a tree
	DataRootField = "data_root"

	// DocumentTypeField represents the doc type property of a tree
	DocumentTypeField = "document_type"

	// SignaturesRootField represents the signatures property of a tree
	SignaturesRootField = "signatures_root"

	// SigningRootField represents the signature root property of a tree
	SigningRootField = "signing_root"

	// idSize represents the size of identifiers, roots etc..
	idSize = 32

	// nftByteCount is the length of combined bytes of registry and tokenID
	nftByteCount = 52

	// DRTreePrefix is the human readable prefix for document root tree props
	DRTreePrefix = "dr_tree"

	// CDTreePrefix is the human readable prefix for core doc tree props
	CDTreePrefix = "cd_tree"

	// SigningTreePrefix is the human readable prefix for signing tree props
	SigningTreePrefix = "signing_tree"

	// SignaturesTreePrefix is the human readable prefix for signature props
	SignaturesTreePrefix = "signatures_tree"
)

// CompactProperties returns the compact property for a given prefix
func CompactProperties(key string) []byte {
	m := map[string][]byte{
		CDRootField:         {0, 0, 0, 7},
		DataRootField:       {0, 0, 0, 5},
		DocumentTypeField:   {0, 0, 0, 100},
		SignaturesRootField: {0, 0, 0, 6},
		SigningRootField:    {0, 0, 0, 10},

		// tree prefixes use the first byte of a 4 byte slice by convention
		CDTreePrefix:         {1, 0, 0, 0},
		SigningTreePrefix:    {2, 0, 0, 0},
		SignaturesTreePrefix: {3, 0, 0, 0},
		DRTreePrefix:         {4, 0, 0, 0},
	}
	return m[key]
}

func allowedAttributeTypes(typ string) (reflect.Type, error) {
	switch typ {
	case "int256":
		return reflect.TypeOf(big.Int{}), nil
	case "bigdecimal":
		return reflect.TypeOf(Decimal{}), nil
	case "string":
		return reflect.TypeOf(""), nil
	case "bytes":
		return reflect.TypeOf([]byte{}), nil
	case "timestamp":
		return reflect.TypeOf(time.Time{}), nil
	default:
		return nil, errors.NewTypedError(ErrCDAttribute, errors.New("can't find the given attribute in allowed attribute types"))
	}
}

func allowedAttributeTypeNames(typ reflect.Type) (string, error) {
	switch typ {
	case reflect.TypeOf(big.Int{}):
		return "int256", nil
	case reflect.TypeOf(Decimal{}):
		return "bigdecimal", nil
	case reflect.TypeOf(""):
		return "string", nil
	case reflect.TypeOf([]byte{}):
		return "bytes", nil
	case reflect.TypeOf(time.Time{}):
		return "timestamp", nil
	default:
		return "", errors.NewTypedError(ErrCDAttribute, errors.New("can't find the given attribute in allowed attribute type names"))
	}
}

// Attribute represents a custom attribute of a document
type Attribute struct {
	Type        reflect.Type
	ReadableKey string
	HashedKey   []byte
	Value       interface{}
}

// NewAttribute creates a new custom attribute
func NewAttribute(readableKey string, attributeType reflect.Type, value interface{}) (*Attribute, error) {
	if readableKey == "" {
		return nil, errors.NewTypedError(ErrCDAttribute, errors.New("can't create attribute with an empty string as name"))
	}

	if value == nil {
		return nil, errors.NewTypedError(ErrCDAttribute, errors.New("can't create attribute with a nil value"))
	}

	if reflect.TypeOf(value) != attributeType {
		return nil, errors.NewTypedError(ErrCDAttribute, errors.New("provided type doesn't match the actual type of the value"))
	}

	h := sha256.New()
	_, err := h.Write([]byte(readableKey))
	if err != nil {
		return nil, errors.NewTypedError(ErrCDAttribute, err)
	}

	hashedKey := h.Sum(nil)
	return &Attribute{
		ReadableKey: readableKey,
		HashedKey:   hashedKey,
		Type:        attributeType,
		Value:       value,
	}, nil
}

// CoreDocument is a wrapper for CoreDocument Protobuf.
type CoreDocument struct {
	// Modified indicates that the CoreDocument has been modified and salts needs to be generated for new fields in coredoc precise-proof tree.
	Modified bool

	// Attributes are the custom attributes added to the document
	Attributes map[string]*Attribute

	Document coredocumentpb.CoreDocument
}

// CollaboratorsAccess allows us to differentiate between the types of access we want to give new collaborators
type CollaboratorsAccess struct {
	ReadCollaborators      []identity.DID
	ReadWriteCollaborators []identity.DID
}

// newCoreDocument returns a new CoreDocument.
func newCoreDocument() (*CoreDocument, error) {
	cd := coredocumentpb.CoreDocument{
		SignatureData: new(coredocumentpb.SignatureData),
	}
	err := populateVersions(&cd, nil)
	if err != nil {
		return nil, err
	}

	return &CoreDocument{Document: cd, Modified: true}, nil
}

// NewCoreDocumentFromProtobuf returns CoreDocument from the CoreDocument Protobuf.
func NewCoreDocumentFromProtobuf(cd coredocumentpb.CoreDocument) *CoreDocument {
	cd.EmbeddedData = nil
	return &CoreDocument{Document: cd}
}

// NewCoreDocumentWithCollaborators generates new core document with a document type specified by the prefix: po or invoice.
// It then adds collaborators, adds read rules and fills salts.
func NewCoreDocumentWithCollaborators(documentPrefix []byte, collaborators CollaboratorsAccess) (*CoreDocument, error) {
	cd, err := newCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(ErrCDCreate, errors.New("failed to create coredoc: %v", err))
	}

	cd.initReadRules(append(collaborators.ReadCollaborators, collaborators.ReadWriteCollaborators...))
	cd.initTransitionRules(documentPrefix, collaborators.ReadWriteCollaborators)
	return cd, nil
}

// ID returns the document identifier
func (cd *CoreDocument) ID() []byte {
	return cd.Document.DocumentIdentifier
}

// CurrentVersion returns the current version of the document
func (cd *CoreDocument) CurrentVersion() []byte {
	return cd.Document.CurrentVersion
}

// CurrentVersionPreimage returns the current version preimage of the document
func (cd *CoreDocument) CurrentVersionPreimage() []byte {
	return cd.Document.CurrentPreimage
}

// PreviousVersion returns the previous version of the document.
func (cd *CoreDocument) PreviousVersion() []byte {
	return cd.Document.PreviousVersion
}

// NextVersion returns the next version of the document.
func (cd *CoreDocument) NextVersion() []byte {
	return cd.Document.NextVersion
}

// AppendSignatures appends signatures to core document.
func (cd *CoreDocument) AppendSignatures(signs ...*coredocumentpb.Signature) {
	if cd.Document.SignatureData == nil {
		cd.Document.SignatureData = new(coredocumentpb.SignatureData)
	}
	cd.Document.SignatureData.Signatures = append(cd.Document.SignatureData.Signatures, signs...)
}

// PrepareNewVersion prepares the next version of the CoreDocument
// if initSalts is true, salts will be generated for new version.
func (cd *CoreDocument) PrepareNewVersion(documentPrefix []byte, collaborators CollaboratorsAccess) (*CoreDocument, error) {
	// get all the old collaborators
	oldCs, err := cd.GetCollaborators()
	if err != nil {
		return nil, err
	}

	rcs := filterCollaborators(collaborators.ReadCollaborators, oldCs.ReadCollaborators...)
	wcs := filterCollaborators(collaborators.ReadWriteCollaborators, oldCs.ReadWriteCollaborators...)
	rcs = append(rcs, wcs...)

	cdp := coredocumentpb.CoreDocument{
		DocumentIdentifier: cd.Document.DocumentIdentifier,
		Roles:              cd.Document.Roles,
		ReadRules:          cd.Document.ReadRules,
		TransitionRules:    cd.Document.TransitionRules,
		Nfts:               cd.Document.Nfts,
		AccessTokens:       cd.Document.AccessTokens,
		SignatureData:      new(coredocumentpb.SignatureData),
	}

	err = populateVersions(&cdp, &cd.Document)
	if err != nil {
		return nil, err
	}

	ncd := &CoreDocument{Document: cdp}
	ncd.addCollaboratorsToReadSignRules(rcs)
	ncd.addCollaboratorsToTransitionRules(documentPrefix, wcs)
	ncd.Modified = true
	return ncd, nil

}

// newRole returns a new role with random role key
func newRole() *coredocumentpb.Role {
	return &coredocumentpb.Role{RoleKey: utils.RandomSlice(idSize)}
}

// newRoleWithCollaborators creates a new Role and adds the given collaborators to this Role.
// The Role is then returned.
// The operation returns a nil Role if no collaborators are provided.
func newRoleWithCollaborators(collaborators ...identity.DID) *coredocumentpb.Role {
	if len(collaborators) == 0 {
		return nil
	}

	// create a role for given collaborators
	role := newRole()
	for _, c := range collaborators {
		c := c
		role.Collaborators = append(role.Collaborators, c[:])
	}
	return role
}

// TreeProof is a helper structure to pass to create proofs
type TreeProof struct {
	tree       *proofs.DocumentTree
	treeHashes [][]byte
}

// newTreeProof returns a TreeProof instance pointer
func newTreeProof(t *proofs.DocumentTree, th [][]byte) *TreeProof {
	return &TreeProof{tree: t, treeHashes: th}
}

// CreateProofs takes document data tree and list to fields and generates proofs.
// we will try generating proofs from the dataTree. If failed, we will generate proofs from CoreDocument.
// errors out when the proof generation is failed on core document tree.
func (cd *CoreDocument) CreateProofs(docType string, dataTree *proofs.DocumentTree, fields []string) (prfs []*proofspb.Proof, err error) {
	treeProofs := make(map[string]*TreeProof, 3)
	drTree, err := cd.DocumentRootTree(docType, dataTree.RootHash())
	if err != nil {
		return nil, err
	}

	signatureTree, err := cd.getSignatureDataTree()
	if err != nil {
		return nil, errors.NewTypedError(ErrCDTree, errors.New("failed to generate signatures tree: %v", err))
	}

	cdTree, err := cd.coredocTree(docType)
	if err != nil {
		return nil, errors.NewTypedError(ErrCDTree, errors.New("failed to generate core document tree: %v", err))
	}

	signingRoot, err := cd.CalculateSigningRoot(docType, dataTree.RootHash())
	if err != nil {
		return nil, errors.NewTypedError(ErrCDTree, errors.New("failed to generate signing root: %v", err))
	}

	dataRoot := dataTree.RootHash()
	cdRoot := cdTree.RootHash()

	dataPrefix, err := getDataTreePrefix(dataTree)
	if err != nil {
		return nil, err
	}

	treeProofs[DRTreePrefix] = newTreeProof(drTree, nil)
	// (dataProof => dataRoot) + cdRoot+ signatureRoot = documentRoot
	treeProofs[dataPrefix] = newTreeProof(dataTree, append([][]byte{cdRoot}, signatureTree.RootHash()))
	// (signatureProof => signatureRoot) + signingRoot = documentRoot
	treeProofs[SignaturesTreePrefix] = newTreeProof(signatureTree, [][]byte{signingRoot})
	// (cdProof => cdRoot) + dataRoot + signatureRoot = documentRoot
	treeProofs[CDTreePrefix] = newTreeProof(cdTree, append([][]byte{dataRoot}, signatureTree.RootHash()))

	return generateProofs(fields, treeProofs)
}

// TODO remove as soon as we have a public method that retrieves the parent prefix
func getDataTreePrefix(dataTree *proofs.DocumentTree) (string, error) {
	props := dataTree.PropertyOrder()
	if len(props) == 0 {
		return "", errors.NewTypedError(ErrCDTree, errors.New("no properties found in data tree"))
	}
	fidx := strings.Split(props[0].ReadableName(), ".")
	if len(fidx) == 1 {
		return "", errors.NewTypedError(ErrCDTree, errors.New("no prefix found in data tree property"))
	}
	return fidx[0], nil
}

// generateProofs creates proofs from fields and trees and hashes provided
func generateProofs(fields []string, treeProofs map[string]*TreeProof) (prfs []*proofspb.Proof, err error) {
	for _, f := range fields {
		fidx := strings.Split(f, ".")
		t, ok := treeProofs[fidx[0]]
		if !ok {
			return nil, errors.New("failed to find prefix tree in supported list")
		}
		tree := t.tree
		proof, err := tree.CreateProof(f)
		if err != nil {
			return nil, err
		}
		thashes := treeProofs[fidx[0]].treeHashes
		proof.SortedHashes = append(proof.SortedHashes, thashes...)
		prfs = append(prfs, &proof)
	}
	return prfs, nil
}

// CalculateSignaturesRoot returns the signatures root of the document.
func (cd *CoreDocument) CalculateSignaturesRoot() ([]byte, error) {
	tree, err := cd.getSignatureDataTree()
	if err != nil {
		return nil, errors.NewTypedError(ErrCDTree, errors.New("failed to get signature tree: %v", err))
	}

	return tree.RootHash(), nil
}

// getSignatureDataTree returns the merkle tree for the Signature Data root.
func (cd *CoreDocument) getSignatureDataTree() (tree *proofs.DocumentTree, err error) {
	tree = cd.DefaultTreeWithPrefix(SignaturesTreePrefix, CompactProperties(SignaturesTreePrefix))
	err = tree.AddLeavesFromDocument(cd.Document.SignatureData)
	if err != nil {
		return nil, err
	}

	err = tree.Generate()
	if err != nil {
		return nil, err
	}

	return tree, nil
}

// DocumentRootTree returns the merkle tree for the document root.
func (cd *CoreDocument) DocumentRootTree(docType string, dataRoot []byte) (tree *proofs.DocumentTree, err error) {
	signingRoot, err := cd.CalculateSigningRoot(docType, dataRoot)
	if err != nil {
		return nil, err
	}

	tree = cd.DefaultTreeWithPrefix(DRTreePrefix, CompactProperties(DRTreePrefix))

	// The first leave added is the signing_root
	err = tree.AddLeaf(proofs.LeafNode{
		Hash:     signingRoot,
		Hashed:   true,
		Property: NewLeafProperty(fmt.Sprintf("%s.%s", DRTreePrefix, SigningRootField), append(CompactProperties(DRTreePrefix), CompactProperties(SigningRootField)...))})
	if err != nil {
		return nil, err
	}
	// Second leaf from the signature data tree
	signatureTree, err := cd.getSignatureDataTree()
	if err != nil {
		return nil, err
	}
	err = tree.AddLeaf(proofs.LeafNode{
		Hash:     signatureTree.RootHash(),
		Hashed:   true,
		Property: NewLeafProperty(fmt.Sprintf("%s.%s", DRTreePrefix, SignaturesRootField), append(CompactProperties(DRTreePrefix), CompactProperties(SignaturesRootField)...))})
	if err != nil {
		return nil, err
	}

	err = tree.Generate()
	if err != nil {
		return nil, err
	}

	cd.Modified = false
	return tree, nil
}

// signingRootTree returns the merkle tree for the signing root.
func (cd *CoreDocument) signingRootTree(docType string, dataRoot []byte) (tree *proofs.DocumentTree, err error) {
	if len(dataRoot) != idSize {
		return nil, errors.NewTypedError(ErrCDTree, errors.New("data root is invalid"))
	}

	cdTree, err := cd.coredocTree(docType)
	if err != nil {
		return nil, err
	}

	// create the signing tree with data root and coredoc root as siblings
	tree = cd.DefaultTreeWithPrefix(SigningTreePrefix, CompactProperties(SigningTreePrefix))
	err = tree.AddLeaves([]proofs.LeafNode{
		{
			Property: NewLeafProperty(fmt.Sprintf("%s.%s", SigningTreePrefix, DataRootField), append(CompactProperties(SigningTreePrefix), CompactProperties(DataRootField)...)),
			Hash:     dataRoot,
			Hashed:   true,
		},
		{
			Property: NewLeafProperty(fmt.Sprintf("%s.%s", SigningTreePrefix, CDRootField), append(CompactProperties(SigningTreePrefix), CompactProperties(CDRootField)...)),
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

// coredocTree returns the merkle tree of the CoreDocument.
func (cd *CoreDocument) coredocTree(docType string) (tree *proofs.DocumentTree, err error) {
	tree = cd.DefaultTreeWithPrefix(CDTreePrefix, CompactProperties(CDTreePrefix))
	err = tree.AddLeavesFromDocument(&cd.Document)
	if err != nil {
		return nil, err
	}

	dtProp := NewLeafProperty(fmt.Sprintf("%s.%s", CDTreePrefix, DocumentTypeField), append(CompactProperties(CDTreePrefix), CompactProperties(DocumentTypeField)...))
	// Adding document type as it is an excluded field in the tree
	documentTypeNode := proofs.LeafNode{
		Property: dtProp,
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
	cs, err := cd.GetCollaborators(filterIDs...)
	if err != nil {
		return nil, err
	}

	return cs.ReadWriteCollaborators, nil
}

// GetCollaborators returns the collaborators excluding the filteredIDs
func (cd *CoreDocument) GetCollaborators(filterIDs ...identity.DID) (CollaboratorsAccess, error) {
	rcs, err := cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ_SIGN, coredocumentpb.Action_ACTION_READ)
	if err != nil {
		return CollaboratorsAccess{}, err
	}

	wcs, err := cd.getWriteCollaborators(coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
	if err != nil {
		return CollaboratorsAccess{}, err
	}

	wc := filterCollaborators(wcs, filterIDs...)
	rc := filterCollaborators(rcs, append(wc, filterIDs...)...)
	return CollaboratorsAccess{
		ReadCollaborators:      rc,
		ReadWriteCollaborators: wc,
	}, nil
}

// getCollaborators returns all the collaborators which have the type of read or read/sign access passed in.
func (cd *CoreDocument) getReadCollaborators(actions ...coredocumentpb.Action) (ids []identity.DID, err error) {
	findReadRole(cd.Document, func(_, _ int, role *coredocumentpb.Role) bool {
		if len(role.Collaborators) < 1 {
			return false
		}

		for _, c := range role.Collaborators {
			var did identity.DID
			did, err = identity.NewDIDFromBytes(c)
			if err != nil {
				return false
			}
			ids = append(ids, did)
		}

		return false
	}, actions...)

	return identity.RemoveDuplicateDIDs(ids), err
}

// getWriteCollaborators returns all the collaborators which have access to the transition actions passed in.
func (cd *CoreDocument) getWriteCollaborators(actions ...coredocumentpb.TransitionAction) (ids []identity.DID, err error) {
	findTransitionRole(cd.Document, func(_, _ int, role *coredocumentpb.Role) bool {
		if len(role.Collaborators) < 1 {
			return false
		}

		for _, c := range role.Collaborators {
			var did identity.DID
			did, err = identity.NewDIDFromBytes(c)
			if err != nil {
				return false
			}
			ids = append(ids, did)
		}

		return false
	}, actions...)

	return identity.RemoveDuplicateDIDs(ids), err
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

// CalculateDocumentRoot calculates the document root of the CoreDocument.
func (cd *CoreDocument) CalculateDocumentRoot(docType string, dataRoot []byte) ([]byte, error) {
	tree, err := cd.DocumentRootTree(docType, dataRoot)
	if err != nil {
		return nil, err
	}

	return tree.RootHash(), nil
}

// CalculateSigningRoot calculates the signing root of the core document.
func (cd *CoreDocument) CalculateSigningRoot(docType string, dataRoot []byte) ([]byte, error) {
	tree, err := cd.signingRootTree(docType, dataRoot)
	if err != nil {
		return nil, err
	}

	return tree.RootHash(), nil
}

// PackCoreDocument prepares the document into a core document.
func (cd *CoreDocument) PackCoreDocument(data *any.Any) coredocumentpb.CoreDocument {
	// lets copy the value so that mutations on the returned doc wont be reflected on document we are holding
	cdp := cd.Document
	cdp.EmbeddedData = data
	return cdp
}

// Signatures returns the copy of the signatures on the document.
func (cd *CoreDocument) Signatures() (signatures []coredocumentpb.Signature) {
	for _, s := range cd.Document.SignatureData.Signatures {
		signatures = append(signatures, *s)
	}
	return signatures
}

// AddUpdateLog adds a log to the model to persist an update related meta data such as author
func (cd *CoreDocument) AddUpdateLog(account identity.DID) (err error) {
	cd.Document.Author = account[:]
	cd.Document.Timestamp, err = utils.ToTimestamp(time.Now().UTC())
	if err != nil {
		return err
	}
	cd.Modified = true
	return nil
}

// Author is the author of the document version represented by the model
func (cd *CoreDocument) Author() (identity.DID, error) {
	did, err := identity.NewDIDFromBytes(cd.Document.Author)
	if err != nil {
		return identity.DID{}, err
	}
	return did, nil
}

// Timestamp is the time of update in UTC of the document version represented by the model
func (cd *CoreDocument) Timestamp() (time.Time, error) {
	return utils.FromTimestamp(cd.Document.Timestamp)
}

// AddAttribute adds a custom attribute to the model with the given value. If an attribute with the given name already exists its updated.
func (cd *CoreDocument) AddAttribute(name string, attributeType string, value string) error {
	// TODO value conversion and type validation of it
	tp, err := allowedAttributeTypes(attributeType)
	if err != nil {
		return err
	}
	// for now its all string
	nAttr, err := NewAttribute(name, tp, value)
	if err != nil {
		return err
	}
	cd.Attributes[name] = nAttr
	return nil
}

// GetAttribute gets the attribute with the given name from the model together with its type, it returns a non-nil error if the attribute doesn't exist or can't be retrieved.
func (cd *CoreDocument) GetAttribute(name string) (hashedKey []byte, attrType string, value interface{}, valueStr string, err error) {
	if attr, ok := cd.Attributes[name]; ok {
		an, err := allowedAttributeTypeNames(attr.Type)
		if err != nil {
			return hashedKey, attrType, value, valueStr, err
		}
		// TODO convert value to its string repr
		return attr.HashedKey, an, attr.Value, "", nil
	}
	return hashedKey, attrType, value, valueStr, errors.NewTypedError(ErrCDAttribute, errors.New("attribute does not exist"))
}

func populateVersions(cd *coredocumentpb.CoreDocument, prevCD *coredocumentpb.CoreDocument) (err error) {
	if prevCD != nil {
		cd.PreviousVersion = prevCD.CurrentVersion
		cd.CurrentVersion = prevCD.NextVersion
		cd.CurrentPreimage = prevCD.NextPreimage
	} else {
		cd.CurrentPreimage, cd.CurrentVersion, err = crypto.GenerateHashPair(idSize)
		cd.DocumentIdentifier = cd.CurrentVersion
		if err != nil {
			return err
		}
	}
	cd.NextPreimage, cd.NextVersion, err = crypto.GenerateHashPair(idSize)
	if err != nil {
		return err
	}
	return nil
}
