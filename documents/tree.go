package documents

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"github.com/centrifuge/go-centrifuge/crypto/pedersen"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
)

// DefaultTree returns a DocumentTree with default opts
func (cd *CoreDocument) DefaultTree() (*proofs.DocumentTree, error) {
	return cd.DefaultTreeWithPrefix("", nil)
}

// DefaultZTree returns a DocumentTree with support for skSnarks calculations
func (cd *CoreDocument) DefaultZTree() (*proofs.DocumentTree, error) {
	t, err := proofs.NewDocumentTree(proofs.TreeOptions{
		CompactProperties: true,
		Hash:              pedersen.New(),
		LeafHash:          sha256.New(),
		Salts:             cd.DocumentSaltsFunc(),
		TreeDepth:         7,
	})
	return &t, err
}

// DefaultTreeWithPrefix returns a DocumentTree with default opts passing a prefix to the tree leaves
func (cd *CoreDocument) DefaultTreeWithPrefix(prefix string, compactPrefix []byte) (*proofs.DocumentTree, error) {
	var prop proofs.Property
	if prefix != "" {
		prop = NewLeafProperty(prefix, compactPrefix)
	}

	t, err := proofs.NewDocumentTree(proofs.TreeOptions{
		CompactProperties: true,
		EnableHashSorting: true,
		Hash:              sha256.New(),
		ParentPrefix:      prop,
		Salts:             cd.DocumentSaltsFunc(),
	})
	return &t, err
}

// NewLeafProperty returns a proof property with the literal and the compact
func NewLeafProperty(literal string, compact []byte) proofs.Property {
	return proofs.NewProperty(literal, compact...)
}

// DocumentSaltsFunc returns a function that fetches and sets salts on the CoreDoc. The boolean `cd.Modified` can be used to define if the salts function should error if a new field is encountered or not.
func (cd *CoreDocument) DocumentSaltsFunc() func(compact []byte) ([]byte, error) {
	salts := cd.Document.Salts
	return func(compact []byte) ([]byte, error) {
		for _, salt := range salts {
			if bytes.Compare(salt.GetCompact(), compact) == 0 {
				return salt.GetValue(), nil
			}
		}

		if !cd.Modified {
			return nil, errors.New("Salt for property %v not found", compact)
		}

		randbytes := make([]byte, 32)
		n, err := rand.Read(randbytes)
		if err != nil {
			return nil, err
		}
		if n != 32 {
			return nil, errors.AppendError(err, errors.New("Only read %d instead of 32 random bytes", n))
		}
		salt := proofspb.Salt{
			Compact: compact,
			Value:   randbytes,
		}

		salts = append(salts, &salt)
		cd.Document.Salts = salts
		return randbytes, nil
	}
}
