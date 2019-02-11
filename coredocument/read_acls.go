package coredocument

import (
	"bytes"

	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// ErrZeroCollaborators error when no collaborators are passed
	ErrZeroCollaborators = errors.Error("require at least one collaborator")

	// nftByteCount is the length of combined bytes of registry and tokenID
	nftByteCount = 52
)

// TokenRegistry defines NFT retrieval functions.
type TokenRegistry interface {
	// OwnerOf to retrieve owner of the tokenID
	OwnerOf(registry common.Address, tokenID []byte) (common.Address, error)
}

// initReadRules initiates the read rules for a given coredocument.
// Collaborators are given Read_Sign action.
// if the rules are created already, this is a no-op.
func initReadRules(cd *coredocumentpb.CoreDocument, collabs []identity.CentID) error {
	if len(cd.Roles) > 0 && len(cd.ReadRules) > 0 {
		return nil
	}

	if len(collabs) < 1 {
		return ErrZeroCollaborators
	}

	return addCollaboratorsToReadSignRules(cd, collabs)
}

func addCollaboratorsToReadSignRules(cd *coredocumentpb.CoreDocument, collabs []identity.CentID) error {
	if len(collabs) == 0 {
		return nil
	}

	// create a role for given collaborators
	role := new(coredocumentpb.Role)
	rk, err := utils.ConvertIntToByte32(len(cd.Roles))
	if err != nil {
		return err
	}
	role.RoleKey = rk[:]
	for _, c := range collabs {
		c := c
		role.Collaborators = append(role.Collaborators, c[:])
	}

	addNewRule(cd, role, coredocumentpb.Action_ACTION_READ_SIGN)
	return nil
}

// addNewRule creates a new rule as per the role and action.
func addNewRule(cd *coredocumentpb.CoreDocument, role *coredocumentpb.Role, action coredocumentpb.Action) {
	cd.Roles = append(cd.Roles, role)

	rule := new(coredocumentpb.ReadRule)
	rule.Roles = append(rule.Roles, role.RoleKey)
	rule.Action = action
	cd.ReadRules = append(cd.ReadRules, rule)
}

// AddNFTToReadRules adds NFT token to the read rules of core document.
func AddNFTToReadRules(cd *coredocumentpb.CoreDocument, registry common.Address, tokenID []byte) error {
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
	addNewRule(cd, role, coredocumentpb.Action_ACTION_READ)
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

// ReadAccessValidator defines validator functions for account .
type ReadAccessValidator interface {
	AccountCanRead(cd *coredocumentpb.CoreDocument, account identity.CentID) bool
	NFTOwnerCanRead(
		cd *coredocumentpb.CoreDocument,
		registry common.Address,
		tokenID []byte,
		account identity.CentID) error
}

// readAccessValidator implements ReadAccessValidator.
type readAccessValidator struct {
	tokenRegistry TokenRegistry
}

// AccountCanRead validate if the core document can be read by the account .
// Returns an error if not.
func (r readAccessValidator) AccountCanRead(cd *coredocumentpb.CoreDocument, account identity.CentID) bool {
	// loop though read rules
	return FindRole(cd, coredocumentpb.Action_ACTION_READ_SIGN, func(_, _ int, role *coredocumentpb.Role) bool {
		_, found := IsAccountInRole(role, account)
		return found
	})
}

// GetRole returns the role matching the key
func GetRole(key []byte, roles []*coredocumentpb.Role) (*coredocumentpb.Role, error) {
	for _, role := range roles {
		if utils.IsSameByteSlice(role.RoleKey, key) {
			return role, nil
		}
	}

	return nil, errors.New("role %d not found", key)
}

// IsAccountInRole returns true and the index of the collaborator if account is in the given role as collaborators.
func IsAccountInRole(role *coredocumentpb.Role, account identity.CentID) (int, bool) {
	for i, id := range role.Collaborators {
		if bytes.Equal(id, account[:]) {
			return i, true
		}
	}

	return 0, false
}

// AccountValidator returns the ReadAccessValidator to verify account .
func AccountValidator() ReadAccessValidator {
	return readAccessValidator{}
}

// NftValidator returns the ReadAccessValidator for nft owner verification.
func NftValidator(tr TokenRegistry) ReadAccessValidator {
	return readAccessValidator{tokenRegistry: tr}
}

// NFTOwnerCanRead checks if the nft owner/account can read the document
// Note: signature should be calculated from the hash which is calculated as
// keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
func (r readAccessValidator) NFTOwnerCanRead(
	cd *coredocumentpb.CoreDocument,
	registry common.Address,
	tokenID []byte,
	account identity.CentID) error {

	// check if the account can read the doc
	if r.AccountCanRead(cd, account) {
		return nil
	}

	// check if the nft is present in read rules
	found := FindRole(cd, coredocumentpb.Action_ACTION_READ, func(_, _ int, role *coredocumentpb.Role) bool {
		_, found := IsNFTInRole(role, registry, tokenID)
		return found
	})

	if !found {
		return errors.New("nft missing")
	}

	// get the owner of the NFT
	owner, err := r.tokenRegistry.OwnerOf(registry, tokenID)
	if err != nil {
		return errors.New("failed to get NFT owner: %v", err)
	}

	// TODO(ved): this will always fail until we roll out identity v2 with CentID type as common.Address
	if !bytes.Equal(owner.Bytes(), account[:]) {
		return errors.New("account (%v) not owner of the NFT", account.String())
	}

	return nil
}

// FindRole calls OnRole for every role,
// if onRole returns true, returns true
// else returns false
func FindRole(
	cd *coredocumentpb.CoreDocument,
	action coredocumentpb.Action,
	onRole func(rridx, ridx int, role *coredocumentpb.Role) bool) bool {
	for i, rule := range cd.ReadRules {
		if rule.Action != action {
			continue
		}

		for j, rk := range rule.Roles {
			role, err := GetRole(rk, cd.Roles)
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

// IsNFTInRole checks if the given nft(registry + token) is part of the core document role.
func IsNFTInRole(role *coredocumentpb.Role, registry common.Address, tokenID []byte) (int, bool) {
	enft, err := ConstructNFT(registry, tokenID)
	if err != nil {
		return 0, false
	}

	for i, n := range role.Nfts {
		if bytes.Equal(n, enft) {
			return i, true
		}
	}

	return 0, false
}
