package coredocument

import (
	"bytes"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// ErrZeroCollaborators error when no collaborators are passed
	ErrZeroCollaborators = errors.Error("require at least one collaborator")

	// ErrPeerNotFound error when peer is not found in the read rules
	ErrPeerNotFound = errors.Error("peer not found")

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

func addCollaboratorsToReadSignRules(cd *coredocumentpb.CoreDocument, collabs []identity.CentID) {
	if len(collabs) == 0 {
		return
	}

	// create a role for given collaborators
	role := new(coredocumentpb.Role)
	for _, c := range collabs {
		c := c
		role.Collaborators = append(role.Collaborators, c[:])
	}

	// add the role to Roles
	cd.Roles = appendRole(cd.Roles, role)

	// create a rule
	rule := new(coredocumentpb.ReadRule)
	rule.Roles = append(rule.Roles, cd.Roles[len(cd.Roles)-1].RoleKey)
	rule.Action = coredocumentpb.Action_ACTION_READ_SIGN
	cd.ReadRules = append(cd.ReadRules, rule)
}

// appendRole appends the roles to role entry
func appendRole(roles []*coredocumentpb.RoleEntry, role *coredocumentpb.Role) []*coredocumentpb.RoleEntry {
	return append(roles, &coredocumentpb.RoleEntry{
		RoleKey: uint32(len(roles)),
		Role:    role,
	})
}

// ReadAccessValidator defines validator functions for peer.
type ReadAccessValidator interface {
	PeerCanRead(cd *coredocumentpb.CoreDocument, peer identity.CentID) error
	NFTOwnerCanRead(
		cd *coredocumentpb.CoreDocument,
		registry common.Address,
		tokenID,
		message,
		signature []byte,
		peer identity.CentID) error
}

// readAccessValidator implements ReadAccessValidator.
type readAccessValidator struct {
	tokenRegistry TokenRegistry
}

// PeerCanRead validate if the core document can be read by the peer.
// Returns an error if not.
func (r readAccessValidator) PeerCanRead(cd *coredocumentpb.CoreDocument, peer identity.CentID) error {
	// lets loop though read rules
	ch := iterateReadRoles(cd)
	for role := range ch {
		if isPeerInRole(role, peer) {
			return nil
		}
	}

	return ErrPeerNotFound
}

func getRole(key uint32, roles []*coredocumentpb.RoleEntry) (*coredocumentpb.Role, error) {
	for _, roleEntry := range roles {
		if roleEntry.RoleKey == key {
			return roleEntry.Role, nil
		}
	}

	return nil, errors.New("role %d not found", key)
}

// isPeerInRole returns true if peer is in the given role as collaborators.
func isPeerInRole(role *coredocumentpb.Role, peer identity.CentID) bool {
	for _, id := range role.Collaborators {
		if bytes.Equal(id, peer[:]) {
			return true
		}
	}

	return false
}

// peerValidator return the
func peerValidator() ReadAccessValidator {
	return readAccessValidator{}
}

// NFTOwnerCanRead checks if the nft owner/peer can read the document
func (r readAccessValidator) NFTOwnerCanRead(
	cd *coredocumentpb.CoreDocument,
	registry common.Address,
	tokenID []byte,
	message []byte,
	signature []byte,
	peer identity.CentID) error {

	// check if the peer can read the doc
	if err := r.PeerCanRead(cd, peer); err == nil {
		return nil
	}

	// check if the nft is present in read rules
	ch := iterateReadRoles(cd)
	found := false
	for role := range ch {
		if isNFTInRole(role, registry, tokenID) {
			found = true
		}
	}

	if !found {
		return errors.New("nft missing")
	}

	// get the owner of the NFT
	owner, err := r.tokenRegistry.OwnerOf(registry, tokenID)
	if err != nil {
		return errors.New("failed to get NFT owner: %v", err)
	}

	// recover owner from the signature
	rowner, err := secp256k1.EcRecover(message, signature)
	if err != nil {
		return errors.New("failed to get owner from signature: %v", err)
	}

	if !bytes.Equal(owner.Bytes(), rowner.Bytes()) {
		return errors.New("identity(%v) doesn't own NFT", rowner.String())
	}

	return nil
}

// iterateReadRoles iterates through each role present in read rule
func iterateReadRoles(cd *coredocumentpb.CoreDocument) <-chan *coredocumentpb.Role {
	ch := make(chan *coredocumentpb.Role)
	go func() {
		for _, rule := range cd.ReadRules {
			for _, rk := range rule.Roles {
				role, err := getRole(rk, cd.Roles)
				if err != nil {
					// seems like roles and rules are not in sync
					// skip to next one
					continue
				}

				ch <- role
			}
		}

		close(ch)
	}()

	return ch
}

func isNFTInRole(role *coredocumentpb.Role, registry common.Address, tokenID []byte) bool {
	var enft []byte
	// firs 20 bytes of registry
	enft = append(enft, registry.Bytes()...)

	// next 32 bytes of the tokenID
	enft = append(enft, tokenID...)

	if len(enft) != nftByteCount {
		return false
	}

	for _, n := range role.Nfts {
		if bytes.Equal(n, enft) {
			return true
		}
	}

	return false
}
