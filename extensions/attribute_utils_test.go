// +build unit

package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: more testing for attribute utils functions

const (
	testFieldKey = "test_agreement[{IDX}]."
	testLabel    = "test_agreement"
)

//type TestData struct {
//	one string
//	two string
//	three string
//	four string
//	five string
//}
//
//func createTestData() TestData {
//	return TestData {
//		one:   "one",
//		two:   "two",
//		three: "three",
//		four:  "four",
//	}
//}

func TestGenerateKey(t *testing.T) {
	assert.Equal(t, "test_agreement[1].days", GenerateLabel(testFieldKey, "1", "days"))
	assert.Equal(t, "test_agreement[0].", GenerateLabel(testFieldKey, "0", ""))
}

//func TestCreateAttributesList(t *testing.T) {
//	testingdocuments.CreateInvoicePayload()
//	inv := new(invoice.Invoice)
//	err := inv.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), testingidentity.GenerateRandomDID())
//	assert.NoError(t, err)
//
//	data := createTestData()
//
//	attributes, err := CreateAttributesList(inv, data, testFieldKey, testLabel)
//	assert.NoError(t, err)
//
//	assert.Equal(t, 5, len(attributes))
//
//	for _, attribute := range attributes {
//		if attribute.KeyLabel == "test_agreement[0].one" {
//			assert.Equal(t, "one", attribute.Value.Str)
//			break
//		}
//
//		// apr was not set
//		assert.NotEqual(t, "funding_agreement[0].five", attribute.KeyLabel)
//	}
//}
