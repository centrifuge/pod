// +build unit

package coredocument

import (
	"reflect"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
)

func TestValidate(t *testing.T) {
	type want struct {
		valid  bool
		errMsg string
		errs   map[string]string
	}

	var (
		id1 = tools.RandomSlice(32)
		id2 = tools.RandomSlice(32)
		id3 = tools.RandomSlice(32)
		id4 = tools.RandomSlice(32)
		id5 = tools.RandomSlice(32)
	)

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
