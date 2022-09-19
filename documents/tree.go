package documents

import (
	"bytes"
	"crypto/rand"
	"hash"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/precise-proofs/proofs"
	proofspb "github.com/centrifuge/precise-proofs/proofs/proto"
	"golang.org/x/crypto/blake2b"
)

func (cd *CoreDocument) defaultTreeWithPrefix(prefix string, compactPrefix []byte, hashSorting bool) (*proofs.DocumentTree, error) {
	var prop proofs.Property
	if prefix != "" {
		prop = NewLeafProperty(prefix, compactPrefix)
	}

	b2bHash, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}

	t, err := proofs.NewDocumentTree(proofs.TreeOptions{
		CompactProperties: true,
		EnableHashSorting: hashSorting,
		Hash:              b2bHash,
		LeafHash:          b2bHash,
		ParentPrefix:      prop,
		Salts:             cd.DocumentSaltsFunc(),
	})
	return &t, err
}

// DefaultTreeWithPrefix returns a DocumentTree with default opts, sorted hashing enabled and passing a prefix to the tree leaves
func (cd *CoreDocument) DefaultTreeWithPrefix(prefix string, compactPrefix []byte) (*proofs.DocumentTree, error) {
	return cd.defaultTreeWithPrefix(prefix, compactPrefix, true)
}

// DefaultOrderedTreeWithPrefix returns a DocumentTree with default opts, sorted hashing disabled and passing a prefix to the tree leaves
func (cd *CoreDocument) DefaultOrderedTreeWithPrefix(prefix string, compactPrefix []byte) (*proofs.DocumentTree, error) {
	return cd.defaultTreeWithPrefix(prefix, compactPrefix, false)
}

// DocumentSaltsFunc returns a function that fetches and sets salts on the CoreDoc. The boolean `cd.Modified` can be used to define if the salts function should error if a new field is encountered or not.
func (cd *CoreDocument) DocumentSaltsFunc() func(compact []byte) ([]byte, error) {
	salts := cd.Document.Salts
	return func(compact []byte) ([]byte, error) {
		for _, salt := range salts {
			if bytes.Equal(salt.GetCompact(), compact) {
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

// NewLeafProperty returns a proof property with the literal and the compact
func NewLeafProperty(literal string, compact []byte) proofs.Property {
	return proofs.NewProperty(literal, compact...)
}

// ValidateProof by comparing it to the provided rootHash
func ValidateProof(proof *proofspb.Proof, rootHash []byte, hashFunc hash.Hash, leafHashFunc hash.Hash) (valid bool, err error) {
	var fieldHash []byte
	if len(proof.Hash) == 0 {
		fieldHash, err = proofs.CalculateHashForProofField(proof, leafHashFunc)
	} else {
		fieldHash = proof.Hash
	}
	if err != nil {
		return false, err
	}
	if len(proof.SortedHashes) > 0 {
		valid, err = proofs.ValidateProofSortedHashes(fieldHash, proof.SortedHashes, rootHash, hashFunc)
	} else {
		valid, err = proofs.ValidateProofHashes(fieldHash, proof.Hashes, rootHash, hashFunc)
	}
	return valid, err
}
