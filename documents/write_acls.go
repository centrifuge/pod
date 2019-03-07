package documents

import (
	"bytes"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
)

// changedField holds the compact property, old and new value of the field that is changed
// if the old is nil, then it is a set operation
// if new is nil, then it is an unset operation
// if both old and new are set, then it is an edit operation
type changedField struct {
	property, old, new []byte
}

// getChangedFields takes two document trees and returns the compact value, old and new value of the fields that are changed in new tree.
// Properties may have been added to the new tree or removed from the new tree.
// In Either case, since the new tree is different from old, that is considered a change.
func getChangedFields(oldTree, newTree *proofs.DocumentTree, lengthSuffix string) (changedFields []changedField) {
	oldProps := oldTree.PropertyOrder()
	newProps := newTree.PropertyOrder()

	props := make(map[string]proofs.Property)
	for _, p := range append(oldProps, newProps...) {
		// we can ignore the length property since any change in slice or map will return in addition or deletion of properties in the new tree
		if p.Text == lengthSuffix {
			continue
		}

		if _, ok := props[p.ReadableName()]; ok {
			continue
		}

		props[p.ReadableName()] = p
	}

	// check each property and append it changed fields if the value is different.
	for k, p := range props {
		_, ol := oldTree.GetLeafByProperty(k)
		_, nl := newTree.GetLeafByProperty(k)

		if ol == nil {
			changedFields = append(changedFields, newChangedField(p, nl, false))
			continue
		}

		if nl == nil {
			changedFields = append(changedFields, newChangedField(p, ol, true))
			continue
		}

		ov := ol.Value
		nv := nl.Value
		if ol.Hashed {
			ov = ol.Hash
			nv = nl.Hash
		}

		if !bytes.Equal(ov, nv) {
			changedFields = append(changedFields, changedField{
				property: p.CompactName(),
				old:      ov,
				new:      nv,
			})
		}
	}

	return changedFields
}

func newChangedField(p proofs.Property, leaf *proofs.LeafNode, old bool) changedField {
	v := leaf.Value
	if leaf.Hashed {
		v = leaf.Hash
	}

	cf := changedField{property: p.CompactName()}
	if old {
		cf.old = v
		return cf
	}

	cf.new = v
	return cf
}

// initTransitionRules initiates the transition rules for a given CoreDocumentModel.
// Collaborators are given default edit capability over all fields of the CoreDocument.
// if the rules are created already, this is a no-op.
// if collaborators are empty, it is a no-op
func (cd *CoreDocument) initTransitionRules(collaborators []identity.DID, compactPrefix []byte) {
	if len(cd.Document.Roles) > 0 && len(cd.Document.TransitionRules) > 0 {
		return
	}
	if len(collaborators) < 0 {
		return
	}
	cd.addCollaboratorsToTransitionRules(collaborators, compactPrefix)
}

// addCollaboratorsToTransitionRules adds the given collaborators to a new transition rule which defaults to
// granting edit capability over all fields of the document.
func (cd *CoreDocument) addCollaboratorsToTransitionRules(collaborators []identity.DID, compactPrefix []byte) {
	role := newRoleWithCollaborators(collaborators)
	if role == nil {
		return
	}
	cd.addNewTransitionRule(role, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX, []byte(compactProperties(CDTreePrefix)), coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
	cd.addNewTransitionRule(role, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX, compactPrefix, coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
}

// addNewTransitionRule creates a new transition rule with the given parameters.
func (cd *CoreDocument) addNewTransitionRule(role *coredocumentpb.Role, matchType coredocumentpb.FieldMatchType, field []byte, action coredocumentpb.TransitionAction) {
	cd.Document.Roles = append(cd.Document.Roles, role)
	rule := new(coredocumentpb.TransitionRule)
	rule.RuleKey = utils.RandomSlice(32)
	rule.Roles = append(rule.Roles, role.RoleKey)
	rule.MatchType = matchType
	rule.Action = action
	rule.Field = field
	cd.Document.TransitionRules = append(cd.Document.TransitionRules, rule)
}
