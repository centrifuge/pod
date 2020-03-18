package documents

import (
	"bytes"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// ChangedField holds the compact property, old and new value of the field that is changed
// if the old is nil, then it is a set operation
// if new is nil, then it is an unset operation
// if both old and new are set, then it is an edit operation
type ChangedField struct {
	Property, Old, New []byte
	Name               string
}

// GetChangedFields takes two document trees and returns the compact property, old and new value of the fields that are changed in new tree.
// Properties may have been added to the new tree or removed from the new tree.
// In Either case, since the new tree is different from old, that is considered a change.
func GetChangedFields(oldTree, newTree *proofs.DocumentTree) (changedFields []ChangedField) {
	oldProps := oldTree.PropertyOrder()
	newProps := newTree.PropertyOrder()

	// check each property and append it changed fields if the value is different.
	props := make(map[string]proofs.Property)
	for _, p := range append(oldProps, newProps...) {
		// we can ignore the length property since any change in slice or map will return in addition or deletion of properties in the new tree
		if p.Text == proofs.DefaultReadablePropertyLengthSuffix {
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
			changedFields = append(changedFields, ChangedField{
				Name:     pn,
				Property: p.CompactName(),
				Old:      ov,
				New:      nv,
			})
		}
	}

	return changedFields
}

func newChangedField(p proofs.Property, leaf *proofs.LeafNode, old bool) ChangedField {
	v := leaf.Value
	if leaf.Hashed {
		v = leaf.Hash
	}

	cf := ChangedField{Property: p.CompactName(), Name: p.ReadableName()}
	if old {
		cf.Old = v
		return cf
	}

	cf.New = v
	return cf
}

// TransitionRulesFor returns a copy all the transition rules for the DID.
func (cd *CoreDocument) TransitionRulesFor(did identity.DID) (rules []coredocumentpb.TransitionRule) {
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

// ValidateTransitions validates the changedFields based on the rules provided.
// returns an error if any ChangedField violates the rules.
func ValidateTransitions(rules []coredocumentpb.TransitionRule, changedFields []ChangedField) error {
	cfMap := make(map[string]struct{})
	for _, cf := range changedFields {
		cfMap[cf.Name] = struct{}{}
	}

	for _, rule := range rules {
		for _, cf := range changedFields {
			if isValidTransition(rule, cf) {
				delete(cfMap, cf.Name)
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

func isValidTransition(rule coredocumentpb.TransitionRule, cf ChangedField) bool {
	// changed property length should be at least equal to rule property
	if len(cf.Property) < len(rule.Field) {
		return false
	}

	// if the match type is prefix, get the compact property till prefix
	v := cf.Property
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
	oldTree, err := cd.coredocTree(docType)
	if err != nil {
		return err
	}

	newTree, err := ncd.coredocTree(docType)
	if err != nil {
		return err
	}

	cf := GetChangedFields(oldTree, newTree)
	rules := cd.TransitionRulesFor(collaborator)
	return ValidateTransitions(rules, cf)
}

// initTransitionRules initiates the transition rules for a given Core document.
// Collaborators are given default edit capability over all fields of the CoreDocument and underlying documents such as invoices or purchase orders.
// if the rules are created already, this is a no-op.
// if collaborators are empty, it is a no-op
func (cd *CoreDocument) initTransitionRules(documentPrefix []byte, collaborators []identity.DID) {
	if len(cd.Document.Roles) > 0 && len(cd.Document.TransitionRules) > 0 {
		return
	}
	if len(collaborators) == 0 {
		return
	}
	cd.addCollaboratorsToTransitionRules(documentPrefix, collaborators)
}

// addCollaboratorsToTransitionRules adds the given collaborators to a new transition rule which defaults to
// granting edit capability over all fields of the document.
func (cd *CoreDocument) addCollaboratorsToTransitionRules(documentPrefix []byte, collaborators []identity.DID) {
	role := newRoleWithCollaborators(collaborators...)
	if role == nil {
		return
	}
	cd.Document.Roles = append(cd.Document.Roles, role)
	cd.addNewTransitionRule(role.RoleKey, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX, CompactProperties(CDTreePrefix), coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
	cd.addNewTransitionRule(role.RoleKey, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX, documentPrefix, coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
	cd.Modified = true
}

// addNewTransitionRule creates a new transition rule with the given parameters and returns the rule
func (cd *CoreDocument) addNewTransitionRule(roleKey []byte, matchType coredocumentpb.FieldMatchType, field []byte, action coredocumentpb.TransitionAction) *coredocumentpb.TransitionRule {
	rule := &coredocumentpb.TransitionRule{
		RuleKey:   utils.RandomSlice(32),
		MatchType: matchType,
		Action:    action,
		Field:     field,
		Roles:     [][]byte{roleKey},
	}
	cd.Document.TransitionRules = append(cd.Document.TransitionRules, rule)
	cd.Modified = true
	return rule
}

// getAttributeFieldPrefix creates a compact property of the attribute key
func getAttributeFieldPrefix(key AttrKey) []byte {
	attrPrefix := append(CompactProperties(CDTreePrefix), []byte{0, 0, 0, 28}...)
	return append(attrPrefix, key[:]...)
}

// defaultRuleFieldProps are the fields that every collaborator should have rule set for to update a document.
func defaultRuleFieldProps() map[string][]byte {
	fields := [][]byte{
		{0, 0, 0, 3},  // current_version
		{0, 0, 0, 4},  // next_version
		{0, 0, 0, 16}, // previous_version
		{0, 0, 0, 22}, // next_preimage
		{0, 0, 0, 23}, // current_preimage
		{0, 0, 0, 25}, // author
		{0, 0, 0, 26}, // timestamp
	}

	fieldMap := make(map[string][]byte)
	for _, f := range fields {
		f := f
		cp := append(CompactProperties(CDTreePrefix), f...)
		fieldMap[hexutil.Encode(cp)] = cp
	}
	return fieldMap
}

// deleteFieldIfRoleExists checks if the role exists in the rule that has a field in the field map.
// will update the fieldMap by deleting fields that already has the role
func deleteFieldIfRoleExists(rule *coredocumentpb.TransitionRule, role []byte, fieldMap map[string][]byte) bool {
	field := hexutil.Encode(rule.Field)
	if rule.MatchType != coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_EXACT {
		// default field rules are exact match
		return false
	}

	if _, ok := fieldMap[field]; !ok {
		// not a match
		return false
	}

	// delete the field from the map since the role is already present or we are going to add one to rule
	delete(fieldMap, field)
	if byteutils.ContainsBytesInSlice(rule.Roles, role) {
		// rule already exists for the role
		return false
	}
	return true
}

// addDefaultRules will update all default rules to include rolekey so that the document can be updated successfully
// Note: assumes that role exists in the document already
func (cd *CoreDocument) addDefaultRules(roleKey []byte) {
	fieldMap := defaultRuleFieldProps()
	for _, rule := range cd.Document.TransitionRules {
		if !deleteFieldIfRoleExists(rule, roleKey, fieldMap) {
			continue
		}

		rule.Roles = append(rule.Roles, roleKey)
		cd.Modified = true
	}

	if len(fieldMap) < 1 {
		// all fields are added
		return
	}

	for _, f := range fieldMap {
		cd.addNewTransitionRule(
			roleKey,
			coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_EXACT,
			f,
			coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
	}
}

// AddTransitionRuleForAttribute adds a new rule with key as fields for the role
// FieldMatchType_FIELD_MATCH_TYPE_PREFIX will be used for the Field match for attributes
// TransitionAction_TRANSITION_ACTION_EDIT is the default action we assign to the rule.
// Role must be present to create a rule.
func (cd *CoreDocument) AddTransitionRuleForAttribute(roleID []byte, key AttrKey) (*coredocumentpb.TransitionRule, error) {
	_, err := cd.GetRole(roleID)
	if err != nil {
		return nil, err
	}

	cd.addDefaultRules(roleID)
	return cd.addNewTransitionRule(
		roleID,
		coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX,
		getAttributeFieldPrefix(key),
		coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT), nil
}

// GetTransitionRule returns the transition rule associated with ruleID in the document.
func (cd *CoreDocument) GetTransitionRule(ruleID []byte) (*coredocumentpb.TransitionRule, error) {
	for _, r := range cd.Document.TransitionRules {
		if bytes.Equal(r.RuleKey, ruleID) {
			return r, nil
		}
	}

	return nil, ErrTransitionRuleMissing
}
