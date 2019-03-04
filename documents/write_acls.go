package documents

import (
	"bytes"

	"github.com/centrifuge/precise-proofs/proofs"
)

// getChangedFields takes two document trees and returns the compact value of the fields that are changed in new tree.
// properties may have been added to the new tree or removed from the new tree. Either case, since the new tree is different from old,
// that is considered a change
func getChangedFields(oldTree, newTree *proofs.DocumentTree, lengthSuffix string) (changedFields [][]byte) {
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
	for _, p := range props {
		_, ol := oldTree.GetLeafByProperty(p.ReadableName())
		_, nl := newTree.GetLeafByProperty(p.ReadableName())

		if ol == nil || nl == nil {
			// this is the property from new tree, add it changed fields
			changedFields = append(changedFields, p.CompactName())
			continue
		}

		ov := ol.Value
		nv := nl.Value
		if ol.Hashed {
			ov = ol.Hash
			nv = nl.Hash
		}

		if !bytes.Equal(ov, nv) {
			changedFields = append(changedFields, p.CompactName())
		}
	}

	return changedFields
}
