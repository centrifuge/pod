// +build unit

package coredocument

import (
	"fmt"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

func TestFillIdentifiers(t *testing.T) {
	id1, id2, id3 := tools.RandomSlice32(), tools.RandomSlice32(), tools.RandomSlice32()

	tests := []struct {
		DocIdentifier     []byte
		CurrentIdentifier []byte
		NextIdentifier    []byte
		err               error
	}{
		// all three different identifiers are filled
		{
			DocIdentifier:     id1,
			CurrentIdentifier: id2,
			NextIdentifier:    id3,
		},

		// Doc and current identifiers are same, different next identifier
		{
			DocIdentifier:     id1,
			CurrentIdentifier: id1,
			NextIdentifier:    id3,
		},

		// Doc and current identifiers are same, missing next identifier
		{
			DocIdentifier:     id1,
			CurrentIdentifier: id1,
		},

		// Doc and next identifiers are same, missing current identifier
		{
			DocIdentifier:  id1,
			NextIdentifier: id3,
		},

		// missing current and next identifier
		{
			DocIdentifier: id1,
		},

		// re-used next identifier
		{
			DocIdentifier:     id1,
			CurrentIdentifier: id1,
			NextIdentifier:    id1,
			err:               fmt.Errorf("reusing old identifier"),
		},

		// re-used next identifier with missing current identifier
		{
			DocIdentifier:  id1,
			NextIdentifier: id1,
			err:            fmt.Errorf("reusing old identifier"),
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
