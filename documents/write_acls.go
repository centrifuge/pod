package documents

import (
	"bytes"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/precise-proofs/proofs"
)

// ChangedField holds the compact property, old and new value of the field that is changed
// if the old is nil, then it is a set operation
// if new is nil, then it is an unset operation
// if both old and new are set, then it is an edit operation
type ChangedField struct {
	property, old, new []byte
	name               string
}

// GetChangedFields takes two document trees and returns the compact value, old and new value of the fields that are changed in new tree.
// Properties may have been added to the new tree or removed from the new tree.
// In Either case, since the new tree is different from old, that is considered a change.
func GetChangedFields(oldTree, newTree *proofs.DocumentTree, lengthSuffix string) (changedFields []ChangedField) {
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
			changedFields = append(changedFields, ChangedField{
				name:     k,
				property: p.CompactName(),
				old:      ov,
				new:      nv,
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

	cf := ChangedField{property: p.CompactName(), name: p.ReadableName()}
	if old {
		cf.old = v
		return cf
	}

	cf.new = v
	return cf
}

// TransitionRulesFor returns a copy all the transition rules for the account.
func (cd *CoreDocument) TransitionRulesFor(account identity.DID) (rules []coredocumentpb.TransitionRule) {
	for _, rule := range cd.Document.TransitionRules {
		for _, rk := range rule.Roles {
			role, err := getRole(rk, cd.Document.Roles)
			if err != nil {
				continue
			}

			if _, ok := isAccountInRole(role, account); !ok {
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

func isValidTransition(rule coredocumentpb.TransitionRule, cf ChangedField) bool {
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

// AccountCanUpdate validates the changes made by the account in the new document.
// returns error if the transitions are not allowed for the account
func (cd *CoreDocument) AccountCanUpdate(ncd *CoreDocument, account identity.DID, docType string) error {
	oldTree, err := cd.documentTree(docType)
	if err != nil {
		return err
	}

	newTree, err := ncd.documentTree(docType)
	if err != nil {
		return err
	}

	cf := GetChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	rules := cd.TransitionRulesFor(account)
	return ValidateTransitions(rules, cf)
}
