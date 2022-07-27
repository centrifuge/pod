//go:build unit

package generic

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_Validate(t *testing.T) {
	srv := service{}
	err := srv.Validate(context.Background(), nil, nil)
	assert.NoError(t, err)
}
