// +build unit

package extensions

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

// TODO: more testing for attribute utils functions

const (
	testFieldKey = "test_agreement[{IDX}]."
	testJSONKey  = "json"
)

func TestGenerateLabel(t *testing.T) {
	assert.Equal(t, "test_agreement[1].days", GenerateLabel(testFieldKey, "1", "days"))
	assert.Equal(t, "test_agreement[0].", GenerateLabel(testFieldKey, "0", ""))
}

func TestLabelFromJSONTag(t *testing.T) {
	assert.Equal(t, "testing_test", LabelFromJSONTag(idxKey, "test", "testing_"))
	assert.Equal(t, "test_agreement[{IDX}].json", LabelFromJSONTag(idxKey, testJSONKey, testFieldKey))
}

func TestToMapAttributes(t *testing.T) {
	attrs := []documents.Attribute{
		{
			KeyLabel: "test_details[0].something",
			Key:      utils.RandomByte32(),
			Value:    documents.AttrVal{Str: "whatever"},
		},
		{
			KeyLabel: "test_details[0].else",
			Key:      utils.RandomByte32(),
			Value:    documents.AttrVal{Str: "nope"},
		},
	}
	mapAttr := ToMapAttributes(nil)
	assert.Empty(t, mapAttr)
	mapAttr = ToMapAttributes(attrs)
	assert.NotEmpty(t, mapAttr[attrs[0].Key])
	assert.Equal(t, attrs[0].KeyLabel, mapAttr[attrs[0].Key].KeyLabel)
	assert.NotEmpty(t, mapAttr[attrs[1].Key])
	assert.Equal(t, attrs[1].KeyLabel, mapAttr[attrs[1].Key].KeyLabel)
}
