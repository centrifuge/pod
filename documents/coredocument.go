package documents

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
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

	// Pending status represents document is in pending state
	Pending status = "pending"

	// Committing status represents document is being committed.
	Committing status = "committing"

	// Committed status represents document is committed/anchored.
	Committed status = "committed"
)

// status represents the document status.
type status string

// CompactProperties returns the compact property for a given prefix
func CompactProperties(key string) []byte {
	switch key {
	case CDRootField:
		return []byte{0, 0, 0, 7}
	case DataRootField:
		return []byte{0, 0, 0, 5}
	case DocumentTypeField:
		return []byte{0, 0, 0, 100}
	case SignaturesRootField:
		return []byte{0, 0, 0, 6}
	case SigningRootField:
		return []byte{0, 0, 0, 10}
	case CDTreePrefix:
		return []byte{1, 0, 0, 0}
	case SigningTreePrefix:
		return []byte{2, 0, 0, 0}
	case SignaturesTreePrefix:
		return []byte{3, 0, 0, 0}
	case DRTreePrefix:
		return []byte{4, 0, 0, 0}
	default:
		return []byte{}
	}
}

// CoreDocument is a wrapper for CoreDocument Protobuf.
type CoreDocument struct {
	// Modified indicates that the CoreDocument has been modified and salts needs to be generated for new fields in coredoc precise-proof tree.
	Modified bool

	// Attributes are the custom attributes added to the document
	Attributes map[AttrKey]Attribute

	// Status represents document status.
	Status status

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

	return &CoreDocument{
		Document:   cd,
		Modified:   true,
		Attributes: make(map[AttrKey]Attribute),
		Status:     Pending,
	}, nil
}

// NewCoreDocumentFromProtobuf returns CoreDocument from the CoreDocument Protobuf.
func NewCoreDocumentFromProtobuf(cd coredocumentpb.CoreDocument) (coreDoc *CoreDocument, err error) {
	cd.EmbeddedData = nil
	coreDoc = &CoreDocument{Document: cd}
	coreDoc.Attributes, err = fromProtocolAttributes(cd.Attributes)
	return coreDoc, err
}

// AccessTokenParams holds details of Grantee and DocumentIdentifier.
type AccessTokenParams struct {
	Grantee, DocumentIdentifier string
}

// NewCoreDocument generates new core document with a document type specified by the prefix: po or invoice.
// It then adds collaborators, adds read rules and fills salts.
func NewCoreDocument(documentPrefix []byte, collaborators CollaboratorsAccess, attributes map[AttrKey]Attribute) (*CoreDocument, error) {
	cd, err := newCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(ErrCDCreate, errors.New("failed to create coredoc: %v", err))
	}

	collaborators.ReadCollaborators = identity.RemoveDuplicateDIDs(collaborators.ReadCollaborators)
	collaborators.ReadWriteCollaborators = identity.RemoveDuplicateDIDs(collaborators.ReadWriteCollaborators)
	cd.initReadRules(append(collaborators.ReadCollaborators, collaborators.ReadWriteCollaborators...))
	cd.initTransitionRules(documentPrefix, collaborators.ReadWriteCollaborators)
	cd.Attributes = attributes
	cd.Document.Attributes, err = toProtocolAttributes(attributes)
	return cd, err
}

// NewCoreDocumentWithAccessToken generates a new core document with a document type specified by the prefix.
// It also adds the targetID as a read collaborator, and adds an access token on this document for the document specified in the documentID parameter
func NewCoreDocumentWithAccessToken(ctx context.Context, documentPrefix []byte, params AccessTokenParams) (*CoreDocument, error) {
	did, err := identity.StringsToDIDs(params.Grantee)
	if err != nil {
		return nil, err
	}

	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, ErrDocumentConfigAccountID
	}

	collaborators := CollaboratorsAccess{
		ReadCollaborators:      []identity.DID{*did[0]},
		ReadWriteCollaborators: []identity.DID{selfDID},
	}
	cd, err := NewCoreDocument(documentPrefix, collaborators, nil)
	if err != nil {
		return nil, err
	}
	at, err := assembleAccessToken(ctx, params, cd.CurrentVersion())
	if err != nil {
		return nil, errors.New("failed to construct access token: %v", err)
	}
	cd.Document.AccessTokens = append(cd.Document.AccessTokens, at)
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

// GetStatus returns document status
// nolint
func (cd *CoreDocument) GetStatus() status {
	return cd.Status
}

// SetStatus set the status of the document.
// if the document is already committed, returns an error if set status is called.
func (cd *CoreDocument) SetStatus(st status) error {
	if cd.Status == Committed && st != Committed {
		return ErrCDStatus
	}

	cd.Status = st
	return nil
}

// AppendSignatures appends signatures to core document.
func (cd *CoreDocument) AppendSignatures(signs ...*coredocumentpb.Signature) {
	if cd.Document.SignatureData == nil {
		cd.Document.SignatureData = new(coredocumentpb.SignatureData)
	}
	cd.Document.SignatureData.Signatures = append(cd.Document.SignatureData.Signatures, signs...)
	cd.Modified = true
}

// PrepareNewData overrides only core document data without provisioning new versions since for document updates
func (cd *CoreDocument) PrepareNewData(documentPrefix []byte, collaborators CollaboratorsAccess, attrs map[AttrKey]Attribute) (*CoreDocument, error) {
	cdp := coredocumentpb.CoreDocument{
		DocumentIdentifier: cd.Document.DocumentIdentifier,
		CurrentVersion:     cd.Document.CurrentVersion,
		PreviousVersion:    cd.Document.PreviousVersion,
		NextVersion:        cd.Document.NextVersion,
		CurrentPreimage:    cd.Document.CurrentPreimage,
		NextPreimage:       cd.Document.NextPreimage,
		Nfts:               cd.Document.Nfts,
		AccessTokens:       cd.Document.AccessTokens,
		SignatureData:      new(coredocumentpb.SignatureData),
	}
	// TODO convert it back to override when we have implemented add/delete for collaborators in API
	// for now it always overrides
	rcs := collaborators.ReadCollaborators
	wcs := collaborators.ReadWriteCollaborators
	rcs = append(rcs, wcs...)

	ncd := &CoreDocument{Document: cdp}
	ncd.addCollaboratorsToReadSignRules(rcs)
	ncd.addCollaboratorsToTransitionRules(documentPrefix, wcs)
	// TODO convert it back to override when we have implemented add/delete for attributes in API
	// for now it always overrides
	p2pAttrs, attrs, err := updateAttributes(nil, attrs)
	if err != nil {
		return nil, errors.NewTypedError(ErrCDNewVersion, err)
	}

	ncd.Document.Attributes = p2pAttrs
	ncd.Attributes = attrs
	ncd.Modified = true
	return ncd, nil
}

// PrepareNewVersion prepares the next version of the CoreDocument
// if initSalts is true, salts will be generated for new version.
func (cd *CoreDocument) PrepareNewVersion(documentPrefix []byte, collaborators CollaboratorsAccess, attrs map[AttrKey]Attribute) (*CoreDocument, error) {
	// get all the old collaborators
	oldCs, err := cd.GetCollaborators()
	if err != nil {
		return nil, errors.NewTypedError(ErrCDNewVersion, err)
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
		return nil, errors.NewTypedError(ErrCDNewVersion, err)
	}

	ncd := &CoreDocument{Document: cdp}
	ncd.addCollaboratorsToReadSignRules(rcs)
	ncd.addCollaboratorsToTransitionRules(documentPrefix, wcs)
	p2pAttrs, attrs, err := updateAttributes(cd.Document.Attributes, attrs)
	if err != nil {
		return nil, errors.NewTypedError(ErrCDNewVersion, err)
	}

	ncd.Document.Attributes = p2pAttrs
	ncd.Attributes = attrs
	ncd.Modified = true
	return ncd, nil
}

// updateAttributes updates the p2p attributes with new ones and returns both formats
func updateAttributes(oldAttrs []*coredocumentpb.Attribute, newAttrs map[AttrKey]Attribute) ([]*coredocumentpb.Attribute, map[AttrKey]Attribute, error) {
	oldAttrsMap, err := fromProtocolAttributes(oldAttrs)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range newAttrs {
		oldAttrsMap[k] = v
	}

	uattrs, err := toProtocolAttributes(oldAttrsMap)
	return uattrs, oldAttrsMap, err
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
// It only generates proofs up to the signingRoot level
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
	// (dataProof => dataRoot) + cdRoot = signingRoot
	treeProofs[dataPrefix] = newTreeProof(dataTree, append([][]byte{cdRoot}))
	// (signatureProof => signatureRoot) + signingRoot = documentRoot
	treeProofs[SignaturesTreePrefix] = newTreeProof(signatureTree, [][]byte{signingRoot})
	// (cdProof => cdRoot) + dataRoot = signingRoot
	treeProofs[CDTreePrefix] = newTreeProof(cdTree, append([][]byte{dataRoot}))

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
	sign, err := cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ_SIGN)
	if err != nil {
		return nil, err
	}

	return filterCollaborators(sign, filterIDs...), nil
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

// AddAttributes adds a custom attribute to the model with the given value. If an attribute with the given name already exists, it's updated.
// Note: The prepareNewVersion flags defines if the returned model should be a new version of the document.
func (cd *CoreDocument) AddAttributes(ca CollaboratorsAccess, prepareNewVersion bool, documentPrefix []byte, attrs ...Attribute) (*CoreDocument, error) {
	if len(attrs) < 1 {
		return nil, errors.NewTypedError(ErrCDAttribute, errors.New("require at least one attribute"))
	}

	var ncd *CoreDocument
	var err error
	if prepareNewVersion {
		ncd, err = cd.PrepareNewVersion(documentPrefix, ca, nil)
		if err != nil {
			return nil, errors.NewTypedError(ErrCDAttribute, errors.New("failed to prepare new version: %v", err))
		}
	} else {
		ncd = cd
	}

	if ncd.Attributes == nil {
		ncd.Attributes = make(map[AttrKey]Attribute)
	}

	for _, attr := range attrs {
		if !isAttrTypeAllowed(attr.Value.Type) {
			return nil, ErrNotValidAttrType
		}

		ncd.Attributes[attr.Key] = attr
	}
	ncd.Document.Attributes, err = toProtocolAttributes(ncd.Attributes)
	return ncd, err
}

// AttributeExists checks if an attribute associated with the key exists.
func (cd *CoreDocument) AttributeExists(key AttrKey) bool {
	_, ok := cd.Attributes[key]
	return ok
}

// GetAttribute gets the attribute with the given name from the model together with its type, it returns a non-nil error if the attribute doesn't exist or can't be retrieved.
func (cd *CoreDocument) GetAttribute(key AttrKey) (attr Attribute, err error) {
	attr, ok := cd.Attributes[key]
	if !ok {
		return attr, errors.NewTypedError(ErrCDAttribute, errors.New("attribute does not exist"))
	}

	return attr, nil
}

// GetAttributes returns all the attributes present in the coredocument.
func (cd *CoreDocument) GetAttributes() (attrs []Attribute) {
	for _, attr := range cd.Attributes {
		attrs = append(attrs, attr)
	}

	return attrs
}

// DeleteAttribute deletes a custom attribute from the model.
// If the attribute is missing, delete returns an error
func (cd *CoreDocument) DeleteAttribute(key AttrKey, prepareNewVersion bool, documentPrefix []byte) (*CoreDocument, error) {
	if _, ok := cd.Attributes[key]; !ok {
		return nil, errors.NewTypedError(ErrCDAttribute, errors.New("missing attribute: %v", key))
	}

	var ncd *CoreDocument
	var err error
	if prepareNewVersion {
		ncd, err = cd.PrepareNewVersion(documentPrefix, CollaboratorsAccess{}, nil)
		if err != nil {
			return nil, errors.NewTypedError(ErrCDAttribute, errors.New("failed to prepare new version: %v", err))
		}
	} else {
		ncd = cd
	}

	delete(ncd.Attributes, key)
	ncd.Document.Attributes, err = toProtocolAttributes(ncd.Attributes)
	return ncd, err
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
	return err
}

// IsDIDCollaborator returns true if the did is a collaborator of the document
func (cd *CoreDocument) IsDIDCollaborator(did identity.DID) (bool, error) {
	collAccess, err := cd.GetCollaborators()
	if err != nil {
		return false, err
	}

	for _, d := range collAccess.ReadWriteCollaborators {
		if d == did {
			return true, nil
		}
	}
	for _, d := range collAccess.ReadCollaborators {
		if d == did {
			return true, nil
		}
	}
	return false, nil
}

// GetAccessTokens returns the access tokens of a core document
func (cd *CoreDocument) GetAccessTokens() ([]*coredocumentpb.AccessToken, error) {
	return cd.Document.AccessTokens, nil
}

// SetUsedAnchorRepoAddress sets used anchor repo address.
func (cd *CoreDocument) SetUsedAnchorRepoAddress(addr common.Address) {
	cd.Document.AnchorRepositoryUsed = addr.Bytes()
}

// AnchorRepoAddress returns the used anchor repo address to which the document is/will be anchored to.
func (cd *CoreDocument) AnchorRepoAddress() common.Address {
	return common.BytesToAddress(cd.Document.AnchorRepositoryUsed)
}

// MarshalJSON marshals the model and returns the json data.
func (cd *CoreDocument) MarshalJSON(m Model) ([]byte, error) {
	pattrs := cd.Document.Attributes
	cd.Document.Attributes = nil
	d, err := json.Marshal(m)
	cd.Document.Attributes = pattrs
	return d, err
}

// UnmarshalJSON unmarshals the data into model and set the attributes back to the document.
// Note: Coredocument should not be nil and should be initialised to the Model before passing to this function.
func (cd *CoreDocument) UnmarshalJSON(data []byte, m Model) error {
	err := json.Unmarshal(data, m)
	if err != nil {
		return err
	}

	cd.Document.Attributes, err = toProtocolAttributes(cd.Attributes)
	return err
}
