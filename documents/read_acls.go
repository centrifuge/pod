package documents

import (
	"bytes"
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// initReadRules initiates the read rules for a given CoreDocumentModel.
// Collaborators are given Read_Sign action.
// if the rules are created already, this is a no-op.
// if collaborators are empty, it is a no-op
func (cd *CoreDocument) initReadRules(collaborators []identity.CentID) {
	if len(cd.Document.Roles) > 0 && len(cd.Document.ReadRules) > 0 {
		return
	}

	if len(collaborators) < 1 {
		return
	}

	cd.addCollaboratorsToReadSignRules(collaborators)
}

// findRole calls OnRole for every role that matches the actions passed in
func findRole(cd coredocumentpb.CoreDocument, onRole func(rridx, ridx int, role *coredocumentpb.Role) bool, actions ...coredocumentpb.Action) bool {
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

// GetExternalCollaborators returns collaborators of a Document without the own centID.
func (cd *CoreDocument) GetExternalCollaborators(self identity.CentID) ([][]byte, error) {
	var cs [][]byte
	for _, c := range cd.Document.Collaborators {
		c := c
		id, err := identity.ToCentID(c)
		if err != nil {
			return nil, errors.New("failed to convert to CentID: %v", err)
		}
		if !self.Equal(id) {
			cs = append(cs, c)
		}
	}

	return cs, nil
}

// NFTOwnerCanRead checks if the nft owner/account can read the Document
func (cd *CoreDocument) NFTOwnerCanRead(tokenRegistry TokenRegistry, registry common.Address, tokenID []byte, account identity.CentID) error {
	// check if the account can read the doc
	if cd.AccountCanRead(account) {
		return nil
	}

	// check if the nft is present in read rules
	found := findRole(cd.Document, func(_, _ int, role *coredocumentpb.Role) bool {
		_, found := isNFTInRole(role, registry, tokenID)
		return found
	}, coredocumentpb.Action_ACTION_READ)

	if !found {
		return errors.New("nft not found in the Document")
	}

	// get the owner of the NFT
	owner, err := tokenRegistry.OwnerOf(registry, tokenID)
	if err != nil {
		return errors.New("failed to get NFT owner: %v", err)
	}

	// TODO(ved): this will always fail until we roll out identity v2 with CentID type as common.Address
	if !bytes.Equal(owner.Bytes(), account[:]) {
		return errors.New("account (%v) not owner of the NFT", account.String())
	}

	return nil
}

// AccountCanRead validate if the core Document can be read by the account .
// Returns an error if not.
func (cd *CoreDocument) AccountCanRead(account identity.CentID) bool {
	// loop though read rules, check all the rules
	return findRole(cd.Document, func(_, _ int, role *coredocumentpb.Role) bool {
		_, found := isAccountInRole(role, account)
		return found
	}, coredocumentpb.Action_ACTION_READ, coredocumentpb.Action_ACTION_READ_SIGN)
}

// addNFTToReadRules adds NFT token to the read rules of core Document.
func (cd *CoreDocument) addNFTToReadRules(registry common.Address, tokenID []byte) error {
	nft, err := ConstructNFT(registry, tokenID)
	if err != nil {
		return errors.New("failed to construct NFT: %v", err)
	}

	role := &coredocumentpb.Role{RoleKey: utils.RandomSlice(32)}
	role.Nfts = append(role.Nfts, nft)
	cd.addNewRule(role, coredocumentpb.Action_ACTION_READ)
	return cd.setSalts()
}

// AddNFT returns a new CoreDocument model with nft added to the Core Document. If grantReadAccess is true, the nft is added
// to the read rules.
func (cd *CoreDocument) AddNFT(grantReadAccess bool, registry common.Address, tokenID []byte) (*CoreDocument, error) {
	ncd, err := cd.PrepareNewVersion(nil, false)
	if err != nil {
		return nil, errors.New("failed to prepare new version: %v", err)
	}

	nft := getStoredNFT(ncd.Document.Nfts, registry.Bytes())
	if nft == nil {
		nft = new(coredocumentpb.NFT)
		// add 12 empty bytes
		eb := make([]byte, 12, 12)
		nft.RegistryId = append(registry.Bytes(), eb...)
		ncd.Document.Nfts = append(ncd.Document.Nfts, nft)
	}
	nft.TokenId = tokenID

	if grantReadAccess {
		err = ncd.addNFTToReadRules(registry, tokenID)
		if err != nil {
			return nil, err
		}
	}

	return ncd, ncd.setSalts()
}

// IsNFTMinted checks if the there is an NFT that is minted against this Document in the given registry.
func (cd *CoreDocument) IsNFTMinted(tokenRegistry TokenRegistry, registry common.Address) bool {
	nft := getStoredNFT(cd.Document.Nfts, registry.Bytes())
	if nft == nil {
		return false
	}

	_, err := tokenRegistry.OwnerOf(registry, nft.TokenId)
	return err == nil
}

// CreateNFTProofs generate proofs returns proofs for NFT minting.
func (cd *CoreDocument) CreateNFTProofs(
	docType string,
	account identity.CentID,
	registry common.Address,
	tokenID []byte,
	nftUniqueProof, readAccessProof bool) (proofs []*proofspb.Proof, err error) {

	if len(cd.Document.DataRoot) != idSize {
		return nil, errors.New("data root is invalid")
	}

	var pfKeys []string
	if nftUniqueProof {
		pk, err := getNFTUniqueProofKey(cd.Document.Nfts, registry)
		if err != nil {
			return nil, err
		}

		pfKeys = append(pfKeys, pk)
	}

	if readAccessProof {
		pks, err := getReadAccessProofKeys(cd.Document, registry, tokenID)
		if err != nil {
			return nil, err
		}

		pfKeys = append(pfKeys, pks...)
	}

	signingRootProofHashes, err := cd.getSigningRootProofHashes()
	if err != nil {
		return nil, errors.New("failed to generate signing root proofs: %v", err)
	}

	cdTree, err := cd.documentTree(docType)
	if err != nil {
		return nil, errors.New("failed to generate core Document tree: %v", err)
	}

	proofs, missedProofs := generateProofs(cdTree, pfKeys, append([][]byte{cd.Document.DataRoot}, signingRootProofHashes...))
	if len(missedProofs) != 0 {
		return nil, errors.New("failed to create proofs for fields %v", missedProofs)
	}

	return proofs, nil
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

// isNFTInRole checks if the given nft(registry + token) is part of the core Document role.
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

func getStoredNFT(nfts []*coredocumentpb.NFT, registry []byte) *coredocumentpb.NFT {
	for _, nft := range nfts {
		if bytes.Equal(nft.RegistryId[:20], registry) {
			return nft
		}
	}

	return nil
}

func getReadAccessProofKeys(cd coredocumentpb.CoreDocument, registry common.Address, tokenID []byte) (pks []string, err error) {
	var rridx int  // index of the read rules which contain the role
	var ridx int   // index of the role
	var nftIdx int // index of the NFT in the above role
	var rk []byte  // role key of the above role

	found := findRole(cd, func(i, j int, role *coredocumentpb.Role) bool {
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
		fmt.Sprintf("read_rules[%d].roles[%d]", rridx, ridx),          // proof that a read rule exists with the nft role
		fmt.Sprintf("roles[%s].nfts[%d]", hexutil.Encode(rk), nftIdx), // proof that role with nft exists
		fmt.Sprintf("read_rules[%d].action", rridx),                   // proof that this read rule has read access
	}, nil
}

func getNFTUniqueProofKey(nfts []*coredocumentpb.NFT, registry common.Address) (pk string, err error) {
	nft := getStoredNFT(nfts, registry.Bytes())
	if nft == nil {
		return pk, errors.New("nft is missing from the Document")
	}

	key := hexutil.Encode(nft.RegistryId)
	return fmt.Sprintf("nfts[%s]", key), nil
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

	return fmt.Sprintf("roles[%s].collaborators[%d]", hexutil.Encode(role.RoleKey), idx), nil
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
