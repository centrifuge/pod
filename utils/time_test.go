//go:build unit

package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToTimestamp(t *testing.T) {
	now := time.Now().UTC()
	ts, err := ToTimestamp(now)
	assert.NoError(t, err)
	assert.NotNil(t, ts, "must be non nil")
	assert.Equal(t, now.Unix(), ts.Seconds)
	assert.Equal(t, now.Nanosecond(), int(ts.Nanos))
}
