// +build unit

package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: more testing for attribute utils functions

const (
	testFieldKey = "test_agreement[{IDX}]."
)

func TestGenerateKey(t *testing.T) {
	assert.Equal(t, "test_agreement[1].days", GenerateLabel(testFieldKey, "1", "days"))
	assert.Equal(t, "test_agreement[0].", GenerateLabel(testFieldKey, "0", ""))
}
