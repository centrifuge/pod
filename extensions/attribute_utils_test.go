// +build unit

package extensions

import (
	"testing"

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
	assert.Equal(t, "test_agreement[{IDX}].test", LabelFromJSONTag(idxKey, testJSONKey, testFieldKey))
}
