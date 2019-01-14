package coredocument

import (
	"bytes"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
)

const (
	// ErrZeroCollaborators error when no collaborators are passed
	ErrZeroCollaborators = errors.Error("require at least one collaborator")

	// ErrPeerNotFound error when peer is not listed in the access list
	ErrPeerNotFound = errors.Error("peer not found in the access list")
)

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
}

// readAccessValidator implements ReadAccessValidator.
type readAccessValidator struct{}

// PeerCanRead validate if the core document can be read by the peer.
// Returns an error if not.
func (r readAccessValidator) PeerCanRead(cd *coredocumentpb.CoreDocument, peer identity.CentID) error {
	// lets loop though read rules
	for _, rule := range cd.ReadRules {
		for _, rk := range rule.Roles {
			role, err := getRole(rk, cd.Roles)
			if err != nil {
				// seems like roles and rules are not in sync
				// skip to next one
				continue
			}

			if isPeerInRole(role, peer) {
				return nil
			}
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
