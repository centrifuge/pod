package documents

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/pkg/errors"
)

// NewDefaultTree returns a DocumentTree with default opts
func NewDefaultTree(coredoc *coredocumentpb.CoreDocument, mutable bool) *proofs.DocumentTree {
	return NewDefaultTreeWithPrefix(coredoc, "", nil, mutable)
}

// NewDefaultTreeWithPrefix returns a DocumentTree with default opts passing a prefix to the tree leaves
func NewDefaultTreeWithPrefix(coredoc *coredocumentpb.CoreDocument, prefix string, compactPrefix []byte, mutable bool) *proofs.DocumentTree {
	var prop proofs.Property
	if prefix != "" {
		prop = NewLeafProperty(prefix, compactPrefix)
	}
	salts, err := DocumentSaltsFunc(coredoc, mutable)
	if err != nil {
		// TODO: return error
		return nil
	}
	t := proofs.NewDocumentTree(proofs.TreeOptions{
		CompactProperties: true,
		EnableHashSorting: true,
		Hash:              sha256.New(),
		ParentPrefix:      prop,
		Salts:             salts,
	})
	return &t
}

// NewLeafProperty returns a proof property with the literal and the compact
func NewLeafProperty(literal string, compact []byte) proofs.Property {
	return proofs.NewProperty(literal, compact...)
}

// DocumentSaltsFunc returns a function that fetches and sets salts on the CoreDoc. The boolean mutable can be used to define if the salts function should error if a new field is encountered or not.
func DocumentSaltsFunc(coredoc *coredocumentpb.CoreDocument, mutable bool) (func(compact []byte) ([]byte, error), error) {
	salts := coredoc.Salts
	return func(compact []byte) ([]byte, error) {
		for _, salt := range salts {
			if bytes.Compare(salt.GetCompact(), compact) == 0 {
				return salt.GetValue(), nil
			}
		}
		if !mutable {
			return nil, fmt.Errorf("Salt for property %v not found", compact)
		}
		randbytes := make([]byte, 32)
		n, err := rand.Read(randbytes)
		if err != nil {
			return nil, err
		} else if n != 32 {
			return nil, errors.Wrapf(err, "Only read %d instead of 32 random bytes", n)
		}
		salt := proofspb.Salt{
			Compact: compact,
			Value:   randbytes,
		}

		salts = append(salts, &salt)
		coredoc.Salts = salts
		return randbytes, nil
	}, nil
}
