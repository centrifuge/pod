// +build unit

package coredocument

import (
	"crypto/sha256"
	"fmt"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var (
	id1 = tools.RandomSlice(32)
	id2 = tools.RandomSlice(32)
	id3 = tools.RandomSlice(32)
	id4 = tools.RandomSlice(32)
	id5 = tools.RandomSlice(32)
)

func TestGetDataProofHashes(t *testing.T) {
	cd := coredocumentpb.CoreDocument{
		DataRoot: tools.RandomSlice(32),
	}
	cd, err := FillIdentifiers(cd)
	assert.Nil(t, err)
	cds := &coredocumentpb.CoreDocumentSalts{}
	proofs.FillSalts(&cd, cds)

	cd.CoredocumentSalts = cds

	err = CalculateSigningRoot(&cd)
	assert.Nil(t, err)

	hashes, err := GetDataProofHashes(&cd)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(hashes))

	valid, err := proofs.ValidateProofSortedHashes(cd.DataRoot, hashes, cd.SigningRoot, sha256.New())
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestGetDocumentSigningTree(t *testing.T) {
	cd := coredocumentpb.CoreDocument{DocumentIdentifier: tools.RandomSlice(32)}
	cd, _ = FillIdentifiers(cd)
	cds := &coredocumentpb.CoreDocumentSalts{}
	proofs.FillSalts(&cd, cds)
	cd.CoredocumentSalts = cds
	tree, err := GetDocumentSigningTree(&cd)
	assert.Nil(t, err)
	assert.NotNil(t, tree)
}

func TestGetDocumentRootTree(t *testing.T) {
	cd := &coredocumentpb.CoreDocument{SigningRoot: tools.RandomSlice(32)}
	tree, err := GetDocumentRootTree(cd)
	assert.Nil(t, err)
	fmt.Println(tree.RootHash(), cd.SigningRoot)
	fmt.Println(tree.PropertyOrder())
	p, err := tree.CreateProof("signatures.length")
	fmt.Println(p.SortedHashes)
	// FAIL: why does below tree have the same root hash as signing root?
	assert.Equal(t, tree.RootHash(), cd.SigningRoot)
	assert.False(t, true)
}

func TestValidate(t *testing.T) {
	type want struct {
		valid  bool
		errMsg string
		errs   map[string]string
	}

	tests := []struct {
		doc  *coredocumentpb.CoreDocument
		want want
	}{
		// empty salts in document
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentIdentifier:  id3,
				NextIdentifier:     id4,
				DataRoot:           id5,
			},
			want: want{
				valid:  false,
				errMsg: "Invalid CoreDocument",
				errs: map[string]string{
					"cd_salts": errors.RequiredField,
				},
			},
		},

		// salts missing previous root
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentIdentifier:  id3,
				NextIdentifier:     id4,
				DataRoot:           id5,
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: id1,
					CurrentIdentifier:  id2,
					NextIdentifier:     id3,
					DataRoot:           id4,
				},
			},
			want: want{
				valid:  false,
				errMsg: "Invalid CoreDocument",
				errs: map[string]string{
					"cd_salts": errors.RequiredField,
				},
			},
		},

		// missing identifiers in core document
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentIdentifier:  id3,
				NextIdentifier:     id4,
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: id1,
					CurrentIdentifier:  id2,
					NextIdentifier:     id3,
					DataRoot:           id4,
					PreviousRoot:       id5,
				},
			},
			want: want{
				valid:  false,
				errMsg: "Invalid CoreDocument",
				errs: map[string]string{
					"cd_data_root": errors.RequiredField,
				},
			},
		},

		// missing identifiers in core document and salts
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentIdentifier:  id3,
				NextIdentifier:     id4,
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: id1,
					CurrentIdentifier:  id2,
					NextIdentifier:     id3,
					DataRoot:           id4,
				},
			},
			want: want{
				valid:  false,
				errMsg: "Invalid CoreDocument",
				errs: map[string]string{
					"cd_data_root": errors.RequiredField,
					"cd_salts":     errors.RequiredField,
				},
			},
		},

		// repeated identifiers
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentIdentifier:  id3,
				NextIdentifier:     id3,
				DataRoot:           id5,
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: id1,
					CurrentIdentifier:  id2,
					NextIdentifier:     id3,
					DataRoot:           id4,
					PreviousRoot:       id5,
				},
			},
			want: want{
				valid:  false,
				errMsg: "Invalid CoreDocument",
				errs: map[string]string{
					"cd_overall": errors.IdentifierReUsed,
				},
			},
		},

		// repeated identifiers
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentIdentifier:  id3,
				NextIdentifier:     id2,
				DataRoot:           id5,
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: id1,
					CurrentIdentifier:  id2,
					NextIdentifier:     id3,
					DataRoot:           id4,
					PreviousRoot:       id5,
				},
			},
			want: want{
				valid:  false,
				errMsg: "Invalid CoreDocument",
				errs: map[string]string{
					"cd_overall": errors.IdentifierReUsed,
				},
			},
		},

		// All okay
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentIdentifier:  id3,
				NextIdentifier:     id4,
				DataRoot:           id5,
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: id1,
					CurrentIdentifier:  id2,
					NextIdentifier:     id3,
					DataRoot:           id4,
					PreviousRoot:       id5,
				},
			},
			want: want{
				valid: true,
			},
		},
	}

	for _, c := range tests {
		valid, err, errs := Validate(c.doc)
		got := want{valid, err, errs}
		if !reflect.DeepEqual(c.want, got) {
			t.Fatalf("Mismatch: %v != %v", c.want, got)
		}
	}
}

func TestFillIdentifiers(t *testing.T) {
	tests := []struct {
		DocIdentifier     []byte
		CurrentIdentifier []byte
		NextIdentifier    []byte
		err               error
	}{
		// all three identifiers are filled
		{
			DocIdentifier:     id1,
			CurrentIdentifier: id2,
			NextIdentifier:    id3,
		},

		// Doc and current identifiers are filled, missing next identifier
		{
			DocIdentifier:     id1,
			CurrentIdentifier: id2,
		},

		// Doc and next identifiers are filled, missing current identifier
		{
			DocIdentifier:  id1,
			NextIdentifier: id3,
		},

		// missing current and next identifier
		{
			DocIdentifier: id1,
		},

		// missing doc identifier and filled up current identifier
		{
			CurrentIdentifier: id2,
			err:               fmt.Errorf("no DocumentIdentifier but has CurrentIdentifier"),
		},

		// missing doc identifier and filled up next identifier
		{
			NextIdentifier: id3,
			err:            fmt.Errorf("no CurrentIdentifier but has NextIdentifier"),
		},

		// missing all identifiers
		{},
	}

	for _, c := range tests {
		doc := coredocumentpb.CoreDocument{
			DocumentIdentifier: c.DocIdentifier,
			CurrentIdentifier:  c.CurrentIdentifier,
			NextIdentifier:     c.NextIdentifier,
		}

		var err error
		doc, err = FillIdentifiers(doc)
		if err != nil {
			if c.err == nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assert.EqualError(t, err, c.err.Error())
			continue
		}

		assert.NotNil(t, doc.DocumentIdentifier)
		assert.NotNil(t, doc.CurrentIdentifier)
		assert.NotNil(t, doc.NextIdentifier)
	}
}
