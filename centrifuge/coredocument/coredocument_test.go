// +build unit

package coredocument

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/signatures"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
)

var (
	id1 = tools.RandomSlice(32)
	id2 = tools.RandomSlice(32)
	id3 = tools.RandomSlice(32)
	id4 = tools.RandomSlice(32)
	id5 = tools.RandomSlice(32)

	centID  = tools.RandomSlice(identity.CentIDLength)
	key1Pub = [...]byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1    = []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
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

	err = CalculateDocumentRoot(&cd)
	assert.Nil(t, err)

	hashes, err := GetDataProofHashes(&cd)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(hashes))

	valid, err := proofs.ValidateProofSortedHashes(cd.DataRoot, hashes, cd.DocumentRoot, sha256.New())
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

// TestGetDocumentRootTree tests that the documentroottree is properly calculated
func TestGetDocumentRootTree(t *testing.T) {
	cd := &coredocumentpb.CoreDocument{SigningRoot: []byte{0x72, 0xee, 0xb8, 0x88, 0x92, 0xf7, 0x6, 0x19, 0x82, 0x76, 0xe9, 0xe7, 0xfe, 0xcc, 0x33, 0xa, 0x66, 0x78, 0xd4, 0xa6, 0x5f, 0xf6, 0xa, 0xca, 0x2b, 0xe4, 0x17, 0xa9, 0xf6, 0x15, 0x67, 0xa1}}
	tree, err := GetDocumentRootTree(cd)

	// Manually constructing the two node tree:
	signaturesLengthLeaf := sha256.Sum256(append([]byte("signatures.length0"), make([]byte, 32)...))
	expectedRootHash := sha256.Sum256(append(signaturesLengthLeaf[:], cd.SigningRoot...))
	assert.Nil(t, err)
	assert.Equal(t, expectedRootHash[:], tree.RootHash())
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
					"cd_salts": centerrors.RequiredField,
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
					"cd_salts": centerrors.RequiredField,
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
					"cd_data_root": centerrors.RequiredField,
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
					"cd_data_root": centerrors.RequiredField,
					"cd_salts":     centerrors.RequiredField,
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
					"cd_overall": centerrors.IdentifierReUsed,
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
					"cd_overall": centerrors.IdentifierReUsed,
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

func TestValidateWithSignature_fail_basic_check(t *testing.T) {
	doc := &coredocumentpb.CoreDocument{
		DocumentRoot:       id1,
		DocumentIdentifier: id2,
		CurrentIdentifier:  id3,
		NextIdentifier:     id4,
		DataRoot:           id5,
	}

	err := ValidateWithSignature(doc)
	assert.NotNil(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "Invalid CoreDocument")
}

func TestValidateWithSignature_fail_missing_signing_root(t *testing.T) {
	doc := testingutils.GenerateCoreDocument()
	err := ValidateWithSignature(doc)
	assert.Error(t, err, "must be a non nil error")
	assert.Contains(t, err.Error(), "signing_root is missing")
}

func TestValidateWithSignature_fail_mismatch_signing_root(t *testing.T) {
	doc := testingutils.GenerateCoreDocument()
	doc.SigningRoot = tools.RandomSlice(32)
	err := ValidateWithSignature(doc)
	assert.Error(t, err, "signing root must mismatch")
	assert.Contains(t, err.Error(), "signing_root mismatch")
}

func TestValidateWithSignature_failed_signature_verification(t *testing.T) {
	sig := &coredocumentpb.Signature{
		EntityId:  centID,
		PublicKey: key1Pub[:],
		Signature: tools.RandomSlice(32)}
	srv := &testingcommons.MockIDService{}
	centID, _ := identity.ToCentID(sig.EntityId)
	srv.On("LookupIdentityForID", centID).Return(nil, fmt.Errorf("failed GetIdentity")).Once()
	identity.IDService = srv
	doc := testingutils.GenerateCoreDocument()
	tree, _ := GetDocumentSigningTree(doc)
	doc.SigningRoot = tree.RootHash()
	doc.Signatures = append(doc.Signatures, sig)
	err := ValidateWithSignature(doc)
	srv.AssertExpectations(t)
	assert.NotNil(t, err, "must be not nil")
	assert.Contains(t, err.Error(), "failed GetIdentity")
}

func TestValidateWithSignature_successful_verification(t *testing.T) {
	sig := &coredocumentpb.Signature{
		EntityId:  centID,
		PublicKey: key1Pub[:],
	}
	centID, _ := identity.ToCentID(sig.EntityId)
	idkey := &identity.EthereumIdentityKey{
		Key:       key1Pub,
		Purposes:  []*big.Int{big.NewInt(identity.KeyPurposeSigning)},
		RevokedAt: big.NewInt(0),
	}
	id := &testingcommons.MockID{}
	srv := &testingcommons.MockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", key1Pub[:]).Return(idkey, nil).Once()
	identity.IDService = srv
	doc := testingutils.GenerateCoreDocument()
	tree, _ := GetDocumentSigningTree(doc)
	doc.SigningRoot = tree.RootHash()
	sig = signatures.Sign(&config.IdentityConfig{
		ID:         sig.EntityId,
		PublicKey:  key1Pub[:],
		PrivateKey: key1,
	}, doc.SigningRoot)
	doc.Signatures = append(doc.Signatures, sig)
	err := ValidateWithSignature(doc)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
}

func TestGetTypeUrl(t *testing.T) {
	coreDocument := testingutils.GenerateCoreDocument()

	documentType, err := GetTypeUrl(coreDocument)
	assert.Nil(t, err, "should not throw an error because coreDocument has a type")
	assert.NotEqual(t, documentType, "", "document type shouldn't be empty")

	_, err = GetTypeUrl(nil)
	assert.Error(t, err, "nil should throw an error")

	coreDocument.EmbeddedData.TypeUrl = ""
	_, err = GetTypeUrl(nil)
	assert.Error(t, err, "should throw an error because typeUrl is not set")
}
