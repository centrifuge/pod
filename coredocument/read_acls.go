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

	addCollaboratorsToReadSignRules(cd, collabs)
	return nil
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

// addNFTToReadRules adds NFT token to the read rules of core document.
func addNFTToReadRules(cd *coredocumentpb.CoreDocument, registry common.Address, tokenID []byte) error {
	nft, err := constructNFT(registry, tokenID)
	if err != nil {
		return errors.New("failed to construct NFT: %v", err)
	}

	role := new(coredocumentpb.Role)
	role.Nfts = append(role.Nfts, nft)
	addNewRule(cd, role, coredocumentpb.Action_ACTION_READ)
	return nil
}

// constructNFT appends registry and tokenID to byte slice
func constructNFT(registry common.Address, tokenID []byte) ([]byte, error) {
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
	return findRole(cd, coredocumentpb.Action_ACTION_READ_SIGN, func(role *coredocumentpb.Role) bool {
		return isAccountInRole(role, account)
	})
}

func getRole(key []byte, roles []*coredocumentpb.Role) (*coredocumentpb.Role, error) {
	for _, role := range roles {
		if utils.IsSameByteSlice(role.RoleKey, key) {
			return role, nil
		}
	}

	return nil, errors.New("role %d not found", key)
}

// isAccountInRole returns true if account is in the given role as collaborators.
func isAccountInRole(role *coredocumentpb.Role, account identity.CentID) bool {
	for _, id := range role.Collaborators {
		if bytes.Equal(id, account[:]) {
			return true
		}
	}

	return false
}

// account returns the ReadAccessValidator tp verify account .
func accountValidator() ReadAccessValidator {
	return readAccessValidator{}
}

// nftValidator returns the ReadAccessValidator for nft owner verification.
func nftValidator(tr TokenRegistry) ReadAccessValidator {
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
	found := findRole(cd, coredocumentpb.Action_ACTION_READ, func(role *coredocumentpb.Role) bool {
		return isNFTInRole(role, registry, tokenID)
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

// findRole calls OnRole for every role,
// if onRole returns true, returns true
// else returns false
func findRole(
	cd *coredocumentpb.CoreDocument,
	action coredocumentpb.Action,
	onRole func(role *coredocumentpb.Role) bool) bool {
	for _, rule := range cd.ReadRules {
		if rule.Action != action {
			continue
		}

		for _, rk := range rule.Roles {
			role, err := getRole(rk, cd.Roles)
			if err != nil {
				// seems like roles and rules are not in sync
				// skip to next one
				continue
			}

			if onRole(role) {
				return true
			}

		}
	}

	return false
}

// isNFTInRole checks if the given nft(registry + token) is part of the core document role.
func isNFTInRole(role *coredocumentpb.Role, registry common.Address, tokenID []byte) bool {
	enft, err := constructNFT(registry, tokenID)
	if err != nil {
		return false
	}

	for _, n := range role.Nfts {
		if bytes.Equal(n, enft) {
			return true
		}
	}

	return false
}
