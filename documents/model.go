package documents

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
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
		CDTreePrefix:      {0, 0, 0, 11},
		SigningTreePrefix: {0, 0, 0, 12},
	}
	return m[key]
}

// Model is an interface to abstract away model specificness like invoice or purchaseOrder
// The interface can cast into the type specified by the model if required
// It should only handle protocol-level Document actions
type Model interface {
	storage.Model
	// Get the ID of the document represented by this model
	ID() ([]byte, error)

	// PackCoreDocument packs the implementing document into a core document
	// should create the identifiers for the core document if not present
	PackCoreDocument() (*CoreDocumentModel, error)

	// UnpackCoreDocument must return the document.Model
	// assumes that core document has valid identifiers set
	UnpackCoreDocument(model *CoreDocumentModel) error

	// CalculateDataRoot calculates the dataroot of precise-proofs tree of the model
	CalculateDataRoot() ([]byte, error)

	// CreateProofs creates precise-proofs for given fields
	CreateProofs(fields []string) (proofs []*proofspb.Proof, err error)
}

// TokenRegistry defines NFT retrieval functions.
type TokenRegistry interface {
	// OwnerOf to retrieve owner of the tokenID
	OwnerOf(registry common.Address, tokenID []byte) (common.Address, error)
}

// CoreDocumentModel contains methods which handle all interactions mutating or reading from a core document
// Access to a core document should always go through this model
type CoreDocumentModel struct {
	Document *coredocumentpb.CoreDocument
	TokenRegistry
}

const (
	// ErrZeroCollaborators error when no collaborators are passed
	ErrZeroCollaborators = errors.Error("require at least one collaborator")

	// ErrNFTRoleMissing errors when role to generate proof doesn't exist
	ErrNFTRoleMissing = errors.Error("NFT Role doesn't exist")

	// nftByteCount is the length of combined bytes of registry and tokenID
	nftByteCount = 52
)

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

// NewCoreDocModel returns a new CoreDocumentModel
// Note: collaborators and salts are to be filled by the caller
// TODO: double check if registry should be initialised as nil
func NewCoreDocModel() *CoreDocumentModel {
	id := utils.RandomSlice(32)
	cd := &coredocumentpb.CoreDocument{
		DocumentIdentifier: id,
		CurrentVersion:     id,
		NextVersion:        utils.RandomSlice(32),
	}
	return &CoreDocumentModel{
		cd,
		nil,
	}
}

// PrepareNewVersion creates a new CoreDocumentModel with the version fields updated
// Adds collaborators and fills salts
// Note: new collaborators are added to the list with old collaborators.
//TODO: this will change when collaborators are moved down to next level
func (m *CoreDocumentModel) PrepareNewVersion(collaborators []string) (*CoreDocumentModel, error) {
	ndm, err := m.prepareNewVersion(collaborators)
	if err != nil {
		return nil, err
	}

	return ndm, ndm.setCoreDocumentSalts()
}

// prepareNewVersion prepares the next version of the CoreDocument
// Note: salts needs to be filled by the caller
func (m *CoreDocumentModel) prepareNewVersion(collaborators []string) (*CoreDocumentModel, error) {
	ndm := NewCoreDocModel()
	ncd := ndm.Document
	ocd := m.Document
	ucs, err := fetchUniqueCollaborators(ocd.Collaborators, collaborators)
	if err != nil {
		return nil, errors.New("failed to decode collaborator: %v", err)
	}

	cs := ocd.Collaborators
	for _, c := range ucs {
		c := c
		cs = append(cs, c[:])
	}

	ncd.Collaborators = cs
	ncd.Roles = m.Document.Roles
	ncd.ReadRules = m.Document.ReadRules
	err = ndm.addCollaboratorsToReadSignRules(ucs)
	if err != nil {
		return nil, err
	}

	if ocd.DocumentIdentifier == nil {
		return nil, errors.New("Document.DocumentIdentifier is nil")
	}
	ncd.DocumentIdentifier = ocd.DocumentIdentifier

	if ocd.CurrentVersion == nil {
		return nil, errors.New("Document.CurrentVersion is nil")
	}
	ncd.PreviousVersion = ocd.CurrentVersion

	if ocd.NextVersion == nil {
		return nil, errors.New("Document.NextVersion is nil")
	}

	ncd.CurrentVersion = ocd.NextVersion
	ncd.NextVersion = utils.RandomSlice(32)
	if ocd.DocumentRoot == nil {
		return nil, errors.New("DocumentRoot is nil")
	}
	ncd.PreviousRoot = ocd.DocumentRoot
	// copy over token registry
	ndm.TokenRegistry = m.TokenRegistry
	return ndm, nil
}

// NewWithCollaborators generates new core document, adds collaborators, adds read rules and fills salts
func (m *CoreDocumentModel) NewWithCollaborators(collaborators []string) (*CoreDocumentModel, error) {
	dm := NewCoreDocModel()
	ids, err := identity.CentIDsFromStrings(collaborators)
	if err != nil {
		return nil, errors.New("failed to decode collaborator: %v", err)
	}

	for i := range ids {
		dm.Document.Collaborators = append(dm.Document.Collaborators, ids[i][:])
	}
	err = dm.initReadRules(ids)
	if err != nil {
		return nil, errors.New("failed to init read rules: %v", err)
	}

	if err := dm.setCoreDocumentSalts(); err != nil {
		return nil, err
	}

	return dm, nil
}

// CreateProofs util function that takes document data tree, coreDocument and a list fo fields and generates proofs
func (m *CoreDocumentModel) CreateProofs(dataTree *proofs.DocumentTree, fields []string) (proofs []*proofspb.Proof, err error) {
	signingRootProofHashes, err := m.getSigningRootProofHashes()
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	cdtree, err := m.GetCoreDocumentTree()
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

// GetCoreDocumentTree returns the merkle tree for the coredoc root
func (m *CoreDocumentModel) GetCoreDocumentTree() (tree *proofs.DocumentTree, err error) {
	document := m.Document
	tree = NewDefaultTreeWithPrefix(ConvertToProofSalts(m.Document.CoredocumentSalts), CDTreePrefix, compactProperties(CDTreePrefix))
	err = tree.AddLeavesFromDocument(document)
	if err != nil {
		return nil, err
	}

	if document.EmbeddedData == nil {
		return nil, errors.New("EmbeddedData cannot be nil when generating signing tree")
	}
	parentProp := NewLeafProperty(CDTreePrefix, compactProperties(CDTreePrefix))
	// Adding document type as it is an excluded field in the tree
	documentTypeNode := proofs.LeafNode{
		Property: parentProp.FieldProp(DocumentTypeField, binary.LittleEndian.Uint32(compactProperties(DocumentTypeField))),
		Salt:     make([]byte, 32),
		Value:    []byte(document.EmbeddedData.TypeUrl),
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

// GetDocumentSigningTree returns the merkle tree for the signing root
func (m *CoreDocumentModel) GetDocumentSigningTree(dataRoot []byte) (tree *proofs.DocumentTree, err error) {
	// coredoc tree
	coreDocTree, err := m.GetCoreDocumentTree()
	if err != nil {
		return nil, err
	}

	// create the signing tree with data root and coredoc root as siblings
	tree = NewDefaultTree(ConvertToProofSalts(m.Document.CoredocumentSalts))
	//tree = NewDefaultTreeWithPrefix(ConvertToProofSalts(m.Document.CoredocumentSalts), SigningTreePrefix, compactProperties(SigningTreePrefix))
	parentProp := NewLeafProperty(SigningTreePrefix, compactProperties(SigningTreePrefix))

	err = tree.AddLeaves([]proofs.LeafNode{
		{
			Property: parentProp.FieldProp(DataRootField, binary.LittleEndian.Uint32(compactProperties(DataRootField))),
			Hash:     dataRoot,
			Hashed:   true,
		},
		{
			Property: parentProp.FieldProp(CDRootField, binary.LittleEndian.Uint32(compactProperties(CDRootField))),
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

// GetDocumentRootTree returns the merkle tree for the document root
func (m *CoreDocumentModel) GetDocumentRootTree() (tree *proofs.DocumentTree, err error) {
	document := m.Document
	tree = NewDefaultTree(ConvertToProofSalts(document.CoredocumentSalts))

	// The first leave added is the signing_root
	err = tree.AddLeaf(proofs.LeafNode{Hash: document.SigningRoot, Hashed: true, Property: NewLeafProperty(SigningRootField, compactProperties(SigningRootField))})
	if err != nil {
		return nil, err
	}

	// For every signature we create a LeafNode
	sigProperty := NewLeafProperty(SignaturesField, compactProperties(SignaturesField))
	sigLeafList := make([]proofs.LeafNode, len(document.Signatures)+1)
	sigLengthNode := proofs.LeafNode{
		Property: sigProperty.LengthProp(proofs.DefaultSaltsLengthSuffix),
		Salt:     make([]byte, 32),
		Value:    []byte(fmt.Sprintf("%d", len(document.Signatures))),
	}

	h := sha256.New()
	err = sigLengthNode.HashNode(h, true)
	if err != nil {
		return nil, err
	}
	sigLeafList[0] = sigLengthNode
	for i, sig := range document.Signatures {
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

// CalculateDocumentRoot calculates the document root of the core document
func (m *CoreDocumentModel) CalculateDocumentRoot() error {
	document := m.Document
	if len(document.SigningRoot) != 32 {
		return errors.New("signing root invalid")
	}

	tree, err := m.GetDocumentRootTree()
	if err != nil {
		return err
	}

	document.DocumentRoot = tree.RootHash()
	return nil
}

// getDataProofHashes returns the hashes needed to create a proof from DataRoot to SigningRoot. This method is used
// to create field proofs
func (m *CoreDocumentModel) getDataProofHashes(dataRoot []byte) (hashes [][]byte, err error) {
	tree, err := m.GetDocumentSigningTree(dataRoot)
	if err != nil {
		return
	}

	signingProof, err := tree.CreateProof(SigningTreePrefix + ".data_root")
	if err != nil {
		return
	}

	rootProofHashes, err := m.getSigningRootProofHashes()
	if err != nil {
		return
	}

	return append(signingProof.SortedHashes, rootProofHashes...), err
}

// getSigningRootProofHashes returns the hashes needed to create a proof for fields from SigningRoot to DocumentRoot. This method is used
// to create field proofs
func (m *CoreDocumentModel) getSigningRootProofHashes() (hashes [][]byte, err error) {
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
func (m *CoreDocumentModel) CalculateSigningRoot(dataRoot []byte) error {
	doc := m.Document
	tree, err := m.GetDocumentSigningTree(dataRoot)
	if err != nil {
		return err
	}

	doc.SigningRoot = tree.RootHash()
	return nil
}

// AccountCanRead validate if the core document can be read by the account .
// Returns an error if not.
func (m *CoreDocumentModel) AccountCanRead(account identity.CentID) bool {
	// loop though read rules, check all the rules
	return m.findRole(func(_, _ int, role *coredocumentpb.Role) bool {
		_, found := isAccountInRole(role, account)
		return found
	}, coredocumentpb.Action_ACTION_READ, coredocumentpb.Action_ACTION_READ_SIGN)
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
	if proofSalts == nil {
		return nil
	}

	protoSalts := make([]*coredocumentpb.DocumentSalt, len(*proofSalts))
	for i, pSalt := range *proofSalts {
		protoSalts[i] = &coredocumentpb.DocumentSalt{Value: pSalt.Value, Compact: pSalt.Compact}
	}

	return protoSalts
}

// ConvertToProofSalts converts protocolSalts into proofSalts
func ConvertToProofSalts(protoSalts []*coredocumentpb.DocumentSalt) *proofs.Salts {
	if protoSalts == nil {
		return nil
	}

	proofSalts := make(proofs.Salts, len(protoSalts))
	for i, pSalt := range protoSalts {
		proofSalts[i] = proofs.Salt{Value: pSalt.Value, Compact: pSalt.Compact}
	}

	return &proofSalts
}

// setCoreDocumentSalts creates a new coredocument.Salts and fills it in case that is not initialized yet
func (m *CoreDocumentModel) setCoreDocumentSalts() error {
	if m.Document.CoredocumentSalts == nil {
		pSalts, err := GenerateNewSalts(m.Document, CDTreePrefix, compactProperties(CDTreePrefix))
		if err != nil {
			return err
		}

		m.Document.CoredocumentSalts = ConvertToProtoSalts(pSalts)
	}

	return nil
}

// PackCoreDocument sets the embed data and embed salts and generates core doc salts if not exists
func (m *CoreDocumentModel) PackCoreDocument(embedData *any.Any, embedSalts []*coredocumentpb.DocumentSalt) error {
	m.Document.EmbeddedData = embedData
	m.Document.EmbeddedDataSalts = embedSalts
	return m.setCoreDocumentSalts()
}

// UnpackCoreDocument sets the embed data and embed salts and generates core doc salts if not exists
func (m *CoreDocumentModel) UnpackCoreDocument() error {
	m.Document.EmbeddedData = nil
	m.Document.EmbeddedDataSalts = nil
	return m.setCoreDocumentSalts()
}

// initReadRules initiates the read rules for a given CoreDocumentModel.
// Collaborators are given Read_Sign action.
// if the rules are created already, this is a no-op.
func (m *CoreDocumentModel) initReadRules(collabs []identity.CentID) error {
	cd := m.Document
	if len(cd.Roles) > 0 && len(cd.ReadRules) > 0 {
		return nil
	}

	if len(collabs) < 1 {
		return ErrZeroCollaborators
	}

	return m.addCollaboratorsToReadSignRules(collabs)
}

// addNewRule creates a new rule as per the role and action.
func (m *CoreDocumentModel) addNewRule(role *coredocumentpb.Role, action coredocumentpb.Action) {
	cd := m.Document
	cd.Roles = append(cd.Roles, role)

	rule := new(coredocumentpb.ReadRule)
	rule.Roles = append(rule.Roles, role.RoleKey)
	rule.Action = action
	cd.ReadRules = append(cd.ReadRules, rule)
}

// findRole calls OnRole for every role that matches the actions passed in
func (m *CoreDocumentModel) findRole(onRole func(rridx, ridx int, role *coredocumentpb.Role) bool, actions ...coredocumentpb.Action) bool {
	cd := m.Document
	am := make(map[int32]struct{})
	for _, a := range actions {
		am[int32(a)] = struct{}{}
	}

	for i, rule := range cd.ReadRules {
		if _, ok := am[int32(rule.Action)]; !ok {
			continue
		}

		for j, rk := range rule.Roles {
			role, err := getRole(rk, cd.Roles)
			if err != nil {
				// seems like roles and rules are not in sync
				// skip to next one
				continue
			}

			if onRole(i, j, role) {
				return true
			}

		}
	}

	return false
}

// isAccountInRole returns the index of the collaborator and true if account is in the given role as collaborators.
func isAccountInRole(role *coredocumentpb.Role, account identity.CentID) (idx int, found bool) {
	for i, id := range role.Collaborators {
		if bytes.Equal(id, account[:]) {
			return i, true
		}
	}

	return idx, false
}

func getRole(key []byte, roles []*coredocumentpb.Role) (*coredocumentpb.Role, error) {
	for _, role := range roles {
		if utils.IsSameByteSlice(role.RoleKey, key) {
			return role, nil
		}
	}

	return nil, errors.New("role %d not found", key)
}

func (m *CoreDocumentModel) addCollaboratorsToReadSignRules(collabs []identity.CentID) error {
	if len(collabs) == 0 {
		return nil
	}
	// create a role for given collaborators
	role := new(coredocumentpb.Role)
	cd := m.Document
	rk, err := utils.ConvertIntToByte32(len(cd.Roles))
	if err != nil {
		return err
	}
	role.RoleKey = rk[:]
	for _, c := range collabs {
		c := c
		role.Collaborators = append(role.Collaborators, c[:])
	}

	m.addNewRule(role, coredocumentpb.Action_ACTION_READ_SIGN)

	return nil
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

// GetExternalCollaborators returns collaborators of a document without the own centID.
func (m *CoreDocumentModel) GetExternalCollaborators(selfCentID identity.CentID) ([][]byte, error) {
	var collabs [][]byte

	for _, collab := range m.Document.Collaborators {
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

// NFTOwnerCanRead checks if the nft owner/account can read the document
func (m *CoreDocumentModel) NFTOwnerCanRead(registry common.Address, tokenID []byte, account identity.CentID) error {
	// check if the account can read the doc
	if m.AccountCanRead(account) {
		return nil
	}

	// check if the nft is present in read rules
	found := m.findRole(func(_, _ int, role *coredocumentpb.Role) bool {
		_, found := isNFTInRole(role, registry, tokenID)
		return found
	}, coredocumentpb.Action_ACTION_READ)

	if !found {
		return errors.New("nft missing")
	}

	// get the owner of the NFT
	owner, err := m.TokenRegistry.OwnerOf(registry, tokenID)
	if err != nil {
		return errors.New("failed to get NFT owner: %v", err)
	}

	// TODO(ved): this will always fail until we roll out identity v2 with CentID type as common.Address
	if !bytes.Equal(owner.Bytes(), account[:]) {
		return errors.New("account (%v) not owner of the NFT", account.String())
	}

	return nil
}

// ConstructNFT appends registry and tokenID to byte slice
func ConstructNFT(registry common.Address, tokenID []byte) ([]byte, error) {
	var nft []byte
	// first 20 bytes of registry
	nft = append(nft, registry.Bytes()...)

	// next 32 bytes of the tokenID
	nft = append(nft, tokenID...)

	if len(nft) != nftByteCount {
		return nil, errors.New("byte length mismatch")
	}

	return nft, nil
}

// AddNFTToReadRules adds NFT token to the read rules of core document.
func (m *CoreDocumentModel) AddNFTToReadRules(registry common.Address, tokenID []byte) error {
	cd := m.Document
	nft, err := ConstructNFT(registry, tokenID)
	if err != nil {
		return errors.New("failed to construct NFT: %v", err)
	}

	role := new(coredocumentpb.Role)
	rk, err := utils.ConvertIntToByte32(len(cd.Roles))
	if err != nil {
		return err
	}
	role.RoleKey = rk[:]
	role.Nfts = append(role.Nfts, nft)
	m.addNewRule(role, coredocumentpb.Action_ACTION_READ)
	if err := m.setCoreDocumentSalts(); err != nil {
		return errors.New("failed to generate CoreDocumentSalts")
	}
	return nil
}

//ValidateDocumentAccess validates the GetDocument request against the AccessType indicated in the request
func (m *CoreDocumentModel) ValidateDocumentAccess(docReq *p2ppb.GetDocumentRequest, requesterCentID identity.CentID) error {
	// checks which access type is relevant for the request
	switch docReq.GetAccessType() {
	case p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION:
		if m.AccountCanRead(requesterCentID) {
			return errors.New("requester does not have access")
		}
	case p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION:
		registry := common.BytesToAddress(docReq.NftRegistryAddress)
		if m.NFTOwnerCanRead(registry, docReq.NftTokenId, requesterCentID) != nil {
			return errors.New("requester does not have access")
		}
		//// case AccessTokenValidation
		// case p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION:
		//
		// case p2ppb.AccessType_ACCESS_TYPE_INVALID:
	default:
		return errors.New("invalid access type ")
	}
	return nil
}

// isNFTInRole checks if the given nft(registry + token) is part of the core document role.
// If found, returns the index of the nft in the role and true
func isNFTInRole(role *coredocumentpb.Role, registry common.Address, tokenID []byte) (nftIdx int, found bool) {
	enft, err := ConstructNFT(registry, tokenID)
	if err != nil {
		return nftIdx, false
	}

	for i, n := range role.Nfts {
		if bytes.Equal(n, enft) {
			return i, true
		}
	}

	return nftIdx, false
}

// AddNFT returns a new CoreDocument model with nft added to the Core document. If grantReadAccess is true, the nft is added
// to the read rules.
func (m *CoreDocumentModel) AddNFT(grantReadAccess bool, registry common.Address, tokenID []byte) (*CoreDocumentModel, error) {
	data := m.Document.EmbeddedData
	ndm, err := m.prepareNewVersion(nil)
	if err != nil {
		return nil, err
	}

	ndm.Document.EmbeddedData = data
	cd := ndm.Document
	nft := getStoredNFT(cd.Nfts, registry.Bytes())
	if nft == nil {
		nft = new(coredocumentpb.NFT)
		// add 12 empty bytes
		eb := make([]byte, 12, 12)
		nft.RegistryId = append(registry.Bytes(), eb...)
		cd.Nfts = append(cd.Nfts, nft)
	}
	nft.TokenId = tokenID

	if grantReadAccess {
		err = ndm.AddNFTToReadRules(registry, tokenID)
		if err != nil {
			return nil, err
		}
	}

	return ndm, ndm.setCoreDocumentSalts()
}

// IsNFTMinted checks if the there is an NFT that is minted against this document in the given registry.
func (m *CoreDocumentModel) IsNFTMinted(tokenRegistry TokenRegistry, registry common.Address) bool {
	nft := getStoredNFT(m.Document.Nfts, registry.Bytes())
	if nft == nil {
		return false
	}

	_, err := tokenRegistry.OwnerOf(registry, nft.TokenId)
	return err == nil
}

func getStoredNFT(nfts []*coredocumentpb.NFT, registry []byte) *coredocumentpb.NFT {
	for _, nft := range nfts {
		if bytes.Equal(nft.RegistryId[:20], registry) {
			return nft
		}
	}

	return nil
}

// IsAccountInRole checks if the the account exists in the role provided
func (m *CoreDocumentModel) IsAccountInRole(roleKey []byte, account identity.CentID) bool {
	role, err := getRole(roleKey, m.Document.Roles)
	if err != nil {
		return false
	}

	_, found := isAccountInRole(role, account)
	return found
}

// GetNFTProofs generate proofs required to mint an NFT
func (m *CoreDocumentModel) GetNFTProofs(
	dataRoot []byte,
	account identity.CentID,
	registry common.Address,
	tokenID []byte,
	nftUniqueProof, readAccessProof bool) (proofs []*proofspb.Proof, err error) {

	var pfKeys []string
	if nftUniqueProof {
		pk, err := getNFTUniqueProofKey(m.Document.Nfts, registry)
		if err != nil {
			return nil, err
		}

		pfKeys = append(pfKeys, pk)
	}

	if readAccessProof {
		pks, err := getReadAccessProofKeys(m, registry, tokenID)
		if err != nil {
			return nil, err
		}

		pfKeys = append(pfKeys, pks...)
	}

	signingRootProofHashes, err := m.getSigningRootProofHashes()
	if err != nil {
		return nil, errors.New("failed to generate signing root proofs: %v", err)
	}

	cdtree, err := m.GetCoreDocumentTree()
	if err != nil {
		return nil, errors.New("failed to generate document tree: %v", err)
	}

	for _, field := range pfKeys {
		proof, err := cdtree.CreateProof(field)
		if err != nil {
			return nil, errors.New("failed to create proof: %v", err)
		}

		proof.SortedHashes = append(proof.SortedHashes, dataRoot)
		proof.SortedHashes = append(proof.SortedHashes, signingRootProofHashes...)
		proofs = append(proofs, &proof)
	}

	return proofs, nil
}

func getReadAccessProofKeys(m *CoreDocumentModel, registry common.Address, tokenID []byte) (pks []string, err error) {
	var rridx int  // index of the read rules which contain the role
	var ridx int   // index of the role
	var nftIdx int // index of the NFT in the above role
	var rk []byte  // role key of the above role

	found := m.findRole(func(i, j int, role *coredocumentpb.Role) bool {
		z, found := isNFTInRole(role, registry, tokenID)
		if found {
			rridx = i
			ridx = j
			rk = role.RoleKey
			nftIdx = z
		}

		return found
	}, coredocumentpb.Action_ACTION_READ)

	if !found {
		return nil, ErrNFTRoleMissing
	}

	return []string{
		fmt.Sprintf(CDTreePrefix+".read_rules[%d].roles[%d]", rridx, ridx),          // proof that a read rule exists with the nft role
		fmt.Sprintf(CDTreePrefix+".roles[%s].nfts[%d]", hexutil.Encode(rk), nftIdx), // proof that role with nft exists
		fmt.Sprintf(CDTreePrefix+".read_rules[%d].action", rridx),                   // proof that this read rule has read access
	}, nil
}

func getNFTUniqueProofKey(nfts []*coredocumentpb.NFT, registry common.Address) (pk string, err error) {
	nft := getStoredNFT(nfts, registry.Bytes())
	if nft == nil {
		return pk, errors.New("nft is missing from the document")
	}

	key := hexutil.Encode(nft.RegistryId)
	return fmt.Sprintf(CDTreePrefix+".nfts[%s]", key), nil
}

func getRoleProofKey(roles []*coredocumentpb.Role, roleKey []byte, account identity.CentID) (pk string, err error) {
	role, err := getRole(roleKey, roles)
	if err != nil {
		return pk, err
	}

	idx, found := isAccountInRole(role, account)
	if !found {
		return pk, ErrNFTRoleMissing
	}

	return fmt.Sprintf(CDTreePrefix+".roles[%s].collaborators[%d]", hexutil.Encode(role.RoleKey), idx), nil
}
