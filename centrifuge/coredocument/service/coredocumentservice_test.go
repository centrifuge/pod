// +build unit

package coredocumentservice

import (
	"os"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

// ----- TESTS -----
var documentIdentifierPrefillCheckData = []struct {
	cd          coredocumentpb.CoreDocument
	expectedDoc coredocumentpb.CoreDocument
	expected    error // expected result
}{
	{
		coredocumentpb.CoreDocument{
			CurrentIdentifier: []byte("abc1"),
		},
		coredocumentpb.CoreDocument{},
		coredocument.NewErrInconsistentState("No DocumentIdentifier but has CurrentIdentifier"),
	},
	{
		coredocumentpb.CoreDocument{
			CurrentIdentifier: []byte("abc1"),
			NextIdentifier:    []byte("abc2"),
		},
		coredocumentpb.CoreDocument{},
		coredocument.NewErrInconsistentState("No DocumentIdentifier but has CurrentIdentifier"),
	},
	{
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
			NextIdentifier:     []byte("abc2"),
		},
		coredocumentpb.CoreDocument{},
		coredocument.NewErrInconsistentState("No CurrentIdentifier but has NextIdentifier"),
	},
	{
		coredocumentpb.CoreDocument{
			NextIdentifier: []byte("abc2"),
		},
		coredocumentpb.CoreDocument{},
		coredocument.NewErrInconsistentState("No CurrentIdentifier but has NextIdentifier"),
	},
	{
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
			CurrentIdentifier:  []byte("abc2"),
			NextIdentifier:     []byte("abc1"),
		},
		coredocumentpb.CoreDocument{},
		coredocument.NewErrInconsistentState("Reusing old Identifier"),
	},
	{
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
			CurrentIdentifier:  []byte("abc2"),
			NextIdentifier:     []byte("abc2"),
		},
		coredocumentpb.CoreDocument{},
		coredocument.NewErrInconsistentState("Reusing old Identifier"),
	},
	{
		coredocumentpb.CoreDocument{},
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("should_be_filled"),
			CurrentIdentifier:  []byte("should_be_filled"),
			NextIdentifier:     []byte("should_be_filled"),
		},
		nil,
	},
	{
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
		},
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
			CurrentIdentifier:  []byte("abc1"),
			NextIdentifier:     []byte("should_be_filled"),
		},
		nil,
	},
	{
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
			CurrentIdentifier:  []byte("abc1"),
		},
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
			CurrentIdentifier:  []byte("abc1"),
			NextIdentifier:     []byte("should_be_filled"),
		},
		nil,
	},
	{
		coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte("abc1"),
			CurrentIdentifier:  []byte("abc1"),
			NextIdentifier:     []byte("abc2"),
		},
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
		actualDoc, actual := AutoFillDocumentIdentifiers(tt.cd)
		if tt.expected == nil {
			assert.Nilf(t, actual, "Expected nil but got [%v] for test [%v]", actual, tt)
			expectedDoc := coredocumentpb.CoreDocument(tt.expectedDoc)

			assert.NotNilf(t, actualDoc.DocumentIdentifier, "Expected [%v] but got [%v]. For test [%v]", expectedDoc.DocumentIdentifier, actualDoc.DocumentIdentifier, tt)
			assert.NotNilf(t, actualDoc.CurrentIdentifier, "Expected [%v] but got [%v]. For test [%v]", expectedDoc.CurrentIdentifier, actualDoc.CurrentIdentifier, tt)
			assert.NotNilf(t, actualDoc.NextIdentifier, "Expected [%v] but got [%v]. For test [%v]", expectedDoc.NextIdentifier, actualDoc.NextIdentifier, tt)
			if !tools.IsSameByteSlice([]byte("should_be_filled"), expectedDoc.DocumentIdentifier) {
				assert.ElementsMatch(t, expectedDoc.DocumentIdentifier, actualDoc.DocumentIdentifier)
			}
			if !tools.IsSameByteSlice([]byte("should_be_filled"), expectedDoc.CurrentIdentifier) {
				assert.ElementsMatch(t, expectedDoc.CurrentIdentifier, actualDoc.CurrentIdentifier)
			}
			if !tools.IsSameByteSlice([]byte("should_be_filled"), expectedDoc.NextIdentifier) {
				assert.ElementsMatch(t, expectedDoc.NextIdentifier, actualDoc.NextIdentifier)
			}
		} else {
			assert.NotNilf(t, actual, "Error when executing test %v", tt)
			assert.Equal(t, tt.expected.Error(), actual.Error())
		}
	}
}

func TestGenerateCoreDocumentIdentifier(t *testing.T) {
	assert.NotNil(t, GenerateCoreDocumentIdentifier())
	assert.Equal(t, 32, len(GenerateCoreDocumentIdentifier()), "CoreDocument identifier should have 32 elements filled")
}

// ----- END TESTS -----
