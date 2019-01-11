package coredocument

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
)

// ErrZeroCollaborators error when no collaborators are passed
const ErrZeroCollaborators = errors.Error("require at least one collaborator")

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
