package documents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/precise-proofs/proofs"
	proofspb "github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"

	"golang.org/x/crypto/sha3"
)

const (
	// idSize represents the size of identifiers, roots etc..
	idSize = 32

	// nftByteCount is the length of combined bytes of registry and tokenID
	nftByteCount = 52

	// Tree fields and prefixes

	// CDRootField represents the coredocument root property of a tree
	CDRootField = "cd_root"

	// DataRootField represents the data root property of a tree
	DataRootField = "data_root"

	// DocumentTypeField represents the doc type property of a tree
	DocumentTypeField = "document_type"

	// BasicDataRootField represents the basic document data tree root
	BasicDataRootField = "bd_root"

	// ZKDataRootField represents the zk document data tree root
	ZKDataRootField = "zkd_root"

	// SignaturesRootField represents the signatures property of a tree
	SignaturesRootField = "signatures_root"

	// SigningRootField represents the signature root property of a tree
	SigningRootField = "signing_root"

	// DRTreePrefix is the human readable prefix for document root tree props
	DRTreePrefix = "dr_tree"

	// CDTreePrefix is the human readable prefix for core doc tree props
	CDTreePrefix = "cd_tree"

	// BasicDataRootPrefix represents the basic document data tree
	BasicDataRootPrefix = "bdr_tree"

	// ZKDataRootPrefix represents the zk document data tree
	ZKDataRootPrefix = "zkdr_tree"

	// SigningTreePrefix is the human readable prefix for signing tree props
	SigningTreePrefix = "signing_tree"

	// SignaturesTreePrefix is the human readable prefix for signature props
	SignaturesTreePrefix = "signatures_tree"

	// Document states

	// Pending status represents document is in pending state
	Pending Status = "pending"

	// Committing status represents document is being committed.
	Committing Status = "committing"

	// Committed status represents document is committed/anchored.
	Committed Status = "committed"
)

// Status represents the document status.
type Status string

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
	case BasicDataRootField:
		return []byte{0, 0, 0, 11}
	case ZKDataRootField:
		return []byte{0, 0, 0, 12}
	case CDTreePrefix:
		return []byte{1, 0, 0, 0}
	case SigningTreePrefix:
		return []byte{2, 0, 0, 0}
	case SignaturesTreePrefix:
		return []byte{3, 0, 0, 0}
	case DRTreePrefix:
		return []byte{4, 0, 0, 0}
	case BasicDataRootPrefix:
		return []byte{5, 0, 0, 0}
	case ZKDataRootPrefix:
		return []byte{6, 0, 0, 0}
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
	Status Status

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

// NewClonedDocument generates new blank core document with a document type specified by the prefix: generic.
// It then copies the Transition rules, Read rules, Roles, and Attributes of a supplied Template document.
func NewClonedDocument(d coredocumentpb.CoreDocument) (*CoreDocument, error) {
	cd, err := newCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(ErrCDCreate, errors.New("failed to create coredoc: %v", err))
	}
	cd.Document.TransitionRules = d.TransitionRules
	cd.Document.ReadRules = d.ReadRules
	cd.Document.Roles = d.Roles
	cd.Attributes, err = fromProtocolAttributes(d.Attributes)
	if err != nil {
		return nil, errors.NewTypedError(ErrCDCreate, errors.New("failed to create coredoc: %v", err))
	}

	cd.Document.Attributes = d.Attributes

	return cd, err
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
	// remove any dids that are present in both read and read write from read.
	collaborators.ReadCollaborators = filterCollaborators(collaborators.ReadCollaborators, collaborators.ReadWriteCollaborators...)
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
func (cd *CoreDocument) GetStatus() Status {
	return cd.Status
}

// SetStatus set the status of the document.
// if the document is already committed, returns an error if set status is called.
func (cd *CoreDocument) SetStatus(st Status) error {
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

// Patch overrides only core document data without provisioning new versions since for document updates
func (cd *CoreDocument) Patch(documentPrefix []byte, collaborators CollaboratorsAccess, attrs map[AttrKey]Attribute) (*CoreDocument, error) {
	if cd.Status == Committing || cd.Status == Committed {
		return nil, ErrDocumentNotInAllowedState
	}

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

	ncd := &CoreDocument{Document: cdp, Status: Pending}
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

	ncd := &CoreDocument{Document: cdp, Status: Pending}
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

// newRoleWithRandomKey returns a new role with random role key
func newRoleWithRandomKey() *coredocumentpb.Role {
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
	role := newRoleWithRandomKey()
	for _, c := range collaborators {
		c := c
		role.Collaborators = append(role.Collaborators, c[:])
	}
	return role
}

// CreateProofs takes document data leaves and list of fields and generates proofs.
func (cd *CoreDocument) CreateProofs(docType string, dataLeaves []proofs.LeafNode, fields []string) (*DocumentProof, error) {
	return cd.createProofs(false, docType, dataLeaves, fields)
}

// CreateProofsFromZKTree takes document data leaves and list of fields and generates proofs from ZK-ready Tree.
func (cd *CoreDocument) CreateProofsFromZKTree(docType string, dataLeaves []proofs.LeafNode, fields []string) (*DocumentProof, error) {
	return cd.createProofs(true, docType, dataLeaves, fields)
}

// createProofs takes document data tree and list to fields and generates proofs.
// it will generate proofs from the dataTree and cdTree.
// It only generates proofs up to the root of the tree/s that correspond to
func (cd *CoreDocument) createProofs(fromZKTree bool, docType string, dataLeaves []proofs.LeafNode, fields []string) (*DocumentProof, error) {
	treeProofs := make(map[string]*proofs.DocumentTree, 4)
	drTree, err := cd.DocumentRootTree(docType, dataLeaves)
	if err != nil {
		return nil, err
	}

	signatureTree, err := cd.GetSignaturesDataTree()
	if err != nil {
		return nil, errors.NewTypedError(ErrCDTree, errors.New("failed to generate signatures tree: %v", err))
	}

	dTrees, sdr, err := cd.SigningDataTrees(docType, dataLeaves)
	if err != nil {
		return nil, errors.NewTypedError(ErrCDTree, errors.New("failed to generate signing data trees: %v", err))
	}
	basicDataTree := dTrees[0]
	zkDataTree := dTrees[1]

	dataPrefix, err := getDataTreePrefix(dataLeaves)
	if err != nil {
		return nil, err
	}

	targetTree := basicDataTree
	if fromZKTree {
		targetTree = zkDataTree
	}

	treeProofs[dataPrefix] = targetTree
	treeProofs[CDTreePrefix] = treeProofs[dataPrefix]
	treeProofs[SignaturesTreePrefix] = signatureTree
	treeProofs[DRTreePrefix] = drTree

	rawProofs, err := generateProofs(fields, treeProofs)
	if err != nil {
		return nil, err
	}

	return &DocumentProof{
		FieldProofs:    rawProofs,
		LeftDataRooot:  basicDataTree.RootHash(),
		RightDataRoot:  zkDataTree.RootHash(),
		SigningRoot:    sdr,
		SignaturesRoot: signatureTree.RootHash(),
	}, nil
}

// CalculateTransitionRulesFingerprint generates a fingerprint for a Core Document
func (cd *CoreDocument) CalculateTransitionRulesFingerprint() ([]byte, error) {
	f := coredocumentpb.TransitionRulesFingerprint{
		Roles:           nil,
		TransitionRules: nil,
	}

	// We only want to copy the roles which are in the transition rules, because there are also instances when roles are also added when a new read rule is created
	// (ie: when a NFT is minted from a document, this means a new read rule is created and a new role created)
	// these roles should not be part of the transition rules fingerprint
	if len(cd.Document.Roles) == 0 || len(cd.Document.TransitionRules) == 0 {
		return []byte{}, nil
	}
	f.TransitionRules = cd.Document.TransitionRules
	var rks [][]byte
	for _, t := range f.TransitionRules {
		for _, r := range t.Roles {
			rks = append(rks, r)
		}
	}
	for _, rk := range rks {
		for _, r := range cd.Document.Roles {
			if bytes.Equal(rk, r.RoleKey) {
				f.Roles = append(f.Roles, r)
			}
		}
	}
	p, err := generateTransitionRulesFingerprintHash(f)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// generateTransitionRulesFingerprintHash takes an assembled fingerprint message and generates the root hash from this message.
// the return value can be used to verify if transition rules or roles have changed across documents
func generateTransitionRulesFingerprintHash(fingerprint coredocumentpb.TransitionRulesFingerprint) ([]byte, error) {
	fm, err := proto.Marshal(&fingerprint)
	if err != nil {
		return nil, err
	}

	s := sha3.NewLegacyKeccak256()
	_, err = s.Write(fm)
	if err != nil {
		return nil, err
	}

	return s.Sum(nil), nil
}

// TODO remove as soon as we have a public method that retrieves the parent prefix
func getDataTreePrefix(dataLeaves []proofs.LeafNode) (string, error) {
	if len(dataLeaves) == 0 {
		return "", errors.NewTypedError(ErrCDTree, errors.New("no properties found in data leaves"))
	}
	fidx := strings.Split(dataLeaves[0].Property.ReadableName(), ".")
	if len(fidx) == 1 {
		return "", errors.NewTypedError(ErrCDTree, errors.New("no prefix found in data leaf property"))
	}
	return fidx[0], nil
}

// generateProofs creates proofs from fields and trees and hashes provided
func generateProofs(fields []string, treeProofs map[string]*proofs.DocumentTree) (prfs []*proofspb.Proof, err error) {
	for _, f := range fields {
		fidx := strings.Split(f, ".")
		tree, ok := treeProofs[fidx[0]]
		if !ok {
			return nil, errors.New("failed to find prefix tree in supported list")
		}
		proof, err := tree.CreateProof(f)
		if err != nil {
			return nil, err
		}
		prfs = append(prfs, &proof)
	}
	return prfs, nil
}

// CalculateSignaturesRoot returns the signatures root of the document.
func (cd *CoreDocument) CalculateSignaturesRoot() ([]byte, error) {
	tree, err := cd.GetSignaturesDataTree()
	if err != nil {
		return nil, errors.NewTypedError(ErrCDTree, errors.New("failed to get signature tree: %v", err))
	}

	return tree.RootHash(), nil
}

// GetSignaturesDataTree returns the merkle tree for the Signature Data root.
func (cd *CoreDocument) GetSignaturesDataTree() (tree *proofs.DocumentTree, err error) {
	tree, err = cd.DefaultTreeWithPrefix(SignaturesTreePrefix, CompactProperties(SignaturesTreePrefix))
	if err != nil {
		return nil, err
	}
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
func (cd *CoreDocument) DocumentRootTree(docType string, dataLeaves []proofs.LeafNode) (tree *proofs.DocumentTree, err error) {
	signingRoot, err := cd.CalculateSigningRoot(docType, dataLeaves)
	if err != nil {
		return nil, err
	}

	tree, err = cd.DefaultOrderedTreeWithPrefix(DRTreePrefix, CompactProperties(DRTreePrefix))
	if err != nil {
		return nil, err
	}

	// The first leave added is the signing_root
	err = tree.AddLeaf(proofs.LeafNode{
		Hash:     signingRoot,
		Hashed:   true,
		Property: NewLeafProperty(fmt.Sprintf("%s.%s", DRTreePrefix, SigningRootField), append(CompactProperties(DRTreePrefix), CompactProperties(SigningRootField)...))})
	if err != nil {
		return nil, err
	}
	// Second leaf from the signature data tree
	signatureTree, err := cd.GetSignaturesDataTree()
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
	tree, err = cd.DefaultTreeWithPrefix(SigningTreePrefix, CompactProperties(SigningTreePrefix))
	if err != nil {
		return nil, err
	}

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

func (cd *CoreDocument) basicDataTree(docType string, dataLeaves []proofs.LeafNode, cdLeaves []proofs.LeafNode) (tree *proofs.DocumentTree, err error) {
	if dataLeaves == nil {
		return nil, errors.NewTypedError(ErrCDTree, errors.New("data tree is invalid"))
	}
	if cdLeaves == nil {
		return nil, errors.NewTypedError(ErrCDTree, errors.New("cd tree is invalid"))
	}
	// create the docDataTrees out of docData and coredoc trees
	tree, err = cd.DefaultTreeWithPrefix(BasicDataRootPrefix, CompactProperties(BasicDataRootPrefix))
	if err != nil {
		return nil, err
	}
	err = tree.AddLeaves(append(dataLeaves, cdLeaves...))
	if err != nil {
		return nil, err
	}
	err = tree.Generate()
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (cd *CoreDocument) zkDataTree(docType string, dataLeaves []proofs.LeafNode, cdLeaves []proofs.LeafNode) (tree *proofs.DocumentTree, err error) {
	if dataLeaves == nil {
		return nil, errors.NewTypedError(ErrCDTree, errors.New("data tree is invalid"))
	}
	if cdLeaves == nil {
		return nil, errors.NewTypedError(ErrCDTree, errors.New("cd tree is invalid"))
	}
	// create the docDataTrees out of docData and coredoc trees
	tree, err = cd.DefaultZTreeWithPrefix(ZKDataRootPrefix, CompactProperties(ZKDataRootPrefix))
	if err != nil {
		return nil, err
	}
	err = tree.AddLeaves(append(dataLeaves, cdLeaves...))
	if err != nil {
		return nil, err
	}
	err = tree.Generate()
	if err != nil {
		return nil, err
	}
	return tree, nil
}

// SigningDataTrees returns the merkle trees (basicData and zkData) + signingRoot Hash for the document data tree provided
func (cd *CoreDocument) SigningDataTrees(docType string, dataLeaves []proofs.LeafNode) (trees []*proofs.DocumentTree, rootHash []byte, err error) {
	if dataLeaves == nil {
		return nil, nil, errors.NewTypedError(ErrCDTree, errors.New("data tree is invalid"))
	}
	cdLeaves, err := cd.coredocLeaves(docType)
	if err != nil {
		return nil, nil, err
	}
	// create the docDataTrees out of docData and coredoc trees
	tree, err := cd.DefaultOrderedTreeWithPrefix(SigningTreePrefix, CompactProperties(SigningTreePrefix))
	if err != nil {
		return nil, nil, err
	}
	basicTree, err := cd.basicDataTree(docType, dataLeaves, cdLeaves)
	if err != nil {
		return nil, nil, err
	}
	zkTree, err := cd.zkDataTree(docType, dataLeaves, cdLeaves)
	if err != nil {
		return nil, nil, err
	}

	// The first leave added is the basic data tree root
	err = tree.AddLeaf(proofs.LeafNode{
		Hash:     basicTree.RootHash(),
		Hashed:   true,
		Property: NewLeafProperty(fmt.Sprintf("%s.%s", SigningTreePrefix, BasicDataRootField), append(CompactProperties(SigningTreePrefix), CompactProperties(BasicDataRootField)...))})
	if err != nil {
		return nil, nil, err
	}

	// The second leave added is the zkSnarks docData tree root
	err = tree.AddLeaf(proofs.LeafNode{
		Hash:     zkTree.RootHash(),
		Hashed:   true,
		Property: NewLeafProperty(fmt.Sprintf("%s.%s", SigningTreePrefix, ZKDataRootField), append(CompactProperties(SigningTreePrefix), CompactProperties(ZKDataRootField)...))})
	if err != nil {
		return nil, nil, err
	}

	err = tree.Generate()
	if err != nil {
		return nil, nil, err
	}

	return []*proofs.DocumentTree{basicTree, zkTree}, tree.RootHash(), nil
}

func (cd *CoreDocument) coredocRawTree(docType string) (*proofs.DocumentTree, error) {
	tree, err := cd.DefaultTreeWithPrefix(CDTreePrefix, CompactProperties(CDTreePrefix))
	if err != nil {
		return nil, err
	}
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

	err = documentTypeNode.HashNode(sha3.NewLegacyKeccak256(), true)
	if err != nil {
		return nil, err
	}

	err = tree.AddLeaf(documentTypeNode)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func (cd *CoreDocument) coredocLeaves(docType string) ([]proofs.LeafNode, error) {
	tree, err := cd.coredocRawTree(docType)
	if err != nil {
		return nil, err
	}
	return tree.GetLeaves(), nil
}

// coredocTree returns the merkle tree of the CoreDocument.
func (cd *CoreDocument) coredocTree(docType string) (tree *proofs.DocumentTree, err error) {
	tree, err = cd.coredocRawTree(docType)
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
// returns collaborators with Action_ACTION_READ_SIGN and TransitionAction_TRANSITION_ACTION_EDIT permissions.
func (cd *CoreDocument) GetSignerCollaborators(filterIDs ...identity.DID) ([]identity.DID, error) {
	sign, err := cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ_SIGN)
	if err != nil {
		return nil, err
	}

	wcs, err := cd.getWriteCollaborators(coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
	if err != nil {
		return nil, err
	}

	wc := filterCollaborators(wcs, filterIDs...)
	rc := filterCollaborators(sign, filterIDs...)
	return identity.RemoveDuplicateDIDs(append(wc, rc...)), nil
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
func (cd *CoreDocument) CalculateDocumentRoot(docType string, dataLeaves []proofs.LeafNode) ([]byte, error) {
	tree, err := cd.DocumentRootTree(docType, dataLeaves)
	if err != nil {
		return nil, err
	}

	return tree.RootHash(), nil
}

// CalculateSigningRoot calculates the signing root of the core document.
func (cd *CoreDocument) CalculateSigningRoot(docType string, dataLeaves []proofs.LeafNode) ([]byte, error) {
	_, treeHash, err := cd.SigningDataTrees(docType, dataLeaves)
	if err != nil {
		return nil, err
	}

	return treeHash, nil
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
func (cd *CoreDocument) MarshalJSON(m Document) ([]byte, error) {
	pattrs := cd.Document.Attributes
	cd.Document.Attributes = nil
	d, err := json.Marshal(m)
	cd.Document.Attributes = pattrs
	return d, err
}

// UnmarshalJSON unmarshals the data into model and set the attributes back to the document.
// Note: Coredocument should not be nil and should be initialised to the Document before passing to this function.
func (cd *CoreDocument) UnmarshalJSON(data []byte, m Document) error {
	err := json.Unmarshal(data, m)
	if err != nil {
		return err
	}

	cd.Document.Attributes, err = toProtocolAttributes(cd.Attributes)
	return err
}

// RemoveCollaborators removes DIDs from the Document.
// Errors out if the document is not in Pending state or collaborators are missing from the document.
func (cd *CoreDocument) RemoveCollaborators(dids []identity.DID) error {
	if cd.Status == Committing || cd.Status == Committed {
		return ErrDocumentNotInAllowedState
	}

	// remove each collaborator from the roles
	for _, did := range dids {
		for _, role := range cd.Document.Roles {
			i, f := isDIDInRole(role, did)
			if !f {
				continue
			}

			cd.Modified = true
			role.Collaborators = byteutils.CutFromSlice(role.Collaborators, i)
		}
	}

	return nil
}

// GetRole returns the role associated with key.
// key has to be 32 bytes long.
func (cd *CoreDocument) GetRole(key []byte) (*coredocumentpb.Role, error) {
	if len(key) != idSize {
		return nil, ErrInvalidRoleKey
	}

	for _, r := range cd.Document.Roles {
		if bytes.Equal(r.RoleKey, key) {
			return r, nil
		}
	}

	return nil, ErrRoleNotExist
}

// AddRole adds a new role to the document.
// key can either be plain text or 32 byte hex string, key cannot be empty
// If key is not 32 byte hex string, then the key is used as pre image for 32 byte key
func (cd *CoreDocument) AddRole(key string, collabs []identity.DID) (*coredocumentpb.Role, error) {
	rk, err := get32ByteKey(key)
	if err != nil {
		return nil, err
	}

	if _, err = cd.GetRole(rk); err == nil {
		return nil, ErrRoleExist
	}

	r := newRoleWithCollaborators(collabs...)
	if r == nil {
		return nil, ErrEmptyCollaborators
	}
	r.RoleKey = rk
	cd.Document.Roles = append(cd.Document.Roles, r)
	cd.Modified = true
	return r, nil
}

// UpdateRole updates existing role with provided collaborators
func (cd *CoreDocument) UpdateRole(rk []byte, collabs []identity.DID) (*coredocumentpb.Role, error) {
	r, err := cd.GetRole(rk)
	if err != nil {
		return nil, err
	}

	if len(collabs) < 1 {
		return nil, ErrEmptyCollaborators
	}

	r.Collaborators = nil
	for _, c := range collabs {
		c := c
		r.Collaborators = append(r.Collaborators, c[:])
	}
	cd.Modified = true
	return r, nil
}

func get32ByteKey(key string) ([]byte, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrEmptyRoleKey
	}

	kb, err := hexutil.Decode(key)
	if err == nil && len(kb) == idSize {
		return kb, nil
	}

	return crypto.Sha256Hash([]byte(key))
}
