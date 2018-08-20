package coredocument

import (
	"reflect"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
)

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
				DocumentRoot:       tools.RandomSlice32(),
				DocumentIdentifier: tools.RandomSlice32(),
				CurrentIdentifier:  tools.RandomSlice32(),
				NextIdentifier:     tools.RandomSlice32(),
				DataRoot:           tools.RandomSlice32(),
			},
			want: want{
				valid:  false,
				errMsg: "Invalid CoreDocument",
				errs: map[string]string{
					"cd_salts": "Empty Document salts",
				},
			},
		},

		// salts missing previous root
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       tools.RandomSlice32(),
				DocumentIdentifier: tools.RandomSlice32(),
				CurrentIdentifier:  tools.RandomSlice32(),
				NextIdentifier:     tools.RandomSlice32(),
				DataRoot:           tools.RandomSlice32(),
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: tools.RandomSlice32(),
					CurrentIdentifier:  tools.RandomSlice32(),
					NextIdentifier:     tools.RandomSlice32(),
					DataRoot:           tools.RandomSlice32(),
				},
			},
			want: want{
				valid:  false,
				errMsg: "Invalid CoreDocument",
				errs: map[string]string{
					"cd_salts": "Empty Document salts",
				},
			},
		},

		// missing identifiers in core document
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       tools.RandomSlice32(),
				DocumentIdentifier: tools.RandomSlice32(),
				CurrentIdentifier:  tools.RandomSlice32(),
				NextIdentifier:     tools.RandomSlice32(),
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: tools.RandomSlice32(),
					CurrentIdentifier:  tools.RandomSlice32(),
					NextIdentifier:     tools.RandomSlice32(),
					DataRoot:           tools.RandomSlice32(),
					PreviousRoot:       tools.RandomSlice32(),
				},
			},
			want: want{
				valid:  false,
				errMsg: "Invalid CoreDocument",
				errs: map[string]string{
					"cd_data_root": "Empty Document Data Root",
				},
			},
		},

		// missing identifiers in core document and salts
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       tools.RandomSlice32(),
				DocumentIdentifier: tools.RandomSlice32(),
				CurrentIdentifier:  tools.RandomSlice32(),
				NextIdentifier:     tools.RandomSlice32(),
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: tools.RandomSlice32(),
					CurrentIdentifier:  tools.RandomSlice32(),
					NextIdentifier:     tools.RandomSlice32(),
					DataRoot:           tools.RandomSlice32(),
				},
			},
			want: want{
				valid:  false,
				errMsg: "Invalid CoreDocument",
				errs: map[string]string{
					"cd_data_root": "Empty Document Data Root",
					"cd_salts":     "Empty Document salts",
				},
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
