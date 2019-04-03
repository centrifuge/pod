package entityrelationship

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateValidator(t *testing.T) {
	cv := CreateValidator(nil)
	assert.Len(t, cv, 1)
}

func TestUpdateValidator(t *testing.T) {
	uv := UpdateValidator(nil)
	assert.Len(t, uv, 2)
}
