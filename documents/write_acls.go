package documents

import (
	"bytes"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
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
	name               string
}

// getChangedFields takes two document trees and returns the compact value, old and new value of the fields that are changed in new tree.
// Properties may have been added to the new tree or removed from the new tree.
// In Either case, since the new tree is different from old, that is considered a change.
func getChangedFields(oldTree, newTree *proofs.DocumentTree, lengthSuffix string) (changedFields []changedField) {
	oldProps := oldTree.PropertyOrder()
	newProps := newTree.PropertyOrder()

	// check each property and append it changed fields if the value is different.
	props := make(map[string]proofs.Property)
	for _, p := range append(oldProps, newProps...) {
		// we can ignore the length property since any change in slice or map will return in addition or deletion of properties in the new tree
		if p.Text == lengthSuffix {
			continue
		}

		pn := p.ReadableName()
		if _, ok := props[pn]; ok {
			continue
		}

		props[pn] = p
		_, ol := oldTree.GetLeafByProperty(pn)
		_, nl := newTree.GetLeafByProperty(pn)

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
				name:     pn,
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

	cf := changedField{property: p.CompactName(), name: p.ReadableName()}
	if old {
		cf.old = v
		return cf
	}

	cf.new = v
	return cf
}

// transitionRulesFor returns a copy all the transition rules for the DID.
func (cd *CoreDocument) transitionRulesFor(did identity.DID) (rules []coredocumentpb.TransitionRule) {
	for _, rule := range cd.Document.TransitionRules {
		for _, rk := range rule.Roles {
			role, err := getRole(rk, cd.Document.Roles)
			if err != nil {
				continue
			}

			if _, ok := isDIDInRole(role, did); !ok {
				continue
			}

			rules = append(rules, coredocumentpb.TransitionRule{
				RuleKey:   copyBytes(rule.RuleKey),
				Roles:     copyByteSlice(rule.Roles),
				MatchType: rule.MatchType,
				Field:     copyBytes(rule.Field),
				Action:    rule.Action,
			})
		}
	}

	return rules
}

func copyBytes(data []byte) []byte {
	if data == nil {
		return nil
	}

	nb := make([]byte, len(data), len(data))
	copy(nb, data)
	return nb
}

func copyByteSlice(data [][]byte) [][]byte {
	nbs := make([][]byte, len(data), len(data))
	for i, b := range data {
		nbs[i] = copyBytes(b)
	}

	return nbs
}

// validateTransitions validates the changedFields based on the rules provided.
// returns an error if any changedField violates the rules.
func validateTransitions(rules []coredocumentpb.TransitionRule, changedFields []changedField) error {
	cfMap := make(map[string]struct{})
	for _, cf := range changedFields {
		cfMap[cf.name] = struct{}{}
	}

	for _, rule := range rules {
		for _, cf := range changedFields {
			if isValidTransition(rule, cf) {
				delete(cfMap, cf.name)
			}
		}
	}

	if len(cfMap) == 0 {
		return nil
	}

	var err error
	for k := range cfMap {
		err = errors.AppendError(err, errors.New("invalid transition: %s", k))
	}

	return err
}

func isValidTransition(rule coredocumentpb.TransitionRule, cf changedField) bool {
	// changed property length should be at least equal to rule property
	if len(cf.property) < len(rule.Field) {
		return false
	}

	// if the match type is prefix, get the compact property till prefix
	v := cf.property
	if rule.MatchType == coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX {
		v = v[:len(rule.Field)]
	}

	// check the properties are equal
	if !bytes.Equal(v, rule.Field) {
		return false
	}

	// check if the action is allowed
	// for now, we have only edit action
	// edit allows following
	// 1. update: editing a value
	// 2. set: setting a new value ex: adding to slice or map
	// 3. delete: deleting the new value ex: removing from slice or map
	// Once we have more actions, like set, increment etc.. we can do those checks here
	return true
}

// CollaboratorCanUpdate validates the changes made by the collaborator in the new document.
// returns error if the transitions are not allowed for the collaborator.
func (cd *CoreDocument) CollaboratorCanUpdate(ncd *CoreDocument, collaborator identity.DID, docType string) error {
	oldTree, err := cd.documentTree(docType)
	if err != nil {
		return err
	}

	newTree, err := ncd.documentTree(docType)
	if err != nil {
		return err
	}

	cf := getChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	rules := cd.transitionRulesFor(collaborator)
	return validateTransitions(rules, cf)
}

// initTransitionRules initiates the transition rules for a given Core Document.
// Collaborators are given default edit capability over all fields of the CoreDocument and underlying documents such as invoices or purchase orders.
// if the rules are created already, this is a no-op.
// if collaborators are empty, it is a no-op
func (cd *CoreDocument) initTransitionRules(collaborators []identity.DID, documentPrefix []byte) {
	if len(cd.Document.Roles) > 0 && len(cd.Document.TransitionRules) > 0 {
		return
	}
	if len(collaborators) < 0 {
		return
	}
	cd.addCollaboratorsToTransitionRules(collaborators, documentPrefix)
}

// addCollaboratorsToTransitionRules adds the given collaborators to a new transition rule which defaults to
// granting edit capability over all fields of the document.
func (cd *CoreDocument) addCollaboratorsToTransitionRules(collaborators []identity.DID, documentPrefix []byte) {
	role := newRoleWithCollaborators(collaborators)
	if role == nil {
		return
	}
	cd.Document.Roles = append(cd.Document.Roles, role)
	cd.addNewTransitionRule(role.RoleKey, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX, compactProperties(CDTreePrefix), coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
	cd.addNewTransitionRule(role.RoleKey, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX, documentPrefix, coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
}

// addNewTransitionRule creates a new transition rule with the given parameters.
func (cd *CoreDocument) addNewTransitionRule(roleKey []byte, matchType coredocumentpb.FieldMatchType, field []byte, action coredocumentpb.TransitionAction) {
	rule := &coredocumentpb.TransitionRule{
		RuleKey:   utils.RandomSlice(32),
		MatchType: matchType,
		Action:    action,
		Field:     field,
		Roles:     [][]byte{roleKey},
	}
	cd.Document.TransitionRules = append(cd.Document.TransitionRules, rule)
}
