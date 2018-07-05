// +build unit

package coredocumentservice

import (
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"os"
	"testing"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/stretchr/testify/assert"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
)

func TestMain(m *testing.M) {
	cc.TestUnitBootstrap()
	result := m.Run()
	cc.TestTearDown()
	os.Exit(result)
}

// ----- TESTS -----
var documentIdentifierPrefillCheckData = []struct {
	cd       coredocumentpb.CoreDocument
	expected error // expected result
}{
	{
		coredocumentpb.CoreDocument{
			CurrentIdentifier: []byte("abc1"),
		},
		coredocument.NewErrInconsistentState("No DocumentIdentifier but has CurrentIdentifier"),
	},
	{
		coredocumentpb.CoreDocument{
			CurrentIdentifier: []byte("abc1"),
			NextIdentifier:    []byte("abc2"),
		},
		coredocument.NewErrInconsistentState("No DocumentIdentifier but has CurrentIdentifier"),
	},
	{
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
			NextIdentifier:     []byte("abc2"),
		},
		coredocument.NewErrInconsistentState("No CurrentIdentifier but has NextIdentifier"),
	},
	{
		coredocumentpb.CoreDocument{
			NextIdentifier: []byte("abc2"),
		},
		coredocument.NewErrInconsistentState("No CurrentIdentifier but has NextIdentifier"),
	},
	{
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
			CurrentIdentifier:  []byte("abc2"),
			NextIdentifier:     []byte("abc1"),
		},
		coredocument.NewErrInconsistentState("Reusing old Identifier"),
	},
	{
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
			CurrentIdentifier:  []byte("abc2"),
			NextIdentifier:     []byte("abc2"),
		},
		coredocument.NewErrInconsistentState("Reusing old Identifier"),
	},
	{
		coredocumentpb.CoreDocument{
		},
		nil,
	},
	{
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
		},
		nil,
	},
	{
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
			CurrentIdentifier:  []byte("abc1"),
		},
		nil,
	},
	{
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
			CurrentIdentifier:  []byte("abc1"),
			NextIdentifier:     []byte("abc2"),
		},
		nil,
	},
	{
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
			CurrentIdentifier:  []byte("abc2"),
			NextIdentifier:     []byte("abc3"),
		},
		nil,
	},
}

func TestCheckForValidDocumentIdentifiers(t *testing.T) {
	for _, tt := range documentIdentifierPrefillCheckData {
		actual := AutoFillDocumentIdentifiers(tt.cd)
		if (tt.expected == nil) {
			assert.Nilf(t, actual,"Expected nil but got [%v] for test [%v]", actual, tt)
		} else {
			assert.NotNilf(t, actual, "Error when executing test %v", tt)
			assert.Equal(t, tt.expected.Error(), actual.Error())
		}
	}
}

// ----- END TESTS -----
