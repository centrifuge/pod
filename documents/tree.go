package documents

import (
	"crypto/sha256"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/gogo/protobuf/proto"
)

// NewDefaultTree returns a DocumentTree with default opts
func NewDefaultTree(salts *proofs.Salts) *proofs.DocumentTree {
	return NewDefaultTreeWithPrefix(salts, "", nil)
}

// NewDefaultTreeWithPrefix returns a DocumentTree with default opts passing a prefix to the tree leaves
func NewDefaultTreeWithPrefix(salts *proofs.Salts, prefix string, compactPrefix []byte) *proofs.DocumentTree {
	var prop proofs.Property
	if prefix != "" {
		prop = NewLeafProperty(prefix, compactPrefix)
	}

	t := proofs.NewDocumentTree(proofs.TreeOptions{CompactProperties: true, EnableHashSorting: true, Hash: sha256.New(), ParentPrefix: prop, Salts: salts})
	return &t
}

// NewLeafProperty returns a proof property with the literal and the compact
func NewLeafProperty(literal string, compact []byte) proofs.Property {
	return proofs.NewProperty(literal, compact...)
}

// GenerateNewSalts generates salts for new document
func GenerateNewSalts(document proto.Message, prefix string, compactPrefix []byte) (*proofs.Salts, error) {
	docSalts := new(proofs.Salts)
	t := NewDefaultTreeWithPrefix(docSalts, prefix, compactPrefix)
	err := t.AddLeavesFromDocument(document)
	if err != nil {
		return nil, err
	}
	return docSalts, nil
}

// ConvertToProtoSalts converts proofSalts into protocolSalts
func ConvertToProtoSalts(proofSalts *proofs.Salts) []*coredocumentpb.DocumentSalt {
	if proofSalts == nil {
		return nil
	}

	protoSalts := make([]*coredocumentpb.DocumentSalt, len(*proofSalts))
	for i, pSalt := range *proofSalts {
		protoSalts[i] = &coredocumentpb.DocumentSalt{Value: pSalt.Value, Compact: pSalt.Compact}
	}

	return protoSalts
}

// ConvertToProofSalts converts protocolSalts into proofSalts
func ConvertToProofSalts(protoSalts []*coredocumentpb.DocumentSalt) *proofs.Salts {
	if protoSalts == nil {
		return nil
	}

	proofSalts := make(proofs.Salts, len(protoSalts))
	for i, pSalt := range protoSalts {
		proofSalts[i] = proofs.Salt{Value: pSalt.Value, Compact: pSalt.Compact}
	}

	return &proofSalts
}
