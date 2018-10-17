// common package only contains structs and functions that are shared by all code documents package. It serves as a way to avoid cycling dependencies.

package common

import (
	proto1 "github.com/centrifuge/precise-proofs/proofs/proto"
)

// DocumentProof is a value to represent a document and its field proofs
type DocumentProof struct {
	DocumentId  []byte
	VersionId   []byte
	State       string
	FieldProofs []*proto1.Proof
}
