// +build unit

package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFreePort(t *testing.T) {
	addr, port, err := GetFreeAddrPort()
	assert.NoError(t, err)
	assert.NotEqual(t, port, 0)
	assert.Equal(t, addr, fmt.Sprintf("127.0.0.1:%d", port))
}
